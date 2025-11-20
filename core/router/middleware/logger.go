package middleware

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"base/core/logger"
	"base/core/router"
)

// LoggerConfig contains logger middleware configuration
type LoggerConfig struct {
	// Logger is the logger instance to use
	Logger logger.Logger

	// SkipPaths lists paths that shouldn't be logged
	SkipPaths []string

	// LogLevel determines what level to log at
	LogLevel string

	// IncludeBody includes request/response body in logs
	IncludeBody bool

	// IncludeHeaders includes headers in logs
	IncludeHeaders bool
}

// DefaultLoggerConfig returns default logger configuration
func DefaultLoggerConfig(log logger.Logger) *LoggerConfig {
	return &LoggerConfig{
		Logger:   log,
		LogLevel: "info",
	}
}

// Logger creates logging middleware
func Logger(config *LoggerConfig) router.MiddlewareFunc {
	if config == nil || config.Logger == nil {
		panic("Logger is required for logger middleware")
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Check if path should be skipped
			for _, path := range config.SkipPaths {
				if c.Request.URL.Path == path {
					return next(c)
				}
			}

			start := time.Now()
			path := c.Request.URL.Path
			raw := c.Request.URL.RawQuery

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Get response status
			status := c.Writer.Status()

			// Build log fields
			fields := []logger.Field{
				logger.String("method", c.Request.Method),
				logger.String("path", path),
				logger.Int("status", status),
				logger.Duration("latency", latency),
				logger.String("ip", c.ClientIP()),
				logger.String("user_agent", c.Request.UserAgent()),
			}

			if raw != "" {
				fields = append(fields, logger.String("query", raw))
			}

			if config.IncludeHeaders {
				headers := make(map[string][]string)
				for k, v := range c.Request.Header {
					headers[k] = v
				}
				fields = append(fields, logger.Any("headers", headers))
			}

			if err != nil {
				fields = append(fields, logger.String("error", err.Error()))
			}

			// Log based on status code
			switch {
			case status >= 500:
				config.Logger.Error("Server error", fields...)
			case status >= 400:
				config.Logger.Warn("Client error", fields...)
			case status >= 300:
				config.Logger.Info("Redirect", fields...)
			default:
				config.Logger.Info("Request", fields...)
			}

			return err
		}
	}
}

// Recovery creates panic recovery middleware
func Recovery(log logger.Logger) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					// Log the panic
					log.Error("Panic recovered",
						logger.Any("panic", r),
						logger.String("path", c.Request.URL.Path),
						logger.String("method", c.Request.Method),
						logger.String("ip", c.ClientIP()),
					)

					// Return 500 error
					err = c.JSON(500, map[string]string{
						"error": "Internal server error",
					})
				}
			}()

			return next(c)
		}
	}
}

// RequestId generates and adds a request Id to the context
func RequestId() router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Generate request Id
			requestId := generateRequestId()

			// Add to context
			c.Set("request_id", requestId)

			// Add to response header
			c.SetHeader("X-Request-Id", requestId)

			return next(c)
		}
	}
}

// generateRequestId generates a unique request Id
func generateRequestId() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// AccessLog creates access log middleware with custom format
func AccessLog(format string, log logger.Logger) router.MiddlewareFunc {
	if format == "" {
		format = ":method :path :status :latency :ip"
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Calculate latency
			latency := time.Since(start)

			// Format log message
			msg := format
			msg = replaceToken(msg, ":method", c.Request.Method)
			msg = replaceToken(msg, ":path", c.Request.URL.Path)
			msg = replaceToken(msg, ":status", fmt.Sprintf("%d", c.Writer.Status()))
			msg = replaceToken(msg, ":latency", latency.String())
			msg = replaceToken(msg, ":ip", c.ClientIP())
			msg = replaceToken(msg, ":user_agent", c.Request.UserAgent())

			// Log the access
			log.Info(msg)

			return err
		}
	}
}

// replaceToken replaces a token in the format string
func replaceToken(format, token, value string) string {
	return strings.ReplaceAll(format, token, value)
}

// Metrics creates metrics collection middleware
func Metrics(collector MetricsCollector) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Collect metrics
			collector.RecordRequest(
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				time.Since(start),
			)

			return err
		}
	}
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	RecordRequest(method, path string, status int, duration time.Duration)
}

// SimpleMetricsCollector is a simple in-memory metrics collector
type SimpleMetricsCollector struct {
	requests map[string]*RequestMetrics
	mu       sync.RWMutex
}

// RequestMetrics contains metrics for a specific endpoint
type RequestMetrics struct {
	Count       int64
	TotalTime   time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	StatusCodes map[int]int64
}

// NewSimpleMetricsCollector creates a new simple metrics collector
func NewSimpleMetricsCollector() *SimpleMetricsCollector {
	return &SimpleMetricsCollector{
		requests: make(map[string]*RequestMetrics),
	}
}

// RecordRequest records a request metric
func (s *SimpleMetricsCollector) RecordRequest(method, path string, status int, duration time.Duration) {
	key := fmt.Sprintf("%s:%s", method, path)

	s.mu.Lock()
	defer s.mu.Unlock()

	metrics, exists := s.requests[key]
	if !exists {
		metrics = &RequestMetrics{
			StatusCodes: make(map[int]int64),
			MinTime:     duration,
			MaxTime:     duration,
		}
		s.requests[key] = metrics
	}

	metrics.Count++
	metrics.TotalTime += duration

	if duration < metrics.MinTime {
		metrics.MinTime = duration
	}
	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}

	metrics.StatusCodes[status]++
}

// GetMetrics returns all collected metrics
func (s *SimpleMetricsCollector) GetMetrics() map[string]*RequestMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	result := make(map[string]*RequestMetrics)
	for k, v := range s.requests {
		result[k] = v
	}
	return result
}
