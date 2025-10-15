# MCP Server for Markdown Navigation - Implementation Prompt

## Problem Statement

When working with large markdown documentation files (2,000+ lines), reading the entire file wastes significant context tokens and makes it difficult to focus on specific sections. For example, implementing "Task 4" from a 2,427-line planning document should only require reading ~600 lines (25%), not the entire file.

**Current Solution**: We built command-line tools (`md-nav`) that use ctags with enhanced fields (`--fields=+KnS`) to enable efficient markdown navigation. These tools work well but require manual bash commands.

**Reference Documentation**: See `README-md-nav.md` for complete details on the current implementation.

## What We Want to Build

**Goal**: Create an MCP (Model Context Protocol) server that exposes markdown navigation capabilities directly to Claude Code, eliminating the need for manual bash commands.

### Required MCP Server Features

#### 1. **`markdown_tree`** Tool
**Purpose**: Display hierarchical document structure (vim-vista style)

**Input**:
- `file_path` (string, required): Path to markdown file
- `tags_file` (string, optional): Path to ctags file (default: "tags")

**Output**: Tree structure showing:
```
route-domain-simplification-refactoring.md

└ Route Domain Simplification Refactoring Plan H1:1
  │ Document Status H2:3
  │ Executive Summary H2:11
  │ Task 4: Update /routes/{id} Access Control H2:550
    │ Business Justification H3:552
    │ Implementation Checklist H3:559
      │ 4.1 Use Case Update H4:561
      │ 4.2 Input Structure Update H4:716
```

**Implementation**: Use the logic from `ctags-tree.py`

#### 2. **`markdown_section_bounds`** Tool
**Purpose**: Find line number boundaries for a specific section

**Input**:
- `file_path` (string, required): Path to markdown file
- `section_query` (string, required): Section name or search query (fuzzy match)
- `tags_file` (string, optional): Path to ctags file (default: "tags")

**Output**: JSON with:
```json
{
  "section_name": "Task 4: Update /routes/{id} Access Control",
  "start_line": 550,
  "end_line": 905,
  "heading_level": "H2",
  "total_lines": 356
}
```

**Implementation**: Use the logic from `ctags-section.py`

#### 3. **`markdown_read_section`** Tool
**Purpose**: Read a specific section's content

**Input**:
- `file_path` (string, required): Path to markdown file
- `section_query` (string, required): Section name or search query
- `tags_file` (string, optional): Path to ctags file (default: "tags")
- `include_subsections` (boolean, optional, default: true): Include child sections

**Output**: Section content as string (ready for Claude to process)

**Implementation**: Combine `markdown_section_bounds` + file reading

#### 4. **`markdown_list_sections`** Tool
**Purpose**: List all top-level sections (or sections matching a pattern)

**Input**:
- `file_path` (string, required): Path to markdown file
- `heading_level` (string, optional): Filter by level (H1, H2, H3, H4)
- `pattern` (string, optional): Search pattern (regex or fuzzy)
- `tags_file` (string, optional): Path to ctags file (default: "tags")

**Output**: Array of sections:
```json
[
  {"name": "Task 1: Remove visibility:organization_controlled", "line": 31, "level": "H2"},
  {"name": "Task 4: Update /routes/{id} Access Control", "line": 550, "level": "H2"},
  {"name": "Testing Strategy", "line": 2123, "level": "H2"}
]
```

### Technical Requirements

#### MCP Server Implementation
1. **Language**: Go 1.21+ (for better performance and single-binary distribution)
2. **Framework**: Use `gomcp` package for MCP protocol (https://gomcp.dev/)
3. **Dependencies**:
   - `github.com/localrivet/gomcp` (Model Context Protocol SDK for Go)
   - Standard library for file I/O and parsing
4. **Code Reuse**: Port the parsing logic from existing scripts to Go:
   - `ctags-tree.py` → Go tree builder
   - `ctags-section.py` → Go section parser

#### MCP Server Configuration
- **Server Name**: `markdown-nav`
- **Description**: "Efficient markdown navigation using ctags for large documentation files"
- **Configuration File**: `mcp-server.json` or similar
- **Working Directory**: Should support relative paths from workspace root

#### Error Handling
- **Tags file not found**: Return clear error message suggesting user run `ctags -R --fields=+KnS`
- **Section not found**: Return helpful message with similar section names
- **File not found**: Standard file not found error
- **Invalid ctags format**: Detect and report if tags file missing required fields

### Expected Usage in Claude Code

#### Scenario 1: Understanding a Planning Document
```
User: "Analyze Task 4 from the route refactoring plan"

Claude:
1. Uses markdown_tree to see document structure
2. Uses markdown_section_bounds to find Task 4 location
3. Uses markdown_read_section to read just Task 4
4. Uses markdown_read_section to read "Testing Strategy" section
5. Provides comprehensive analysis using only 25% of document
```

#### Scenario 2: Finding Related Sections
```
User: "Show me all tasks in the planning document"

Claude:
1. Uses markdown_list_sections with heading_level="H2" and pattern="Task"
2. Returns list of all Task sections with line numbers
3. User can then ask to read specific tasks
```

#### Scenario 3: Implementation Workflow
```
User: "Implement Task 4"

Claude (using general-purpose agent):
1. markdown_tree - Get overview
2. markdown_read_section "Executive Summary" - Context
3. markdown_read_section "Task 4" - Implementation details
4. markdown_read_section "Testing Strategy" - Validation
5. markdown_read_section "Dependencies" - Prerequisites
6. Autonomous understanding with minimal context usage
```

## Implementation Plan

### Phase 1: Core MCP Server (Priority: High)
1. Set up MCP server boilerplate with Python `mcp` package
2. Implement ctags parsing utilities (extract from existing scripts)
3. Implement `markdown_tree` tool
4. Implement `markdown_section_bounds` tool
5. Implement `markdown_read_section` tool
6. Test with route-domain-simplification-refactoring.md

### Phase 2: Enhanced Features (Priority: Medium)
1. Implement `markdown_list_sections` tool
2. Add fuzzy section matching
3. Add caching for tags file parsing (performance)
4. Add scope/hierarchy information to outputs

### Phase 3: Integration & Testing (Priority: High)
1. Test with Claude Code general-purpose agent
2. Verify autonomous section discovery works
3. Document usage examples
4. Write integration tests

### Phase 4: Distribution (Priority: Low)
1. Package as installable MCP server
2. Document installation in project README
3. Consider publishing to MCP registry (optional)

## Success Criteria

The MCP server is complete when:

✅ Claude Code can navigate large markdown files efficiently
✅ Agents can autonomously discover and read relevant sections
✅ Context usage reduced by 70-80% for large documents
✅ All 4 core tools working correctly
✅ Error handling is clear and helpful
✅ Documentation complete with usage examples

## Getting Started

1. **Review existing implementation**: Study `ctags-tree.py` and `ctags-section.py` for parsing logic
2. **Review gomcp API documentation**:
   - First, try using context7 MCP tool to fetch gomcp library documentation
   - If context7 unavailable, use markitdown MCP tool to fetch docs from https://gomcp.dev/
3. **Set up development environment**: Go 1.21+, `github.com/localrivet/gomcp` package
4. **Start with `markdown_tree`**: Implement the simplest tool first
5. **Iterate and test**: Verify each tool works with real markdown files

## Reference Files

- **Current Tools**: `md-nav`, `ctags-tree.py`, `ctags-section.py`
- **Documentation**: `README-md-nav.md`
- **Test File**: `../follow-api/ai-docs/planning/backlog/route-domain-simplification-refactoring.md` (2,427 lines)
- **Tags File**: `tags` (generated with `ctags -R --fields=+KnS`)

## Notes

- The MCP server should be **stateless** - each tool call is independent
- Tags file parsing should be **efficient** - cache parsed data if possible
- Error messages should be **actionable** - tell user exactly what to do
- The server should **work out of the box** with default configuration
- Consider **backward compatibility** with standard ctags (without -S flag) if possible

---

**Objective**: Replace manual bash commands with seamless MCP integration, enabling Claude Code to navigate large markdown files as efficiently as vim-vista, with full support for autonomous agent workflows.
