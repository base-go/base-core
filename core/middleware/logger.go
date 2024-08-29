package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// LogrusLogger is a middleware that logs requests using Logrus
func LogrusLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Parse user agent
		ua := c.Request.UserAgent()
		os, browser := parseUserAgent(ua)

		entry := logger.WithFields(logrus.Fields{
			"status":  c.Writer.Status(),
			"method":  c.Request.Method,
			"path":    path,
			"ip":      c.ClientIP(),
			"latency": latency.String(),
			"os":      os,
			"browser": browser,
			"error":   c.Errors.ByType(gin.ErrorTypePrivate).String(),
		})

		if c.Writer.Status() >= 500 {
			entry.Error("Server error")
		} else {
			entry.Info("Request")
		}
	}
}

// parseUserAgent extracts OS and browser information from the user agent string
func parseUserAgent(ua string) (os, browser string) {
	ua = strings.ToLower(ua)

	// OS detection
	switch {
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "mac os"):
		os = "macOS"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "ios"):
		os = "iOS"
	default:
		os = "Unknown"
	}

	// Browser detection
	switch {
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "chrome"):
		browser = "Chrome"
	case strings.Contains(ua, "safari"):
		browser = "Safari"
	case strings.Contains(ua, "opera"):
		browser = "Opera"
	case strings.Contains(ua, "edge"):
		browser = "Edge"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident"):
		browser = "Internet Explorer"
	default:
		browser = "Unknown"
	}

	return
}
