package middleware

import (
	"base/core/router"
	"net/http"
	"os"
	"strings"
)

func Api() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Skip middleware for non-JSON requests
			if !isJSONRequest(c) {
				return next(c)
			}

			// Exceptions
			if c.IsWebSocket() || c.Request.Method == "OPTIONS" || c.Request.URL.Path == "/public" {
				return next(c)
			}

			apiKey := c.GetHeader("X-Api-Key")
			expectedAPIKey := os.Getenv("API_KEY")
			if apiKey == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: API key is required"})
				return nil
			}

			if apiKey != expectedAPIKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized: Invalid API key"})
				return nil
			}

			return next(c)
		}
	}
}

// Helper function to determine if the request is a JSON request
func isJSONRequest(c *router.Context) bool {
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
