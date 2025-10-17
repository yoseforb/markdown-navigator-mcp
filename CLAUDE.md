# Markdown Navigation MCP Server - Development Guide

## Project Overview

**Purpose**: An MCP (Model Context Protocol) server that provides efficient navigation of large markdown files using ctags. This enables Claude Code and other MCP clients to work with large documentation files (2,000+ lines) without wasting context tokens by allowing targeted section reading.

**Key Innovation**: Context-efficient navigation that enables autonomous agents to understand and work with large markdown documents by reading only the sections they need, reducing context usage by 70-80%.

**Language**: Go 1.25.3
**Framework**: [gomcp](https://github.com/localrivet/gomcp) v1.7.2 by LocalRivet
**Module**: `github.com/yoseforb/markdown-nav-mcp`

## Architecture

### Current Implementation

The project uses **on-demand ctags execution with mtime-based caching**:

1. **JSON-based ctags parsing** (`--output-format=json`)
2. **On-demand ctags execution** per file with automatic caching
3. **In-memory caching** with file modification time (mtime) tracking
4. **Concurrent-safe cache** with `sync.RWMutex` protection
5. **Automatic cache invalidation** when files change
6. **Zero configuration** - no manual ctags file generation required

### Architecture History

Previous versions (pre-v0.1.0) used a **pre-generated ctags file** approach that required manual tags file generation. This was replaced with the current on-demand execution model for zero-configuration operation.

### Performance Characteristics

Based on benchmark testing:
- **mtime checking**: 528 nanoseconds (~1,900,000 ops/sec)
- **MD5 hashing**: 483 microseconds (~2,070 ops/sec)
- **ctags execution**: 12.6 milliseconds (~79 ops/sec)

**Cache Strategy**: mtime-based (915x faster than MD5, reliable for markdown editing workflows)
- Cache hit: Instant (~528ns)
- Cache miss: ~13ms (ctags execution + JSON parsing + cache update)

### Package Structure

```
markdown-mcp/
├── cmd/
│   ├── server/
│   │   └── main.go             # MCP server entry point
│   └── test_tools/
│       └── main.go             # Tool testing utility
├── pkg/
│   ├── ctags/
│   │   ├── parser.go           # Current: Tab-separated ctags parsing
│   │   ├── tree.go             # Tree structure building from tags
│   │   ├── parser_test.go      # Tests for current parser
│   │   ├── tree_test.go        # Tests for tree building
│   │   ├── json_parser.go      # NEW: JSON ctags parsing
│   │   ├── cache.go            # NEW: Mtime-based cache manager
│   │   └── executor.go         # NEW: Ctags command execution
│   └── tools/
│       ├── errors.go           # Error definitions
│       ├── tree.go             # markdown_tree tool
│       ├── section_bounds.go   # markdown_section_bounds tool
│       ├── read_section.go     # markdown_read_section tool
│       └── list_sections.go    # markdown_list_sections tool
├── testdata/                   # Test fixtures
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── .golangci.yml               # Linter configuration (strict)
├── README.md                   # User-facing documentation
├── CLAUDE.md                   # This file - development guide
└── ai-docs/                    # AI agent documentation
    └── planning/               # Planning documents
        ├── backlog/           # Features awaiting implementation
        ├── active/            # Current implementation work
        ├── completed/         # Recently completed (last 30 days)
        └── archived/          # Historical reference
```

## MCP Tools

The server provides 4 MCP tools for markdown navigation:

### 1. markdown_tree

**Purpose**: Display hierarchical document structure (vim-vista style)

**Parameters**:
- `file_path` (required): Path to markdown file
- `format` (optional): Output format - "json" or "ascii" (default: "json")
- `section_name_pattern` (optional): Regex pattern to filter sections
- `max_depth` (optional): Maximum tree depth (1-6, 0=all, default: 2)

**Output Format**:
```
planning.md

└ Planning Document H1:1
  │ Overview H2:5
  │ Task 1: Authentication H2:50
    │ Requirements H3:52
    │ Implementation H3:75
  │ Task 2: Database Schema H2:150
```

**Use Case**: Quick overview of document structure before targeted section reading

### 2. markdown_section_bounds

**Purpose**: Find line number boundaries for a specific section

**Parameters**:
- `file_path` (required): Path to markdown file
- `section_heading` (required): Exact heading text (case-sensitive, without # symbols)

**Response**:
```json
{
  "section_name": "Task 1: Authentication",
  "start_line": 50,
  "end_line": 149,
  "heading_level": "H2",
  "total_lines": 100
}
```

**Use Case**: Determine section boundaries for targeted reading

### 3. markdown_read_section

**Purpose**: Read a specific section's content

**Parameters**:
- `file_path` (required): Path to markdown file
- `section_heading` (required): Exact heading text (case-sensitive, without # symbols)
- `max_subsection_levels` (optional): Limit subsection depth (omit for unlimited, 0=no subsections, 1=immediate children, 2=children+grandchildren)

**Response**:
```json
{
  "content": "## Task 1: Authentication\n\n### Requirements\n...",
  "section_name": "Task 1: Authentication",
  "start_line": 50,
  "end_line": 149,
  "lines_read": 100
}
```

**Use Case**: Read only the relevant section without loading entire document

### 4. markdown_list_sections

**Purpose**: List all sections matching filters

**Parameters**:
- `file_path` (required): Path to markdown file
- `max_depth` (optional): Maximum heading depth (1-6, 0=all, default: 2)
- `section_name_pattern` (optional): Regex pattern to filter section names

**Response**:
```json
{
  "sections": [
    {"name": "Task 1: Authentication", "line": 50, "level": "H2"},
    {"name": "Task 2: Database Schema", "line": 150, "level": "H2"}
  ],
  "count": 2
}
```

**Use Case**: Discover available sections and their locations

## Development Workflow

### Quality Gates (MANDATORY)

All code changes MUST pass these quality gates before commit:

1. **Linting**: `golangci-lint run` (zero errors/warnings)
2. **Formatting**: `gofumpt -w .` (code must be formatted)
3. **Testing**: `go test ./...` (all tests must pass)
4. **Coverage**: Maintain or improve test coverage
5. **Build**: `go build -o mdnav-server ./cmd/server` (must compile)

### Go Development Standards

#### Code Style

- **Formatting**: Use `gofumpt` (stricter than `gofmt`)
- **Line Length**: Maximum 80 characters (enforced by `golines`)
- **Import Organization**: `goimports` with local prefix grouping
- **Comment Style**: All exported functions/types require godoc comments
- **Error Handling**: Wrap errors with context using `fmt.Errorf` with `%w`

#### Naming Conventions

- **Packages**: Short, lowercase, single word (no underscores)
- **Files**: Lowercase with underscores (e.g., `json_parser.go`)
- **Types**: PascalCase for exported, camelCase for unexported
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase (not SCREAMING_SNAKE_CASE)
- **Test files**: `*_test.go` suffix

#### Error Handling

```go
// BAD - loses context
if err != nil {
    return nil, err
}

// GOOD - wraps with context
if err != nil {
    return nil, fmt.Errorf("failed to parse tags file: %w", err)
}

// GOOD - custom error wrapping
if err != nil {
    return nil, fmt.Errorf("%w: %s", ErrSectionNotFound, sectionQuery)
}
```

#### MCP Tool Parameter Handling

**Making Parameters Optional**

In gomcp (v1.7.2), **all struct fields are treated as required by default**, regardless of JSON tags. To make a parameter optional, you MUST use **pointer types**.

**Struct Tag Format**

gomcp requires **separate tags** for `description:` and `required:`:

```go
type ToolArgs struct {
    // CORRECT - Required parameter with separate tags
    FilePath string `json:"file_path" description:"Path to file" required:"true"`

    // CORRECT - Optional parameter (pointer type, no required tag)
    Pattern *string `json:"pattern,omitempty" description:"Regex pattern to filter"`
}
```

**WRONG - These formats don't work:**
```go
type ToolArgs struct {
    // WRONG - Combined jsonschema tag (old format)
    FilePath string `json:"file_path" jsonschema:"required,description=Path to file"`

    // WRONG - Non-pointer optional (still required!)
    Pattern  string `json:"pattern,omitempty" description:"Pattern"`
}
```

**Handling Optional Parameters in Code:**
```go
func MyTool(ctx *server.Context, args ToolArgs) (interface{}, error) {
    // Check if optional parameter was provided
    if args.Pattern != nil && *args.Pattern != "" {
        // Use the pattern
        filteredData = filterByPattern(data, *args.Pattern)
    }
    // If Pattern is nil, it was not provided
}
```

**Generated MCP Schema:**

With string field:
```json
{
  "required": ["file_path", "pattern"]  // Both required!
}
```

With pointer field:
```json
{
  "required": ["file_path"]  // Only file_path required
}
```

**Rule of Thumb:**
- Required parameter: Use value type (`string`, `int`, `bool`) with `required:"true"` tag
- Optional parameter: Use pointer type (`*string`, `*int`, `*bool`) without required tag
- Use separate `description:` and `required:` tags (not combined `jsonschema:` tag)
- Always check for `nil` before dereferencing pointers

#### Testing Standards

- **Table-driven tests** for multiple scenarios
- **Descriptive test names**: `TestFunctionName_Scenario_ExpectedBehavior`
- **Test coverage**: Aim for >80% coverage on new code
- **Test fixtures**: Use `testdata/` directory for test files
- **Parallel tests**: Use `t.Parallel()` when tests are independent

Example:
```go
func TestParseTags_ValidInput_ReturnsEntries(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name        string
        input       string
        want        []*TagEntry
        wantErr     bool
    }{
        {
            name:  "single section",
            input: "section1\tfile.md\t/^## Section 1$/\tsection\tline:10\n",
            want:  []*TagEntry{{Name: "section1", Line: 10}},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseTags(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseTags() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseTags() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Linting Configuration

The project uses a **strict golangci-lint configuration** (`.golangci.yml`):

#### Enabled Linters (50+)

Key linters to be aware of:
- **errcheck**: All errors must be checked
- **govet**: Standard Go vet checks
- **staticcheck**: Advanced static analysis
- **exhaustruct**: Struct fields must be explicitly initialized (with exclusions)
- **gochecknoglobals**: No global variables allowed
- **gochecknoinits**: No `init()` functions allowed
- **sloglint**: Enforces `log/slog` usage (no global loggers)
- **wrapcheck**: Errors from external packages must be wrapped
- **cyclop**: Cyclomatic complexity <= 30
- **gocognit**: Cognitive complexity <= 20
- **funlen**: Functions <= 100 lines, <= 50 statements

#### Exclusions

- Test files (`*_test.go`): Relaxed requirements for test utilities
- Anonymous structs: Exhaustruct checking disabled

#### Common Issues and Fixes

**Issue**: Struct initialization missing fields
```go
// BAD
entry := TagEntry{Name: "test"}

// GOOD
entry := TagEntry{
    Name:    "test",
    File:    "",
    Pattern: "",
    Kind:    "",
    Line:    0,
    Scope:   "",
    Level:   0,
}
```

**Issue**: Unchecked error
```go
// BAD
file.Close()

// GOOD
if err := file.Close(); err != nil {
    return fmt.Errorf("failed to close file: %w", err)
}
```

**Issue**: Global variable
```go
// BAD
var cache = make(map[string]*CacheEntry)

// GOOD - use struct method
type CacheManager struct {
    cache map[string]*CacheEntry
}
```

**Issue**: Using `log` package instead of `slog`
```go
// BAD
log.Printf("error: %v", err)

// GOOD
logger.Error("operation failed", "error", err)
```

### Git Commit Standards

#### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring without behavior change
- `perf`: Performance improvement
- `test`: Adding or modifying tests
- `docs`: Documentation changes
- `chore`: Build, tooling, dependencies

**Examples**:
```
feat(ctags): add JSON parser for ctags output

Implement JSON parsing for ctags --output-format=json.
This replaces tab-separated parsing and provides structured
data that's easier to work with.

Closes #123
```

```
refactor(cache): implement mtime-based caching

Replace direct ctags file parsing with cached execution.
Uses file modification time for cache invalidation (915x
faster than MD5 hashing).

Performance: ~528ns cache hits, ~13ms cache misses
```

### Development Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run linter
golangci-lint run

# Auto-fix linter issues
golangci-lint run --fix

# Format code
gofumpt -w .

# Build server
go build -o mdnav-server ./cmd/server

# Run integration tests
./test-integration.sh

# Install dependencies
go mod download

# Update dependencies
go get -u ./...
go mod tidy
```

### Testing Strategy

#### Unit Tests

- **Location**: Same package as code (`*_test.go`)
- **Coverage**: Each public function should have tests
- **Focus**: Input validation, edge cases, error conditions

#### Integration Tests

- **Location**: `cmd/test_tools/` and shell scripts
- **Focus**: End-to-end tool behavior with real markdown files
- **Fixtures**: Use `testdata/` directory

#### Cache Testing

New caching layer requires specific tests:

1. **Cache Hit Behavior**:
   - Verify cached data is returned without ctags execution
   - Measure performance (should be <1µs)

2. **Cache Miss Behavior**:
   - Verify ctags is executed on first access
   - Verify cache is populated correctly

3. **Cache Invalidation**:
   - Verify cache is invalidated when file mtime changes
   - Verify stale data is not returned

4. **Concurrent Access**:
   - Verify multiple goroutines can safely access cache
   - No race conditions (test with `go test -race`)

Example cache test:
```go
func TestCache_MtimeInvalidation(t *testing.T) {
    t.Parallel()

    cache := NewCacheManager()
    testFile := createTempMarkdown(t, "# Test")

    // First access - cache miss
    tags1, err := cache.GetTags(testFile)
    require.NoError(t, err)

    // Second access - cache hit
    tags2, err := cache.GetTags(testFile)
    require.NoError(t, err)
    assert.Equal(t, tags1, tags2)

    // Modify file
    time.Sleep(10 * time.Millisecond) // Ensure mtime changes
    modifyMarkdown(t, testFile, "# Modified")

    // Third access - cache miss (invalidated)
    tags3, err := cache.GetTags(testFile)
    require.NoError(t, err)
    assert.NotEqual(t, tags1, tags3)
}
```

## Recent Breaking Changes (2025-10-17)

**Commit**: `0a78d45` - Parameter renames and struct tag fixes

### Parameter Name Changes

All MCP tools have been updated with clearer, more explicit parameter names:

**markdown_tree:**
- `pattern` → `section_name_pattern`
- Added `format` parameter (json/ascii)

**markdown_section_bounds:**
- `section_query` → `section_heading` (now requires exact match, case-sensitive)

**markdown_read_section:**
- `section_query` → `section_heading` (now requires exact match, case-sensitive)
- `depth` → `max_subsection_levels` (clearer semantics)

**markdown_list_sections:**
- `heading_level` (enum: H1/H2/H3/ALL) → `max_depth` (integer: 0-6, cumulative)
- `pattern` → `section_name_pattern`

### Struct Tag Format Change

All tools now use separate `description:` and `required:` tags instead of combined `jsonschema:` tags:

```go
// NEW format (required)
FilePath string `json:"file_path" description:"Path to file" required:"true"`

// OLD format (no longer works)
FilePath string `json:"file_path" jsonschema:"required,description=Path to file"`
```

### Migration Guide

If you have code or scripts using the old parameter names:
1. Replace `section_query` with `section_heading` (and use exact heading text)
2. Replace `pattern` with `section_name_pattern`
3. Replace `heading_level` values (H1/H2/etc) with `max_depth` numbers (1/2/etc)
4. Replace `depth` with `max_subsection_levels`

## Project Context

### Problem Statement

Claude Code and other MCP clients have limited context windows. Large markdown documentation files (2,000+ lines) consume significant context tokens when loaded entirely. This makes it difficult to work with planning documents, technical specifications, and comprehensive documentation without hitting context limits.

### Solution Approach

Provide **structural navigation** of markdown files using ctags:
1. Parse markdown heading structure (H1-H6, all 6 levels supported as of 2025-10-17)
2. Expose hierarchical tree view
3. Enable targeted section reading
4. Support section querying and filtering

This allows agents to:
- View document structure without reading content (tree view)
- Read only relevant sections (section reading)
- Navigate large documents efficiently (section bounds)
- Discover available sections (section listing)

### Target Use Cases

**Use Case 1: Planning Document Analysis**
- Agent receives: "Analyze Task 4 from the route refactoring plan"
- Agent workflow:
  1. `markdown_tree` - Get document overview
  2. `markdown_section_bounds` with `section_heading="Task 4"` - Find Task 4 location
  3. `markdown_read_section` with `section_heading="Task 4"` - Read only Task 4 content
- Result: Comprehensive analysis using only 25% of document tokens

**Use Case 2: Implementation Workflow**
- Agent receives: "Implement Task 4 from planning document"
- Agent workflow:
  1. `markdown_tree` - Understand document structure
  2. `markdown_read_section` with `section_heading="Executive Summary"` - Get context
  3. `markdown_read_section` with `section_heading="Task 4"` - Get implementation details
  4. `markdown_read_section` with `section_heading="Testing Strategy"` - Get validation requirements
- Result: Autonomous implementation with minimal context usage

**Use Case 3: Documentation Discovery**
- Agent receives: "What testing tasks are documented?"
- Agent workflow:
  1. `markdown_list_sections` with `section_name_pattern="test"` - Find all testing sections
  2. For each section: `markdown_read_section` with exact `section_heading` - Read details
- Result: Comprehensive understanding of documented testing requirements

### Design Decisions

**Decision 1: Use ctags instead of custom parser**
- **Rationale**: ctags is battle-tested, widely available, handles edge cases
- **Alternative**: Custom markdown parser (more maintenance, reinventing wheel)
- **Trade-off**: External dependency vs. implementation complexity

**Decision 2: Mtime-based caching instead of MD5 hashing**
- **Rationale**: 915x faster, reliable for typical markdown editing workflows
- **Alternative**: MD5 hashing (more reliable but significantly slower)
- **Trade-off**: Performance vs. theoretical cache reliability
- **Context**: Markdown editors update mtime on save (standard behavior)

**Decision 3: JSON ctags output instead of tab-separated**
- **Rationale**: Structured data, easier parsing, extensible, less error-prone
- **Alternative**: Continue with tab-separated parsing
- **Trade-off**: Requires newer ctags version vs. simpler parsing

**Decision 4: In-memory cache instead of persistent cache**
- **Rationale**: Simpler implementation, no disk I/O, adequate for typical usage
- **Alternative**: Persistent cache on disk
- **Trade-off**: Cache lost on restart vs. implementation complexity

**Decision 5: Per-file caching instead of workspace caching**
- **Rationale**: More granular invalidation, simpler cache key management
- **Alternative**: Workspace-level caching (one tags file for all files)
- **Trade-off**: Multiple ctags executions vs. cache complexity

### Dependencies

**Runtime Dependencies**:
- Universal Ctags (external binary) - markdown parsing
- gomcp v1.7.2 - MCP server framework
- Go standard library only (no external Go dependencies)

**Development Dependencies**:
- golangci-lint - comprehensive linting
- gofumpt - strict formatting
- golines - line length enforcement

### Future Enhancements

**Potential features** (not currently planned):

1. **LRU Cache Eviction**: Implement size-limited cache with LRU eviction
2. **Cache Statistics API**: Expose cache hit/miss metrics via MCP tool
3. **Workspace-Level Caching**: Optional workspace-wide tags file support
4. **Custom Section Delimiters**: Support for non-standard markdown structures
5. **Section Diffing**: Compare sections across document versions
6. **Section Search**: Full-text search within sections
7. **Cross-Reference Detection**: Detect links between sections

## Quick Reference

### Common Development Tasks

**Start new feature**:
```bash
git checkout -b feat/feature-name
# Implement feature
go test ./...
golangci-lint run
gofumpt -w .
git commit -m "feat(scope): description"
```

**Fix a bug**:
```bash
git checkout -b fix/bug-description
# Fix bug
# Add regression test
go test ./...
golangci-lint run
git commit -m "fix(scope): description"
```

**Run full validation**:
```bash
# Format, lint, test, build
gofumpt -w . && golangci-lint run && go test ./... && go build ./cmd/server
```

**Debug ctags output**:
```bash
# See tab-separated format
ctags -R --fields=+KnS --languages=markdown -f - file.md

# See JSON format
ctags --output-format=json --fields=+KnS --languages=markdown -f - file.md
```

### Key Files Reference

- **cmd/server/main.go**: Server initialization, tool registration
- **pkg/ctags/parser.go**: Current tab-separated parsing logic
- **pkg/ctags/tree.go**: Tree structure building from tags
- **pkg/tools/tree.go**: markdown_tree tool implementation
- **pkg/tools/section_bounds.go**: markdown_section_bounds tool
- **pkg/tools/read_section.go**: markdown_read_section tool
- **pkg/tools/list_sections.go**: markdown_list_sections tool
- **.golangci.yml**: Strict linting configuration
- **ai-docs/planning/**: Agent planning documents

### Important Links

- **MCP Specification**: https://modelcontextprotocol.io/
- **gomcp Framework**: https://github.com/localrivet/gomcp
- **Universal Ctags**: https://ctags.io/
- **golangci-lint**: https://golangci-lint.run/
- **Go Testing**: https://pkg.go.dev/testing

---

**Last Updated**: 2025-10-17
**Status**: Active Development - JSON-based caching implemented, continuous improvements
