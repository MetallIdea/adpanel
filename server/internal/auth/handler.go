package auth

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// loginLimiter provides simple rate limiting for the login endpoint.
type loginLimiter struct {
	mu        sync.Mutex
	tokens    float64
	maxTokens float64
	refill    float64
	lastTime  time.Time
}

func newLoginLimiter(rps float64) *loginLimiter {
	if rps <= 0 {
		rps = 5
	}
	return &loginLimiter{
		tokens:    rps,
		maxTokens: rps,
		refill:    rps,
		lastTime:  time.Now(),
	}
}

func (l *loginLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastTime).Seconds()
	l.lastTime = now
	l.tokens += elapsed * l.refill
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

var loginLimiterInstance = newLoginLimiter(5)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Login(c *gin.Context) {
	// Rate limit login attempts.
	if !loginLimiterInstance.allow() {
		c.JSON(429, gin.H{"error": "too many login attempts, try again later"})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	// Use constant-time comparison to prevent timing attacks.
	expectedLogin := os.Getenv("ADMIN_LOGIN")
	expectedPass := os.Getenv("ADMIN_PASSWORD")
	if req.Login == "" || expectedLogin == "" || expectedPass == "" {
		log.Println("[auth] missing ADMIN_LOGIN or ADMIN_PASSWORD configuration")
		c.JSON(500, gin.H{"error": "server configuration error"})
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.Login), []byte(expectedLogin)) != 1 ||
		!comparePasswords(req.Password, expectedPass) {
		c.JSON(401, gin.H{"error": "invalid login or password"})
		return
	}

	token, err := GenerateToken(req.Login)
	if err != nil {
		log.Println("[auth] failed to generate token:", err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	secure := os.Getenv("COOKIE_SECURE") == "true"
	var sameSite http.SameSite
	if secure {
		sameSite = http.SameSiteStrictMode
	} else {
		sameSite = http.SameSiteLaxMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie("jwt_token", token, 3600*24, "", "", secure, true)

	c.JSON(200, gin.H{
		"message": "login successful",
	})
}

// comparePasswords uses constant-time comparison to prevent timing attacks.
func comparePasswords(provided, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
