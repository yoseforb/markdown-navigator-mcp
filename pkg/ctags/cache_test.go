package ctags

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestMarkdownFile creates a temporary markdown file for testing.
func createTestMarkdownFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	err := os.WriteFile(mdFile, []byte(content), 0o644)
	require.NoError(t, err)

	return mdFile
}

// modifyMarkdownFile modifies an existing markdown file.
func modifyMarkdownFile(t *testing.T, filePath string, newContent string) {
	t.Helper()

	// Sleep briefly to ensure mtime changes
	time.Sleep(10 * time.Millisecond)

	err := os.WriteFile(filePath, []byte(newContent), 0o644)
	require.NoError(t, err)
}

func TestCacheManager_Hit(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "# Test\n## Section\n")

	// First access - cache miss
	tags1, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	require.NotEmpty(t, tags1)

	// Second access - cache hit
	tags2, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	require.NotEmpty(t, tags2)

	// Verify same tags returned
	assert.Len(t, tags2, len(tags1))
	assert.Equal(t, tags1[0].Name, tags2[0].Name)

	// Verify statistics
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(1), hits, "Should have 1 cache hit")
	assert.Equal(t, uint64(1), misses, "Should have 1 cache miss")
}

func TestCacheManager_Miss(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "# Initial\n")

	// First access - cache miss
	tags1, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	assert.Len(t, tags1, 1)
	assert.Equal(t, "Initial", tags1[0].Name)

	// Verify statistics after first access
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(0), hits)
	assert.Equal(t, uint64(1), misses)
}

func TestCacheManager_Invalidation(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "# Original\n")

	// First access - cache miss
	tags1, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	require.Len(t, tags1, 1)
	assert.Equal(t, "Original", tags1[0].Name)

	// Modify file (change mtime)
	modifyMarkdownFile(t, file, "# Modified\n")

	// Second access - cache miss (invalidated by mtime)
	tags2, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	require.Len(t, tags2, 1)
	assert.Equal(t, "Modified", tags2[0].Name)

	// Verify statistics: 2 misses, 0 hits
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(0), hits)
	assert.Equal(t, uint64(2), misses)
}

func TestCacheManager_ManualInvalidation(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "# Test\n")

	// First access - populate cache
	_, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)

	// Verify cache has entry
	assert.Equal(t, 1, cache.Size())

	// Manually invalidate
	cache.InvalidateFile(file)

	// Verify cache is empty
	assert.Equal(t, 0, cache.Size())

	// Next access should be cache miss
	_, err = cache.GetTags(context.Background(), file)
	require.NoError(t, err)

	// Verify statistics: 1 miss (initial), 1 miss (after invalidation)
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(0), hits)
	assert.Equal(t, uint64(2), misses)
}

func TestCacheManager_Clear(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file1 := createTestMarkdownFile(t, "# File1\n")
	file2 := createTestMarkdownFile(t, "# File2\n")

	// Populate cache with multiple files
	_, err := cache.GetTags(context.Background(), file1)
	require.NoError(t, err)
	_, err = cache.GetTags(context.Background(), file2)
	require.NoError(t, err)

	// Verify cache has entries
	assert.Equal(t, 2, cache.Size())

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	assert.Equal(t, 0, cache.Size())
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "# Concurrent Test\n## Section\n")

	const numGoroutines = 100
	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Launch multiple goroutines accessing the same file
	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := cache.GetTags(context.Background(), file)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Verify no errors occurred
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify statistics: should have many hits, 1 miss
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(99), hits, "Should have 99 cache hits")
	assert.Equal(t, uint64(1), misses, "Should have 1 cache miss")
}

func TestCacheManager_ConcurrentDifferentFiles(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()

	// Create multiple files
	files := make([]string, 10)
	for i := range 10 {
		files[i] = createTestMarkdownFile(t, "# File\n")
	}

	const accessesPerFile = 10
	var wg sync.WaitGroup
	errors := make(chan error, len(files)*accessesPerFile)

	// Access each file multiple times concurrently
	for _, file := range files {
		for range accessesPerFile {
			wg.Add(1)
			go func(f string) {
				defer wg.Done()
				_, err := cache.GetTags(context.Background(), f)
				if err != nil {
					errors <- err
				}
			}(file)
		}
	}

	wg.Wait()
	close(errors)

	// Verify no errors
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}

	// Verify cache has all files
	assert.Equal(t, 10, cache.Size())

	// Verify statistics
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(90), hits, "Should have 90 cache hits (9 per file)")
	assert.Equal(
		t,
		uint64(10),
		misses,
		"Should have 10 cache misses (1 per file)",
	)
}

func TestCacheManager_FileNotFound(t *testing.T) {
	cache := NewCacheManager()

	_, err := cache.GetTags(context.Background(), "/nonexistent/file.md")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrFileNotFound)

	// Should not increment miss counter for errors
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(0), hits)
	assert.Equal(t, uint64(0), misses)
}

func TestCacheManager_GlobalCache(t *testing.T) {
	// Get global cache instance
	cache1 := GetGlobalCache()
	cache2 := GetGlobalCache()

	// Verify it's the same instance
	assert.Same(t, cache1, cache2, "Global cache should be singleton")
}

func TestCacheManager_EmptyFile(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "")

	// Access empty file
	tags, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	assert.Empty(t, tags, "Empty file should have no tags")

	// Verify cache was populated
	assert.Equal(t, 1, cache.Size())

	// Second access should hit cache
	tags2, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	assert.Empty(t, tags2)

	// Verify statistics
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(1), hits)
	assert.Equal(t, uint64(1), misses)
}

func TestCacheManager_MultipleModifications(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	cache := NewCacheManager()
	file := createTestMarkdownFile(t, "# Version 1\n")

	// Access 1
	tags1, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	assert.Equal(t, "Version 1", tags1[0].Name)

	// Modify 1
	modifyMarkdownFile(t, file, "# Version 2\n")
	tags2, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	assert.Equal(t, "Version 2", tags2[0].Name)

	// Modify 2
	modifyMarkdownFile(t, file, "# Version 3\n")
	tags3, err := cache.GetTags(context.Background(), file)
	require.NoError(t, err)
	assert.Equal(t, "Version 3", tags3[0].Name)

	// Verify statistics: 3 misses (each modification), 0 hits
	hits, misses := cache.Stats()
	assert.Equal(t, uint64(0), hits)
	assert.Equal(t, uint64(3), misses)
}

// Benchmark tests.
func BenchmarkCache_Hit(b *testing.B) {
	if !IsCtagsInstalled() {
		b.Skip("ctags not installed, skipping benchmark")
	}

	cache := NewCacheManager()
	tmpDir := b.TempDir()
	file := filepath.Join(tmpDir, "bench.md")

	content := "# Benchmark Test\n## Section\n### Subsection\n"
	err := os.WriteFile(file, []byte(content), 0o644)
	require.NoError(b, err)

	// Warm cache
	_, err = cache.GetTags(context.Background(), file)
	require.NoError(b, err)

	b.ResetTimer()
	for range b.N {
		_, _ = cache.GetTags(context.Background(), file)
	}
}

func BenchmarkCache_Miss(b *testing.B) {
	if !IsCtagsInstalled() {
		b.Skip("ctags not installed, skipping benchmark")
	}

	cache := NewCacheManager()
	tmpDir := b.TempDir()

	content := "# Benchmark Test\n## Section\n"

	b.ResetTimer()
	for i := range b.N {
		b.StopTimer()
		file := filepath.Join(tmpDir, fmt.Sprintf("bench-%d.md", i))
		err := os.WriteFile(file, []byte(content), 0o644)
		require.NoError(b, err)
		b.StartTimer()

		_, _ = cache.GetTags(context.Background(), file)
	}
}
