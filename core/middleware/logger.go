package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger returns a gin.HandlerFunc (middleware) that logs requests using Zap.
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if path != "/ping" && path != "/health" {
			end := time.Now()
			latency := end.Sub(start)

			if len(c.Errors) > 0 {
				// Collect all error messages
				errorMessages := make([]string, len(c.Errors))
				for i, err := range c.Errors {
					errorMessages[i] = err.Error()
				}

				logger.Error("Request failed",
					zap.String("path", path),
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.Duration("latency", latency),
					zap.String("ip", c.ClientIP()),
					zap.String("query", query),
					zap.Strings("errors", errorMessages),
					zap.String("user-agent", c.Request.UserAgent()),
				)
			} else {
				logger.Info("Request processed",
					zap.String("path", path),
					zap.Int("status", c.Writer.Status()),
					zap.String("method", c.Request.Method),
					zap.Duration("latency", latency),
					zap.String("ip", c.ClientIP()),
					zap.String("query", query),
					zap.String("user-agent", c.Request.UserAgent()),
				)
			}
		}
	}
}
