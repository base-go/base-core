package middleware

import (
	"base/core/cache"
	"base/core/router"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"
)

// CacheMiddlewareConfig holds configuration for cache middleware
type CacheMiddlewareConfig struct {
	TTL                time.Duration
	CacheableStatuses  []int
	CacheableMethods   []string
	SkipPaths          []string
	VaryByQueryParams  bool
	VaryByHeaders      []string
}

// DefaultCacheMiddlewareConfig returns default cache middleware configuration
func DefaultCacheMiddlewareConfig() CacheMiddlewareConfig {
	return CacheMiddlewareConfig{
		TTL:               5 * time.Minute,
		CacheableStatuses: []int{200},
		CacheableMethods:  []string{"GET"},
		SkipPaths:         []string{},
		VaryByQueryParams: true,
		VaryByHeaders:     []string{},
	}
}

// CacheMiddleware creates a middleware that caches HTTP responses
func CacheMiddleware(c cache.Cache, config CacheMiddlewareConfig) router.MiddlewareFunc {
	if config.CacheableStatuses == nil {
		config = DefaultCacheMiddlewareConfig()
	}

	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(ctx *router.Context) error {
			// Check if method is cacheable
			if !contains(config.CacheableMethods, ctx.Request.Method) {
				return next(ctx)
			}

			// Check if path should be skipped
			for _, skipPath := range config.SkipPaths {
				if ctx.Request.URL.Path == skipPath {
					return next(ctx)
				}
			}

			// Generate cache key
			query := ""
			if config.VaryByQueryParams {
				query = ctx.Request.URL.RawQuery
			}
			cacheKey := cache.HTTPResponseKey(ctx.Request.Method, ctx.Request.URL.Path, query)

			// Try to get cached response
			var cachedResp CachedResponse
			err := c.Get(context.Background(), cacheKey, &cachedResp)
			if err == nil {
				// Cache hit - return cached response
				ctx.Writer.WriteHeader(cachedResp.Status)
				for key, values := range cachedResp.Headers {
					for _, value := range values {
						ctx.Writer.Header().Add(key, value)
					}
				}
				ctx.Writer.Header().Set("X-Cache", "HIT")
				ctx.Writer.Write(cachedResp.Body)
				return nil
			}

			// Cache miss - execute handler and cache response
			// Create a response recorder to capture the response
			recorder := &responseRecorder{
				ResponseWriter: ctx.Writer,
				status:         200,
				body:           &bytes.Buffer{},
			}
			ctx.Writer = recorder

			// Execute the next handler
			err = next(ctx)
			if err != nil {
				return err
			}

			// Check if response should be cached
			if contains(config.CacheableStatuses, recorder.status) {
				cachedResp = CachedResponse{
					Status:  recorder.status,
					Headers: recorder.Header().Clone(),
					Body:    recorder.body.Bytes(),
				}

				// Cache the response
				c.Set(context.Background(), cacheKey, cachedResp, config.TTL)
			}

			// Set cache miss header
			ctx.Writer.Header().Set("X-Cache", "MISS")

			return nil
		}
	}
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}

// responseRecorder is a custom ResponseWriter that records the response
type responseRecorder struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
	body    *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.written {
		return
	}
	r.status = status
	r.written = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Size() int {
	return r.size
}

func (r *responseRecorder) Written() bool {
	return r.written
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *responseRecorder) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// contains checks if a slice contains a specific value
func contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// CacheStats returns cache statistics as JSON
func CacheStatsHandler(c cache.Cache) router.HandlerFunc {
	return func(ctx *router.Context) error {
		stats := c.GetStats()

		// Calculate hit rate
		total := stats.Hits + stats.Misses
		hitRate := 0.0
		if total > 0 {
			hitRate = float64(stats.Hits) / float64(total) * 100
		}

		response := map[string]interface{}{
			"provider":    stats.Provider,
			"hits":        stats.Hits,
			"misses":      stats.Misses,
			"hit_rate":    hitRate,
			"sets":        stats.Sets,
			"deletes":     stats.Deletes,
			"size":        stats.Size,
			"evictions":   stats.Evictions,
			"last_cleanup": stats.LastCleanup,
		}

		ctx.Writer.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(ctx.Writer).Encode(response)
	}
}

// CacheClearHandler clears the cache
func CacheClearHandler(c cache.Cache) router.HandlerFunc {
	return func(ctx *router.Context) error {
		if err := c.Clear(context.Background()); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to clear cache"})
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Cache cleared successfully",
		}

		return ctx.JSON(http.StatusOK, response)
	}
}

// CacheInvalidateHandler invalidates cache entries matching a pattern
func CacheInvalidateHandler(c cache.Cache) router.HandlerFunc {
	return func(ctx *router.Context) error {
		// Read pattern from request body
		var req struct {
			Pattern string `json:"pattern"`
		}

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "Failed to read request body"})
		}

		if err := json.Unmarshal(body, &req); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "Invalid JSON"})
		}

		if req.Pattern == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]any{"error": "Pattern is required"})
		}

		// Invalidate matching cache entries
		if err := c.DeletePattern(context.Background(), req.Pattern); err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]any{"error": "Failed to invalidate cache"})
		}

		response := map[string]interface{}{
			"success": true,
			"message": "Cache invalidated successfully",
			"pattern": req.Pattern,
		}

		return ctx.JSON(http.StatusOK, response)
	}
}
