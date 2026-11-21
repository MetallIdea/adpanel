package filesystem

import (
	"os"
	"path/filepath"
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
