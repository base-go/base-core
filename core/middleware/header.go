package middleware

import (
	"base/core/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// HeaderPrefix is the prefix for all Base headers
	HeaderPrefix = "Base-"
	// ContextKeyPrefix is the prefix for context keys
	ContextKeyPrefix = "base_"
)

// HeaderMiddleware extracts headers with the Base- prefix and stores them in the context
// This allows for generic header handling without hardcoding specific headers
func HeaderMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process all headers
		for key, values := range c.Request.Header {
			// Check if this is a Base header
			if strings.HasPrefix(key, HeaderPrefix) {
				// Strip the prefix and convert to lowercase for context key
				contextKey := strings.ToLower(strings.Replace(key, HeaderPrefix, ContextKeyPrefix, 1))
				
				// Store the first value in the context
				if len(values) > 0 {
					log.Debug("Setting context key", logger.String("key", contextKey), logger.String("value", values[0]))
					c.Set(contextKey, values[0])
				}
			}
		}
		
		c.Next()
	}
}
