package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Common errors
var (
	ErrKeyNotFound = errors.New("key not found in cache")
	ErrCacheFull   = errors.New("cache is full")
)

// Cache defines the interface for cache implementations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// MemoryCache implements an in-memory cache
type MemoryCache struct {
	items     map[string]cacheItem
	maxItems  int
	mu        sync.RWMutex
	janitorOn bool
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxItems int) *MemoryCache {
	cache := &MemoryCache{
		items:    make(map[string]cacheItem),
		maxItems: maxItems,
	}

	// Start the janitor if maxItems > 0
	if maxItems > 0 {
		go cache.janitor()
		cache.janitorOn = true
	}

	return cache
}

// Get retrieves a value from the cache
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, ErrKeyNotFound
	}

	// Check if the item has expired
	if !item.expiration.IsZero() && item.expiration.Before(time.Now()) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, ErrKeyNotFound
	}

	// Return a copy of the value to prevent modification
	value := make([]byte, len(item.value))
	copy(value, item.value)
	return value, nil
}

// Set stores a value in the cache
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the cache is full
	if c.maxItems > 0 && len(c.items) >= c.maxItems && c.items[key].value == nil {
		return ErrCacheFull
	}

	// Calculate expiration time
	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	// Store a copy of the value to prevent modification
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	c.items[key] = cacheItem{
		value:      valueCopy,
		expiration: expiration,
	}

	return nil
}

// Delete removes a value from the cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Clear removes all values from the cache
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cacheItem)
	return nil
}

// janitor periodically removes expired items from the cache
func (c *MemoryCache) janitor() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if !item.expiration.IsZero() && item.expiration.Before(now) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
