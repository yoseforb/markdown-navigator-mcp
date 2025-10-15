# JSON-Based Ctags Caching Refactoring

**Status**: Backlog
**Priority**: High
**Effort**: Medium (4-6 development sessions)
**Type**: Architecture Refactoring

## Executive Summary

Refactor the markdown-mcp server from **pre-generated tags file parsing** to **on-demand ctags execution with intelligent caching**. This change eliminates manual ctags generation, improves user experience, and provides better performance through mtime-based caching.

## Current State Analysis

### What We Have Today

1. **Manual ctags generation required**:
   - Users must run: `ctags -R --fields=+KnS --languages=markdown`
   - Tags file becomes stale when markdown files change
   - No automatic regeneration or invalidation

2. **Tab-separated parsing**:
   - Custom parser in `pkg/ctags/parser.go`
   - Manual field extraction with `strings.Split()`
   - Error-prone parsing logic

3. **No caching**:
   - Tags file re-parsed on every tool invocation
   - Redundant I/O and parsing work
   - No performance optimization

### Pain Points

**User Experience Issues**:
- Users forget to regenerate tags after editing markdown
- Stale tags lead to incorrect line numbers and missing sections
- Extra setup step reduces adoption

**Performance Issues**:
- Every tool call re-parses entire tags file
- No benefit from repeated queries on same files
- Wasted CPU cycles on parsing

**Maintenance Issues**:
- Custom tab-separated parser is complex
- Hard to extend with new ctags fields
- Test coverage requires mock tags files

## What We Want to Build

### Target Architecture

```
User Request
    ↓
MCP Tool (e.g., markdown_tree)
    ↓
CacheManager.GetTags(filePath)
    ↓
    ├─→ [Cache Hit] Return cached tags (~528ns)
    │   └─→ Mtime matches? → Use cache
    │
    └─→ [Cache Miss] Execute ctags (~13ms)
        ├─→ Run: ctags --output-format=json -f - file.md
        ├─→ Parse JSON → []TagEntry
        ├─→ Store in cache with current mtime
        └─→ Return tags
```

### Key Features

1. **Zero-Configuration User Experience**:
   - No manual ctags generation required
   - Just point tools at markdown files
   - Automatic cache management

2. **JSON-Based Parsing**:
   - Use `ctags --fields=+KnSe --output-format=json`
   - Structured data with Go unmarshaling
   - Easier to extend and maintain

3. **Intelligent Caching**:
   - In-memory cache with mtime tracking
   - Automatic invalidation on file changes
   - Concurrent-safe with `sync.RWMutex`

4. **Performance Characteristics**:
   - Cache hit: ~528ns (instant)
   - Cache miss: ~13ms (ctags + parse)
   - Mtime check: ~528ns overhead

### Benefits

**For Users**:
- Simpler setup (no manual ctags generation)
- Always fresh data (automatic invalidation)
- Faster repeated queries (caching)

**For Developers**:
- Cleaner code (JSON unmarshaling)
- Better testability (mock cache)
- Extensible architecture (easy to add features)

**For Performance**:
- 915x faster cache validation vs MD5 hashing
- Near-instant cache hits
- Minimal overhead for unchanged files

## Performance Analysis

**Complete benchmark analysis and testing methodology**:
See [`ai-docs/research/caching-performance-analysis.md`](../research/caching-performance-analysis.md)

### Bottom Line Decision: Mtime-Based Caching

**Performance Comparison**:
- **mtime checking**: 528ns (~1,900,000 ops/sec)
- MD5 hashing: 483µs (~2,070 ops/sec) - 915x slower
- ctags execution: 12.6ms (~79 ops/sec)

**Decision Rationale**:
- 915x faster than MD5 for cache validation
- Reliable for markdown editing workflows (editors update mtime)
- Simple implementation (single `stat()` syscall)
- Industry-proven (make, ninja, go build)

**Expected Performance**:
- Cache hit: ~600ns (mtime check + map lookup)
- Cache miss: ~13ms (ctags execution + JSON parsing)
- Typical usage: 90%+ cache hit rate

## Implementation Plan

### Phase 1: JSON Parser (Foundation)

**Objective**: Implement JSON ctags parsing

**Files to Create**:
- `pkg/ctags/json_parser.go` - JSON parsing logic
- `pkg/ctags/json_parser_test.go` - JSON parser tests

**Implementation Steps**:

1. **Define JSON Structs**:
```go
type CtagsJSONEntry struct {
    Type      string `json:"_type"`
    Name      string `json:"name"`
    Path      string `json:"path"`
    Pattern   string `json:"pattern"`
    Line      int    `json:"line"`
    Kind      string `json:"kind"`
    Scope     string `json:"scope"`
    ScopeKind string `json:"scopeKind"`
    End       int    `json:"end"`
}
```

2. **Implement Parser**:
```go
func ParseJSONTags(jsonData []byte, targetFile string) ([]*TagEntry, error)
```

3. **Map JSON to TagEntry**:
   - Convert `kind` to heading level (chapter=1, section=2, etc.)
   - Extract scope hierarchy
   - Handle missing/optional fields

4. **Unit Tests**:
   - Valid JSON input → parsed entries
   - Invalid JSON → error handling
   - Missing fields → default values
   - File filtering → only target file entries

**Acceptance Criteria**:
- ✅ Can parse real ctags JSON output
- ✅ Maps to existing `TagEntry` struct
- ✅ Handles all markdown heading levels
- ✅ Test coverage >90%

**Time Estimate**: 1-2 hours

---

### Phase 2: Ctags Executor

**Objective**: Implement ctags command execution

**Files to Create**:
- `pkg/ctags/executor.go` - Ctags execution logic
- `pkg/ctags/executor_test.go` - Executor tests

**Implementation Steps**:

1. **Implement Executor Function**:
```go
func ExecuteCtags(filePath string) ([]byte, error)
```

2. **Build Command**:
```go
cmd := exec.Command(
    "ctags",
    "--output-format=json",
    "--fields=+KnSe",
    "--languages=markdown",
    "-f", "-", // Output to stdout
    filePath,
)
```

3. **Error Handling**:
   - Check if ctags is installed (`exec.LookPath`)
   - Validate file exists before execution
   - Set execution timeout (5 seconds)
   - Capture and return descriptive errors

4. **Timeout Protection**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
cmd := exec.CommandContext(ctx, "ctags", ...)
```

5. **Unit Tests**:
   - Valid markdown file → JSON output
   - File not found → descriptive error
   - Ctags not installed → clear error message
   - Timeout scenario → context deadline error

**Acceptance Criteria**:
- ✅ Executes ctags with JSON output
- ✅ Returns raw JSON for parsing
- ✅ Handles ctags not installed gracefully
- ✅ Timeout protection works
- ✅ Test coverage >80%

**Time Estimate**: 1-2 hours

---

### Phase 3: Cache Manager (Core Logic)

**Objective**: Implement mtime-based caching with concurrent safety

**Files to Create**:
- `pkg/ctags/cache.go` - Cache manager implementation
- `pkg/ctags/cache_test.go` - Cache tests including race detection

**Implementation Steps**:

1. **Define Cache Structures**:
```go
type CacheEntry struct {
    FilePath string
    ModTime  time.Time
    Tags     []*TagEntry
}

type CacheManager struct {
    cache  map[string]*CacheEntry
    mu     sync.RWMutex
    hits   uint64 // atomic counter
    misses uint64 // atomic counter
}
```

2. **Implement GetTags Method**:
```go
func (cm *CacheManager) GetTags(filePath string) ([]*TagEntry, error) {
    // 1. Get file mtime
    stat, err := os.Stat(filePath)
    currentMtime := stat.ModTime()

    // 2. Check cache (read lock)
    cm.mu.RLock()
    entry, exists := cm.cache[filePath]
    cm.mu.RUnlock()

    // 3. Cache hit: return cached data
    if exists && entry.ModTime.Equal(currentMtime) {
        atomic.AddUint64(&cm.hits, 1)
        return entry.Tags, nil
    }

    // 4. Cache miss: execute ctags
    atomic.AddUint64(&cm.misses, 1)
    jsonData, err := ExecuteCtags(filePath)
    tags, err := ParseJSONTags(jsonData, filePath)

    // 5. Update cache (write lock)
    cm.mu.Lock()
    cm.cache[filePath] = &CacheEntry{
        FilePath: filePath,
        ModTime:  currentMtime,
        Tags:     tags,
    }
    cm.mu.Unlock()

    return tags, nil
}
```

3. **Implement Cache Operations**:
```go
func (cm *CacheManager) InvalidateFile(filePath string)
func (cm *CacheManager) Clear()
func (cm *CacheManager) Stats() (hits, misses uint64)
```

4. **Global Cache Instance**:
```go
var globalCache = NewCacheManager()

func GetGlobalCache() *CacheManager {
    return globalCache
}
```

5. **Comprehensive Tests**:

**Test: Cache Hit**:
```go
func TestCacheManager_Hit(t *testing.T) {
    cache := NewCacheManager()
    file := createTestMarkdown(t, "# Test")

    // First access - miss
    tags1, err := cache.GetTags(file)
    require.NoError(t, err)

    // Second access - hit
    tags2, err := cache.GetTags(file)
    require.NoError(t, err)
    assert.Equal(t, tags1, tags2)

    // Verify stats
    hits, misses := cache.Stats()
    assert.Equal(t, uint64(1), hits)
    assert.Equal(t, uint64(1), misses)
}
```

**Test: Cache Invalidation**:
```go
func TestCacheManager_Invalidation(t *testing.T) {
    cache := NewCacheManager()
    file := createTestMarkdown(t, "# Test")

    // First access
    tags1, _ := cache.GetTags(file)

    // Modify file (change mtime)
    time.Sleep(10 * time.Millisecond)
    modifyMarkdown(t, file, "# Modified")

    // Second access - should invalidate
    tags2, _ := cache.GetTags(file)
    assert.NotEqual(t, tags1[0].Name, tags2[0].Name)
}
```

**Test: Concurrent Access**:
```go
func TestCacheManager_Concurrent(t *testing.T) {
    cache := NewCacheManager()
    file := createTestMarkdown(t, "# Test")

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, err := cache.GetTags(file)
            require.NoError(t, err)
        }()
    }
    wg.Wait()

    // Should have 1 miss, 99 hits
    hits, misses := cache.Stats()
    assert.Equal(t, uint64(99), hits)
    assert.Equal(t, uint64(1), misses)
}
```

**Test: Race Conditions**:
```bash
go test -race ./pkg/ctags/
```

6. **Performance Benchmarks**:
```go
func BenchmarkCache_Hit(b *testing.B) {
    cache := NewCacheManager()
    file := setupTestFile(b)
    cache.GetTags(file) // Warm cache

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = cache.GetTags(file)
    }
}
```

**Acceptance Criteria**:
- ✅ Cache correctly handles hits/misses
- ✅ Mtime-based invalidation works
- ✅ Concurrent access is safe (no races)
- ✅ Statistics tracking works
- ✅ Performance: cache hit <1µs
- ✅ Test coverage >90%

**Time Estimate**: 2-3 hours

---

### Phase 4: Tool Integration

**Objective**: Refactor all tools to use cache manager

**Files to Modify**:
- `pkg/tools/tree.go`
- `pkg/tools/section_bounds.go`
- `pkg/tools/read_section.go`
- `pkg/tools/list_sections.go`

**Implementation Steps**:

1. **Replace ParseTagsFile Calls**:

**Before**:
```go
entries, err := ctags.ParseTagsFile(tagsFile, args.FilePath)
```

**After**:
```go
cache := ctags.GetGlobalCache()
entries, err := cache.GetTags(args.FilePath)
```

2. **Remove tags_file Parameter**:

Remove from all tool argument structs:
```go
// REMOVE THIS:
TagsFile string `json:"tags_file,omitempty" jsonschema:"description=..."`

// Tool args become simpler:
type MarkdownTreeArgs struct {
    FilePath string `json:"file_path" jsonschema:"required,description=Path to markdown file"`
}
```

3. **Error Message Updates**:

**Before**:
```
failed to parse tags file: failed to open tags file
```

**After**:
```
failed to execute ctags: ensure universal-ctags is installed
```

4. **Update Integration Tests**:

Test each tool with new cache manager:
```go
func TestMarkdownTree_WithCache(t *testing.T) {
    file := createTestMarkdown(t, "# Title\n## Section")

    args := MarkdownTreeArgs{
        FilePath: file,
    }

    resp, err := markdownTreeHandler(args)
    require.NoError(t, err)
    assert.Contains(t, resp.Tree, "Title")
    assert.Contains(t, resp.Tree, "Section")
}
```

**Acceptance Criteria**:
- ✅ All tools use cache manager
- ✅ `tags_file` parameter removed from all tools
- ✅ Tool argument structs simplified
- ✅ All tests updated and passing
- ✅ Error messages updated

**Time Estimate**: 1-2 hours

---

### Phase 5: Documentation Updates

**Objective**: Update all documentation to reflect new architecture

**Files to Modify**:
- `README.md` - User-facing documentation
- `CLAUDE.md` - Already updated with refactoring plan
- `pkg/ctags/*.go` - Add godoc comments

**Implementation Steps**:

1. **README.md Changes**:

**Remove**:
```markdown
### Generate ctags file

Before using the MCP server, generate a ctags file:

```bash
ctags -R --fields=+KnS --languages=markdown
```
```

**Add**:
```markdown
### Prerequisites

- Universal Ctags must be installed and available in PATH
- No manual ctags generation required (automatic)

### Installation

Install Universal Ctags:
```bash
# macOS
brew install universal-ctags

# Ubuntu/Debian
sudo apt-get install universal-ctags

# Fedora
sudo dnf install universal-ctags
```
```

**Update**:
```markdown
## Features

- **Zero-configuration setup**: No manual ctags generation required
- **Automatic caching**: Intelligent mtime-based cache for performance
- **Always fresh**: Automatic cache invalidation on file changes
```

2. **Update Tool Documentation**:

Remove references to `tags_file` parameter from examples:

**Before**:
```json
{
  "file_path": "docs/planning.md",
  "tags_file": "tags"
}
```

**After**:
```json
{
  "file_path": "docs/planning.md"
}
```

3. **Add Performance Section**:
```markdown
## Performance

- **Cache hits**: Sub-microsecond response time
- **Cache misses**: ~13ms (ctags execution + parsing)
- **Cache validation**: Mtime-based, ~528ns overhead
- **Typical usage**: 90%+ cache hit rate
```

4. **Update Troubleshooting**:

**Remove**:
```markdown
### Tags file not found

**Error**: `failed to parse tags file: failed to open tags file`

**Solution**: Generate a tags file in the current directory
```

**Add**:
```markdown
### Ctags not found

**Error**: `failed to execute ctags: exec: "ctags" not found`

**Solution**: Install Universal Ctags (see Prerequisites section)

### Cache Issues

To manually clear the cache (rare), restart the MCP server.
```

5. **Add Godoc Comments**:

Add comprehensive package and function documentation:
```go
// Package ctags provides markdown document structure parsing using
// Universal Ctags with intelligent mtime-based caching.
//
// The package automatically executes ctags on markdown files and caches
// the parsed results in memory. Cache invalidation is based on file
// modification time (mtime), providing fast cache validation with
// automatic invalidation when files change.
//
// Performance characteristics:
//   - Cache hit: ~528ns (mtime check only)
//   - Cache miss: ~13ms (ctags execution + JSON parsing)
//   - Typical usage: 90%+ cache hit rate
//
// Example usage:
//
//	cache := ctags.GetGlobalCache()
//	tags, err := cache.GetTags("/path/to/file.md")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, tag := range tags {
//	    fmt.Printf("%s (line %d)\n", tag.Name, tag.Line)
//	}
package ctags
```

**Acceptance Criteria**:
- ✅ README reflects zero-configuration setup
- ✅ Examples updated (no tags_file)
- ✅ Performance characteristics documented
- ✅ Godoc comments complete
- ✅ Troubleshooting updated

**Time Estimate**: 1-2 hours

---

### Phase 6: Testing & Quality Assurance

**Objective**: Comprehensive testing and validation

**Tasks**:

1. **Unit Test Coverage**:
```bash
go test -cover ./...
```
Target: >80% overall, >90% for cache.go

2. **Race Detection**:
```bash
go test -race ./...
```
Must pass with zero race conditions

3. **Linting**:
```bash
golangci-lint run
```
Must pass with zero errors/warnings

4. **Integration Testing**:
```bash
./test-integration.sh
```
All integration tests must pass

5. **Performance Benchmarks**:
```bash
go test -bench=. -benchmem ./pkg/ctags/
```
Verify cache hit performance sub-1µs

6. **Manual Testing**:
- Test with large markdown files (2000+ lines)
- Test with multiple concurrent requests
- Test cache invalidation by editing files
- Test error handling (ctags not installed)

**Acceptance Criteria**:
- ✅ All tests pass
- ✅ No race conditions
- ✅ Linting passes
- ✅ Coverage targets met
- ✅ Performance benchmarks acceptable
- ✅ Manual testing scenarios verified

**Time Estimate**: 1-2 hours

---

## Implementation Milestones

### Milestone 1: Foundation (Phases 1-2)
**Deliverables**:
- JSON parser implemented and tested
- Ctags executor implemented and tested
- Can execute ctags and parse JSON end-to-end

**Acceptance**:
- Unit tests pass for parser and executor
- Can run: `ExecuteCtags(file)` → `ParseJSONTags(json)` → `[]TagEntry`

**Time**: 2-4 hours

---

### Milestone 2: Caching Layer (Phase 3)
**Deliverables**:
- Cache manager implemented
- Concurrent safety verified
- Performance benchmarks run

**Acceptance**:
- Cache hit/miss logic works correctly
- No race conditions detected
- Cache hit performance <1µs

**Time**: 2-3 hours

---

### Milestone 3: Integration (Phases 4-5)
**Deliverables**:
- All tools use cache manager
- Documentation updated
- `tags_file` parameter removed

**Acceptance**:
- All tests updated and passing
- README reflects new architecture
- Tool interfaces simplified

**Time**: 2-4 hours

---

### Milestone 4: Quality Assurance (Phase 6)
**Deliverables**:
- Full test suite passes
- Performance validated
- Manual testing complete

**Acceptance**:
- All quality gates pass
- Performance meets targets
- Ready for production use

**Time**: 1-2 hours

---

**Total Estimated Time**: 7-13 hours (4-6 development sessions)

## Testing Strategy

### Unit Tests

**JSON Parser Tests** (`pkg/ctags/json_parser_test.go`):
- Valid JSON input → parsed entries
- Invalid JSON → error handling
- Missing optional fields → default values
- File filtering → only target file entries
- Heading level mapping → correct H1-H4 mapping

**Executor Tests** (`pkg/ctags/executor_test.go`):
- Valid markdown file → JSON output returned
- File not found → descriptive error
- Ctags not installed → clear error message
- Timeout scenario → context deadline error
- Empty file → empty JSON array

**Cache Tests** (`pkg/ctags/cache_test.go`):
- Cache hit → returns cached data without ctags execution
- Cache miss → executes ctags and updates cache
- Mtime invalidation → detects file changes
- Concurrent access → no race conditions
- Statistics tracking → correct hit/miss counts
- Manual invalidation → clears specific entry
- Full cache clear → removes all entries

### Integration Tests

**Tool Integration Tests**:
- Each tool works with cache manager
- Error handling propagates correctly
- Performance acceptable for real files
- Simplified tool interfaces (no tags_file parameter)

### Performance Tests

**Benchmarks**:
```go
BenchmarkCache_Hit           // Target: <1µs
BenchmarkCache_Miss          // Target: ~13ms
BenchmarkCache_Concurrent    // Verify scaling
```

**Load Testing**:
- 1000 sequential queries (verify cache effectiveness)
- 100 concurrent queries (verify thread safety)
- Memory usage monitoring (verify no leaks)

### Manual Testing Scenarios

1. **Happy Path**:
   - Start server
   - Query markdown file (cache miss)
   - Query same file (cache hit)
   - Verify response times

2. **Cache Invalidation**:
   - Query file (cache miss)
   - Edit file (change mtime)
   - Query file again (cache miss, fresh data)

3. **Error Handling**:
   - Query non-existent file → error
   - Run without ctags installed → clear error
   - Query invalid markdown → graceful handling

4. **Concurrent Usage**:
   - Multiple Claude Code instances
   - Query same file from multiple tools
   - Verify no race conditions or cache corruption

## Risk Assessment

### Risk 1: Ctags Not Installed
**Impact**: High (tool unusable)
**Likelihood**: Medium (users may not have ctags)
**Mitigation**:
- Clear error message with installation instructions
- Check for ctags in PATH during server startup
- Log warning if ctags not found
- Document installation prominently in README

### Risk 2: Cache Memory Growth
**Impact**: Medium (memory usage)
**Likelihood**: Low (typical usage has few files)
**Mitigation**:
- Monitor cache size in production
- Future: Implement LRU eviction if needed
- Cache only parsed tags, not raw JSON
- Document expected memory usage

### Risk 3: Mtime Unreliable
**Impact**: Low (rare cache staleness)
**Likelihood**: Very Low (standard editor behavior)
**Mitigation**:
- Document filesystem requirements
- Provide manual cache clear mechanism
- Consider MD5 fallback in future if needed
- Monitor for edge case reports

### Risk 4: Performance Regression
**Impact**: Medium (user experience)
**Likelihood**: Low (caching improves performance)
**Mitigation**:
- Benchmark before/after implementation
- Test with various file sizes
- Optimize cache lookup path (read lock only)
- Monitor real-world performance

### Risk 5: Race Conditions
**Impact**: High (data corruption, panics)
**Likelihood**: Low (careful RWMutex usage)
**Mitigation**:
- Comprehensive race detection testing
- Use RWMutex correctly (read lock for reads)
- Atomic counters for statistics
- Code review focusing on concurrency

### Risk 6: Breaking Changes for Users
**Impact**: Low (early development, no production users yet)
**Likelihood**: High (API parameters removed)
**Mitigation**:
- Project is in early development phase
- Version bump to v1.0.0 (breaking change)
- Clear documentation of new zero-configuration approach
- No backward compatibility needed (no production users)

## Success Metrics

### Performance Metrics
- ✅ Cache hit latency: <1µs (target: ~528ns)
- ✅ Cache miss latency: <20ms (target: ~13ms)
- ✅ Cache hit rate: >80% in typical usage
- ✅ Memory usage: <10MB for 100 cached files

### Quality Metrics
- ✅ Test coverage: >80% overall, >90% for cache
- ✅ Zero race conditions detected
- ✅ Zero linting errors/warnings
- ✅ All integration tests pass

### User Experience Metrics
- ✅ Zero-configuration setup (no manual ctags generation)
- ✅ Automatic cache invalidation (no stale data)
- ✅ Clear error messages (ctags installation)
- ✅ Simplified tool interfaces (removed tags_file parameter)

### Code Quality Metrics
- ✅ Reduced complexity (JSON vs tab-separated parsing)
- ✅ Better testability (mock cache vs mock files)
- ✅ Comprehensive documentation (godoc + README)
- ✅ Maintainability (cleaner architecture)

## Rollout Plan

### Development Phase
1. Implement on feature branch: `feat/json-cache-refactoring`
2. Incremental commits for each phase
3. Code review after each milestone
4. Merge to main after all phases complete

### Testing Phase
1. Run full test suite (unit + integration)
2. Performance benchmarking
3. Manual testing scenarios
4. Race detection validation

### Release Phase
1. Version bump: `v0.x.x` → `v1.0.0` (breaking change - removed tags_file parameter)
2. Update CHANGELOG.md with changes and benefits
3. Tag release with comprehensive release notes
4. Update MCP registry (if applicable)

### Post-Release
1. Monitor for issue reports
2. Collect performance data
3. Address any edge cases discovered
4. Consider future enhancements (LRU, etc.)

## Future Enhancements

**Not included in this refactoring**, but potential future features:

1. **LRU Cache Eviction**: Implement size-limited cache with LRU policy
2. **Persistent Cache**: Optional disk-based cache for server restarts
3. **Cache Statistics API**: Expose cache metrics via MCP tool
4. **Workspace-Level Caching**: Optional shared cache for related files
5. **MD5 Fallback**: Use MD5 if mtime proves unreliable
6. **Cache Warming**: Pre-populate cache for known files
7. **Parallel Ctags Execution**: Execute ctags for multiple files concurrently

## Conclusion

This refactoring transforms the markdown-mcp server from a manual, file-based approach to an intelligent, cache-aware system. The benefits include:

- **Better UX**: Zero-configuration, automatic cache management
- **Better Performance**: 915x faster cache validation, near-instant cache hits
- **Better Code**: JSON parsing, cleaner architecture, better tests
- **Better Maintainability**: Easier to extend, comprehensive documentation

The implementation is well-scoped, thoroughly planned, and ready for execution.

---

**Status**: Backlog - Ready for Implementation
**Next Steps**: Move to `active/` when development begins
**Estimated Completion**: 4-6 development sessions
**Dependencies**: None (can start immediately)
