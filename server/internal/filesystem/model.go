package filesystem

import "time"

// FileType определяет тип файла.
type FileType string

const (
	FileTypeFile    FileType = "file"
	FileTypeFolder   FileType = "folder"
)

// FileInfo содержит информацию о файле или директории.
type FileInfo struct {
	Name      string    `json:"name"`
	Type      FileType  `json:"type"`
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}
