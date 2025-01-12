package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"base/core/logger"
	"base/core/types"

	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Logger middleware for logging requests and responses
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Read the request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a new response body writer
		w := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		// Process request
		c.Next()

		// Get response status and body
		status := c.Writer.Status()
		responseBody := w.body.String()

		// Parse response body to get error message if it exists
		var errorMsg string
		if status >= 400 {
			var errResp types.ErrorResponse
			if err := json.Unmarshal([]byte(responseBody), &errResp); err == nil {
				errorMsg = errResp.Error
			}
		}

		// Log the request details
		log.Info("HTTP Request",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.Int("status", status),
			logger.String("latency", time.Since(start).String()),
			logger.String("request", string(requestBody)),
			logger.String("response", responseBody),
			logger.String("error", errorMsg),
		)
	}
}
