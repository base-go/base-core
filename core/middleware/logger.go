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

const maxBodyLength = 1000 // Maximum length for logged bodies

func truncateBody(body []byte) string {
	if body == nil {
		return ""
	}
	if len(body) == 0 {
		return ""
	}
	if len(body) > maxBodyLength {
		return string(body[:maxBodyLength]) + "... [truncated]"
	}
	return string(body)
}

func truncateOrSkipResponse(body string, contentType string, path string) string {
	if body == "" {
		return ""
	}

	// Skip logging HTML responses
	if contentType == "text/html" || contentType == "text/html; charset=utf-8" {
		return "[html content]"
	}

	// Skip logging binary responses
	if contentType == "application/octet-stream" {
		return "[binary content]"
	}

	// Skip logging image responses
	if len(contentType) >= 6 && contentType[:6] == "image/" {
		return "[image content]"
	}

	// Skip logging Swagger documentation
	if path == "/swagger/doc.json" || path == "/swagger/swagger.json" {
		return "[swagger doc]"
	}

	if len(body) > maxBodyLength {
		return body[:maxBodyLength] + "... [truncated]"
	}
	return body
}

// Logger middleware for logging requests and responses
func Logger(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c == nil || c.Request == nil {
			return
		}

		// Start timer
		start := time.Now()

		// Read the request body
		var requestBody []byte
		if c.Request.Body != nil {
			var err error
			requestBody, err = io.ReadAll(c.Request.Body)
			if err == nil {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}
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
		if status >= 400 && responseBody != "" {
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
			logger.String("request", truncateBody(requestBody)),
			logger.String("response", truncateOrSkipResponse(responseBody, c.Writer.Header().Get("Content-Type"), c.Request.URL.Path)),
			logger.String("error", errorMsg),
		)
	}
}
