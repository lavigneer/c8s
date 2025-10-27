package handlers

import (
	"log"

	"github.com/org/c8s/pkg/dashboard"
)

// Global cache instance - initialized at startup
var globalCache *dashboard.CacheLayer

// InitCache initializes the global cache layer
func InitCache(cache *dashboard.CacheLayer) {
	globalCache = cache
}

// GetCache returns the global cache instance
func GetCache() *dashboard.CacheLayer {
	return globalCache
}

// InvalidatePipelineListCache invalidates the pipeline list cache for a project
func InvalidatePipelineListCache(projectID string) {
	if globalCache == nil {
		return
	}
	keyBuilder := &dashboard.CacheKeyBuilder{}
	key := keyBuilder.PipelineListKey(projectID)
	globalCache.Invalidate(key)
	log.Printf("Invalidated cache: %s", key)
}

// InvalidatePipelineRunCache invalidates a specific pipeline run cache
func InvalidatePipelineRunCache(runID string) {
	if globalCache == nil {
		return
	}
	keyBuilder := &dashboard.CacheKeyBuilder{}
	key := keyBuilder.PipelineRunKey(runID)
	globalCache.Invalidate(key)
	log.Printf("Invalidated cache: %s", key)
}

// InvalidateProjectListCache invalidates the project list cache for a user
func InvalidateProjectListCache(userID string) {
	if globalCache == nil {
		return
	}
	keyBuilder := &dashboard.CacheKeyBuilder{}
	key := keyBuilder.ProjectListKey(userID)
	globalCache.Invalidate(key)
	log.Printf("Invalidated cache: %s", key)
}

// InvalidateProjectCache invalidates a specific project cache
func InvalidateProjectCache(projectID string) {
	if globalCache == nil {
		return
	}
	keyBuilder := &dashboard.CacheKeyBuilder{}
	key := keyBuilder.ProjectKey(projectID)
	globalCache.Invalidate(key)
	log.Printf("Invalidated cache: %s", key)
}

// InvalidateLogCache invalidates log snapshot cache
func InvalidateLogCache(runID, stepID string) {
	if globalCache == nil {
		return
	}
	keyBuilder := &dashboard.CacheKeyBuilder{}
	key := keyBuilder.LogSnapshotKey(runID, stepID)
	globalCache.Invalidate(key)
	log.Printf("Invalidated cache: %s", key)
}

// InvalidateUserPermissionsCache invalidates user permissions cache
func InvalidateUserPermissionsCache(userID string) {
	if globalCache == nil {
		return
	}
	keyBuilder := &dashboard.CacheKeyBuilder{}
	key := keyBuilder.UserPermissionsKey(userID)
	globalCache.Invalidate(key)
	log.Printf("Invalidated cache: %s", key)
}

// BroadcastCacheInvalidation broadcasts cache invalidation event via SSE
// This is called when cache needs to be cleared on all connected clients
func BroadcastCacheInvalidation(projectID string, cachePattern string) {
	broadcaster := getOrCreateBroadcaster(projectID)

	// Create SSE event for cache invalidation
	event := dashboard.NewEventBuilder().
		WithEvent("cache_invalidated").
		WithData("{\"pattern\":\"" + cachePattern + "\"}").
		Build()

	broadcaster.BroadcastAsync(event)
	log.Printf("Broadcast cache invalidation event for pattern: %s", cachePattern)
}

// OnPipelineStatusChanged handles pipeline status change events
// Called when a pipeline run status changes - invalidates relevant caches
func OnPipelineStatusChanged(projectID, runID string) {
	InvalidatePipelineListCache(projectID)
	InvalidatePipelineRunCache(runID)

	// Broadcast invalidation to connected clients
	BroadcastCacheInvalidation(projectID, "pipeline:*")
}

// OnProjectCreated handles project creation events
// Called when a new project is created - invalidates project list cache
func OnProjectCreated(userID, projectID string) {
	InvalidateProjectListCache(userID)
	InvalidateProjectCache(projectID)
}

// OnProjectUpdated handles project update events
// Called when project is updated - invalidates project cache
func OnProjectUpdated(projectID string) {
	InvalidateProjectCache(projectID)
}

// OnProjectDeleted handles project deletion events
// Called when project is deleted - invalidates caches and closes broadcaster
func OnProjectDeleted(projectID, userID string) {
	InvalidateProjectCache(projectID)
	InvalidateProjectListCache(userID)
	CloseBroadcaster(projectID)
}

// OnLogUpdated handles log update events
// Called when new logs are streamed - invalidates log cache
func OnLogUpdated(runID, stepID string) {
	InvalidateLogCache(runID, stepID)
}

// ClearAllCaches clears all caches
func ClearAllCaches() {
	if globalCache == nil {
		return
	}
	globalCache.InvalidateAll()
	log.Printf("Cleared all caches")
}

// GetCacheStats returns cache statistics for monitoring
func GetCacheStats() map[string]interface{} {
	if globalCache == nil {
		return map[string]interface{}{"error": "cache not initialized"}
	}
	return globalCache.GetStats()
}
