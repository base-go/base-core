package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"base/core/router"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow returns true if the request should be allowed
	Allow(key string) bool

	// Reset resets the rate limiter for a specific key
	Reset(key string)
}

// TokenBucket implements token bucket rate limiting
type TokenBucket struct {
	rate      int           // tokens per interval
	interval  time.Duration // interval duration
	maxTokens int           // maximum tokens in bucket
	buckets   map[string]*bucket
	mu        sync.RWMutex
	cleanup   *time.Ticker
}

type bucket struct {
	tokens   int
	lastFill time.Time
	mu       sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(rate int, interval time.Duration, maxTokens int) *TokenBucket {
	tb := &TokenBucket{
		rate:      rate,
		interval:  interval,
		maxTokens: maxTokens,
		buckets:   make(map[string]*bucket),
		cleanup:   time.NewTicker(5 * time.Minute),
	}

	// Start cleanup goroutine
	go tb.cleanupRoutine()

	return tb
}

// Allow checks if a request should be allowed
func (tb *TokenBucket) Allow(key string) bool {
	tb.mu.RLock()
	b, exists := tb.buckets[key]
	tb.mu.RUnlock()

	if !exists {
		tb.mu.Lock()
		b = &bucket{
			tokens:   tb.maxTokens,
			lastFill: time.Now(),
		}
		tb.buckets[key] = b
		tb.mu.Unlock()
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(b.lastFill)
	tokensToAdd := int(elapsed/tb.interval) * tb.rate

	if tokensToAdd > 0 {
		b.tokens = min(b.tokens+tokensToAdd, tb.maxTokens)
		b.lastFill = now
	}

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// Reset resets the rate limiter for a specific key
func (tb *TokenBucket) Reset(key string) {
	tb.mu.Lock()
	delete(tb.buckets, key)
	tb.mu.Unlock()
}

// cleanupRoutine removes old buckets periodically
func (tb *TokenBucket) cleanupRoutine() {
	for range tb.cleanup.C {
		tb.mu.Lock()
		now := time.Now()
		for key, b := range tb.buckets {
			b.mu.Lock()
			if now.Sub(b.lastFill) > 1*time.Hour {
				delete(tb.buckets, key)
			}
			b.mu.Unlock()
		}
		tb.mu.Unlock()
	}
}

// Stop stops the cleanup routine
func (tb *TokenBucket) Stop() {
	tb.cleanup.Stop()
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	// Limiter is the rate limiter implementation
	Limiter RateLimiter

	// KeyFunc extracts the key from the request
	KeyFunc func(*router.Context) string

	// ErrorHandler handles rate limit errors
	ErrorHandler func(*router.Context) error

	// SkipPaths lists paths that don't require rate limiting
	SkipPaths []string
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		Limiter: NewTokenBucket(60, time.Minute, 60), // 60 requests per minute
		KeyFunc: func(c *router.Context) string {
			return c.ClientIP()
		},
		ErrorHandler: func(c *router.Context) error {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": "Rate limit exceeded",
			})
		},
	}
}

// RateLimit creates rate limiting middleware
func RateLimit(config *RateLimitConfig) router.MiddlewareFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Check if path should be skipped
			for _, path := range config.SkipPaths {
				if c.Request.URL.Path == path {
					return next(c)
				}
			}

			// Get rate limit key
			key := config.KeyFunc(c)

			// Check rate limit
			if !config.Limiter.Allow(key) {
				return config.ErrorHandler(c)
			}

			return next(c)
		}
	}
}

// PerEndpointRateLimit creates per-endpoint rate limiting
func PerEndpointRateLimit(requests int, duration time.Duration) router.MiddlewareFunc {
	limiter := NewTokenBucket(requests, duration, requests)

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			// Create key from IP + path
			key := fmt.Sprintf("%s:%s:%s", c.ClientIP(), c.Request.Method, c.Request.URL.Path)

			if !limiter.Allow(key) {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Rate limit exceeded for this endpoint",
				})
			}

			return next(c)
		}
	}
}

// SlidingWindow implements sliding window rate limiting
type SlidingWindow struct {
	windowSize  time.Duration
	maxRequests int
	requests    map[string][]time.Time
	mu          sync.RWMutex
}

// NewSlidingWindow creates a new sliding window rate limiter
func NewSlidingWindow(windowSize time.Duration, maxRequests int) *SlidingWindow {
	sw := &SlidingWindow{
		windowSize:  windowSize,
		maxRequests: maxRequests,
		requests:    make(map[string][]time.Time),
	}

	// Start cleanup routine
	go sw.cleanup()

	return sw
}

// Allow checks if a request should be allowed
func (sw *SlidingWindow) Allow(key string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-sw.windowSize)

	// Get or create request history
	history, exists := sw.requests[key]
	if !exists {
		sw.requests[key] = []time.Time{now}
		return true
	}

	// Remove old requests outside window
	validRequests := []time.Time{}
	for _, t := range history {
		if t.After(windowStart) {
			validRequests = append(validRequests, t)
		}
	}

	// Check if under limit
	if len(validRequests) < sw.maxRequests {
		validRequests = append(validRequests, now)
		sw.requests[key] = validRequests
		return true
	}

	sw.requests[key] = validRequests
	return false
}

// Reset resets the rate limiter for a specific key
func (sw *SlidingWindow) Reset(key string) {
	sw.mu.Lock()
	delete(sw.requests, key)
	sw.mu.Unlock()
}

// cleanup removes old entries periodically
func (sw *SlidingWindow) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sw.mu.Lock()
		now := time.Now()
		for key, history := range sw.requests {
			if len(history) == 0 || now.Sub(history[len(history)-1]) > sw.windowSize {
				delete(sw.requests, key)
			}
		}
		sw.mu.Unlock()
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
