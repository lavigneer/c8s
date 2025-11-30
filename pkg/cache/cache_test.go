package cache

import (
	"context"
	"testing"
	"time"
)

// TestCacheSet tests setting and getting values from cache
func TestCacheSet(t *testing.T) {
	cache := NewCacheLayer(1*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	// Set a value
	cache.Set("test_key", "test_value")

	// Get the value
	value, found := cache.Get("test_key")
	if !found {
		t.Fatal("Expected to find cached value")
	}

	if value != "test_value" {
		t.Fatalf("Expected 'test_value', got %v", value)
	}
}

// TestCacheSetWithTTL tests setting values with custom TTL
func TestCacheSetWithTTL(t *testing.T) {
	cache := NewCacheLayer(10*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	// Set with short TTL
	cache.SetWithTTL("short_ttl", "value", 100*time.Millisecond)

	// Should exist immediately
	value, found := cache.Get("short_ttl")
	if !found {
		t.Fatal("Expected to find cached value immediately")
	}
	if value != "value" {
		t.Fatalf("Expected 'value', got %v", value)
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("short_ttl")
	if found {
		t.Fatal("Expected value to expire")
	}
}

// TestCacheInvalidate tests invalidating a specific key
func TestCacheInvalidate(t *testing.T) {
	cache := NewCacheLayer(10*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Invalidate key1
	cache.Invalidate("key1")

	// key1 should not exist
	_, found := cache.Get("key1")
	if found {
		t.Fatal("Expected key1 to be invalidated")
	}

	// key2 should still exist
	_, found = cache.Get("key2")
	if !found {
		t.Fatal("Expected key2 to still exist")
	}
}

// TestCacheInvalidatePattern tests pattern-based invalidation
func TestCacheInvalidatePattern(t *testing.T) {
	cache := NewCacheLayer(10*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	// Set multiple keys
	cache.Set("pipeline:list:proj1", "value1")
	cache.Set("pipeline:list:proj2", "value2")
	cache.Set("project:proj1", "value3")

	// Invalidate all pipeline:list:* keys
	err := cache.InvalidatePattern("pipeline:list:.*")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Both pipeline:list keys should be gone
	_, found := cache.Get("pipeline:list:proj1")
	if found {
		t.Fatal("Expected pipeline:list:proj1 to be invalidated")
	}

	_, found = cache.Get("pipeline:list:proj2")
	if found {
		t.Fatal("Expected pipeline:list:proj2 to be invalidated")
	}

	// project:proj1 should still exist
	_, found = cache.Get("project:proj1")
	if !found {
		t.Fatal("Expected project:proj1 to still exist")
	}
}

// TestCacheInvalidateAll tests clearing entire cache
func TestCacheInvalidateAll(t *testing.T) {
	cache := NewCacheLayer(10*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// Clear all
	cache.InvalidateAll()

	// All keys should be gone
	_, found := cache.Get("key1")
	if found {
		t.Fatal("Expected key1 to be cleared")
	}
	_, found = cache.Get("key2")
	if found {
		t.Fatal("Expected key2 to be cleared")
	}
	_, found = cache.Get("key3")
	if found {
		t.Fatal("Expected key3 to be cleared")
	}
}

// TestCacheCleanup tests automatic cleanup of expired entries
func TestCacheCleanup(t *testing.T) {
	cache := NewCacheLayer(1*time.Second, 100*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start cleanup routine
	cache.StartCleanup(ctx)

	// Set value that will expire
	cache.SetWithTTL("expire_me", "value", 50*time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(200 * time.Millisecond)

	// Value should be removed by cleanup
	_, found := cache.Get("expire_me")
	if found {
		t.Fatal("Expected expired value to be cleaned up")
	}

	cache.StopCleanup()
}

// TestCacheGetStats tests cache statistics
func TestCacheGetStats(t *testing.T) {
	cache := NewCacheLayer(10*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	// Set some values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.SetWithTTL("expire_soon", "value", 50*time.Millisecond)

	// Get stats
	stats := cache.GetStats()

	if stats["total_entries"] != 3 {
		t.Fatalf("Expected 3 total entries, got %v", stats["total_entries"])
	}

	// Wait for one to expire
	time.Sleep(100 * time.Millisecond)

	// Get updated stats
	stats = cache.GetStats()
	if stats["expired_entries"] != 1 {
		t.Fatalf("Expected 1 expired entry, got %v", stats["expired_entries"])
	}
}

// TestCacheKeyBuilder tests cache key construction
func TestCacheKeyBuilder(t *testing.T) {
	builder := &KeyBuilder{}

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "PipelineListKey",
			fn:       func() string { return builder.PipelineListKey("proj123") },
			expected: "pipeline:list:proj123",
		},
		{
			name:     "PipelineRunKey",
			fn:       func() string { return builder.PipelineRunKey("run123") },
			expected: "pipeline:run:run123",
		},
		{
			name:     "ProjectListKey",
			fn:       func() string { return builder.ProjectListKey("user456") },
			expected: "project:list:user456",
		},
		{
			name:     "ProjectKey",
			fn:       func() string { return builder.ProjectKey("proj789") },
			expected: "project:proj789",
		},
		{
			name:     "LogSnapshotKey",
			fn:       func() string { return builder.LogSnapshotKey("run123", "step456") },
			expected: "log:run123:step456",
		},
		{
			name:     "UserPermissionsKey",
			fn:       func() string { return builder.UserPermissionsKey("user123") },
			expected: "user:perms:user123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.expected {
				t.Fatalf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

// TestCacheThreadSafety tests concurrent access to cache
func TestCacheThreadSafety(t *testing.T) {
	cache := NewCacheLayer(10*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("concurrent_key", i)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			cache.Get("concurrent_key")
		}
		done <- true
	}()

	// Invalidator goroutine
	go func() {
		for i := 0; i < 10; i++ {
			cache.Invalidate("concurrent_key")
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	<-done
	<-done
	<-done

	// Cache should still be functional
	cache.Set("final_test", "success")
	value, found := cache.Get("final_test")
	if !found || value != "success" {
		t.Fatal("Cache is not functional after concurrent access")
	}
}

// TestCacheExpiration tests that expired values are not returned
func TestCacheExpiration(t *testing.T) {
	cache := NewCacheLayer(1*time.Second, 1*time.Minute)
	defer cache.StopCleanup()

	cache.SetWithTTL("will_expire", "value", 50*time.Millisecond)

	// Should exist immediately
	_, found := cache.Get("will_expire")
	if !found {
		t.Fatal("Expected value to exist immediately")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should not exist after expiration
	_, found = cache.Get("will_expire")
	if found {
		t.Fatal("Expected value to expire")
	}
}
