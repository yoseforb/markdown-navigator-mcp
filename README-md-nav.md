# Markdown Navigation Tools

A collection of tools to navigate large markdown files using ctags with enhanced fields.

## Setup

1. **Generate ctags with enhanced fields:**
   ```bash
   ctags -R --fields=+KnS --languages=markdown
   ```

2. **Tools are ready to use:**
   - `ctags-tree.py` - Generate vim-vista-like tree structure
   - `ctags-section.py` - Find section line bounds
   - `md-nav` - All-in-one navigation helper

## Usage

### 1. Show Full Document Structure

```bash
./scripts/md-nav tree ai-docs/planning/backlog/route-domain-simplification-refactoring.md
```

**Output:**
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
      ...
```

### 2. Find Section Line Bounds

```bash
./scripts/md-nav section ai-docs/planning/backlog/route-domain-simplification-refactoring.md "Task 4"
```

**Output:**
```
Section: Task 4: Update /routes/{id} Access Control
Lines: 550-905

To read this section:
Read(file_path='...', offset=549, limit=356)
```

### 3. Read Specific Section

```bash
./scripts/md-nav read ai-docs/planning/backlog/route-domain-simplification-refactoring.md "Task 4"
```

**Output:** Full content of Task 4 section (lines 550-905)

## Integration with Claude Code

Now you can efficiently work with large markdown files:

1. **Get overview:**
   ```bash
   ./scripts/md-nav tree <file>
   ```

2. **Find section bounds:**
   ```bash
   ./scripts/md-nav section <file> "query"
   ```

3. **Read only what you need:**
   - Use the line numbers from step 2
   - Read specific sections without loading entire file
   - Save context tokens!

## Examples

```bash
# Show all tasks in planning doc
./scripts/md-nav tree ai-docs/planning/backlog/route-domain-simplification-refactoring.md | grep "Task"

# Find and read Task 5
./scripts/md-nav read ai-docs/planning/backlog/route-domain-simplification-refactoring.md "Task 5"

# Get line numbers for Task 3
./scripts/md-nav section ai-docs/planning/backlog/route-domain-simplification-refactoring.md "Task 3"
```

## Benefits

✅ **Context-efficient** - Read only the sections you need
✅ **Fast navigation** - Instant lookups using ctags
✅ **Hierarchical view** - vim-vista-like tree structure
✅ **Line-accurate** - Exact line numbers for Read tool
✅ **Flexible search** - Fuzzy section name matching

## How It Works

1. **ctags** parses markdown and generates tags file with:
   - Line numbers (`line:XXX`)
   - Heading levels (`chapter`, `section`, `subsection`, `subsubsection`)
   - Scope/hierarchy information

2. **Python scripts** parse the tags file and:
   - Build hierarchical tree structure
   - Calculate section boundaries
   - Enable targeted content extraction

3. **md-nav wrapper** provides a simple CLI interface for all operations
