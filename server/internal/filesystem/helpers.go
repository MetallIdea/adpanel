package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

// listDirInfo возвращает os.FileInfo для указанного пути.
func (c Config) listDirInfo(absPath string) (os.FileInfo, error) {
	return os.Stat(absPath)
}

// osIsNotExist проверяет, является ли ошибкой отсутствие файла/директории.
func osIsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// sanitizePathForDisplay формирует отображаемый путь для FileInfo.Path.
func sanitizePathForDisplay(absDirPath string, name string) string {
	return filepath.Join(filepath.Base(absDirPath), name)
}

// sanitizeFileName безопасно обрабатывает имя файла из пользовательского ввода.
//
// Механизмы защиты:
//  1. filepath.Clean — убирает .., //, лишние точки.
//  2. filepath.Base — оставляет только последний компонент (убирает пути).
//  3. Проверка на пустой результат и path separators.
//  4. Проверка на ".." после очистки.
//
// Возвращает ошибку, если имя файла небезопасно использовать.
func sanitizeFileName(originalName string) (string, error) {
	// Убираем пробелы и управляющие символы.
	name := strings.TrimSpace(originalName)
	if name == "" {
		return "", ErrInvalidMultipartRequest
	}

	// filepath.Base оставляет только последний компонент пути.
	// Это безопасно: даже если originalName = "../../../etc/passwd",
	// base будет "passwd" — просто имя файла без пути.
	name = filepath.Base(name)

	// filepath.Base может вернуть "." или ".." для специальных путей.
	if name == "." || name == ".." || name == "" {
		return "", ErrInvalidMultipartRequest
	}

	// Проверка на path separators (на случай edge cases).
	if strings.ContainsAny(name, `\/`) {
		return "", ErrInvalidMultipartRequest
	}

	return name, nil
}
