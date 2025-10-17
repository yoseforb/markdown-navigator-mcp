# Markdown Navigation MCP Server

Efficiently navigate large markdown files (2,000+ lines) without loading entire documents into context. Reduces token usage by 50-80% when working with documentation, planning files, and technical specifications.

## Quick Start

**Prerequisites**: Universal Ctags and Go 1.21+

```bash
# Install ctags
brew install universal-ctags         # macOS
sudo apt install universal-ctags     # Ubuntu/Debian
sudo dnf install universal-ctags     # Fedora

# Build and install
git clone <repo-url>
cd markdown-mcp
go build -o mdnav-server ./cmd/server
sudo cp mdnav-server /usr/local/bin/
```

**Configure Claude Code** (`~/claude.json`):
```json
{
  "mcpServers": {
    "markdown-nav": {
      "command": "mdnav-server"
    }
  }
}
```

## Features

- **Zero-configuration**: Automatic ctags execution on-demand
- **Smart caching**: Sub-microsecond responses for repeated queries
- **Auto-invalidation**: Cache updates when files change
- **Selective reading**: Load only the sections you need
- **Tree navigation**: View document structure without reading content
- **Pattern matching**: Find sections by regex patterns
- **Depth control**: Limit tree/section depth for focused views

## Tools

### markdown_tree
Display document structure as tree (ASCII or JSON format).

**Key parameters:**
- `file_path`: Path to markdown file
- `format`: "ascii" or "json" (default: "json")
- `max_depth`: Limit tree depth 1-6 (default: 2 shows H1+H2)
- `section_name_pattern`: Regex to filter sections

### markdown_section_bounds
Get line number boundaries for a specific section.

**Key parameters:**
- `file_path`: Path to markdown file
- `section_heading`: Exact heading text (without # symbols)

### markdown_read_section
Read content from a specific section.

**Key parameters:**
- `file_path`: Path to markdown file
- `section_heading`: Exact heading text (without # symbols)
- `max_subsection_levels`: Limit subsection depth (omit for all)

### markdown_list_sections
List all sections with filters.

**Key parameters:**
- `file_path`: Path to markdown file
- `max_depth`: Maximum heading level to show (default: 2)
- `section_name_pattern`: Regex to filter section names

## Usage Examples

### Finding and reading a specific task

```
User: "Review Task 4 from the planning document"

Claude uses:
1. markdown_tree to see document structure
2. markdown_section_bounds to find Task 4 location
3. markdown_read_section to read only Task 4 content

Result: Complete task analysis using only relevant section (~200 lines instead of 2000)
```

### Discovering documentation sections

```
User: "What testing strategies are documented?"

Claude uses:
1. markdown_list_sections with pattern="test" to find testing sections
2. markdown_read_section for each relevant section

Result: Comprehensive overview without loading entire document
```

For more detailed tool usage examples with real output, see [examples/EXAMPLES.md](examples/EXAMPLES.md).

## Configuration

### Custom ctags path

If ctags is not in PATH, specify location:

```json
{
  "mcpServers": {
    "markdown-nav": {
      "command": "mdnav-server",
      "args": ["-ctags-path", "/custom/path/to/ctags"]
    }
  }
}
```

## Troubleshooting

**"ctags not found in PATH"**
- Install Universal Ctags or use `-ctags-path` flag

**"section not found"**
- Use exact heading text (case-sensitive, no # symbols)
- Run `markdown_list_sections` to see available sections

**"no entries found"**
- Ensure file has markdown headings (#, ##, ###, ####)
- Verify Universal Ctags (not Exuberant) is installed

**Cache issues**
- Restart MCP server to clear cache (auto-invalidates on file changes)

## Development

For implementation details, architecture, and contributing guidelines, see [CLAUDE.md](CLAUDE.md).

**Quick development commands:**
```bash
go test ./...              # Run tests
golangci-lint run         # Lint code
go build ./cmd/server     # Build server
```

## License

[Your license here]
