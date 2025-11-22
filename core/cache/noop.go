package cache

import (
	"context"
	"time"
)

// NoopCache is a cache implementation that does nothing (for when caching is disabled)
type NoopCache struct{}

// NewNoopCache creates a new no-op cache
func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (n *NoopCache) Get(ctx context.Context, key string, dest interface{}) error {
	return ErrCacheMiss
}

func (n *NoopCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (n *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (n *NoopCache) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (n *NoopCache) Has(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (n *NoopCache) Clear(ctx context.Context) error {
	return nil
}

func (n *NoopCache) Close() error {
	return nil
}

func (n *NoopCache) GetStats() Stats {
	return Stats{Provider: "noop"}
}
