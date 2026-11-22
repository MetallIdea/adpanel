package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{RootDir: tmpDir}

	tests := []struct {
		name        string
		requested   string
		wantInRoot  bool
		wantErr     error
	}{
		{"empty path", "", true, nil},
		{"current dir", ".", true, nil},
		{"simple relative", "subdir", true, nil},
		{"nested relative", "a/b/c", true, nil},
		{"path traversal", "../etc", false, ErrPathTraversal},
		{"deep traversal", "a/../../etc", false, ErrPathTraversal},
		{"double slash", "//etc", false, ErrPathTraversal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём структуру для относительных путей.
			if tt.requested != "" && tt.requested != "." {
				parts := filepath.SplitList(tt.requested)
				for _, part := range parts {
					if part != "" && part != "." {
						os.MkdirAll(filepath.Join(tmpDir, part), 0755)
					}
				}
			}

			got, err := cfg.ResolvePath(tt.requested)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ResolvePath(%q) error = %v, want %v", tt.requested, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("ResolvePath(%q) unexpected error: %v", tt.requested, err)
				return
			}
			if !tt.wantInRoot {
				t.Errorf("ResolvePath(%q) = %q, expected error", tt.requested, got)
				return
			}
			// Проверяем, что результат внутри RootDir.
			rel, err := filepath.Rel(tmpDir, got)
			if err != nil || rel == ".." || filepath.HasPrefix(rel, ".."+string(filepath.Separator)) {
				t.Errorf("ResolvePath(%q) = %q is outside RootDir", tt.requested, got)
			}
		})
	}
}

func TestResolvePath_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()
	cfg := Config{RootDir: tmpDir}

	// Абсолютный путь внутри RootDir — должен разрешиться.
	insidePath := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(insidePath, []byte("test"), 0644)
	got, err := cfg.ResolvePath(insidePath)
	if err != nil {
		t.Errorf("ResolvePath(inside) unexpected error: %v", err)
	}
	if got != insidePath {
		t.Errorf("ResolvePath(inside) = %q, want %q", got, insidePath)
	}

	// Абсолютный путь вне RootDir — должен вернуть ошибку.
	outsidePath := filepath.Join(otherDir, "file.txt")
	_, err = cfg.ResolvePath(outsidePath)
	if err != ErrPathTraversal {
		t.Errorf("ResolvePath(outside) error = %v, want %v", err, ErrPathTraversal)
	}
}

func TestResolvePath_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()
	cfg := Config{RootDir: tmpDir}

	// Создаём symlink внутри RootDir, указывающий на другую директорию.
	symlinkPath := filepath.Join(tmpDir, "link")
	if err := os.Symlink(otherDir, symlinkPath); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	// Запрос через symlink, ведущий за пределы RootDir.
	_, err := cfg.ResolvePath("link")
	if err != ErrSymlinkTraversal {
		t.Errorf("ResolvePath(symlink outside) error = %v, want %v", err, ErrSymlinkTraversal)
	}

	// Создаём symlink внутри RootDir, указывающий на поддиректорию.
	subDir := filepath.Join(tmpDir, "sub")
	os.MkdirAll(subDir, 0755)
	safeLink := filepath.Join(tmpDir, "safeLink")
	if err := os.Symlink(subDir, safeLink); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	got, err := cfg.ResolvePath("safeLink")
	if err != nil {
		t.Errorf("ResolvePath(symlink inside) unexpected error: %v", err)
	}
	// Разрешённый путь должен быть реальным каталогом.
	realPath, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Errorf("EvalSymlinks failed: %v", err)
	}
	if realPath != subDir {
		t.Errorf("ResolvePath(symlink inside) resolved to %q, want %q", realPath, subDir)
	}
}

func TestResolvePath_SymlinkChain(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()
	cfg := Config{RootDir: tmpDir}

	// Цепочка: link1 -> link2 -> outside.
	link2 := filepath.Join(tmpDir, "link2")
	if err := os.Symlink(otherDir, link2); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}
	link1 := filepath.Join(tmpDir, "link1")
	if err := os.Symlink(link2, link1); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	_, err := cfg.ResolvePath("link1")
	if err != ErrSymlinkTraversal {
		t.Errorf("ResolvePath(symlink chain) error = %v, want %v", err, ErrSymlinkTraversal)
	}
}
