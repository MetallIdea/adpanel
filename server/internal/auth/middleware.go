package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("jwt_token")
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid cookie",
			})
			c.Abort()
			return
		}

		err = ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
