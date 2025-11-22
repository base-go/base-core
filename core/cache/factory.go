package cache

import (
	"fmt"
	"time"
)

// Provider represents the cache provider type
type Provider string

const (
	ProviderMemory Provider = "memory"
	ProviderRedis  Provider = "redis"
	ProviderNoop   Provider = "none"
)

// Config holds unified cache configuration
type Config struct {
	Enabled         bool
	Provider        Provider
	DefaultTTL      time.Duration
	Memory          MemoryConfig
	Redis           RedisConfig
	FallbackEnabled bool // If true, falls back to memory if Redis fails
}

// New creates a new cache instance based on configuration
func New(config Config) (Cache, error) {
	if !config.Enabled {
		return NewNoopCache(), nil
	}

	switch config.Provider {
	case ProviderMemory:
		return NewMemoryCache(config.Memory), nil

	case ProviderRedis:
		cache, err := NewRedisCache(config.Redis)
		if err != nil {
			if config.FallbackEnabled {
				// Fall back to memory cache
				fmt.Printf("Redis cache failed, falling back to memory: %v\n", err)
				return NewMemoryCache(config.Memory), nil
			}
			return nil, err
		}
		return cache, nil

	case ProviderNoop:
		return NewNoopCache(), nil

	default:
		return nil, fmt.Errorf("unknown cache provider: %s", config.Provider)
	}
}
