package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"axiomod/internal/examples/example/entity"
	"axiomod/internal/platform/observability"

	"go.uber.org/zap"
)

// ExampleCache provides caching functionality for Example entities
type ExampleCache struct {
	cache  Cache
	logger *observability.Logger
	ttl    time.Duration
}

// Cache defines the interface for a cache implementation
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// NewExampleCache creates a new ExampleCache
func NewExampleCache(cache Cache, logger *observability.Logger) *ExampleCache {
	return &ExampleCache{
		cache:  cache,
		logger: logger,
		ttl:    time.Hour, // Default TTL
	}
}

// SetTTL sets the time-to-live for cached items
func (c *ExampleCache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// Get retrieves an Example entity from the cache
func (c *ExampleCache) Get(ctx context.Context, id string) (*entity.Example, error) {
	key := c.buildKey(id)

	data, err := c.cache.Get(ctx, key)
	if err != nil {
		c.logger.Debug("Cache miss", zap.String("key", key), zap.Error(err))
		return nil, err
	}

	var example entity.Example
	if err := json.Unmarshal(data, &example); err != nil {
		c.logger.Error("Failed to unmarshal cached example", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal cached example: %w", err)
	}

	c.logger.Debug("Cache hit", zap.String("key", key))
	return &example, nil
}

// Set stores an Example entity in the cache
func (c *ExampleCache) Set(ctx context.Context, example *entity.Example) error {
	key := c.buildKey(example.ID)

	data, err := json.Marshal(example)
	if err != nil {
		c.logger.Error("Failed to marshal example for cache", zap.Error(err))
		return fmt.Errorf("failed to marshal example for cache: %w", err)
	}

	if err := c.cache.Set(ctx, key, data, c.ttl); err != nil {
		c.logger.Error("Failed to set cache", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to set cache: %w", err)
	}

	c.logger.Debug("Cache set", zap.String("key", key))
	return nil
}

// Delete removes an Example entity from the cache
func (c *ExampleCache) Delete(ctx context.Context, id string) error {
	key := c.buildKey(id)

	if err := c.cache.Delete(ctx, key); err != nil {
		c.logger.Error("Failed to delete from cache", zap.String("key", key), zap.Error(err))
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	c.logger.Debug("Cache delete", zap.String("key", key))
	return nil
}

// buildKey builds a cache key for an Example entity
func (c *ExampleCache) buildKey(id string) string {
	return fmt.Sprintf("example:%s", id)
}
