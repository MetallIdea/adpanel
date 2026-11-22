package filesystem

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config определяет конфигурацию файлового менеджера.
type Config struct {
	// RootDir — корневая директория, в пределах которой разрешён доступ.
	RootDir string
	// MaxUploadSize — максимальный размер загружаемого файла в байтах.
	MaxUploadSize int64
}

// DefaultConfig создаёт конфигурацию с значениями по умолчанию.
// Значения переопределяются через переменные окружения:
//
//	FILES_ROOT_DIR    — путь к корневой директории (по умолчанию: /var/www/files).
//	MAX_UPLOAD_SIZE   — максимальный размер файла в байтах (по умолчанию: 104857600 = 100MB).
func DefaultConfig() Config {
	rootDir := os.Getenv("FILES_ROOT_DIR")
	if rootDir == "" {
		rootDir = "/var/www/files"
	}

	maxUploadSize := int64(100 * 1024 * 1024) // 100MB по умолчанию
	if s := os.Getenv("MAX_UPLOAD_SIZE"); s != "" {
		if n, err := parseSize(s); err == nil && n > 0 {
			maxUploadSize = n
		}
	}

	return Config{
		RootDir:       rootDir,
		MaxUploadSize: maxUploadSize,
	}
}

// parseSize преобразует строковое представление размера в байты.
// Поддерживает числовые значения и суффиксы: K, M, G (case-insensitive).
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty size string")
	}

	multiplier := int64(1)
	switch strings.ToUpper(string(s[len(s)-1])) {
	case "K":
		multiplier = 1024
		s = s[:len(s)-1]
	case "M":
		multiplier = 1024 * 1024
		s = s[:len(s)-1]
	case "G":
		multiplier = 1024 * 1024 * 1024
		s = s[:len(s)-1]
	}

	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, fmt.Errorf("invalid size value %q: %w", s, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("size value must be non-negative, got %d", n)
	}
	return n * multiplier, nil
}
