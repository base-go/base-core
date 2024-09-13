package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func APIKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip middleware for non-JSON requests
		if !isJSONRequest(c) {
			c.Next()
			return
		}

		// Existing exceptions
		if c.IsWebsocket() || c.Request.Method == "OPTIONS" || c.Request.URL.Path == "/public" {
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

// Helper function to determine if the request is a JSON request
func isJSONRequest(c *gin.Context) bool {
	// Check Accept header
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		return true
	}

	// Check Content-Type header for POST, PUT, PATCH requests
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
		if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			return true
		}
	}

	// Check if the URL ends with .json
	if strings.HasSuffix(c.Request.URL.Path, ".json") {
		return true
	}

	// Check if there's a format=json query parameter
	if c.Query("format") == "json" {
		return true
	}

	return false
}
