package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	rootDir string
}

func NewHandler() *Handler {
	rootDir := "/var/www"
	if d := os.Getenv("FILE_ROOT_DIR"); d != "" {
		rootDir = d
	}
	return &Handler{rootDir: rootDir}
}

type FileEntry struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "file" or "folder"
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
	Modified  time.Time `json:"modified"`
	IsDir     bool      `json:"-"`
}

func (h *Handler) ListFiles(c *gin.Context) {
	path := c.Query("path")

	// Clean and validate path
	path = filepath.Clean(path)
	if path == "." {
		path = ""
	}

	// Security: prevent path traversal
	fullPath := filepath.Join(h.rootDir, path)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}

	// Ensure the path is within rootDir
	if !strings.HasPrefix(absPath, h.rootDir) {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(404, gin.H{"error": "path not found"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if !info.IsDir() {
		c.JSON(400, gin.H{"error": "path is not a directory"})
		return
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var files []FileEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := filepath.Join(path, entry.Name())
		if path != "" {
			filePath = filepath.Join("/", path, entry.Name())
		} else {
			filePath = "/" + entry.Name()
		}

		files = append(files, FileEntry{
			Name:     entry.Name(),
			Type:     "file",
			Size:     info.Size(),
			Path:     filePath,
			Modified: info.ModTime(),
			IsDir:    entry.IsDir(),
		})
	}

	// Folders first, then files, both sorted by name
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ftype := "file"
		if entry.IsDir() {
			ftype = "folder"
		}

		filePath := filepath.Join(path, entry.Name())
		if path != "" {
			filePath = filepath.Join("/", path, entry.Name())
		} else {
			filePath = "/" + entry.Name()
		}

		files = append(files, FileEntry{
			Name:     entry.Name(),
			Type:     ftype,
			Size:     info.Size(),
			Path:     filePath,
			Modified: info.ModTime(),
			IsDir:    entry.IsDir(),
		})
	}

	// Deduplicate: rebuild from entries properly
	files = nil
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ftype := "file"
		if entry.IsDir() {
			ftype = "folder"
		}

		filePath := entry.Name()
		if path != "" {
			filePath = filepath.Join(path, entry.Name())
		}

		files = append(files, FileEntry{
			Name:     entry.Name(),
			Type:     ftype,
			Size:     info.Size(),
			Path:     filePath,
			Modified: info.ModTime(),
			IsDir:    entry.IsDir(),
		})
	}

	// Sort: folders first, then files, alphabetically within each group
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			swap := false
			if files[i].Type == "file" && files[j].Type == "folder" {
				swap = true
			} else if files[i].Type == files[j].Type {
				swap = strings.ToLower(files[i].Name) > strings.ToLower(files[j].Name)
			}
			if swap {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	c.JSON(200, files)
}

func (h *Handler) UploadFile(c *gin.Context) {
	path := c.PostForm("path")
	if path == "" {
		path = ""
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Security: sanitize filename
	filename := filepath.Base(header.Filename)

	// Build target path
	targetPath := filepath.Join(h.rootDir, path, filename)
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}

	// Ensure the path is within rootDir
	if !strings.HasPrefix(absPath, h.rootDir) {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create directory: %v", err)})
		return
	}

	// Create destination file
	dst, err := os.Create(absPath)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create file: %v", err)})
		return
	}
	defer dst.Close()

	// Copy file content
	if _, err := file.Copy(dst); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to save file: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message": "file uploaded successfully",
		"name":    filename,
		"path":    filepath.Join(path, filename),
		"size":    header.Size,
	})
}

func parsePath(p string) string {
	if p == "" || p == "/" {
		return ""
	}
	p = strings.TrimPrefix(p, "/")
	return filepath.Clean(p)
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func (h *Handler) DeleteFile(c *gin.Context) {
	path := c.Param("path")
	if path == "" {
		c.JSON(400, gin.H{"error": "path is required"})
		return
	}

	fullPath := filepath.Join(h.rootDir, path)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}

	if !strings.HasPrefix(absPath, h.rootDir) {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	if err := os.RemoveAll(absPath); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "deleted successfully"})
}

func (h *Handler) CreateDirectory(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	dirPath := filepath.Join(h.rootDir, req.Path, req.Name)
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid path"})
		return
	}

	if !strings.HasPrefix(absPath, h.rootDir) {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "directory created", "path": filepath.Join(req.Path, req.Name)})
}

func (h *Handler) GetBreadcrumb(c *gin.Context) {
	path := parsePath(c.Query("path"))

	var parts []string
	if path != "" {
		parts = strings.Split(path, string(filepath.Separator))
	}

	var breadcrumb []gin.H
	var currentPath string

	for _, part := range parts {
		if part == "" {
			continue
		}
		currentPath = filepath.Join(currentPath, part)
		breadcrumb = append(breadcrumb, gin.H{
			"label": part,
			"path":  currentPath,
		})
	}

	c.JSON(200, breadcrumb)
}
