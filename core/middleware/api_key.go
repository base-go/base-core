package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.IsWebsocket() {
			c.Next()
			return
		}

		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
		if c.Request.URL.Path == "/public" {
			c.Next()
			return
		}

		apiKey := c.GetHeader("X-Api-Key")
		expectedAPIKey := os.Getenv("API_KEY")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if apiKey != expectedAPIKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}

		c.Next()
	}
}
