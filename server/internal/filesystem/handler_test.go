package filesystem

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/MetallIdea/adpanel/server/internal/auth"
)

func setupTestHandler(t *testing.T) (*Handler, *gin.Engine, string) {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 1024, // 1KB for testing
	}
	handler := NewHandler(cfg)

	// Создаём тестовую структуру файлов.
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("world"), 0644)

	r := gin.New()
	r.GET("/files", handler.ListFiles)
	r.POST("/files/upload", handler.UploadFile)

	return handler, r, tmpDir
}

// ====== Группы 1: базовые тесты (восстановленные) ======

func TestListFiles_Success(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		return
	}

	var files []FileInfo
	if err := json.Unmarshal(w.Body.Bytes(), &files); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(files) < 2 {
		t.Errorf("expected at least 2 files, got %d", len(files))
	}

	typeMap := make(map[string]FileType)
	for _, f := range files {
		typeMap[f.Name] = f.Type
	}

	if typeMap["file1.txt"] != FileTypeFile {
		t.Errorf("file1.txt type = %v, want %v", typeMap["file1.txt"], FileTypeFile)
	}
	if typeMap["subdir"] != FileTypeFolder {
		t.Errorf("subdir type = %v, want %v", typeMap["subdir"], FileTypeFolder)
	}
}

func TestListFiles_NonExistentDir(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files?path=nonexistent", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestListFiles_PathTraversal(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	outsideDir, err := os.MkdirTemp("", "outside_test_*")
	if err != nil {
		t.Fatalf("failed to create outside dir: %v", err)
	}
	defer os.RemoveAll(outsideDir)

	outsideFile := filepath.Join(outsideDir, "outside_test.txt")
	os.WriteFile(outsideFile, []byte("secret"), 0644)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files?path=../../../"+strings.TrimPrefix(outsideFile, "/"), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestListFiles_NonDirectory(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files?path=file1.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestListFiles_SubDirectory(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files?path=subdir", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
		return
	}

	var files []FileInfo
	if err := json.Unmarshal(w.Body.Bytes(), &files); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(files) != 1 || files[0].Name != "file2.txt" {
		t.Errorf("expected [file2.txt], got %v", files)
	}
}

func TestUploadFile_Success(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("upload test content"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		return
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if resp["message"] != "file uploaded successfully" {
		t.Errorf("message = %v, want %v", resp["message"], "file uploaded successfully")
	}
}

func TestUploadFile_Overwrite(t *testing.T) {
	_, r, tmpDir := setupTestHandler(t)

	body1 := &bytes.Buffer{}
	writer1 := multipart.NewWriter(body1)
	part1, _ := writer1.CreateFormFile("file", "file1.txt")
	part1.Write([]byte("new content"))
	writer1.Close()

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/files/upload", body1)
	req1.Header.Set("Content-Type", writer1.FormDataContentType())
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusConflict {
		t.Errorf("first upload status = %d, want %d", w1.Code, http.StatusConflict)
	}

	content, _ := os.ReadFile(filepath.Join(tmpDir, "file1.txt"))
	if string(content) != "hello" {
		t.Errorf("file content = %q, want %q", string(content), "hello")
	}
}

func TestUploadFile_TooLarge(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	largeContent := strings.Repeat("x", 2048)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.txt")
	part.Write([]byte(largeContent))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestUploadFile_InvalidMultipart(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", strings.NewReader("not multipart"))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUploadFile_NoFile(t *testing.T) {
	_, r, _ := setupTestHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("path", "")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestUploadFile_PathTraversal(t *testing.T) {
	_, r, tmpDir := setupTestHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "../../../etc/passwd")
	part.Write([]byte("should not be written"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
		return
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "passwd")); err != nil {
		t.Errorf("file 'passwd' should be created inside RootDir: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "..", "passwd")); err == nil {
		t.Error("file was created outside RootDir")
	}
}

func TestUploadFile_SymlinkTarget(t *testing.T) {
	tmpDir := t.TempDir()
	otherDir := t.TempDir()

	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 1024,
	}
	handler := NewHandler(cfg)

	symlinkPath := filepath.Join(tmpDir, "link")
	if err := os.Symlink(otherDir, symlinkPath); err != nil {
		t.Skipf("symlink creation failed: %v", err)
	}

	r := gin.New()
	r.POST("/files/upload", handler.UploadFile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test"))
	writer.WriteField("path", "link")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestUploadFile_ConcurrentUploads(t *testing.T) {
	_, r, tmpDir := setupTestHandler(t)

	var results []int
	var mu sync.Mutex

	const fileName = "concurrent.txt"

	for i := 0; i < 5; i++ {
		go func(n int) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", fileName)
			part.Write([]byte(strings.Repeat("x", n)))
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/files/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			r.ServeHTTP(w, req)

			mu.Lock()
			results = append(results, w.Code)
			mu.Unlock()
		}(i)
	}

	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	created := 0
	conflict := 0
	for _, code := range results {
		switch code {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			conflict++
		}
	}

	if created == 0 {
		t.Errorf("expected at least one 201, got 0. Results: %v", results)
	}
	if conflict == 0 {
		t.Errorf("expected at least one 409, got 0. Results: %v", results)
	}

	files, _ := os.ReadDir(tmpDir)
	fileCount := 0
	for _, f := range files {
		if !f.IsDir() {
			fileCount++
		}
	}
	if fileCount != 2 {
		t.Errorf("expected 2 files, got %d", fileCount)
	}
}

func TestUploadFile_UploadToSubdir(t *testing.T) {
	_, r, tmpDir := setupTestHandler(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "new_in_subdir.txt")
	part.Write([]byte("subdir content"))
	writer.WriteField("path", "subdir")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
		return
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "subdir", "new_in_subdir.txt"))
	if err != nil {
		t.Fatalf("file not found in subdir: %v", err)
	}
	if string(content) != "subdir content" {
		t.Errorf("content = %q, want %q", string(content), "subdir content")
	}
}

func TestUploadFile_SanitizeFileName(t *testing.T) {
	_, r, tmpDir := setupTestHandler(t)

	tests := []struct {
		name     string
		fileName string
		wantCode int
		wantFile string
	}{
		{"path traversal in name", "../../../etc/passwd", http.StatusCreated, "passwd"},
		{"dot only", ".", http.StatusBadRequest, ""},
		{"normal file", "normal.txt", http.StatusCreated, "normal.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", tt.fileName)
			part.Write([]byte("content"))
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/files/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			r.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", w.Code, tt.wantCode)
			}

			if tt.wantFile != "" {
				if _, err := os.Stat(filepath.Join(tmpDir, tt.wantFile)); err != nil {
					t.Errorf("file %q not created: %v", tt.wantFile, err)
				}
			}
		})
	}
}

func TestUploadFile_LimitReaderProtection(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 100,
	}
	handler := NewHandler(cfg)

	r := gin.New()
	r.POST("/files/upload", handler.UploadFile)

	largeContent := strings.Repeat("x", 200)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.txt")
	part.Write([]byte(largeContent))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "large.txt")); err == nil {
		t.Error("large file was not removed after exceeding limit")
	}
}

// ====== Группа 2: авторизация и multipart limits ======

func TestListFiles_Unauthorized(t *testing.T) {
	cfg := Config{RootDir: t.TempDir()}
	handler := NewHandler(cfg)

	r := gin.New()
	r.GET("/files", handler.ListFiles)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files", nil)
	r.ServeHTTP(w, req)

	// Без middleware авторизации запрос проходит — handler не enforce auth.
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (handler should not enforce auth itself)", w.Code, http.StatusOK)
	}
}

func TestUploadFile_Unauthorized(t *testing.T) {
	cfg := Config{RootDir: t.TempDir(), MaxUploadSize: 1024}
	handler := NewHandler(cfg)

	r := gin.New()
	r.MaxMultipartMemory = 32 << 20
	r.POST("/files/upload", handler.UploadFile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d (handler should not enforce auth itself)", w.Code, http.StatusCreated)
	}
}

func TestAuthMiddleware_Integration(t *testing.T) {
	// Устанавливаем SECRET_KEY для JWT валидации.
	t.Setenv("SECRET_KEY", "super-secret-key")

	cfg := Config{RootDir: t.TempDir()}
	handler := NewHandler(cfg)

	r := gin.New()
	authorized := r.Group("/api")
	authorized.Use(auth.AuthMiddleware())
	authorized.GET("/files", handler.ListFiles)
	authorized.POST("/files/upload", handler.UploadFile)

	// Запрос без токена.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/files", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /api/files without token: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Запрос с невалидным токеном.
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/files", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: "invalid.token.here"})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /api/files with invalid token: status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	// Запрос с валидным токеном.
	claims := jwt.MapClaims{
		"login": "admin",
		"exp":   jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, _ := token.SignedString([]byte("super-secret-key"))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/files", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: signed})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/files with valid token: status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestUploadFile_MultipartBodyLimit(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 100,
	}
	handler := NewHandler(cfg)

	r := gin.New()
	r.MaxMultipartMemory = 32 << 20
	r.POST("/files/upload", handler.UploadFile)

	largeContent := strings.Repeat("x", 500)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.txt")
	part.Write([]byte(largeContent))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "large.txt")); err == nil {
		t.Error("large file was not removed after exceeding limit")
	}
}

func TestUploadFile_FileSizeBoundary(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 100,
	}
	handler := NewHandler(cfg)

	r := gin.New()
	r.MaxMultipartMemory = 32 << 20
	r.POST("/files/upload", handler.UploadFile)

	tests := []struct {
		name       string
		fileSize   int
		wantStatus int
	}{
		{"exactly at limit", 100, http.StatusCreated},
		{"1 byte over limit", 101, http.StatusRequestEntityTooLarge},
		{"half limit", 50, http.StatusCreated},
		{"empty file", 0, http.StatusCreated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "boundary_"+tt.name+".txt")
			if tt.fileSize > 0 {
				part.Write([]byte(strings.Repeat("x", tt.fileSize)))
			}
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/files/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("fileSize=%d: status = %d, want %d", tt.fileSize, w.Code, tt.wantStatus)
			}
		})
	}
}

func TestUploadFile_MultipartBodyTooLarge(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 1024, // 1KB file limit
	}
	handler := NewHandler(cfg)

	r := gin.New()
	r.MaxMultipartMemory = 1024

	r.POST("/files/upload", handler.UploadFile)

	// Создаём файл размером 2KB (больше MaxUploadSize и MaxMultipartMemory).
	largeContent := strings.Repeat("x", 2048)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.txt")
	part.Write([]byte(largeContent))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	// ParseMultipartForm сохранит файл на диск, но UploadFile вернёт 413
	// из-за io.LimitReader (файл > MaxUploadSize).
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d (LimitReader blocks oversized file)", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestUploadFile_MultipartRequestBodyLimit(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := Config{
		RootDir:       tmpDir,
		MaxUploadSize: 1024,
	}
	handler := NewHandler(cfg)

	r := gin.New()
	r.MaxMultipartMemory = 32 << 20
	r.POST("/files/upload", handler.UploadFile)

	// Создаём multipart-запрос размером 50MB (значительно больше 32MB лимита).
	// MaxBytesReader должен отсечь его на уровне HTTP.
	largeContent := strings.Repeat("x", 50*1024*1024)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "huge.txt")
	part.Write([]byte(largeContent))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	// MaxBytesReader вернёт 413 (Request Entity Too Large) при превышении лимита.
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d (MaxBytesReader blocks oversized body)", w.Code, http.StatusRequestEntityTooLarge)
	}
}
