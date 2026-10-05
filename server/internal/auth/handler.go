package auth

import (
	"os"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Login != os.Getenv("ADMIN_LOGIN") ||
		req.Password != os.Getenv("ADMIN_PASSWORD") {

		c.JSON(401, gin.H{"error": "invalid login or password"})
		return
	}

	token, err := GenerateToken(req.Login)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("jwt_token", token, 3600*24, "", "", false, true)

	c.JSON(200, gin.H{
		"message": "login successful",
	})
}
