package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setEnv(t *testing.T, key, val string) {
	t.Helper()
	os.Setenv(key, val)
	t.Cleanup(func() { os.Unsetenv(key) })
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	setEnv(t, "ADMIN_LOGIN", "admin")
	setEnv(t, "ADMIN_PASSWORD", "password")
	setEnv(t, "SECRET_KEY", "test-secret")

	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/login", NewHandler().Login)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	setEnv(t, "ADMIN_LOGIN", "admin")
	setEnv(t, "ADMIN_PASSWORD", "password")
	setEnv(t, "SECRET_KEY", "test-secret")

	body, _ := json.Marshal(LoginRequest{Login: "admin", Password: "wrong"})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/login", NewHandler().Login)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	setEnv(t, "ADMIN_LOGIN", "admin")
	setEnv(t, "ADMIN_PASSWORD", "password")
	setEnv(t, "SECRET_KEY", "test-secret")

	body, _ := json.Marshal(LoginRequest{Login: "admin", Password: "password"})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/login", NewHandler().Login)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "jwt_token" {
			found = true
			if c.HttpOnly == false {
				t.Error("cookie should be HttpOnly")
			}
			break
		}
	}
	if !found {
		t.Error("expected jwt_token cookie to be set")
	}
}

func TestLoginHandler_RateLimitExceeded(t *testing.T) {
	setEnv(t, "ADMIN_LOGIN", "admin")
	setEnv(t, "ADMIN_PASSWORD", "password")
	setEnv(t, "SECRET_KEY", "test-secret")

	// Reset the limiter for this test
	loginLimiterInstance = newLoginLimiter(3)

	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(LoginRequest{Login: "admin", Password: "password"})
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r := gin.New()
		r.POST("/api/login", NewHandler().Login)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i+1, w.Code, http.StatusOK)
		}
	}

	// Next request should be rate limited
	body, _ := json.Marshal(LoginRequest{Login: "admin", Password: "password"})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/login", NewHandler().Login)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d (429)", w.Code, http.StatusTooManyRequests)
	}
}

func TestLoginHandler_MissingConfig(t *testing.T) {
	// Reset limiter to ensure we don't get 429 from previous tests
	loginLimiterInstance = newLoginLimiter(5)

	// No ADMIN_LOGIN set
	setEnv(t, "ADMIN_PASSWORD", "password")
	setEnv(t, "SECRET_KEY", "test-secret")

	body, _ := json.Marshal(LoginRequest{Login: "admin", Password: "password"})
	req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := gin.New()
	r.POST("/api/login", NewHandler().Login)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	setEnv(t, "SECRET_KEY", "test-secret")

	req := httptest.NewRequest("GET", "/api/sites", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/api/sites", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	setEnv(t, "SECRET_KEY", "test-secret")

	req := httptest.NewRequest("GET", "/api/sites", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: "invalid.token.here"})
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/api/sites", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	setEnv(t, "SECRET_KEY", "test-secret")

	// Generate a valid token
	claims := jwt.MapClaims{
		"login": "admin",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest("GET", "/api/sites", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: signed})
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/api/sites", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	setEnv(t, "SECRET_KEY", "test-secret")

	claims := jwt.MapClaims{
		"login": "admin",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest("GET", "/api/sites", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: signed})
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/api/sites", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_WrongAlgorithm(t *testing.T) {
	setEnv(t, "SECRET_KEY", "test-secret")

	// Create token with HS256 (wrong algorithm)
	claims := jwt.MapClaims{
		"login": "admin",
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest("GET", "/api/sites", nil)
	req.AddCookie(&http.Cookie{Name: "jwt_token", Value: signed})
	w := httptest.NewRecorder()

	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/api/sites", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
