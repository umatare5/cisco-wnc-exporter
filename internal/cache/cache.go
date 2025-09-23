// Package cache provides generic caching functionality with TTL support.
package cache

import (
	"log/slog"
	"sync"
	"time"
)

// Cache provides generic caching with TTL support and double-checked locking pattern.
// T represents the type of data being cached.
type Cache[T any] struct {
	mu       sync.RWMutex
	data     T
	cachedAt time.Time
	ttl      time.Duration
	name     string // for logging
}

// New creates a new cache with the specified TTL and name for logging.
func New[T any](ttl time.Duration, name string) *Cache[T] {
	return &Cache[T]{
		ttl:  ttl,
		name: name,
	}
}

// Get retrieves cached data if fresh, otherwise calls refreshFunc to update the cache.
// Uses double-checked locking pattern to avoid race conditions.
func (c *Cache[T]) Get(refreshFunc func() (T, error)) (T, error) {
	// Fast path: check cache with read lock
	c.mu.RLock()
	if c.isFresh() {
		data := c.data
		c.mu.RUnlock()
		return data, nil
	}
	c.mu.RUnlock()

	// Slow path: refresh cache with write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check pattern to avoid race conditions
	if c.isFresh() {
		return c.data, nil
	}

	slog.Info("refreshing cache", "cache_name", c.name)
	start := time.Now()

	data, err := refreshFunc()
	if err != nil {
		var zero T
		return zero, err
	}

	c.data = data
	c.cachedAt = time.Now()

	slog.Info("cache refreshed", "cache_name", c.name, "duration_ms", time.Since(start).Milliseconds())
	return data, nil
}

// isFresh checks if the cached data is still within TTL.
// Must be called with at least read lock held.
func (c *Cache[T]) isFresh() bool {
	return !c.cachedAt.IsZero() && time.Since(c.cachedAt) < c.ttl
}

// GetCached returns cached data without refreshing if available and fresh.
// Returns zero value and false if cache is stale or empty.
func (c *Cache[T]) GetCached() (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.isFresh() {
		return c.data, true
	}

	var zero T
	return zero, false
}

// Invalidate clears the cache, forcing the next Get() call to refresh.
func (c *Cache[T]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cachedAt = time.Time{}
	slog.Debug("cache invalidated", "cache_name", c.name)
}

// SetTTL updates the cache TTL. Useful for dynamic TTL adjustment.
func (c *Cache[T]) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
	slog.Debug("cache TTL updated", "cache_name", c.name, "ttl_seconds", ttl.Seconds())
}
