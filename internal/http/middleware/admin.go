package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func Admin() gin.HandlerFunc {
	token := os.Getenv("ADMIN_TOKEN")
	return func(c *gin.Context) {
		if token == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized, 
				gin.H{"code": "INVALID_TOKEN", "message": "no admin token"},
			)
			return
		}
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized, 
				gin.H{"code": "INVALID_TOKEN", "message": "no token header"},
			)
			return
		}
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized, 
				gin.H{"code": "INVALID_TOKEN", "message": "invalid token header"},
			)
			return
		}
		if strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")) != token {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized, 
				gin.H{"code": "INVALID_TOKEN", "message": "invalid token"},
			)
			return
		}
		c.Next()
	}
}