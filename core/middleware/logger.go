package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
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

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

var (
	// Regex patterns for sensitive data
	passwordPattern = regexp.MustCompile("(\"password\"\\s*:\\s*\")[^\"]*(\")")
	tokenPattern    = regexp.MustCompile("(\"accessToken\"\\s*:\\s*\")[^\"]*(\")")
	keyPattern      = regexp.MustCompile("(\"api_key\"\\s*:\\s*\")[^\"]*(\")")
	authPattern     = regexp.MustCompile("(\"authorization\"\\s*:\\s*\")[^\"]*(\")")
)

// sanitizeBody removes sensitive information from request/response bodies
func sanitizeBody(body string) string {
	body = passwordPattern.ReplaceAllString(body, `${1}***${2}`)
	body = tokenPattern.ReplaceAllString(body, `${1}***${2}`)
	body = keyPattern.ReplaceAllString(body, `${1}***${2}`)
	body = authPattern.ReplaceAllString(body, `${1}***${2}`)
	return body
}

func truncateBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	bodyStr := sanitizeBody(string(body))
	if len(bodyStr) > maxBodyLength {
		return bodyStr[:maxBodyLength] + "... [truncated]"
	}
	return bodyStr
}

func truncateOrSkipResponse(body string, contentType string, path string) string {
	if body == "" {
		return ""
	}

	// Skip logging HTML responses
	if strings.Contains(contentType, "text/html") {
		return "[html content]"
	}

	// Skip logging binary responses
	if contentType == "application/octet-stream" {
		return "[binary content]"
	}

	// Skip logging image responses
	if strings.HasPrefix(contentType, "image/") {
		return "[image content]"
	}

	// Skip logging large file responses
	if strings.Contains(contentType, "application/pdf") || strings.Contains(contentType, "application/zip") {
		return "[file content]"
	}

	// Skip logging Swagger documentation
	if strings.Contains(path, "/swagger/") {
		return "[swagger doc]"
	}

	// Sanitize sensitive data
	body = sanitizeBody(body)

	if len(body) > maxBodyLength {
		return body[:maxBodyLength] + "... [truncated]"
	}
	return body
}

// getStatusEmoji returns an emoji based on HTTP status code
func getStatusEmoji(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "✅"
	case status >= 300 && status < 400:
		return "🔄"
	case status >= 400 && status < 500:
		return "❌"
	case status >= 500:
		return "💥"
	default:
		return "❓"
	}
}

// getLatencyEmoji returns an emoji based on request latency
func getLatencyEmoji(latency time.Duration) string {
	switch {
	case latency < 10*time.Millisecond:
		return "🚀" // Very fast
	case latency < 100*time.Millisecond:
		return "⚡" // Fast
	case latency < 500*time.Millisecond:
		return "🐎" // Medium
	case latency < 2*time.Second:
		return "🐌" // Slow
	default:
		return "🦥" // Very slow
	}
}

// getMethodEmoji returns an emoji based on HTTP method
func getMethodEmoji(method string) string {
	switch method {
	case "GET":
		return "📥"
	case "POST":
		return "📤"
	case "PUT":
		return "🔄"
	case "DELETE":
		return "🗑️"
	case "PATCH":
		return "🔧"
	default:
		return "❓"
	}
}

// getStatusColor returns ANSI color code based on HTTP status
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return ColorGreen
	case status >= 300 && status < 400:
		return ColorYellow
	case status >= 400 && status < 500:
		return ColorRed
	case status >= 500:
		return ColorRed + ColorBold
	default:
		return ColorWhite
	}
}

// getMethodColor returns ANSI color code based on HTTP method
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return ColorBlue
	case "POST":
		return ColorGreen
	case "PUT":
		return ColorYellow
	case "DELETE":
		return ColorRed
	case "PATCH":
		return ColorPurple
	default:
		return ColorWhite
	}
}

// getLatencyColor returns ANSI color code based on latency
func getLatencyColor(latency time.Duration) string {
	switch {
	case latency < 10*time.Millisecond:
		return ColorGreen + ColorBold
	case latency < 100*time.Millisecond:
		return ColorGreen
	case latency < 500*time.Millisecond:
		return ColorYellow
	case latency < 2*time.Second:
		return ColorRed
	default:
		return ColorRed + ColorBold
	}
}

// shouldLogDetails determines if we should log request/response details
func shouldLogDetails(path string) bool {
	// Skip health checks and static files
	skipPaths := []string{"/health", "/static/", "/favicon.ico", "/swagger/"}
	for _, skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			return false
		}
	}
	return true
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

		// Calculate latency
		latency := time.Since(start)
		status := c.Writer.Status()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Get client IP
		clientIP := c.ClientIP()
		if c.Request.Header.Get("X-Forwarded-For") != "" {
			clientIP = c.Request.Header.Get("X-Forwarded-For")
		}

		// Get User-Agent (truncated)
		userAgent := c.Request.Header.Get("User-Agent")
		if len(userAgent) > 50 {
			userAgent = userAgent[:50] + "..."
		}

		// Create a colorful and human-readable log message
		humanDetails := fmt.Sprintf("\nfrom %s%s%s", ColorDim, clientIP, ColorReset)

		logMessage := fmt.Sprintf("%s%s%s %s%s%s %s%s%s %s%s%s %s%s %s%s %s",
			getStatusColor(status), getStatusEmoji(status), ColorReset,
			getMethodColor(method), getMethodEmoji(method), ColorReset,
			getMethodColor(method), method, ColorReset,
			ColorCyan, path, ColorReset,
			getLatencyColor(latency), getLatencyEmoji(latency), latency.String(), ColorReset,
			humanDetails,
		)

		// Add more details for important requests or errors
		if shouldLogDetails(path) {
			responseBody := w.body.String()
			var additionalInfo []string

			// Parse response body to get error message if it exists
			var errorMsg string
			if status >= 400 && responseBody != "" {
				var errResp types.ErrorResponse
				if err := json.Unmarshal([]byte(responseBody), &errResp); err == nil {
					errorMsg = errResp.Error
				}
			}

			// Add query parameters if they exist
			if queryParams := c.Request.URL.RawQuery; queryParams != "" {
				additionalInfo = append(additionalInfo, fmt.Sprintf("\nquery: %s%s%s", ColorDim, queryParams, ColorReset))
			}

			// Add request body for POST, PUT, PATCH requests
			if method == "POST" || method == "PUT" || method == "PATCH" {
				if reqBody := truncateBody(requestBody); reqBody != "" {
					additionalInfo = append(additionalInfo, fmt.Sprintf("\nrequest: %s%s%s", ColorDim, reqBody, ColorReset))
				}
			}

			// Add response body (especially for errors)
			if responseBody := truncateOrSkipResponse(responseBody, c.Writer.Header().Get("Content-Type"), path); responseBody != "" {
				additionalInfo = append(additionalInfo, fmt.Sprintf("\nresponse: %s%s%s", ColorDim, responseBody, ColorReset))
			}

			// Add error message if exists
			if errorMsg != "" {
				additionalInfo = append(additionalInfo, fmt.Sprintf("\nerror: %s%s%s", ColorRed, errorMsg, ColorReset))
			}

			// Add User-Agent for debugging
			if userAgent != "" {
				additionalInfo = append(additionalInfo, fmt.Sprintf("\nuser-agent: %s%s%s", ColorDim, userAgent, ColorReset))
			}

			// Add content length
			if c.Request.ContentLength > 0 {
				additionalInfo = append(additionalInfo, fmt.Sprintf("\nsize: %s%d bytes%s", ColorDim, c.Request.ContentLength, ColorReset))
			}

			// Append additional info to the log message
			if len(additionalInfo) > 0 {
				logMessage += fmt.Sprintf(" | %s", strings.Join(additionalInfo, " | "))
			}
		}

		// Choose log level based on status - no structured fields for clean output
		switch {
		case status >= 500:
			log.Error(logMessage)
		case status >= 400:
			log.Warn(logMessage)
		case status >= 300:
			log.Info(logMessage)
		default:
			log.Info(logMessage)
		}
	}
}
