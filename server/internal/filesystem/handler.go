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

// ListFilesResponse — ответ на запрос списка файлов.
type ListFilesResponse struct {
	Files []FileBrief `json:"files"`
}

// FileBrief — краткая информация о файле (только name и type).
type FileBrief struct {
	Name string     `json:"name"`
	Type FileType   `json:"type"`
}

// ListFiles обрабатывает GET /api/files.
//
// Query parameters:
//   - path — относительный путь внутри корневой директории (по умолчанию: "").
//
// Возвращает список файлов и директорий в запрошенной папке,
// завернутый в объект с полем "files".
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
	fileList, err := h.config.ListDir(absPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read directory",
		})
		return
	}

	// Формируем ответ только с name и type.
	files := make([]FileBrief, 0, len(fileList))
	for _, f := range fileList {
		files = append(files, FileBrief{
			Name: f.Name,
			Type: f.Type,
		})
	}

	c.JSON(http.StatusOK, ListFilesResponse{Files: files})
}

// UploadFile обрабатывает POST /api/files/upload.
//
// Ожидаемый multipart/form-data:
//   - file — файл для загрузки (обязательный)
//   - path — относительный путь внутри корневой директории (по умолчанию: "")
//   - overwrite — "true" для разрешения перезаписи существующего файла
//
// Ограничения:
//   - Максимальный размер файла: MAX_UPLOAD_SIZE (по умолчанию 100MB)
//   - Перезапись существующих файлов запрещена (если overwrite не указан)
//   - Файлы сохраняются только внутри RootDir
func (h *Handler) UploadFile(c *gin.Context) {
	// Ограничиваем общий размер multipart-запроса для защиты от DoS.
	// http.MaxBytesReader ограничивает чтение тела запроса до заданного
	// лимита на уровне HTTP-уровня. Если клиент отправит больше данных,
	// сервер вернёт 413 и разорвёт соединение.
	const maxRequestBody = 32 << 20 // 32 MB

	http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBody)

	// Ограничиваем размер файлов в памяти при парсинге multipart.
	// Файлы больше этого размера сохраняются на диск.
	const maxMemory = 32 << 20 // 32 MB

	if err := c.Request.ParseMultipartForm(maxMemory); err != nil {
		// Исправление #10, #11: ParseMultipartForm автоматически
		// очищает временные файлы при ошибке. Если MaxBytesReader
		// отсёк тело до полного чтения, ParseMultipartForm вернёт
		// io.EOF — это корректная ошибка, временные файлы не создавались.
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

	// Безопасная обработка имени файла.
	// header.Filename из multipart может содержать "../" или другие артефакты.
	safeName, err := sanitizeFileName(header.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid file name",
		})
		return
	}

	// Формируем полный путь к файлу.
	// safeName — только basename, без path separators, поэтому filepath.Join
	// гарантирует, что targetPath находится внутри dirAbsPath.
	targetPath := filepath.Join(dirAbsPath, safeName)

	// Проверяем, не существует ли уже файл с таким именем.
	// Используем Lstat, чтобы не следовать за symlink.
	existingInfo, err := os.Lstat(targetPath)
	if err == nil {
		// Файл существует — проверяем параметр overwrite.
		overwrite := c.Query("overwrite") == "true" || c.PostForm("overwrite") == "true"
		if !overwrite {
			c.JSON(http.StatusConflict, gin.H{
				"error": "file already exists",
			})
			return
		}
		// Overwrite разрешён — удаляем существующий файл.
		if existingInfo.Mode()&os.ModeSymlink != 0 {
			// Не перезаписываем symlink — это небезопасно.
			c.JSON(http.StatusForbidden, gin.H{
				"error": "cannot overwrite symbolic link",
			})
			return
		}
		os.Remove(targetPath)
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
		c.JSON(http.StatusConflict, gin.H{
			"error": "file already exists",
		})
		return
	}
	defer f.Close()

	// Ограничиваем чтение: даже если multipart body больше MaxUploadSize,
	// мы прочитаем только до лимита. Это защищает от DoS через oversized body.
	reader := io.LimitReader(file, h.config.MaxUploadSize+1)

	// Копируем данные из запроса в файл с проверкой размера.
	written, err := io.Copy(f, reader)
	if err != nil {
		// Удаляем частично записанный файл при ошибке.
		// targetPath был санитизирован и находится внутри dirAbsPath,
		// который проверен через ResolvePath. os.Remove безопасен.
		os.Remove(targetPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save file",
		})
		return
	}

	// Проверяем фактический размер записанных данных.
	// io.LimitReader ограничивает чтение до MaxUploadSize+1,
	// поэтому если written > MaxUploadSize — файл слишком большой.
	if written > h.config.MaxUploadSize {
		os.Remove(targetPath)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file size exceeds maximum allowed size",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "file uploaded successfully",
		"name":    safeName,
		"size":    written,
	})
}
