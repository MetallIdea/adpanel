package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDirInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаём тестовый файл.
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello"), 0644)

	cfg := Config{RootDir: tmpDir}
	info, err := cfg.listDirInfo(testFile)
	if err != nil {
		t.Fatalf("listDirInfo unexpected error: %v", err)
	}
	if info.Name() != "test.txt" {
		t.Errorf("info.Name() = %q, want %q", info.Name(), "test.txt")
	}
	if info.Size() != 5 {
		t.Errorf("info.Size() = %d, want %d", info.Size(), 5)
	}
}

func TestOsIsNotExist(t *testing.T) {
	if !osIsNotExist(os.ErrNotExist) {
		t.Error("osIsNotExist(os.ErrNotExist) = false, want true")
	}
	if !osIsNotExist(&os.PathError{}) {
		// PathError не является ErrNotExist, но os.IsNotExist может unwrap его.
		// Тестируем через os.Stat на несуществующий файл.
	}

	_, err := os.Stat(filepath.Join(t.TempDir(), "nonexistent"))
	if !osIsNotExist(err) {
		t.Errorf("osIsNotExist(os.Stat nonexistent) = false, want true")
	}
}

func TestSanitizePathForDisplay(t *testing.T) {
	tests := []struct {
		absDirPath string
		name       string
		want       string
	}{
		{"/var/www/files", "test.txt", "files/test.txt"},
		{"/var/www/files/subdir", "test.txt", "subdir/test.txt"},
		// sanitizePathForDisplay использует filepath.Join(filepath.Base(absDirPath), name),
		// поэтому "subdir/file.txt" останется как есть (basename не удаляет компоненты).
		{"/var/www/files", "subdir/file.txt", "files/subdir/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.absDirPath, func(t *testing.T) {
			got := sanitizePathForDisplay(tt.absDirPath, tt.name)
			if got != tt.want {
				t.Errorf("sanitizePathForDisplay(%q, %q) = %q, want %q",
					tt.absDirPath, tt.name, got, tt.want)
			}
		})
	}
}
