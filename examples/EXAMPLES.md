# Markdown Navigation MCP Server - Usage Examples

This file demonstrates all 4 tools using `sample-project.md` as the example document. The example file is a realistic project planning document with 261 lines, multiple heading levels, and nested sections.

## Example File

See [sample-project.md](./sample-project.md) for the complete example document (a route refactoring project plan with 9 major sections).

---

## Tool 1: markdown_tree

**Purpose**: View hierarchical document structure without reading content. Perfect for getting a quick overview of a large document's organization.

**Example Call**:
```
mcp__markdown-nav__markdown_tree
  file_path: /path/to/sample-project.md
  format: ascii
```

**Output**:
```
sample-project.md

└ Route Refactoring Project H1:1:261
  │ Executive Summary H2:3:21
  │ Task 1: Extract Authentication Routes H2:22:62
  │ Task 2: Extract User Management Routes H2:63:87
  │ Task 3: Implement Middleware Framework H2:88:127
  │ Task 4: Implement API Versioning H2:128:156
  │ Testing Strategy H2:157:186
  │ Rollout Plan H2:187:217
  │ Monitoring and Metrics H2:218:239
  │ Post-Launch H2:240:261
```

**Explanation**: The tree shows the document structure in a vim-vista style. Each line shows:
- Section name
- Heading level (H1, H2, H3, etc.)
- Line range (start:end)

This allows you to see the entire document structure at a glance without loading any content, saving significant context tokens.

---

## Tool 2: markdown_list_sections

**Purpose**: Get a structured list of sections with their locations. Useful for discovering what sections exist and where they are located.

**Example Call**:
```
mcp__markdown-nav__markdown_list_sections
  file_path: /path/to/sample-project.md
  max_depth: 2
```

**Output**:
```json
{
  "sections": [
    {
      "name": "Route Refactoring Project",
      "start_line": 1,
      "end_line": 261,
      "level": "H1"
    },
    {
      "name": "Executive Summary",
      "start_line": 3,
      "end_line": 21,
      "level": "H2"
    },
    {
      "name": "Task 1: Extract Authentication Routes",
      "start_line": 22,
      "end_line": 62,
      "level": "H2"
    },
    {
      "name": "Task 2: Extract User Management Routes",
      "start_line": 63,
      "end_line": 87,
      "level": "H2"
    },
    {
      "name": "Task 3: Implement Middleware Framework",
      "start_line": 88,
      "end_line": 127,
      "level": "H2"
    },
    {
      "name": "Task 4: Implement API Versioning",
      "start_line": 128,
      "end_line": 156,
      "level": "H2"
    },
    {
      "name": "Testing Strategy",
      "start_line": 157,
      "end_line": 186,
      "level": "H2"
    },
    {
      "name": "Rollout Plan",
      "start_line": 187,
      "end_line": 217,
      "level": "H2"
    },
    {
      "name": "Monitoring and Metrics",
      "start_line": 218,
      "end_line": 239,
      "level": "H2"
    },
    {
      "name": "Post-Launch",
      "start_line": 240,
      "end_line": 261,
      "level": "H2"
    }
  ],
  "count": 10
}
```

**Explanation**: Returns a JSON array of all sections up to the specified depth (max_depth: 2 shows H1 and H2 headings). Each section includes:
- Exact section name
- Start and end line numbers
- Heading level

The `max_depth` parameter lets you control how much detail you see. Set it to 0 to see all heading levels (H1-H6), or use 1-6 to limit the depth.

**Optional Parameters**:
- `section_name_pattern`: Regex filter (e.g., "Task.*" to show only task sections)

---

## Tool 3: markdown_section_bounds

**Purpose**: Find the exact line boundaries for a specific section. Use this when you need to know WHERE a section is without reading its content.

**Example Call**:
```
mcp__markdown-nav__markdown_section_bounds
  file_path: /path/to/sample-project.md
  section_heading: Task 1: Extract Authentication Routes
```

**Output**:
```json
{
  "section_name": "Task 1: Extract Authentication Routes",
  "start_line": 22,
  "end_line": 62,
  "heading_level": "H2",
  "total_lines": 41
}
```

**Explanation**: Returns the precise boundaries of the section:
- `start_line`: Where the section begins (line 22)
- `end_line`: Where the section ends (line 62)
- `total_lines`: Total lines in the section (41 lines)
- `heading_level`: The heading level (H2)

This is useful for determining section size before reading, or for tools that need line number ranges for other operations.

**Note**: The `section_heading` must match exactly (case-sensitive), but without the `#` symbols. For example, use "Task 1: Extract Authentication Routes" not "## Task 1: Extract Authentication Routes".

---

## Tool 4: markdown_read_section

**Purpose**: Read a specific section's content without loading the entire document. This is the most powerful tool for context-efficient reading.

**Example Call**:
```
mcp__markdown-nav__markdown_read_section
  file_path: /path/to/sample-project.md
  section_heading: Task 1: Extract Authentication Routes
  max_subsection_levels: 1
```

**Output**:
```json
{
  "content": "## Task 1: Extract Authentication Routes\n\n### Overview\n\nExtract all authentication-related routes from the monolithic router into a dedicated authentication module. This includes login, logout, registration, password reset, and token management endpoints.\n\n### Requirements\n\n1. Create new `routes/auth` package\n2. Move authentication handlers to dedicated files\n3. Implement authentication middleware\n4. Update route registration to use new structure\n5. Maintain backward compatibility with existing API\n\n### Implementation Details\n\nThe authentication module should expose a single `RegisterRoutes` function that takes a router group and registers all auth endpoints. This allows for clean separation and easy testing.\n\n```go\n// Example structure\npackage auth\n\nfunc RegisterRoutes(group *router.Group) {\n    group.POST(\"/login\", handleLogin)\n    group.POST(\"/logout\", handleLogout)\n    group.POST(\"/register\", handleRegister)\n}\n```\n\n### Testing Requirements\n\n- Unit tests for each handler\n- Integration tests for authentication flow\n- Test authentication middleware with valid/invalid tokens\n- Test rate limiting on login endpoints\n\n### Dependencies\n\n- Task 3 (middleware framework must be ready)\n- Database migration for session storage",
  "section_name": "Task 1: Extract Authentication Routes",
  "start_line": 22,
  "end_line": 62,
  "lines_read": 41
}
```

**Explanation**: Returns the complete content of the section as markdown text. The `content` field contains the actual section text, including:
- The section heading
- All subsections up to the specified depth
- All text, code blocks, lists, etc.

In this example, we read Task 1 which is 41 lines, instead of loading the entire 261-line document. This saves approximately 84% of the context tokens that would be used by reading the full file.

**Optional Parameters**:
- `max_subsection_levels`: Control subsection depth (omit to read entire section with all subsections)
  - `0`: Just the section heading and immediate content (no subsections)
  - `1`: Include immediate child subsections (H3 if parent is H2)
  - `2`: Include children and grandchildren
  - Omitted: Read entire section tree (recommended default)

---

## Typical Workflow Examples

### Scenario 1: "Analyze Task 3 from the planning document"

**Agent Workflow**:
```
1. markdown_tree (format: ascii)
   → Get document overview, identify sections

2. markdown_section_bounds (section_heading: "Task 3: Implement Middleware Framework")
   → Verify section exists and get size (40 lines)

3. markdown_read_section (section_heading: "Task 3: Implement Middleware Framework")
   → Read only Task 3 content (15% of document)

Result: Complete task analysis using only ~40 lines instead of 261 lines (85% token savings)
```

### Scenario 2: "What testing tasks are documented?"

**Agent Workflow**:
```
1. markdown_list_sections (section_name_pattern: ".*[Tt]est.*")
   → Find all sections with "test" in the name

2. For each matching section:
   markdown_read_section (section_heading: <matched_section>)
   → Read only testing-related sections

Result: Targeted reading of only relevant sections
```

### Scenario 3: "Implement the rollout plan"

**Agent Workflow**:
```
1. markdown_tree (format: ascii)
   → Understand document structure

2. markdown_read_section (section_heading: "Executive Summary")
   → Get project context

3. markdown_read_section (section_heading: "Rollout Plan")
   → Get implementation details

4. markdown_read_section (section_heading: "Monitoring and Metrics")
   → Understand success criteria

Result: Comprehensive understanding using only 3 targeted section reads
```

---

## Performance Benefits

Using this markdown navigation approach instead of reading entire files:

- **Token Usage**: Reduce context usage by 70-85% for targeted section reading
- **Efficiency**: Read only what you need, when you need it
- **Scalability**: Handle documents of 2,000+ lines without context limits
- **Navigation**: Quickly explore document structure before committing to reading

## Tips for Claude Code and MCP Clients

1. **Start with structure**: Use `markdown_tree` or `markdown_list_sections` to understand the document before reading
2. **Read selectively**: Use `markdown_read_section` to read only relevant sections
3. **Verify before reading**: Use `markdown_section_bounds` to check section size before reading
4. **Use filters**: The `section_name_pattern` parameter helps narrow down large section lists
5. **Control depth**: Use `max_depth` and `max_subsection_levels` to control how much you read
