# Markdown Navigation MCP Server - Development Guide

## Project Overview

**Purpose**: An MCP (Model Context Protocol) server that provides efficient navigation of large markdown files using ctags. This enables Claude Code and other MCP clients to work with large documentation files (2,000+ lines) without wasting context tokens by allowing targeted section reading.

**Key Innovation**: Context-efficient navigation that enables autonomous agents to understand and work with large markdown documents by reading only the sections they need, reducing context usage by 70-80%.

**Language**: Go 1.25.3
**Framework**: [gomcp](https://github.com/localrivet/gomcp) v1.7.2 by LocalRivet
**Module**: `github.com/yoseforb/markdown-nav-mcp`

## Architecture

The project uses **on-demand ctags execution with mtime-based caching**:

1. **JSON-based ctags parsing** (`--output-format=json`)
2. **On-demand ctags execution** per file with automatic caching
3. **In-memory caching** with file modification time (mtime) tracking
4. **Concurrent-safe cache** with `sync.RWMutex` protection
5. **Automatic cache invalidation** when files change
6. **Zero configuration** - no manual ctags file generation required

### Performance Characteristics

- **Cache hit**: Instant (~528ns)
- **Cache miss**: ~13ms (ctags execution + JSON parsing + cache update)
- **Cache strategy**: mtime-based (915x faster than MD5)

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
│   │   ├── parser.go           # Tab-separated ctags parsing
│   │   ├── tree.go             # Tree structure building from tags
│   │   ├── json_parser.go      # JSON ctags parsing
│   │   ├── cache.go            # Mtime-based cache manager
│   │   └── executor.go         # Ctags command execution
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

### 2. markdown_section_bounds

**Purpose**: Find line number boundaries for a specific section

**Parameters**:
- `file_path` (required): Path to markdown file
- `section_heading` (required): Exact heading text (case-sensitive, without # symbols)

### 3. markdown_read_section

**Purpose**: Read a specific section's content

**Parameters**:
- `file_path` (required): Path to markdown file
- `section_heading` (required): Exact heading text (case-sensitive, without # symbols)
- `max_subsection_levels` (optional): Limit subsection depth (omit for unlimited)

### 4. markdown_list_sections

**Purpose**: List all sections matching filters

**Parameters**:
- `file_path` (required): Path to markdown file
- `max_depth` (optional): Maximum heading depth (1-6, 0=all, default: 2)
- `section_name_pattern` (optional): Regex pattern to filter section names

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

In gomcp (v1.7.2), **all struct fields are treated as required by default**. To make a parameter optional, you MUST use **pointer types**.

```go
type ToolArgs struct {
    // Required parameter with separate tags
    FilePath string `json:"file_path" description:"Path to file" required:"true"`

    // Optional parameter (pointer type, no required tag)
    Pattern *string `json:"pattern,omitempty" description:"Regex pattern to filter"`
}

func MyTool(ctx *server.Context, args ToolArgs) (interface{}, error) {
    // Check if optional parameter was provided
    if args.Pattern != nil && *args.Pattern != "" {
        filteredData = filterByPattern(data, *args.Pattern)
    }
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

### Linting Configuration

The project uses a **strict golangci-lint configuration** (`.golangci.yml`) with 50+ linters.

Key linters:
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

### Git Commit Standards

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

# Run full validation
gofumpt -w . && golangci-lint run && go test ./... && go build ./cmd/server

# Debug ctags output
ctags --output-format=json --fields=+KnS --languages=markdown -f - file.md
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

Required tests for caching layer:

1. **Cache Hit Behavior**: Verify cached data is returned without ctags execution
2. **Cache Miss Behavior**: Verify ctags is executed on first access
3. **Cache Invalidation**: Verify cache is invalidated when file mtime changes
4. **Concurrent Access**: Verify multiple goroutines can safely access cache

## Design Decisions

**Decision 1: Use ctags instead of custom parser**
- **Rationale**: ctags is battle-tested, widely available, handles edge cases
- **Trade-off**: External dependency vs. implementation complexity

**Decision 2: Mtime-based caching instead of MD5 hashing**
- **Rationale**: 915x faster, reliable for typical markdown editing workflows
- **Trade-off**: Performance vs. theoretical cache reliability

**Decision 3: JSON ctags output instead of tab-separated**
- **Rationale**: Structured data, easier parsing, extensible, less error-prone
- **Trade-off**: Requires newer ctags version vs. simpler parsing

**Decision 4: In-memory cache instead of persistent cache**
- **Rationale**: Simpler implementation, no disk I/O, adequate for typical usage
- **Trade-off**: Cache lost on restart vs. implementation complexity

**Decision 5: Per-file caching instead of workspace caching**
- **Rationale**: More granular invalidation, simpler cache key management
- **Trade-off**: Multiple ctags executions vs. cache complexity

## Dependencies

**Runtime Dependencies**:
- Universal Ctags (external binary) - markdown parsing
- gomcp v1.7.2 - MCP server framework
- Go standard library only (no external Go dependencies)

**Development Dependencies**:
- golangci-lint - comprehensive linting
- gofumpt - strict formatting
- golines - line length enforcement

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
# Fix bug and add regression test
go test ./...
golangci-lint run
git commit -m "fix(scope): description"
```

### Key Files Reference

- **cmd/server/main.go**: Server initialization, tool registration
- **pkg/ctags/**: Ctags execution, parsing, caching
- **pkg/tools/**: MCP tool implementations
- **.golangci.yml**: Strict linting configuration
- **ai-docs/**: AI agent documentation

### Important Links

- **MCP Specification**: https://modelcontextprotocol.io/
- **gomcp Framework**: https://github.com/localrivet/gomcp
- **Universal Ctags**: https://ctags.io/
- **golangci-lint**: https://golangci-lint.run/
- **Go Testing**: https://pkg.go.dev/testing