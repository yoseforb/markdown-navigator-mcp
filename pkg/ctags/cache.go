package ctags

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// CacheEntry represents a cached set of tags for a file.
// It stores the file path, modification time, and parsed tags.
type CacheEntry struct {
	FilePath string
	ModTime  time.Time
	Tags     []*TagEntry
}

// CacheManager manages in-memory caching of ctags output with mtime-based invalidation.
// It provides concurrent-safe access to cached tags with automatic invalidation
// when files change. The cache uses per-file mutexes to prevent duplicate
// ctags executions for the same file when multiple goroutines request it simultaneously.
type CacheManager struct {
	cache      map[string]*CacheEntry // Cached entries by file path
	mu         sync.RWMutex           // Protects cache map
	hits       atomic.Uint64          // Cache hit counter
	misses     atomic.Uint64          // Cache miss counter
	inProgress map[string]*sync.Mutex // Track in-progress operations per file
	progressMu sync.Mutex             // Protects inProgress map
}

// NewCacheManager creates a new cache manager.
func NewCacheManager() *CacheManager {
	return &CacheManager{
		cache:      make(map[string]*CacheEntry),
		mu:         sync.RWMutex{},
		hits:       atomic.Uint64{},
		misses:     atomic.Uint64{},
		inProgress: make(map[string]*sync.Mutex),
		progressMu: sync.Mutex{},
	}
}

// globalCache is the singleton cache instance used throughout the application.
// This pattern is acceptable for caches as it provides a single point of
// coordination for cache operations and avoids passing cache instances through
// multiple layers of the application.
var globalCache = NewCacheManager() //nolint:gochecknoglobals // singleton cache pattern

// GetGlobalCache returns the global cache instance.
func GetGlobalCache() *CacheManager {
	return globalCache
}

// GetTags retrieves tags for a file, using cache if available and valid.
// Cache validation is based on file modification time (mtime).
// Concurrent requests for the same file are serialized to prevent duplicate work.
func (cm *CacheManager) GetTags(filePath string) ([]*TagEntry, error) {
	// Get file modification time
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, filePath)
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	currentMtime := stat.ModTime()

	// Check cache with read lock
	cm.mu.RLock()
	entry, exists := cm.cache[filePath]
	cm.mu.RUnlock()

	// Cache hit: return cached data if mtime matches
	if exists && entry.ModTime.Equal(currentMtime) {
		cm.hits.Add(1)
		return entry.Tags, nil
	}

	// Cache miss: need to execute ctags
	// Use per-file mutex to prevent multiple concurrent executions for same file
	cm.progressMu.Lock()
	fileMutex, exists := cm.inProgress[filePath]
	if !exists {
		fileMutex = &sync.Mutex{}
		cm.inProgress[filePath] = fileMutex
	}
	cm.progressMu.Unlock()

	// Lock for this specific file
	fileMutex.Lock()
	defer func() {
		fileMutex.Unlock()
		// Clean up the in-progress entry
		cm.progressMu.Lock()
		delete(cm.inProgress, filePath)
		cm.progressMu.Unlock()
	}()

	// Check cache again in case another goroutine just populated it
	cm.mu.RLock()
	entry, exists = cm.cache[filePath]
	cm.mu.RUnlock()

	if exists && entry.ModTime.Equal(currentMtime) {
		cm.hits.Add(1)
		return entry.Tags, nil
	}

	// Execute ctags (only one goroutine reaches here per file)
	cm.misses.Add(1)

	jsonData, err := ExecuteCtags(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ctags: %w", err)
	}

	// Parse JSON output
	tags, err := ParseJSONTags(jsonData, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ctags JSON: %w", err)
	}

	// Sort tags by line number to ensure document order
	SortByLine(tags)

	// Update cache with write lock
	cm.mu.Lock()
	cm.cache[filePath] = &CacheEntry{
		FilePath: filePath,
		ModTime:  currentMtime,
		Tags:     tags,
	}
	cm.mu.Unlock()

	return tags, nil
}

// InvalidateFile removes a specific file from the cache.
// This is useful for manually clearing cache when file changes are detected
// through external means, though the cache automatically invalidates based on mtime.
func (cm *CacheManager) InvalidateFile(filePath string) {
	cm.mu.Lock()
	delete(cm.cache, filePath)
	cm.mu.Unlock()
}

// Clear removes all entries from the cache.
func (cm *CacheManager) Clear() {
	cm.mu.Lock()
	cm.cache = make(map[string]*CacheEntry)
	cm.mu.Unlock()
}

// Stats returns cache hit and miss statistics.
// Useful for monitoring cache effectiveness and performance tuning.
func (cm *CacheManager) Stats() (hits, misses uint64) {
	return cm.hits.Load(), cm.misses.Load()
}

// Size returns the number of entries currently cached.
// Useful for monitoring memory usage and cache capacity.
func (cm *CacheManager) Size() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.cache)
}
