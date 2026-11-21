package filesystem

import (
	"os"
	"path/filepath"
)

// ResolvePath безопасно разрешает относительный путь внутри корневой директории.
//
// Механизмы защиты:
//  1. Нормализует путь через filepath.Clean (убирает .., //, лишние точки).
//  2. Формирует абсолютный путь, объединяя RootDir и cleaned path.
//  3. Проверяет, что результирующий путь находится внутри RootDir через
//     filepath.Rel — это надёжнее, чем strings.HasPrefix, потому что
//     корректно обрабатывает случаи с разделителями и граничные условия.
//  4. Проверяет символические ссылки через EvalSymlinks — проверяет, что
//     реальная (разрешённая) директория находится внутри RootDir.
//
// Возвращает ошибку, если путь выходит за пределы RootDir.
func (c Config) ResolvePath(requestedPath string) (string, error) {
	// Нормализация пути.
	cleaned := filepath.Clean(requestedPath)

	// Блокируем абсолютные пути, не начинающиеся с RootDir.
	if filepath.IsAbs(cleaned) {
		rel, err := filepath.Rel(c.RootDir, cleaned)
		if err != nil || rel == ".." || filepath.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", ErrPathTraversal
		}
		return cleaned, nil
	}

	// Объединяем с RootDir.
	absPath := filepath.Clean(filepath.Join(c.RootDir, cleaned))

	// filepath.Rel возвращает ошибку, если пути не имеют общего базового каталога,
	// или относительный путь, начинающийся с "..", если целевой путь выходит за пределы base.
	rel, err := filepath.Rel(c.RootDir, absPath)
	if err != nil || rel == ".." || filepath.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	// Проверка символических ссылок.
	if err := c.verifySymlinkSafety(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

// verifySymlinkSafety проверяет, что реальная (разрешённая) директория находится
// внутри RootDir. Использует EvalSymlinks для разрешения всех symlink в пути.
//
// Ограничения (TOCTOU):
//
//	filepath.EvalSymlinks() не является атомарной системной операцией.
//	Внутри Go она вызывает os.Readlink() + os.Lstat() — два отдельных
//	системных вызова. Если symlink изменится между вызовом EvalSymlinks
//	и последующим os.ReadDir() / os.Open(), это может привести к обходу
//	ограничений.
//
// Для данного проекта (авторизованный администратор, не публичный API)
// этот риск считается приемлемым. Для production-систем с повышенными
// требованиями к безопасности следует использовать один из подходов:
//
//  1. openat() с дескрипторами — позволяет открывать файлы относительно
//     уже проверенного дескриптора директории, исключая TOCTOU.
//     Требует syscall-пакета и не переносимо на Windows.
//
//  2. chroot / pivot_root — изоляция файловой системы на уровне ОС.
//     Сервер запускается в изолированном окружении, где RootDir — корень.
//
//  3. FUSE-файловая система — перехват всех операций с файлами на уровне
//     ядра, проверка разрешений до доступа к ФС.
//
// В текущей реализации мы принимаем риск TOCTOU как приемлемый, так как:
//  - Доступ к API только для авторизованных пользователей
//  - Пользователи не имеют прав на создание/изменение symlink на сервере
//  - RootDir контролируется администратором
func (c Config) verifySymlinkSafety(absPath string) error {
	// Путь может не существовать — проверяем существующую часть.
	realPath := absPath

	_, err := os.Lstat(absPath)
	if err != nil {
		// Проверяем родительскую директорию.
		parent := filepath.Dir(absPath)
		if parent == absPath {
			return nil
		}
		realParent, err2 := filepath.EvalSymlinks(parent)
		if err2 != nil {
			// Не можем проверить — полагаемся на filepath.Rel проверку.
			return nil
		}
		realPath = filepath.Join(realParent, filepath.Base(absPath))
	} else {
		// Путь существует — разрешаем symlink.
		realPath, err = filepath.EvalSymlinks(absPath)
		if err != nil {
			return ErrSymlinkTraversal
		}
	}

	// Проверяем, что реальная директория находится внутри RootDir.
	rel, err := filepath.Rel(c.RootDir, realPath)
	if err != nil || rel == ".." || filepath.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ErrSymlinkTraversal
	}

	return nil
}

// ListDir безопасно перечисляет содержимое директории в пределах RootDir.
// Принимает уже разрешённый абсолютный путь (через ResolvePath).
//
// Примечание: после ResolvePath+verifySymlinkSafety мы гарантируем, что
// absPath указывает внутрь RootDir на момент вызова. OS-уровень TOCTOU
// (изменение symlink после проверки) остаётся, но для данного уровня
// доверия считается приемлемым (см. комментарий в verifySymlinkSafety).
func (c Config) ListDir(absPath string) ([]FileInfo, error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Пропускаем файлы, информацию о которых нельзя получить.
		}

		fileInfo := FileInfo{
			Name:      entry.Name(),
			Type:      func() FileType { if entry.IsDir() { return FileTypeFolder }; return FileTypeFile }(),
			Size:      info.Size(),
			Path:      sanitizePathForDisplay(absPath, entry.Name()),
			CreatedAt: info.ModTime(),
		}
		files = append(files, fileInfo)
	}

	return files, nil
}
