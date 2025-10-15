# Markdown Navigation MCP Server

An MCP (Model Context Protocol) server that provides efficient navigation of large markdown files using ctags. This allows Claude Code and other MCP clients to work with large documentation files (2,000+ lines) without wasting context tokens.

## Features

- **Zero-configuration setup**: No manual ctags generation required - automatic on-demand execution
- **Intelligent caching**: Mtime-based cache for instant responses on repeated queries
- **Always fresh data**: Automatic cache invalidation when files change
- **markdown_tree**: Display hierarchical document structure (vim-vista style)
- **markdown_section_bounds**: Find line number boundaries for specific sections
- **markdown_read_section**: Read a specific section's content
- **markdown_list_sections**: List all sections matching filters (level, pattern)

## Prerequisites

- Go 1.21 or higher
- Universal Ctags (installed and available in PATH)

**Install Universal Ctags:**
```bash
# macOS
brew install universal-ctags

# Ubuntu/Debian
sudo apt-get install universal-ctags

# Fedora
sudo dnf install universal-ctags
```

**Note**: Universal Ctags is required for the server to function. The server will execute ctags automatically when needed - no manual ctags file generation required.

## Installation

### Build from source

```bash
# Clone or navigate to the repository
cd markdown-mcp

# Build the server
go build -o markdown-nav-server ./cmd/server

# Optional: Install to your PATH
sudo cp markdown-nav-server /usr/local/bin/
```

## Usage

### Configure Claude Code

Add the server to your Claude Code MCP configuration (`~/.config/claude-code/mcp.json` or similar):

```json
{
  "mcpServers": {
    "markdown-nav": {
      "command": "/path/to/markdown-nav-server"
    }
  }
}
```

Or if installed in PATH:

```json
{
  "mcpServers": {
    "markdown-nav": {
      "command": "markdown-nav-server"
    }
  }
}
```

### Configuration Options

The server supports the following command-line flags:

#### `-ctags-path`

Specify a custom path to the ctags executable. By default, the server looks for `ctags` in your PATH.

**Usage examples:**

```json
{
  "mcpServers": {
    "markdown-nav": {
      "command": "/path/to/markdown-nav-server",
      "args": ["-ctags-path", "/usr/local/bin/universal-ctags"]
    }
  }
}
```

Or for custom ctags installation:

```json
{
  "mcpServers": {
    "markdown-nav": {
      "command": "markdown-nav-server",
      "args": ["-ctags-path", "/opt/ctags/bin/ctags"]
    }
  }
}
```

**When to use:**
- Ctags is installed in a non-standard location
- Multiple ctags versions are installed and you need to specify which one
- Using a custom-built ctags binary

**Command-line usage:**

```bash
# Use default ctags from PATH
./markdown-nav-server

# Specify custom ctags path
./markdown-nav-server -ctags-path /usr/local/bin/universal-ctags

# View help
./markdown-nav-server -h
```

## Tools

### markdown_tree

Display hierarchical document structure.

**Parameters:**
- `file_path` (required): Path to markdown file

**Example:**
```json
{
  "file_path": "docs/planning.md"
}
```

**Response:**
```
planning.md

└ Planning Document H1:1
  │ Overview H2:5
  │ Task 1: Authentication H2:50
    │ Requirements H3:52
    │ Implementation H3:75
  │ Task 2: Database Schema H2:150
```

### markdown_section_bounds

Find line number boundaries for a specific section.

**Parameters:**
- `file_path` (required): Path to markdown file
- `section_query` (required): Section name or search query (fuzzy match)

**Example:**
```json
{
  "file_path": "docs/planning.md",
  "section_query": "Task 1"
}
```

**Response:**
```json
{
  "section_name": "Task 1: Authentication",
  "start_line": 50,
  "end_line": 149,
  "heading_level": "H2",
  "total_lines": 100
}
```

### markdown_read_section

Read a specific section's content.

**Parameters:**
- `file_path` (required): Path to markdown file
- `section_query` (required): Section name or search query
- `include_subsections` (optional): Include child sections (default: true)

**Example:**
```json
{
  "file_path": "docs/planning.md",
  "section_query": "Task 1"
}
```

**Response:**
```json
{
  "content": "## Task 1: Authentication\n\n### Requirements\n...",
  "section_name": "Task 1: Authentication",
  "start_line": 50,
  "end_line": 149,
  "lines_read": 100
}
```

### markdown_list_sections

List all sections matching filters.

**Parameters:**
- `file_path` (required): Path to markdown file
- `heading_level` (optional): Filter by level (H1, H2, H3, H4)
- `pattern` (optional): Search pattern (fuzzy match)

**Example:**
```json
{
  "file_path": "docs/planning.md",
  "heading_level": "H2",
  "pattern": "Task"
}
```

**Response:**
```json
{
  "sections": [
    {"name": "Task 1: Authentication", "line": 50, "level": "H2"},
    {"name": "Task 2: Database Schema", "line": 150, "level": "H2"},
    {"name": "Task 3: API Endpoints", "line": 300, "level": "H2"}
  ],
  "count": 3
}
```

## Usage Examples

### Scenario 1: Understanding a planning document

```
User: "Analyze Task 4 from the route refactoring plan"

Claude:
1. Uses markdown_tree to see document structure
2. Uses markdown_section_bounds to find Task 4 location
3. Uses markdown_read_section to read just Task 4
4. Provides comprehensive analysis using only 25% of document
```

### Scenario 2: Finding related sections

```
User: "Show me all tasks in the planning document"

Claude:
1. Uses markdown_list_sections with heading_level="H2" and pattern="Task"
2. Returns list of all Task sections with line numbers
```

### Scenario 3: Implementation workflow

```
User: "Implement Task 4"

Claude:
1. markdown_tree - Get overview
2. markdown_read_section "Executive Summary" - Context
3. markdown_read_section "Task 4" - Implementation details
4. markdown_read_section "Testing Strategy" - Validation
5. Autonomous understanding with minimal context usage
```

## Benefits

- **Zero configuration**: No manual ctags generation - works automatically
- **Reduced context usage**: Read only the sections you need (70-80% reduction)
- **Faster navigation**: Jump directly to relevant sections
- **High performance**: Sub-microsecond cache hits, ~13ms cache misses
- **Always fresh**: Automatic cache invalidation on file changes
- **Better organization**: Tree view shows document structure at a glance
- **Autonomous agent support**: Agents can discover and navigate documents independently

## Performance

- **Cache hits**: Sub-microsecond response time (~528ns)
- **Cache misses**: ~13ms (ctags execution + JSON parsing)
- **Cache validation**: Mtime-based, ~528ns overhead per query
- **Typical usage**: 90%+ cache hit rate for repeated queries

## Development

### Project structure

```
markdown-mcp/
├── cmd/
│   ├── server/
│   │   └── main.go           # MCP server entry point
│   └── test_tools/
│       └── main.go           # Tool testing utility
├── pkg/
│   ├── ctags/
│   │   ├── cache.go          # Mtime-based caching
│   │   ├── executor.go       # Ctags execution
│   │   ├── json_parser.go    # JSON ctags parsing
│   │   ├── parser.go         # Legacy parser (deprecated)
│   │   ├── tree.go           # Tree structure building
│   │   └── errors.go         # Error definitions
│   └── tools/
│       ├── errors.go         # Tool error definitions
│       ├── tree.go           # markdown_tree tool
│       ├── section_bounds.go # markdown_section_bounds tool
│       ├── read_section.go   # markdown_read_section tool
│       └── list_sections.go  # markdown_list_sections tool
├── go.mod
├── go.sum
└── README.md
```

### Running tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Linting

```bash
# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

## Troubleshooting

### Ctags not found

**Error**: `ctags not found in PATH: install universal-ctags`

**Solutions**:
1. Install Universal Ctags (see Prerequisites section)
2. If ctags is installed in a non-standard location, use the `-ctags-path` flag:
   ```bash
   markdown-nav-server -ctags-path /path/to/ctags
   ```
3. Update your MCP configuration to include the custom path:
   ```json
   {
     "mcpServers": {
       "markdown-nav": {
         "command": "markdown-nav-server",
         "args": ["-ctags-path", "/usr/local/bin/universal-ctags"]
       }
     }
   }
   ```

### Cache issues

If you experience stale data or cache-related issues, restart the MCP server to clear the cache. The cache automatically invalidates when files change based on modification time.

### Section not found

**Error**: `section not found: 'Task X'`

**Solution**:
1. Verify the section exists in the markdown file
2. Try a shorter search query (fuzzy matching is supported)
3. Use `markdown_list_sections` to see all available sections

### No entries found

**Error**: `no entries found`

**Solutions**:
1. Ensure you're using Universal Ctags (not Exuberant Ctags)
2. Verify the markdown file contains heading markers (#, ##, ###, ####)
3. Check that Universal Ctags supports markdown language

## Contributing

Contributions are welcome! Please ensure:
1. Code passes `golangci-lint` with no errors
2. All tests pass
3. New features include documentation
4. Commit messages are clear and descriptive

## License

[Your license here]

## Acknowledgments

- Built with [gomcp](https://github.com/localrivet/gomcp) by LocalRivet
- Inspired by vim-vista and ctags workflows
- Part of the markdown navigation toolkit for efficient documentation handling
