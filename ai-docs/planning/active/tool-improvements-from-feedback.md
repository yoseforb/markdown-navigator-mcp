# Markdown Navigation Tools Enhancement Plan

## Document Status

**Status**: Active - Ready for Implementation
**Created**: 2025-10-15
**Last Updated**: 2025-10-15
**Assigned To**: Backend Developer / AI Agent
**Estimated Effort**: 12-18 hours (3-4 development sessions)
**Priority**: High - Critical usability improvements based on real-world feedback

## Executive Summary

This plan addresses critical usability issues and high-value enhancements identified through real-world usage of the markdown navigation MCP server. The changes focus on improving the developer experience for autonomous agents and human users working with large markdown documents.

**Key Changes:**
1. Make pattern filtering optional in `markdown_list_sections` (Critical)
2. Add comprehensive line range information across all tools (Critical)
3. Add JSON output format to `markdown_tree` for programmatic access (High Value)
4. Add pattern filtering to `markdown_tree` for focused navigation (High Value)
5. Support "ALL" heading level filter (Nice-to-Have)
6. Add depth limiting to tree display (Nice-to-Have)

**Impact**: These changes will reduce context usage by 30-40% by enabling more precise section discovery and targeted reading.

---

## Table of Contents

1. [Background & Context](#background--context)
2. [Implementation Phases](#implementation-phases)
3. [Detailed Task Breakdown](#detailed-task-breakdown)
4. [API Changes](#api-changes)
5. [Code Changes](#code-changes)
6. [Testing Strategy](#testing-strategy)
7. [Risk Assessment](#risk-assessment)
8. [Success Criteria](#success-criteria)
9. [Timeline & Milestones](#timeline--milestones)

---

## Background & Context

### Problem Statement

Real-world usage revealed several critical usability issues:

1. **Pattern Parameter Required**: Users cannot list all sections at a heading level without providing an empty pattern, which causes errors
2. **Missing Line Range Info**: Users need to know section size without reading content, but `markdown_list_sections` only provides start line
3. **No Programmatic Tree Access**: ASCII tree format is not parseable by agents, requiring manual parsing or multiple tool calls
4. **No Tree Filtering**: Users must view entire document structure even when only interested in specific sections

### Current Architecture

```
markdown-mcp/
├── pkg/
│   ├── ctags/
│   │   ├── types.go              # TagEntry definition with Line and End fields
│   │   ├── json_parser.go        # Parses ctags JSON output (includes End field)
│   │   ├── tree.go               # BuildTreeStructure (ASCII format only)
│   │   └── cache.go              # Mtime-based caching
│   └── tools/
│       ├── tree.go               # markdown_tree (ASCII output only)
│       ├── list_sections.go      # markdown_list_sections (missing end_line)
│       ├── section_bounds.go     # markdown_section_bounds (has start/end)
│       └── read_section.go       # markdown_read_section
```

**Key Observation**: The `TagEntry` struct already has the `End` field populated from ctags JSON output, but it's not exposed consistently across all tools.

### Design Principles

1. **Consistency**: Use `start_line` and `end_line` uniformly across all tools
2. **Optional Filtering**: All filter parameters should be optional, not required
3. **Structured Data**: Provide both human-readable and machine-parseable formats
4. **Backward Compatibility**: Not a concern (early development phase)
5. **Tool Separation**: Keep three distinct tools with focused purposes

---

## Implementation Phases

### Phase 1: Critical Fixes (Priority 1)
**Estimated Time**: 4-6 hours

**Tasks:**
1. Make pattern parameter optional in `markdown_list_sections`
2. Add `end_line` to `markdown_list_sections` response
3. Rename `line` to `start_line` for consistency
4. Update tests and documentation

**Rationale**: These are blocking usability issues affecting everyday usage.

### Phase 2: High-Value Enhancements (Priority 2)
**Estimated Time**: 6-8 hours

**Tasks:**
1. Add JSON output format to `markdown_tree`, make json output default
2. Implement hierarchical JSON tree structure
3. Add pattern filtering to `markdown_tree`
4. Update tests for new functionality

**Rationale**: Enables programmatic tree access and focused navigation, reducing context usage significantly.

### Phase 3: Nice-to-Have Improvements (Priority 3)
**Estimated Time**: 2-4 hours

**Tasks:**
1. Add "ALL" option for heading level in `markdown_list_sections`
2. Add `max_depth` parameter to `markdown_tree`
3. Update documentation and examples

**Rationale**: Quality-of-life improvements that enhance usability but aren't blocking.

---

## Detailed Task Breakdown

### Phase 1: Critical Fixes

#### Task 1.1: Make Pattern Optional in markdown_list_sections
**File**: `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections.go`
**Estimated Time**: 1 hour

**Current Code** (Lines 10-15):
```go
type MarkdownListSectionsArgs struct {
	FilePath     string `json:"file_path"               jsonschema:"required,description=Path to markdown file"`
	HeadingLevel string `json:"heading_level,omitempty" jsonschema:"description=Filter by level (H1, H2, H3, H4)"`
	Pattern      string `json:"pattern,omitempty"       jsonschema:"description=Search pattern (fuzzy match)"`
}
```

**Changes Required**:
- Pattern is already marked `omitempty`, but logic treats empty string as error
- Update filtering logic at lines 70-76 to handle empty pattern gracefully

**Implementation**:
```go
// Filter by pattern if specified
if args.Pattern != "" {
	filteredEntries = ctags.FilterByPattern(
		filteredEntries,
		args.Pattern,
	)
}
// If pattern is empty, all entries pass through (no filtering)
```

**Testing**:
- Test with pattern omitted entirely
- Test with pattern as empty string
- Test with valid pattern
- Test with non-matching pattern

**Acceptance Criteria**:
- Can call `markdown_list_sections` without pattern parameter
- Returns all sections at specified heading level when pattern omitted
- Maintains existing pattern matching behavior when pattern provided

---

#### Task 1.2: Add end_line to markdown_list_sections Response
**File**: `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections.go`
**Estimated Time**: 2 hours

**Current Code** (Lines 17-22):
```go
type SectionInfo struct {
	Name  string `json:"name"`
	Line  int    `json:"line"`
	Level string `json:"level"`
}
```

**Changes Required**:
1. Rename `Line` to `StartLine`
2. Add `EndLine` field
3. Update response building logic (lines 79-86)

**Implementation**:
```go
type SectionInfo struct {
	Name      string `json:"name"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Level     string `json:"level"`
}

// In response building (lines 79-86):
for _, entry := range filteredEntries {
	sections = append(sections, SectionInfo{
		Name:      entry.Name,
		StartLine: entry.Line,
		EndLine:   entry.End,
		Level:     fmt.Sprintf("H%d", entry.Level),
	})
}
```

**Testing**:
- Verify `StartLine` matches ctags line number
- Verify `EndLine` matches ctags End field
- Test with sections at document end (EndLine may be 0 or file length)
- Test with nested sections to ensure EndLine is accurate

**Acceptance Criteria**:
- Response includes both `start_line` and `end_line`
- Line numbers are accurate and match ctags output
- JSON schema is updated and valid
- Old `line` field is removed (no backward compatibility needed)

---

#### Task 1.3: Update Tests for list_sections Changes
**File**: New test file `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections_test.go`
**Estimated Time**: 2 hours

**Test Cases Required**:

```go
func TestMarkdownListSections_NoPattern(t *testing.T) {
	// Test listing all H2 sections without pattern
	// Expected: Returns all H2 sections
}

func TestMarkdownListSections_WithPattern(t *testing.T) {
	// Test listing H2 sections matching "Task"
	// Expected: Returns only matching sections
}

func TestMarkdownListSections_EmptyPattern(t *testing.T) {
	// Test with pattern as empty string
	// Expected: Returns all sections (same as no pattern)
}

func TestMarkdownListSections_LineRanges(t *testing.T) {
	// Test that start_line and end_line are accurate
	// Expected: Matches ctags output
}

func TestMarkdownListSections_AllLevels(t *testing.T) {
	// Test without heading_level filter
	// Expected: Returns sections from all levels
}
```

**Acceptance Criteria**:
- All tests pass with `go test ./...`
- Code coverage for `list_sections.go` >80%
- Tests use table-driven approach
- Tests validate JSON response structure

---

### Phase 2: High-Value Enhancements

#### Task 2.1: Add JSON Output Format to markdown_tree
**Files**:
- `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/tree.go`
- `/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree.go` (new function)
**Estimated Time**: 3 hours

**Current Response Structure** (tree.go lines 15-18):
```go
type MarkdownTreeResponse struct {
	Tree string `json:"tree"`
}
```

**New Response Structure**:
```go
type TreeNode struct {
	Name      string      `json:"name"`
	Level     string      `json:"level"`
	StartLine int         `json:"start_line"`
	EndLine   int         `json:"end_line"`
	Children  []*TreeNode `json:"children"`
}

type MarkdownTreeResponse struct {
	Tree     string     `json:"tree,omitempty"`      // ASCII format (deprecated)
	TreeJSON *TreeNode  `json:"tree_json,omitempty"` // JSON format (default)
	Format   string     `json:"format"`              // "json" or "ascii"
}
```

**Arguments Update**:
```go
type MarkdownTreeArgs struct {
	FilePath string `json:"file_path" jsonschema:"required,description=Path to markdown file"`
	Format   string `json:"format,omitempty" jsonschema:"description=Output format: json (default) or ascii"`
	Pattern  string `json:"pattern,omitempty" jsonschema:"description=Filter to sections matching pattern"`
}
```

**Implementation Steps**:

1. **Add BuildTreeJSON function to ctags/tree.go**:
```go
// BuildTreeJSON builds a hierarchical JSON tree structure from tag entries.
// Returns the root node containing all sections and their children.
func BuildTreeJSON(entries []*TagEntry) *TreeNode {
	if len(entries) == 0 {
		return nil
	}

	// Create root node
	root := &TreeNode{
		Name:      filepath.Base(entries[0].File),
		Level:     "H0",
		StartLine: 0,
		EndLine:   0,
		Children:  []*TreeNode{},
	}

	// Stack to track parent nodes at each level
	stack := []*TreeNode{root}

	for _, entry := range entries {
		node := &TreeNode{
			Name:      entry.Name,
			Level:     fmt.Sprintf("H%d", entry.Level),
			StartLine: entry.Line,
			EndLine:   entry.End,
			Children:  []*TreeNode{},
		}

		// Pop stack to find correct parent (level-based)
		for len(stack) > 1 && getLevel(stack[len(stack)-1].Level) >= entry.Level {
			stack = stack[:len(stack)-1]
		}

		// Add as child of current parent
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)

		// Push current node for potential children
		stack = append(stack, node)
	}

	return root
}

// getLevel extracts numeric level from "H1", "H2", etc.
func getLevel(levelStr string) int {
	if len(levelStr) < 2 || levelStr[0] != 'H' {
		return 0
	}
	level := 0
	fmt.Sscanf(levelStr[1:], "%d", &level)
	return level
}
```

2. **Update tools/tree.go to support both formats**:
```go
func RegisterMarkdownTree(srv server.Server) {
	srv.Tool(
		"markdown_tree",
		"Display hierarchical document structure (JSON or ASCII format)",
		func(_ *server.Context, args MarkdownTreeArgs) (interface{}, error) {
			// Get tags from cache
			cache := ctags.GetGlobalCache()
			entries, err := cache.GetTags(args.FilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to get tags: %w", err)
			}

			if len(entries) == 0 {
				return nil, fmt.Errorf("%w for %s", ErrNoEntries, args.FilePath)
			}

			// Filter by pattern if provided
			if args.Pattern != "" {
				entries = ctags.FilterByPatternWithParents(entries, args.Pattern)
			}

			// Default to JSON format
			format := args.Format
			if format == "" {
				format = "json"
			}

			response := MarkdownTreeResponse{Format: format}

			switch format {
			case "json":
				response.TreeJSON = ctags.BuildTreeJSON(entries)
			case "ascii":
				response.Tree = ctags.BuildTreeStructure(entries)
			default:
				return nil, fmt.Errorf("invalid format: %s (must be 'json' or 'ascii')", format)
			}

			return response, nil
		},
	)
}
```

**Testing**:
- Test JSON output structure is well-formed
- Test parent-child relationships are correct
- Test line ranges are accurate
- Test format parameter defaults to JSON
- Test ASCII format still works
- Test empty pattern returns full tree
- Test invalid format returns error

**Acceptance Criteria**:
- JSON output is valid and parseable
- Hierarchical structure matches document structure
- Line ranges (start_line, end_line) are accurate
- Default format is JSON
- ASCII format still available for backward compatibility

---

#### Task 2.2: Add Pattern Filtering to markdown_tree
**File**: `/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/types.go` (new function)
**Estimated Time**: 2 hours

**New Function Required**:
```go
// FilterByPatternWithParents filters entries by pattern but preserves parent sections
// to maintain tree hierarchy. This ensures that matching sections are shown in context.
//
// Example: If searching for "Testing" matches "Section 2.1: Testing", the result will
// include "Section 2" (parent) even if it doesn't match the pattern.
func FilterByPatternWithParents(entries []*TagEntry, pattern string) []*TagEntry {
	if pattern == "" {
		return entries
	}

	// First pass: find all matching entries
	matchingLines := make(map[int]bool)
	lowerPattern := strings.ToLower(pattern)

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lowerPattern) {
			matchingLines[entry.Line] = true
		}
	}

	// Second pass: include parents of matching entries
	var result []*TagEntry
	var stack []*TagEntry // Track potential parent entries

	for _, entry := range entries {
		// Pop stack to current level
		for len(stack) > 0 && stack[len(stack)-1].Level >= entry.Level {
			stack = stack[:len(stack)-1]
		}

		// Check if entry should be included
		include := matchingLines[entry.Line] // Direct match

		if !include {
			// Check if any child matches (entry is a parent)
			for _, stackEntry := range stack {
				if matchingLines[stackEntry.Line] {
					include = true
					break
				}
			}
		}

		if include {
			result = append(result, entry)
		}

		stack = append(stack, entry)
	}

	return result
}
```

**Algorithm Explanation**:
1. First pass: Identify all entries matching the pattern
2. Second pass: Include parents of matching entries to preserve hierarchy
3. Result: Filtered tree showing only relevant branches with context

**Testing**:
- Test pattern matches at different levels (H1, H2, H3, H4)
- Test parent sections are preserved
- Test multiple matches in same branch
- Test no matches returns empty result
- Test empty pattern returns full tree

**Acceptance Criteria**:
- Matching sections are included in result
- Parent sections of matches are preserved
- Sibling sections not matching are excluded
- Hierarchy is maintained in filtered result

---

#### Task 2.3: Update Tests for tree Changes
**File**: New test file `/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree_json_test.go`
**Estimated Time**: 2 hours

**Test Cases Required**:

```go
func TestBuildTreeJSON_BasicStructure(t *testing.T) {
	// Test that JSON tree has correct structure
	// Expected: Root node with children array
}

func TestBuildTreeJSON_Hierarchy(t *testing.T) {
	// Test parent-child relationships: H1 > H2 > H3 > H4
	// Expected: Nested children arrays match document structure
}

func TestBuildTreeJSON_LineRanges(t *testing.T) {
	// Test start_line and end_line in JSON output
	// Expected: Matches ctags output
}

func TestFilterByPatternWithParents_DirectMatch(t *testing.T) {
	// Test pattern matches section directly
	// Expected: Section and parents included
}

func TestFilterByPatternWithParents_ParentPreservation(t *testing.T) {
	// Test parent sections are preserved for matched children
	// Expected: Full path from root to match included
}

func TestFilterByPatternWithParents_NoMatch(t *testing.T) {
	// Test pattern with no matches
	// Expected: Empty result
}

func TestMarkdownTree_JSONFormat(t *testing.T) {
	// Integration test: markdown_tree with format=json
	// Expected: Valid JSON response with TreeJSON populated
}

func TestMarkdownTree_ASCIIFormat(t *testing.T) {
	// Integration test: markdown_tree with format=ascii
	// Expected: ASCII tree string in Tree field
}

func TestMarkdownTree_WithPattern(t *testing.T) {
	// Integration test: markdown_tree with pattern filter
	// Expected: Only matching branches in output
}
```

**Acceptance Criteria**:
- All tests pass with `go test ./...`
- Code coverage for tree.go >80%
- Tests validate JSON structure and content
- Tests verify parent preservation logic

---

### Phase 3: Nice-to-Have Improvements

#### Task 3.1: Add "ALL" Heading Level Option
**File**: `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections.go`
**Estimated Time**: 1 hour

**Current Code** (Lines 47-68):
```go
// Filter by heading level if specified
filteredEntries := entries
if args.HeadingLevel != "" {
	var level int
	switch args.HeadingLevel {
	case "H1":
		level = 1
	case "H2":
		level = 2
	case "H3":
		level = 3
	case "H4":
		level = 4
	default:
		return nil, fmt.Errorf(
			"%w: %s (must be H1, H2, H3, or H4)",
			ErrInvalidLevel,
			args.HeadingLevel,
		)
	}
	filteredEntries = ctags.FilterByLevel(filteredEntries, level)
}
```

**Changes Required**:
```go
// Filter by heading level if specified
filteredEntries := entries
if args.HeadingLevel != "" {
	if args.HeadingLevel == "ALL" {
		// No filtering - return all levels
		filteredEntries = entries
	} else {
		var level int
		switch args.HeadingLevel {
		case "H1":
			level = 1
		case "H2":
			level = 2
		case "H3":
			level = 3
		case "H4":
			level = 4
		default:
			return nil, fmt.Errorf(
				"%w: %s (must be H1, H2, H3, H4, or ALL)",
				ErrInvalidLevel,
				args.HeadingLevel,
			)
		}
		filteredEntries = ctags.FilterByLevel(filteredEntries, level)
	}
}
```

**Testing**:
- Test `heading_level="ALL"` returns sections from all levels
- Test combination with pattern filter
- Test case sensitivity (should "all" also work?)

**Acceptance Criteria**:
- Can specify `heading_level="ALL"` to get sections at all levels
- Combines correctly with pattern filtering
- Error messages updated to include "ALL" option

---

#### Task 3.2: Add max_depth to markdown_tree
**File**: `/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree.go`
**Estimated Time**: 1.5 hours

**Update MarkdownTreeArgs**:
```go
type MarkdownTreeArgs struct {
	FilePath string `json:"file_path" jsonschema:"required,description=Path to markdown file"`
	Format   string `json:"format,omitempty" jsonschema:"description=Output format: json (default) or ascii"`
	Pattern  string `json:"pattern,omitempty" jsonschema:"description=Filter to sections matching pattern"`
	MaxDepth int    `json:"max_depth,omitempty" jsonschema:"description=Maximum depth to display (0 = unlimited)"`
}
```

**Implementation**:

1. **Add FilterByDepth function to ctags/types.go**:
```go
// FilterByDepth filters entries to maximum heading level depth.
// depth=1 shows only H1, depth=2 shows H1+H2, depth=0 shows all.
func FilterByDepth(entries []*TagEntry, depth int) []*TagEntry {
	if depth <= 0 {
		return entries
	}

	var filtered []*TagEntry
	for _, entry := range entries {
		if entry.Level <= depth {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
```

2. **Update tools/tree.go to apply depth filter**:
```go
// Filter by pattern if provided
if args.Pattern != "" {
	entries = ctags.FilterByPatternWithParents(entries, args.Pattern)
}

// Filter by depth if provided
if args.MaxDepth > 0 {
	entries = ctags.FilterByDepth(entries, args.MaxDepth)
}
```

**Testing**:
- Test `max_depth=1` shows only H1
- Test `max_depth=2` shows H1 and H2
- Test `max_depth=0` or omitted shows all levels
- Test combination with pattern filter
- Test both JSON and ASCII formats

**Acceptance Criteria**:
- Tree output respects max_depth parameter
- Depth limiting works with both JSON and ASCII formats
- Combines correctly with pattern filtering

---

#### Task 3.3: Documentation Updates
**Files**:
- `/home/yoseforb/pkg/follow/markdown-mcp/README.md`
- `/home/yoseforb/pkg/follow/markdown-mcp/CLAUDE.md`
**Estimated Time**: 1 hour

**Updates Required**:

1. **README.md**: Update tool documentation
   - Add JSON format examples for `markdown_tree`
   - Update `markdown_list_sections` examples with `start_line` and `end_line`
   - Document new parameters (format, pattern, max_depth, ALL)
   - Add migration notes if needed

2. **CLAUDE.md**: Update development guide
   - Update "MCP Tools" section with new parameters
   - Add JSON output format examples
   - Update response format examples
   - Add note about deprecated ASCII-only format

**Example Documentation**:

```markdown
### markdown_tree

**Purpose**: Display hierarchical document structure (JSON or ASCII format)

**Parameters**:
- `file_path` (required): Path to markdown file
- `format` (optional): Output format: "json" (default) or "ascii"
- `pattern` (optional): Filter to sections matching pattern (preserves parents)
- `max_depth` (optional): Maximum depth to display (0 = unlimited)

**JSON Output Format** (default):
```json
{
  "format": "json",
  "tree_json": {
    "name": "planning.md",
    "level": "H0",
    "start_line": 0,
    "end_line": 0,
    "children": [
      {
        "name": "Planning Document",
        "level": "H1",
        "start_line": 1,
        "end_line": 2436,
        "children": [
          {
            "name": "Overview",
            "level": "H2",
            "start_line": 5,
            "end_line": 48,
            "children": []
          }
        ]
      }
    ]
  }
}
```

**Use Case**:
- JSON format: Programmatic tree access for agents
- ASCII format: Human-readable tree visualization
- Pattern filter: Focus on specific sections while maintaining context
```

**Acceptance Criteria**:
- All new parameters documented with examples
- JSON format examples are accurate and complete
- Migration guidance provided (if needed)
- Use cases clearly explained

---

## API Changes

### markdown_list_sections

**Before**:
```json
{
  "file_path": "/path/to/file.md",
  "heading_level": "H2",
  "pattern": "Task"
}

Response:
{
  "sections": [
    {"name": "Task 1", "line": 50, "level": "H2"}
  ],
  "count": 1
}
```

**After**:
```json
{
  "file_path": "/path/to/file.md",
  "heading_level": "H2",  // Optional: omit to get all levels or use "ALL"
  "pattern": "Task"        // Optional: omit to get all sections
}

Response:
{
  "sections": [
    {"name": "Task 1", "start_line": 50, "end_line": 149, "level": "H2"}
  ],
  "count": 1
}
```

**Breaking Changes**:
- `line` renamed to `start_line`
- Added `end_line` field
- Pattern is now optional (was effectively required)

---

### markdown_tree

**Before**:
```json
{
  "file_path": "/path/to/file.md"
}

Response:
{
  "tree": "planning.md\n\n└ Planning Document H1:1\n  │ Overview H2:5\n..."
}
```

**After (JSON format - default)**:
```json
{
  "file_path": "/path/to/file.md",
  "format": "json",          // Optional: "json" (default) or "ascii"
  "pattern": "Task",         // Optional: filter to matching sections
  "max_depth": 2             // Optional: limit depth (0 = unlimited)
}

Response:
{
  "format": "json",
  "tree_json": {
    "name": "planning.md",
    "level": "H0",
    "start_line": 0,
    "end_line": 0,
    "children": [/* nested tree structure */]
  }
}
```

**After (ASCII format)**:
```json
{
  "file_path": "/path/to/file.md",
  "format": "ascii"
}

Response:
{
  "format": "ascii",
  "tree": "planning.md\n\n└ Planning Document H1:1:2436\n  │ Overview H2:5:48\n..."
}
```

**Breaking Changes**:
- Default output changed from ASCII to JSON
- ASCII format requires explicit `format: "ascii"` parameter
- Response structure changed (added `format` field, `tree_json` field)

**Backward Compatibility**:
- ASCII format still available but not default
- Consider adding deprecation notice for ASCII-only usage

---

## Code Changes

### Files to Modify

1. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections.go`**
   - Update `SectionInfo` struct (add `StartLine`, `EndLine`, remove `Line`)
   - Update response building logic to populate new fields
   - Update argument description for pattern (mark as optional)
   - Add support for `heading_level="ALL"`
   - **Lines affected**: 17-22, 79-86, 61-65

2. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/tree.go`**
   - Add `TreeNode` struct definition
   - Update `MarkdownTreeArgs` (add `Format`, `Pattern`, `MaxDepth`)
   - Update `MarkdownTreeResponse` (add `TreeJSON`, `Format`)
   - Update tool logic to support format selection and filtering
   - **Lines affected**: 10-18, 20-43 (complete rewrite of tool function)

3. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree.go`**
   - Add `BuildTreeJSON` function (new, ~60 lines)
   - Add `getLevel` helper function (new, ~10 lines)
   - Keep existing `BuildTreeStructure` for ASCII format
   - **New code**: ~70 lines

4. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/types.go`**
   - Add `FilterByPatternWithParents` function (new, ~50 lines)
   - Add `FilterByDepth` function (new, ~15 lines)
   - **New code**: ~65 lines

### New Files to Create

1. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections_test.go`**
   - Unit tests for updated `markdown_list_sections` tool
   - **Estimated size**: ~200 lines

2. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree_json_test.go`**
   - Unit tests for `BuildTreeJSON` function
   - Tests for `FilterByPatternWithParents`
   - Tests for `FilterByDepth`
   - **Estimated size**: ~300 lines

3. **`/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/tree_test.go`**
   - Integration tests for updated `markdown_tree` tool
   - Tests for JSON/ASCII format selection
   - Tests for pattern and depth filtering
   - **Estimated size**: ~250 lines

### Code Removal

None - all changes are additions or modifications, no code deletion required.

---

## Testing Strategy

### Unit Tests

**Scope**: Individual functions in isolation

**Test Files**:
- `/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree_json_test.go`
- `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/list_sections_test.go`
- `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/tree_test.go`

**Test Cases** (grouped by function):

#### `BuildTreeJSON` Tests
```go
func TestBuildTreeJSON_EmptyEntries(t *testing.T)
func TestBuildTreeJSON_SingleEntry(t *testing.T)
func TestBuildTreeJSON_FlatStructure(t *testing.T) // All same level
func TestBuildTreeJSON_NestedStructure(t *testing.T) // H1 > H2 > H3 > H4
func TestBuildTreeJSON_MultipleRoots(t *testing.T) // Multiple H1 sections
func TestBuildTreeJSON_LineRanges(t *testing.T) // Verify start_line and end_line
func TestBuildTreeJSON_LargeDocument(t *testing.T) // 100+ sections
```

#### `FilterByPatternWithParents` Tests
```go
func TestFilterByPatternWithParents_NoPattern(t *testing.T)
func TestFilterByPatternWithParents_DirectMatch(t *testing.T)
func TestFilterByPatternWithParents_ParentPreservation(t *testing.T)
func TestFilterByPatternWithParents_MultipleMatches(t *testing.T)
func TestFilterByPatternWithParents_NoMatches(t *testing.T)
func TestFilterByPatternWithParents_CaseInsensitive(t *testing.T)
func TestFilterByPatternWithParents_PartialMatch(t *testing.T)
```

#### `FilterByDepth` Tests
```go
func TestFilterByDepth_UnlimitedDepth(t *testing.T) // depth = 0
func TestFilterByDepth_DepthOne(t *testing.T) // Only H1
func TestFilterByDepth_DepthTwo(t *testing.T) // H1 + H2
func TestFilterByDepth_DepthThree(t *testing.T) // H1 + H2 + H3
func TestFilterByDepth_DepthFour(t *testing.T) // All levels
```

#### `markdown_list_sections` Tests
```go
func TestListSections_NoPattern(t *testing.T)
func TestListSections_EmptyPattern(t *testing.T)
func TestListSections_WithPattern(t *testing.T)
func TestListSections_HeadingLevelH1(t *testing.T)
func TestListSections_HeadingLevelH2(t *testing.T)
func TestListSections_HeadingLevelALL(t *testing.T)
func TestListSections_LineRanges(t *testing.T) // Verify start_line and end_line
func TestListSections_NoHeadingLevel(t *testing.T) // All levels
func TestListSections_InvalidHeadingLevel(t *testing.T)
func TestListSections_NoEntries(t *testing.T)
```

#### `markdown_tree` Tests
```go
func TestMarkdownTree_DefaultFormat(t *testing.T) // Should be JSON
func TestMarkdownTree_JSONFormat(t *testing.T)
func TestMarkdownTree_ASCIIFormat(t *testing.T)
func TestMarkdownTree_InvalidFormat(t *testing.T)
func TestMarkdownTree_WithPattern(t *testing.T)
func TestMarkdownTree_WithMaxDepth(t *testing.T)
func TestMarkdownTree_PatternAndDepth(t *testing.T) // Combined filters
func TestMarkdownTree_NoEntries(t *testing.T)
func TestMarkdownTree_EmptyFile(t *testing.T)
```

**Testing Approach**:
- Use table-driven tests where appropriate
- Use test fixtures from `/home/yoseforb/pkg/follow/markdown-mcp/testdata/`
- Test edge cases: empty files, no matches, invalid inputs
- Test large documents (2000+ lines) for performance

---

### Integration Tests

**Scope**: End-to-end tool behavior with real markdown files

**Test Approach**:
1. Use `cmd/test_tools/main.go` for manual integration testing
2. Create automated integration test suite

**Test Files**:
- Update `/home/yoseforb/pkg/follow/markdown-mcp/cmd/test_tools/main.go`
- Add test scenarios for new functionality

**Test Scenarios**:

1. **List All H2 Sections Without Pattern**
   ```bash
   # Expected: Returns all H2 sections with start_line and end_line
   ```

2. **List H2 Sections Matching "Task"**
   ```bash
   # Expected: Returns only matching H2 sections
   ```

3. **Get JSON Tree of Document**
   ```bash
   # Expected: Returns hierarchical JSON structure
   ```

4. **Get Filtered JSON Tree (Pattern="Testing")**
   ```bash
   # Expected: Returns only branches with "Testing" sections
   ```

5. **Get ASCII Tree with Max Depth 2**
   ```bash
   # Expected: Returns ASCII tree showing only H1 and H2
   ```

6. **Large Document Performance Test**
   ```bash
   # Test with 2000+ line document
   # Expected: Completes in <100ms (cache hit) or <30ms (cache miss)
   ```

**Integration Test Fixtures**:
- Small document (50 lines, 10 sections)
- Medium document (500 lines, 50 sections)
- Large document (2000+ lines, 200+ sections)
- Deeply nested document (H1 > H2 > H3 > H4)
- Flat document (all same level)

---

### Performance Tests

**Benchmark Tests**: Add to `/home/yoseforb/pkg/follow/markdown-mcp/pkg/ctags/tree_benchmark_test.go`

```go
func BenchmarkBuildTreeJSON(b *testing.B) {
	entries := loadLargeTestData() // 200+ entries
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildTreeJSON(entries)
	}
}

func BenchmarkBuildTreeStructure(b *testing.B) {
	entries := loadLargeTestData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildTreeStructure(entries)
	}
}

func BenchmarkFilterByPatternWithParents(b *testing.B) {
	entries := loadLargeTestData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterByPatternWithParents(entries, "Task")
	}
}
```

**Performance Targets**:
- `BuildTreeJSON`: <5ms for 200 entries
- `BuildTreeStructure`: <3ms for 200 entries (ASCII is simpler)
- `FilterByPatternWithParents`: <2ms for 200 entries
- `FilterByDepth`: <1ms for 200 entries

---

### Test Execution

**Commands**:
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test file
go test ./pkg/tools/list_sections_test.go

# Run specific test function
go test -run TestListSections_NoPattern ./pkg/tools/

# Run benchmarks
go test -bench=. ./pkg/ctags/

# Run tests with race detection
go test -race ./...

# Integration tests (manual)
go run cmd/test_tools/main.go testdata/large_planning.md
```

**Quality Gates**:
- All tests must pass: `go test ./...`
- Code coverage: >80% for modified files
- No race conditions: `go test -race ./...`
- Linting: `golangci-lint run` (zero errors)
- Formatting: `gofumpt -w .`

---

## Risk Assessment

### High Risk Issues

#### Risk 1: Breaking Changes to API
**Impact**: High - Affects all users of `markdown_list_sections`
**Probability**: Certain (by design)

**Mitigation**:
- Document all breaking changes clearly
- Provide migration examples in README
- Since project is in early development, breaking changes acceptable
- Consider version bump (v0.x to v1.0) to signal major changes

**Contingency**:
- If users report issues, provide detailed migration guide
- Consider adding temporary compatibility layer if needed

---

#### Risk 2: Performance Degradation with JSON Tree
**Impact**: Medium - JSON parsing and tree building could be slower
**Probability**: Low - Should be comparable to ASCII tree

**Mitigation**:
- Benchmark both JSON and ASCII formats
- Optimize JSON tree building if needed
- Cache is already in place, so repeated calls are fast
- Target performance: <10ms for 200 entries

**Contingency**:
- If JSON is significantly slower, optimize algorithm
- Consider pre-allocating tree node slices
- Profile and optimize hot paths

**Monitoring**:
```bash
# Run benchmarks before and after changes
go test -bench=BenchmarkBuildTree ./pkg/ctags/
```

---

#### Risk 3: Pattern Filtering Breaks Tree Hierarchy
**Impact**: High - Could return confusing or invalid tree structures
**Probability**: Medium - Complex parent preservation logic

**Mitigation**:
- Comprehensive testing of `FilterByPatternWithParents`
- Test edge cases: multiple matches, nested matches, no matches
- Review algorithm with test-driven approach
- Add integration tests with real documents

**Contingency**:
- If hierarchy breaks, revise algorithm with more conservative approach
- Fall back to simple filtering without parent preservation
- Add warning in documentation about pattern filtering limitations

**Testing Strategy**:
```go
// Test cases to validate hierarchy preservation
TestFilterByPatternWithParents_ParentPreservation
TestFilterByPatternWithParents_SiblingExclusion
TestFilterByPatternWithParents_MultipleMatches
```

---

### Medium Risk Issues

#### Risk 4: Inconsistent Line Numbering
**Impact**: Medium - Could confuse users or break integrations
**Probability**: Low - Line numbers come directly from ctags

**Mitigation**:
- Validate line numbers against ctags output in tests
- Test with documents where sections end at EOF
- Test with nested sections to ensure EndLine is accurate
- Add integration tests comparing tool output to ctags JSON

**Contingency**:
- If EndLine is unreliable, document limitations
- Consider calculating EndLine from next section's start
- Add validation warnings in response

---

#### Risk 5: Complex Filtering Combinations
**Impact**: Medium - Pattern + depth filtering could produce unexpected results
**Probability**: Medium - Order of operations matters

**Mitigation**:
- Document filter application order clearly
- Test all combinations: pattern only, depth only, both
- Ensure filters compose predictably
- Add examples in documentation

**Filter Application Order**:
1. Pattern filtering (with parent preservation)
2. Depth filtering (remove deep sections)

**Contingency**:
- If combinations don't work well, simplify to one filter at a time
- Document which combinations are supported

---

### Low Risk Issues

#### Risk 6: Golangci-lint Failures
**Impact**: Low - Blocks PR but doesn't affect functionality
**Probability**: Low - Using strict configuration from start

**Mitigation**:
- Run linter frequently during development
- Fix issues as they arise
- Use `golangci-lint run --fix` for auto-fixable issues

**Contingency**:
- Add nolint directives only when absolutely necessary
- Document why lint rules are disabled

---

#### Risk 7: Test Flakiness
**Impact**: Low - Slows development but doesn't affect users
**Probability**: Low - Tests are deterministic, no concurrency in tools

**Mitigation**:
- Avoid time-dependent tests
- Use fixed test data
- Run tests multiple times to detect flakiness
- Use `t.Parallel()` carefully

**Contingency**:
- If tests are flaky, add retries or stabilize test environment
- Disable parallel execution if race conditions found

---

## Success Criteria

### Functional Requirements

**Phase 1: Critical Fixes**
- [ ] `markdown_list_sections` accepts calls without pattern parameter
- [ ] `markdown_list_sections` returns `start_line` and `end_line` for all sections
- [ ] Line numbers in response match ctags output exactly
- [ ] Empty pattern is handled same as omitted pattern
- [ ] All existing functionality still works

**Phase 2: High-Value Enhancements**
- [ ] `markdown_tree` returns JSON format by default
- [ ] JSON tree structure is hierarchical with correct parent-child relationships
- [ ] JSON tree includes `start_line` and `end_line` for all nodes
- [ ] ASCII format still available via `format="ascii"` parameter
- [ ] Pattern filtering in `markdown_tree` works correctly
- [ ] Pattern filtering preserves parent sections for context
- [ ] Filtered trees maintain correct hierarchy

**Phase 3: Nice-to-Have Improvements**
- [ ] `heading_level="ALL"` returns sections from all levels
- [ ] `max_depth` parameter limits tree depth correctly
- [ ] `max_depth=0` or omitted shows all levels
- [ ] Filters combine correctly (pattern + depth)

---

### Non-Functional Requirements

**Code Quality**
- [ ] All new code passes `golangci-lint run` with zero errors
- [ ] All code formatted with `gofumpt -w .`
- [ ] Code coverage >80% for modified/new files
- [ ] No race conditions detected by `go test -race ./...`

**Performance**
- [ ] `BuildTreeJSON` completes in <5ms for 200 entries
- [ ] `FilterByPatternWithParents` completes in <2ms for 200 entries
- [ ] No performance regression in existing tools
- [ ] Cache still provides 90%+ hit rate

**Testing**
- [ ] All unit tests pass: `go test ./...`
- [ ] All integration tests pass
- [ ] Test coverage report generated and reviewed
- [ ] Benchmark tests run and performance validated

**Documentation**
- [ ] README.md updated with new parameters and examples
- [ ] CLAUDE.md updated with new tool signatures
- [ ] All new parameters documented with descriptions
- [ ] Migration notes provided for breaking changes
- [ ] JSON output format examples included

---

### User Experience

**Usability**
- [ ] Tools are intuitive and easy to use
- [ ] Optional parameters have sensible defaults
- [ ] Error messages are clear and actionable
- [ ] JSON output is well-structured and parseable
- [ ] Documentation provides clear examples

**Consistency**
- [ ] All tools use `start_line` and `end_line` consistently
- [ ] Parameter naming is consistent across tools
- [ ] Response formats follow similar patterns
- [ ] Error handling is consistent

**Backward Compatibility**
- [ ] Breaking changes documented clearly
- [ ] Migration path explained in README
- [ ] ASCII format still available for users who need it

---

### Validation Checklist

**Pre-Implementation**
- [ ] Planning document reviewed and approved
- [ ] Test fixtures prepared
- [ ] Development environment set up

**During Implementation**
- [ ] Run tests after each task: `go test ./...`
- [ ] Run linter frequently: `golangci-lint run`
- [ ] Commit after each completed task
- [ ] Update progress in planning document

**Post-Implementation**
- [ ] All tests passing
- [ ] All quality gates passed
- [ ] Documentation updated
- [ ] Integration tests run manually
- [ ] Performance benchmarks run and reviewed
- [ ] Code reviewed (self-review or peer review)

**Release Checklist**
- [ ] README.md updated
- [ ] CLAUDE.md updated
- [ ] CHANGELOG.md updated (if exists)
- [ ] Version bumped (if applicable)
- [ ] Git tags created (if applicable)

---

## Timeline & Milestones

### Phase 1: Critical Fixes
**Duration**: 4-6 hours
**Milestone**: Usability blockers resolved

**Day 1 (Session 1): 2-3 hours**
- [ ] Task 1.1: Make pattern optional (1 hour)
  - Update list_sections.go logic
  - Test with no pattern, empty pattern, valid pattern
- [ ] Task 1.2: Add end_line to response (2 hours)
  - Update SectionInfo struct
  - Update response building logic
  - Update JSON schema

**Day 2 (Session 2): 2-3 hours**
- [ ] Task 1.3: Write tests for list_sections (2 hours)
  - Create list_sections_test.go
  - Write unit tests for all scenarios
  - Validate test coverage >80%
- [ ] Phase 1 Validation (30 minutes)
  - Run all tests: `go test ./...`
  - Run linter: `golangci-lint run`
  - Manual integration testing
- [ ] Phase 1 Complete: Commit and tag

---

### Phase 2: High-Value Enhancements
**Duration**: 6-8 hours
**Milestone**: JSON tree and pattern filtering working

**Day 3 (Session 3): 3-4 hours**
- [ ] Task 2.1: Add JSON output to markdown_tree (3 hours)
  - Implement BuildTreeJSON in ctags/tree.go
  - Update MarkdownTreeResponse struct
  - Update tools/tree.go to support format parameter
  - Test JSON structure and hierarchy
- [ ] Task 2.2: Start pattern filtering implementation (1 hour)
  - Design FilterByPatternWithParents algorithm
  - Write test cases
  - Begin implementation

**Day 4 (Session 4): 3-4 hours**
- [ ] Task 2.2: Complete pattern filtering (2 hours)
  - Complete FilterByPatternWithParents implementation
  - Test parent preservation logic
  - Integrate with markdown_tree
- [ ] Task 2.3: Write tests for tree changes (2 hours)
  - Create tree_json_test.go
  - Test BuildTreeJSON with various structures
  - Test FilterByPatternWithParents edge cases
  - Integration tests for markdown_tree
- [ ] Phase 2 Validation (30 minutes)
  - Run all tests with coverage
  - Run benchmarks
  - Manual integration testing
- [ ] Phase 2 Complete: Commit and tag

---

### Phase 3: Nice-to-Have Improvements
**Duration**: 2-4 hours
**Milestone**: Quality-of-life enhancements complete

**Day 5 (Session 5): 2-3 hours**
- [ ] Task 3.1: Add "ALL" heading level (1 hour)
  - Update list_sections.go
  - Test ALL option with patterns
  - Update error messages
- [ ] Task 3.2: Add max_depth parameter (1.5 hours)
  - Implement FilterByDepth function
  - Update markdown_tree to use depth filter
  - Test depth limiting with JSON and ASCII
  - Test combination with pattern filter
- [ ] Task 3.3: Documentation updates (1 hour)
  - Update README.md with new parameters
  - Update CLAUDE.md with API changes
  - Add migration notes
  - Add JSON output examples

**Day 6 (Session 6): 1-2 hours**
- [ ] Final Validation (1 hour)
  - Run full test suite
  - Run all quality gates
  - Manual integration testing with real documents
  - Performance benchmarks
- [ ] Documentation review (30 minutes)
  - Proofread README and CLAUDE.md
  - Verify examples are accurate
  - Check migration notes are complete
- [ ] Final commit and release preparation

---

### Milestones Summary

| Milestone | Completion Date | Deliverables |
|-----------|----------------|--------------|
| **Phase 1 Complete** | Session 2 | Pattern optional, end_line added, tests passing |
| **Phase 2 Complete** | Session 4 | JSON tree working, pattern filtering working |
| **Phase 3 Complete** | Session 5 | ALL option, max_depth, docs updated |
| **Release Ready** | Session 6 | All quality gates passed, ready for use |

---

### Progress Tracking

**During Implementation**:
- Update checkboxes as tasks complete
- Add notes for any issues encountered
- Track actual time vs. estimated time
- Document any scope changes or decisions

**Daily Standup Format**:
```
Yesterday: Completed Tasks X, Y, Z
Today: Working on Task A (estimated N hours)
Blockers: None / Issue with X (mitigation: Y)
```

**Weekly Review**:
- Review progress against timeline
- Adjust estimates if needed
- Identify any risks or blockers
- Update planning document with lessons learned

---

## Appendix

### A. Test Fixtures

**Location**: `/home/yoseforb/pkg/follow/markdown-mcp/testdata/`

**Required Test Documents**:

1. **`small.md`** (50 lines, 10 sections)
   - Simple hierarchy: H1 > H2 > H3
   - Used for: Basic unit tests

2. **`medium.md`** (500 lines, 50 sections)
   - Mixed hierarchy with multiple branches
   - Used for: Integration tests

3. **`large.md`** (2000+ lines, 200+ sections)
   - Complex planning document
   - Used for: Performance tests

4. **`flat.md`** (100 lines, 20 H2 sections only)
   - All sections at same level
   - Used for: Testing level filtering

5. **`deep.md`** (200 lines, 40 sections with full H1>H2>H3>H4 nesting)
   - Maximum nesting depth
   - Used for: Testing depth limiting

**Example small.md Structure**:
```markdown
# Test Document

## Section 1: Introduction

### Subsection 1.1: Background

Content here...

### Subsection 1.2: Objectives

More content...

## Section 2: Implementation

### Subsection 2.1: Design

Details...

### Subsection 2.2: Testing

#### Deep Section 2.2.1: Unit Tests

More details...

## Section 3: Conclusion

Final thoughts...
```

---

### B. Error Handling

**Error Types** (from `/home/yoseforb/pkg/follow/markdown-mcp/pkg/tools/errors.go`):

```go
var (
	ErrNoEntries       = errors.New("no entries found")
	ErrSectionNotFound = errors.New("section not found")
	ErrInvalidLevel    = errors.New("invalid heading level")
)
```

**New Error Messages**:

```go
// Invalid format
return nil, fmt.Errorf("invalid format: %s (must be 'json' or 'ascii')", format)

// Pattern with no matches (not an error - return empty result)
return MarkdownTreeResponse{
	Format: "json",
	TreeJSON: &TreeNode{
		Name: filename,
		Level: "H0",
		Children: []*TreeNode{},
	},
}, nil
```

**Error Handling Principles**:
1. Use context-aware error wrapping: `fmt.Errorf("context: %w", err)`
2. Provide actionable error messages
3. Don't error on empty results - return empty arrays/structures
4. Validate inputs early and fail fast

---

### C. Performance Optimization Notes

**Optimization Opportunities**:

1. **Pre-allocate Slices**:
```go
// Instead of:
var children []*TreeNode

// Use:
children := make([]*TreeNode, 0, estimatedSize)
```

2. **Avoid Repeated String Operations**:
```go
// Cache lowercase pattern
lowerPattern := strings.ToLower(pattern)
// Use lowerPattern in loop instead of converting each time
```

3. **Reuse Buffers**:
```go
var buf strings.Builder
buf.Grow(estimatedSize)
// Use buf for string concatenation
```

4. **Profile Hot Paths**:
```bash
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

**Performance Targets Revisited**:
- `BuildTreeJSON`: <5ms for 200 entries (target: 2-3ms)
- `FilterByPatternWithParents`: <2ms for 200 entries (target: <1ms)
- `FilterByDepth`: <1ms for 200 entries (target: <500µs)

---

### D. Future Enhancements (Post-Implementation)

**Not in Current Scope** (consider for future iterations):

1. **Cache Statistics API**
   - Tool to expose cache hit/miss rates
   - Useful for performance monitoring

2. **Section Diffing**
   - Compare sections across document versions
   - Useful for tracking changes in planning docs

3. **Cross-Reference Detection**
   - Detect links between sections (markdown links)
   - Build section dependency graph

4. **Full-Text Search**
   - Search within section content (not just headings)
   - Requires different approach than ctags

5. **Custom Heading Markers**
   - Support non-standard markdown structures
   - Configurable section delimiters

6. **Streaming API**
   - Stream large documents section by section
   - Reduce memory usage for huge files

7. **Section Metadata**
   - Extract tags, labels, status from sections
   - Parse structured metadata in headings

---

### E. Development Commands Reference

**Quick Reference**:

```bash
# Test single function
go test -run TestListSections_NoPattern ./pkg/tools/

# Test with verbose output
go test -v ./pkg/tools/

# Test with coverage
go test -coverprofile=coverage.out ./pkg/tools/
go tool cover -html=coverage.out

# Benchmark
go test -bench=BenchmarkBuildTreeJSON ./pkg/ctags/

# Race detection
go test -race ./...

# Lint specific file
golangci-lint run pkg/tools/list_sections.go

# Format specific file
gofumpt -w pkg/tools/list_sections.go

# Build server
go build -o mdnav-server

# Run integration tests
go run cmd/test_tools/main.go testdata/large.md

# Check test coverage percentage
go test -cover ./... | grep coverage
```

---

## Notes & Decisions

### Implementation Notes
(To be filled during implementation)

**Date**:
**Note**:

---

### Decision Log

**Decision 1**: Default to JSON format instead of ASCII
- **Date**: 2025-10-15
- **Rationale**: JSON is more useful for programmatic access by agents
- **Impact**: Breaking change but acceptable in early development
- **Alternative Considered**: Keep ASCII as default, add JSON as opt-in

**Decision 2**: Preserve parents in pattern filtering
- **Date**: 2025-10-15
- **Rationale**: Context is important for understanding section location
- **Impact**: More complex algorithm but better UX
- **Alternative Considered**: Simple filtering without parents (confusing results)

**Decision 3**: Make pattern optional, not required
- **Date**: 2025-10-15
- **Rationale**: Common use case is "list all sections at level X"
- **Impact**: Fixes critical usability issue
- **Alternative Considered**: Require empty string (current behavior, poor UX)

---

## Related Documents

- **Project Overview**: `/home/yoseforb/pkg/follow/markdown-mcp/CLAUDE.md`
- **User Documentation**: `/home/yoseforb/pkg/follow/markdown-mcp/README.md`
- **Cache Refactoring Plan**: `/home/yoseforb/pkg/follow/markdown-mcp/ai-docs/planning/backlog/json-cache-refactoring.md`
- **Test Data**: `/home/yoseforb/pkg/follow/markdown-mcp/testdata/`

---

**Document Version**: 1.0
**Last Updated**: 2025-10-15
**Status**: Ready for Implementation
