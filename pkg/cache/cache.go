// Package cache provides caching functionality for the C8S web dashboard.
package cache

import (
	"context"
	"regexp"
	"sync"
	"time"
)

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CacheLayer provides thread-safe caching with TTL support
type CacheLayer struct {
	mu              sync.RWMutex
	entries         map[string]*CacheEntry
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewCacheLayer creates a new cache with default TTL and cleanup interval
func NewCacheLayer(defaultTTL, cleanupInterval time.Duration) *CacheLayer {
	return &CacheLayer{
		entries:         make(map[string]*CacheEntry),
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		stopCleanup:     make(chan struct{}),
	}
}

// Get retrieves a value from cache, returning false if not found or expired
func (c *CacheLayer) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in cache with default TTL
func (c *CacheLayer) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (c *CacheLayer) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
	}
}

// Invalidate removes a specific key from cache
func (c *CacheLayer) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// InvalidatePattern removes all keys matching a regex pattern
func (c *CacheLayer) InvalidatePattern(pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	for key := range c.entries {
		if re.MatchString(key) {
			delete(c.entries, key)
		}
	}

	return nil
}

// InvalidateAll clears the entire cache
func (c *CacheLayer) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

// StartCleanup runs a background goroutine that periodically removes expired entries
func (c *CacheLayer) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.stopCleanup:
				return
			case <-ticker.C:
				c.cleanup()
			}
		}
	}()
}

// cleanup removes expired entries from cache
func (c *CacheLayer) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// StopCleanup stops the background cleanup goroutine
func (c *CacheLayer) StopCleanup() {
	select {
	case c.stopCleanup <- struct{}{}:
	default:
	}
}

// GetStats returns cache statistics for monitoring
func (c *CacheLayer) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	for _, entry := range c.entries {
		if time.Now().After(entry.ExpiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_entries":   len(c.entries),
		"expired_entries": expired,
		"active_entries":  len(c.entries) - expired,
	}
}

// CacheKeyBuilder provides helper methods for constructing cache keys
type CacheKeyBuilder struct{}

// PipelineListKey returns cache key for pipeline list
func (b *CacheKeyBuilder) PipelineListKey(projectID string) string {
	return "pipeline:list:" + projectID
}

// PipelineRunKey returns cache key for pipeline run
func (b *CacheKeyBuilder) PipelineRunKey(runID string) string {
	return "pipeline:run:" + runID
}

// ProjectListKey returns cache key for project list
func (b *CacheKeyBuilder) ProjectListKey(userID string) string {
	return "project:list:" + userID
}

// ProjectKey returns cache key for project metadata
func (b *CacheKeyBuilder) ProjectKey(projectID string) string {
	return "project:" + projectID
}

// LogSnapshotKey returns cache key for log snapshot
func (b *CacheKeyBuilder) LogSnapshotKey(runID, stepID string) string {
	return "log:" + runID + ":" + stepID
}

// UserPermissionsKey returns cache key for user permissions
func (b *CacheKeyBuilder) UserPermissionsKey(userID string) string {
	return "user:perms:" + userID
}
