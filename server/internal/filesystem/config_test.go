package filesystem

import (
	"os"
	"testing"
)

func TestDefaultConfig_EnvOverride(t *testing.T) {
	// Сохраняем и восстанавливаем оригинальные значения env.
	origRoot := os.Getenv("FILES_ROOT_DIR")
	origSize := os.Getenv("MAX_UPLOAD_SIZE")
	t.Cleanup(func() {
		os.Setenv("FILES_ROOT_DIR", origRoot)
		os.Setenv("MAX_UPLOAD_SIZE", origSize)
	})

	// Проверяем значения по умолчанию.
	cfg := DefaultConfig()
	if cfg.RootDir != "/var/www/files" {
		t.Errorf("default RootDir = %q, want %q", cfg.RootDir, "/var/www/files")
	}
	if cfg.MaxUploadSize != 100*1024*1024 {
		t.Errorf("default MaxUploadSize = %d, want %d", cfg.MaxUploadSize, 100*1024*1024)
	}

	// Проверяем переопределение через env.
	os.Setenv("FILES_ROOT_DIR", "/custom/root")
	os.Setenv("MAX_UPLOAD_SIZE", "50M")
	cfg = DefaultConfig()
	if cfg.RootDir != "/custom/root" {
		t.Errorf("env RootDir = %q, want %q", cfg.RootDir, "/custom/root")
	}
	if cfg.MaxUploadSize != 50*1024*1024 {
		t.Errorf("env MaxUploadSize = %d, want %d", cfg.MaxUploadSize, 50*1024*1024)
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		want     int64
		wantErr  bool
	}{
		{"1024", 1024, false},
		{"1K", 1024, false},
		{"1k", 1024, false},
		{"1M", 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"50M", 50 * 1024 * 1024, false},
		{"", 0, true},
		{"abc", 0, true},
		{"-1", 0, true},
		{"K", 0, true},       // суффикс без цифры
		{"M", 0, true},       // суффикс без цифры
		{"100X", 0, true},    // неизвестный суффикс
		{"1.5M", 0, true},    // дробное число
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseSize(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("parseSize(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
