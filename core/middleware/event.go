package middleware

import (
	"time"

	"base/core/event"

	"github.com/gin-gonic/gin"
)

// EventTrackingMiddleware creates a middleware that tracks HTTP requests as events
func EventTrackingMiddleware(eventService *event.EventService) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Skip certain paths
		if c.Request.URL.Path == "/ping" || c.Request.URL.Path == "/health" {
			return
		}

		// Determine status
		status := "success"
		if len(c.Errors) > 0 || c.Writer.Status() >= 400 {
			status = "failed"
		}

		// Get user ID from context if available
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		// Create event
		_, err := eventService.Track(c.Request.Context(), event.EventOptions{
			Type:     "http_request",
			Category: "api",
			Actor:    "user",
			ActorID:  userID,
			Target:   c.Request.URL.Path,
			Action:   c.Request.Method,
			Status:   status,
			Metadata: map[string]interface{}{
				"path":        c.Request.URL.Path,
				"method":      c.Request.Method,
				"ip":         c.ClientIP(),
				"user_agent": c.Request.UserAgent(),
				"status":     c.Writer.Status(),
				"latency_ms": time.Since(startTime).Milliseconds(),
				"query":      c.Request.URL.RawQuery,
				"errors":     c.Errors.Errors(),
			},
		})

		if err != nil {
			// Log error but don't affect the request
			c.Error(err)
		}
	}
}
