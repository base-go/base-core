package cache

import (
	"context"
	"time"
)

// Cache defines the interface for all cache implementations
type Cache interface {
	// Get retrieves a value from cache and unmarshals it into dest
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in cache with the specified TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a single key from cache
	Delete(ctx context.Context, key string) error

	// DeletePattern removes all keys matching the pattern (supports wildcards like "user:*")
	DeletePattern(ctx context.Context, pattern string) error

	// Has checks if a key exists in cache
	Has(ctx context.Context, key string) (bool, error)

	// Clear removes all keys from cache
	Clear(ctx context.Context) error

	// Close cleanly shuts down the cache provider
	Close() error

	// GetStats returns cache statistics (hits, misses, size)
	GetStats() Stats
}

// Stats represents cache statistics
type Stats struct {
	Hits        int64
	Misses      int64
	Sets        int64
	Deletes     int64
	Size        int64
	Evictions   int64
	Provider    string
	LastCleanup time.Time
}

// ErrCacheMiss is returned when a key is not found in cache
var ErrCacheMiss = &CacheError{Message: "cache miss", Type: "miss"}

// ErrCacheNotAvailable is returned when the cache provider is not available
var ErrCacheNotAvailable = &CacheError{Message: "cache not available", Type: "unavailable"}

// CacheError represents a cache-specific error
type CacheError struct {
	Message string
	Type    string
	Err     error
}

func (e *CacheError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

// IsMiss returns true if the error is a cache miss
func IsMiss(err error) bool {
	if err == nil {
		return false
	}
	ce, ok := err.(*CacheError)
	return ok && ce.Type == "miss"
}
