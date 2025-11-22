package cache

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MemoryCache implements an in-memory cache with TTL and LRU eviction
type MemoryCache struct {
	data            sync.Map
	maxSize         int
	currentSize     int64
	cleanupInterval time.Duration
	stopCleanup     chan struct{}

	// Stats
	hits      int64
	misses    int64
	sets      int64
	deletes   int64
	evictions int64
	lastCleanup time.Time
	mu        sync.RWMutex
}

type cacheEntry struct {
	value      []byte
	expiration time.Time
	lastAccess time.Time
}

// MemoryConfig holds configuration for in-memory cache
type MemoryConfig struct {
	MaxSize         int
	CleanupInterval time.Duration
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config MemoryConfig) *MemoryCache {
	if config.MaxSize <= 0 {
		config.MaxSize = 10000 // Default max size
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	cache := &MemoryCache{
		maxSize:         config.MaxSize,
		cleanupInterval: config.CleanupInterval,
		stopCleanup:     make(chan struct{}),
		lastCleanup:     time.Now(),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a value from cache
func (m *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, ok := m.data.Load(key)
	if !ok {
		atomic.AddInt64(&m.misses, 1)
		return ErrCacheMiss
	}

	entry := val.(*cacheEntry)

	// Check expiration
	if time.Now().After(entry.expiration) {
		m.data.Delete(key)
		atomic.AddInt64(&m.currentSize, -1)
		atomic.AddInt64(&m.misses, 1)
		return ErrCacheMiss
	}

	// Update last access time for LRU
	entry.lastAccess = time.Now()

	// Unmarshal value
	if err := json.Unmarshal(entry.value, dest); err != nil {
		return &CacheError{Message: "unmarshal failed", Type: "unmarshal", Err: err}
	}

	atomic.AddInt64(&m.hits, 1)
	return nil
}

// Set stores a value in cache
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Marshal value
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Message: "marshal failed", Type: "marshal", Err: err}
	}

	entry := &cacheEntry{
		value:      data,
		expiration: time.Now().Add(ttl),
		lastAccess: time.Now(),
	}

	// Check if we need to evict entries
	currentSize := atomic.LoadInt64(&m.currentSize)
	if int(currentSize) >= m.maxSize {
		m.evictLRU()
	}

	// Store entry
	_, existed := m.data.LoadOrStore(key, entry)
	if !existed {
		atomic.AddInt64(&m.currentSize, 1)
	}

	atomic.AddInt64(&m.sets, 1)
	return nil
}

// Delete removes a key from cache
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	_, existed := m.data.LoadAndDelete(key)
	if existed {
		atomic.AddInt64(&m.currentSize, -1)
		atomic.AddInt64(&m.deletes, 1)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern
func (m *MemoryCache) DeletePattern(ctx context.Context, pattern string) error {
	// Convert pattern to Go-style matching (simple wildcard support)
	deleted := 0
	m.data.Range(func(key, value interface{}) bool {
		keyStr := key.(string)
		if matchPattern(keyStr, pattern) {
			m.data.Delete(key)
			deleted++
		}
		return true
	})

	atomic.AddInt64(&m.currentSize, -int64(deleted))
	atomic.AddInt64(&m.deletes, int64(deleted))
	return nil
}

// Has checks if a key exists
func (m *MemoryCache) Has(ctx context.Context, key string) (bool, error) {
	val, ok := m.data.Load(key)
	if !ok {
		return false, nil
	}

	entry := val.(*cacheEntry)
	if time.Now().After(entry.expiration) {
		m.data.Delete(key)
		atomic.AddInt64(&m.currentSize, -1)
		return false, nil
	}

	return true, nil
}

// Clear removes all entries
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.data.Range(func(key, value interface{}) bool {
		m.data.Delete(key)
		return true
	})
	atomic.StoreInt64(&m.currentSize, 0)
	return nil
}

// Close stops the cleanup goroutine
func (m *MemoryCache) Close() error {
	close(m.stopCleanup)
	return nil
}

// GetStats returns cache statistics
func (m *MemoryCache) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return Stats{
		Hits:        atomic.LoadInt64(&m.hits),
		Misses:      atomic.LoadInt64(&m.misses),
		Sets:        atomic.LoadInt64(&m.sets),
		Deletes:     atomic.LoadInt64(&m.deletes),
		Size:        atomic.LoadInt64(&m.currentSize),
		Evictions:   atomic.LoadInt64(&m.evictions),
		Provider:    "memory",
		LastCleanup: m.lastCleanup,
	}
}

// cleanup periodically removes expired entries
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.removeExpired()
			m.mu.Lock()
			m.lastCleanup = time.Now()
			m.mu.Unlock()
		case <-m.stopCleanup:
			return
		}
	}
}

// removeExpired removes all expired entries
func (m *MemoryCache) removeExpired() {
	now := time.Now()
	deleted := 0

	m.data.Range(func(key, value interface{}) bool {
		entry := value.(*cacheEntry)
		if now.After(entry.expiration) {
			m.data.Delete(key)
			deleted++
		}
		return true
	})

	atomic.AddInt64(&m.currentSize, -int64(deleted))
}

// evictLRU evicts the least recently used entry
func (m *MemoryCache) evictLRU() {
	var oldestKey interface{}
	var oldestTime time.Time

	m.data.Range(func(key, value interface{}) bool {
		entry := value.(*cacheEntry)
		if oldestKey == nil || entry.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastAccess
		}
		return true
	})

	if oldestKey != nil {
		m.data.Delete(oldestKey)
		atomic.AddInt64(&m.currentSize, -1)
		atomic.AddInt64(&m.evictions, 1)
	}
}

// matchPattern performs simple wildcard pattern matching
func matchPattern(key, pattern string) bool {
	// Support for * wildcard
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix := parts[0]
			suffix := parts[1]
			return strings.HasPrefix(key, prefix) && strings.HasSuffix(key, suffix)
		}
	}

	return key == pattern
}
