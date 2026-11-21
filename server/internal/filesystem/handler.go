package filesystem

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// Handler обрабатывает запросы файлового менеджера.
type Handler struct {
	config Config
}

// NewHandler создаёт новый handler с заданной конфигурацией.
func NewHandler(config Config) *Handler {
	return &Handler{config: config}
}

// ListFiles обрабатывает GET /api/files.
//
// Query parameters:
//   - path — относительный путь внутри корневой директории (по умолчанию: "").
//
// Возвращает список файлов и директорий в запрошенной папке.
func (h *Handler) ListFiles(c *gin.Context) {
	requestedPath := c.Query("path")

	// Безопасное разрешение пути.
	absPath, err := h.config.ResolvePath(requestedPath)
	if err != nil {
		if err == ErrPathTraversal || err == ErrSymlinkTraversal {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid path",
		})
		return
	}

	// Проверяем, что путь — директория.
	info, err := h.config.listDirInfo(absPath)
	if err != nil {
		if osIsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "directory not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to access directory",
		})
		return
	}

	if !info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "path is not a directory",
		})
		return
	}

	// Перечисляем содержимое.
	files, err := h.config.ListDir(absPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read directory",
		})
		return
	}

	c.JSON(http.StatusOK, files)
}

// UploadFile обрабатывает POST /api/files/upload.
//
// Ожидаемый multipart/form-data:
//   - file — файл для загрузки (обязательный)
//   - path — относительный путь внутри корневой директории (по умолчанию: "")
//
// Ограничения:
//   - Максимальный размер файла: MAX_UPLOAD_SIZE (по умолчанию 100MB)
//   - Перезапись существующих файлов запрещена
//   - Файлы сохраняются только внутри RootDir
func (h *Handler) UploadFile(c *gin.Context) {
	// Разрешаем до 32MB в памяти, остальное — на диск.
	const maxMemory = 32 << 20 // 32 MB

	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid multipart request: " + err.Error(),
		})
		return
	}

	// Получаем имя целевой директории.
	requestedDir := c.PostForm("path")

	// Безопасное разрешение директории.
	dirAbsPath, err := h.config.ResolvePath(requestedDir)
	if err != nil {
		if err == ErrPathTraversal || err == ErrSymlinkTraversal {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid path",
		})
		return
	}

	// Проверяем, что путь — директория.
	dirInfo, err := h.config.listDirInfo(dirAbsPath)
	if err != nil {
		if osIsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "directory not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to access directory",
		})
		return
	}

	if !dirInfo.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "target path is not a directory",
		})
		return
	}

	// Получаем файл из multipart-запроса.
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to get file from request: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Проверяем размер заголовка (Content-Length).
	// Это первая проверка — фактивный размер будет проверен при чтении.
	if header.Size > h.config.MaxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file size exceeds maximum allowed size",
		})
		return
	}

	// Формируем полный путь к файлу.
	targetPath := filepath.Join(dirAbsPath, header.Filename)

	// Проверяем, не существует ли уже файл с таким именем.
	// Используем Lstat, чтобы не следовать за symlink.
	existingInfo, err := os.Lstat(targetPath)
	if err == nil {
		// Файл существует — запрещаем перезапись.
		if existingInfo.Mode()&os.ModeSymlink != 0 {
			// Если это symlink — потенциально опасно, возвращаем ту же ошибку,
			// что и для обычного файла (не раскрываем детали).
			c.JSON(http.StatusConflict, gin.H{
				"error": "file already exists",
			})
			return
		}
		c.JSON(http.StatusConflict, gin.H{
			"error": "file already exists",
		})
		return
	}

	// Создаём файл с O_EXCL — атомарная операция "создать или ошибка".
	// O_EXCL гарантирует, что если файл появится между проверкой Lstat
	// и вызовом OpenFile, операция завершится ошибкой.
	//
	// Ограничение: O_EXCL не защищает от TOCTOU, если symlink в родительской
	// директории меняется между ResolvePath и OpenFile. Для полной защиты
	// нужен chroot или openat().
	f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		// Duplicated entry — файл появился между Lstat и OpenFile.
		c.JSON(http.StatusConflict, gin.H{
			"error": "file already exists",
		})
		return
	}
	defer f.Close()

	// Копируем данные из запроса в файл с проверкой размера.
	written, err := io.Copy(f, file)
	if err != nil {
		// Удаляем частично записанный файл при ошибке.
		os.Remove(targetPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save file",
		})
		return
	}

	// Проверяем фактический размер записанных данных.
	if written > h.config.MaxUploadSize {
		os.Remove(targetPath)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file size exceeds maximum allowed size",
		})
		return
	}

	// Проверяем, что созданный файл не является symlink.
	finalInfo, err := os.Lstat(targetPath)
	if err != nil {
		os.Remove(targetPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to verify uploaded file",
		})
		return
	}

	if finalInfo.Mode()&os.ModeSymlink != 0 {
		os.Remove(targetPath)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	// Проверяем, что реальная директория файла внутри RootDir.
	realPath, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		// Файл не symlink — EvalSymlinks возвращает тот же путь.
		realPath = targetPath
	}

	rel, err := filepath.Rel(h.config.RootDir, realPath)
	if err != nil || rel == ".." || filepath.HasPrefix(rel, ".."+string(filepath.Separator)) {
		os.Remove(targetPath)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "access denied",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "file uploaded successfully",
		"name":    header.Filename,
		"size":    written,
	})
}
