package cache

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements a Redis-backed cache
type RedisCache struct {
	client *redis.Client
	prefix string

	// Stats
	hits    int64
	misses  int64
	sets    int64
	deletes int64
}

// RedisConfig holds configuration for Redis cache
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Prefix   string
}

// NewRedisCache creates a new Redis-backed cache
func NewRedisCache(config RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, &CacheError{Message: "redis connection failed", Type: "connection", Err: err}
	}

	prefix := config.Prefix
	if prefix == "" {
		prefix = "base:"
	}

	return &RedisCache{
		client: client,
		prefix: prefix,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := r.prefix + key

	val, err := r.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		atomic.AddInt64(&r.misses, 1)
		return ErrCacheMiss
	}
	if err != nil {
		return &CacheError{Message: "redis get failed", Type: "get", Err: err}
	}

	// Unmarshal value
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return &CacheError{Message: "unmarshal failed", Type: "unmarshal", Err: err}
	}

	atomic.AddInt64(&r.hits, 1)
	return nil
}

// Set stores a value in Redis
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.prefix + key

	// Marshal value
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Message: "marshal failed", Type: "marshal", Err: err}
	}

	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return &CacheError{Message: "redis set failed", Type: "set", Err: err}
	}

	atomic.AddInt64(&r.sets, 1)
	return nil
}

// Delete removes a key from Redis
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.prefix + key

	deleted, err := r.client.Del(ctx, fullKey).Result()
	if err != nil {
		return &CacheError{Message: "redis delete failed", Type: "delete", Err: err}
	}

	if deleted > 0 {
		atomic.AddInt64(&r.deletes, deleted)
	}

	return nil
}

// DeletePattern removes all keys matching a pattern
func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := r.prefix + pattern

	// Use SCAN to find matching keys
	var cursor uint64
	var deletedCount int64

	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return &CacheError{Message: "redis scan failed", Type: "scan", Err: err}
		}

		if len(keys) > 0 {
			deleted, err := r.client.Del(ctx, keys...).Result()
			if err != nil {
				return &CacheError{Message: "redis delete failed", Type: "delete", Err: err}
			}
			deletedCount += deleted
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	atomic.AddInt64(&r.deletes, deletedCount)
	return nil
}

// Has checks if a key exists in Redis
func (r *RedisCache) Has(ctx context.Context, key string) (bool, error) {
	fullKey := r.prefix + key

	exists, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, &CacheError{Message: "redis exists failed", Type: "exists", Err: err}
	}

	return exists > 0, nil
}

// Clear removes all keys with the prefix
func (r *RedisCache) Clear(ctx context.Context) error {
	return r.DeletePattern(ctx, "*")
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GetStats returns cache statistics
func (r *RedisCache) GetStats() Stats {
	// Get Redis info for size
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var size int64
	if dbSize, err := r.client.DBSize(ctx).Result(); err == nil {
		size = dbSize
	}

	return Stats{
		Hits:     atomic.LoadInt64(&r.hits),
		Misses:   atomic.LoadInt64(&r.misses),
		Sets:     atomic.LoadInt64(&r.sets),
		Deletes:  atomic.LoadInt64(&r.deletes),
		Size:     size,
		Provider: "redis",
	}
}
