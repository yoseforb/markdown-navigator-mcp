# Markdown Navigation MCP Server - Project Status

## Quick Metrics

- **Version**: 0.1.0-dev (pre-release)
- **Binary**: `mdnav-server`
- **Language**: Go 1.21+
- **Framework**: gomcp v1.7.2
- **Tools**: 4 MCP tools
- **Build**: ✅ PASSING
- **Tests**: ✅ PASSING
- **Linting**: ✅ COMPLIANT

## Overview

**markdown-mcp** is an MCP (Model Context Protocol) server that provides efficient navigation of large markdown files using ctags. The project enables Claude Code and other MCP clients to work with large documentation files without wasting context tokens by allowing targeted section reading.

## Features

### MCP Tools

1. **markdown_tree** - Hierarchical document structure display (JSON/ASCII)
2. **markdown_section_bounds** - Line number boundary detection for sections
3. **markdown_read_section** - Targeted section content reading with depth control
4. **markdown_list_sections** - Section discovery with filtering and level limits

### Technical Features

- ✅ Zero-configuration operation (automatic ctags execution)
- ✅ On-demand ctags execution with JSON parsing
- ✅ Mtime-based cache invalidation (sub-microsecond cache hits)
- ✅ Context propagation for cancellation support
- ✅ Support for all 6 heading levels (H1-H6)
- ✅ Graceful shutdown (SIGINT/SIGTERM)

## Architecture

```
User → MCP Tool → CacheManager (mtime check) → On-demand ctags execution → JSON parsing → TagEntry[]
```

### Performance

- **Cache hit**: ~528ns (sub-microsecond)
- **Cache miss**: ~13ms (ctags execution + JSON parsing)
- **Context reduction**: 50-70% token savings vs reading entire files
- **Cache strategy**: mtime-based (915x faster than MD5)

## Project Structure

```
markdown-mcp/
├── cmd/server/             # MCP server entry point
├── pkg/
│   ├── ctags/             # Ctags execution, parsing, caching
│   └── tools/             # MCP tool implementations
├── testdata/              # Test fixtures
├── ai-docs/               # AI agent documentation
├── .golangci.yml          # Linter configuration (50+ linters)
├── README.md              # User documentation
└── CLAUDE.md              # Development guide
```

## Quality Standards

### Code Quality Gates

1. **Linting**: golangci-lint (50+ linters, zero tolerance)
2. **Formatting**: gofumpt (stricter than gofmt)
3. **Testing**: All tests must pass
4. **Coverage**: >80% for new code
5. **Build**: Must compile successfully

### Testing

- ✅ Unit tests with table-driven approach
- ✅ Integration test scripts
- ✅ Race detection verified (`go test -race`)
- ✅ Test fixtures in `testdata/`

## Dependencies

### Runtime
- **gomcp** v1.7.2 - MCP server framework
- **Universal Ctags** - External binary for markdown parsing
- **Go standard library** - No external Go dependencies

### Development
- **golangci-lint** - Comprehensive linting
- **gofumpt** - Strict code formatting
- **golines** - Line length enforcement (80 chars)

## Known Limitations

### Implementation
- No persistent cache (cleared on server restart)
- No LRU eviction (unlimited cache size)
- No cache statistics API

### Documentation
- No git tags/releases yet (pre-v1.0)
- No automated MCP client integration tests

## Next Priorities

### Active
- Monitor user feedback on tool parameters
- Address edge cases discovered in production

### Backlog
1. **LRU Cache Eviction** - Size-limited cache
2. **Cache Statistics API** - Hit/miss metrics
3. **Integration Tests** - MCP client test suite
4. **Release Process** - Git tags, versioning, changelog

## Development Commands

```bash
# Run tests
go test ./...

# Run linter
golangci-lint run

# Format code
gofumpt -w .

# Build server
go build -o mdnav-server ./cmd/server

# Full validation
gofumpt -w . && golangci-lint run && go test ./... && go build ./cmd/server
```

## Health Summary

**Status**: ✅ **Healthy - Stable**

**Strengths**:
- Zero-configuration user experience
- Strong quality standards enforced
- Excellent cache performance
- Comprehensive documentation

**Risks**:
- Unlimited cache size (memory concern)
- No versioning process (pre-v1.0)

**Confidence**: **High** - Implementation is solid, well-tested, and performant.