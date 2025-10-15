package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownReadSectionArgs defines the input arguments.
type MarkdownReadSectionArgs struct {
	FilePath     string `json:"file_path"       jsonschema:"required,description=Path to markdown file"`
	SectionQuery string `json:"section_query"   jsonschema:"required,description=Section name or search query"`
	Depth        *int   `json:"depth,omitempty" jsonschema:"description=Maximum subsection depth to include (default: unlimited, 0 = no subsections, 1 = immediate children only)"`
}

// MarkdownReadSectionResponse defines the response structure.
type MarkdownReadSectionResponse struct {
	Content     string `json:"content"`
	SectionName string `json:"section_name"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	LinesRead   int    `json:"lines_read"`
}

// RegisterMarkdownReadSection registers the markdown_read_section tool.
func RegisterMarkdownReadSection(srv server.Server) {
	srv.Tool(
		"markdown_read_section",
		"Read a specific section's content",
		handleReadSection,
	)
}

// handleReadSection implements the markdown_read_section tool logic.
func handleReadSection(
	_ *server.Context,
	args MarkdownReadSectionArgs,
) (interface{}, error) {
	// Note: gomcp's server.Context does not provide request-level context.
	// Application-level cancellation is handled via signal handling in main.go.
	reqCtx := context.Background()

	// Get tags from cache with context
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(reqCtx, args.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("%w for %s", ErrNoEntries, args.FilePath)
	}

	// Find section bounds
	startLine, endLine, sectionName, found := ctags.FindSectionBounds(
		entries,
		args.SectionQuery,
	)
	if !found {
		return nil, fmt.Errorf(
			"%w: '%s'",
			ErrSectionNotFound,
			args.SectionQuery,
		)
	}

	// Read the full section content (without depth filtering at boundary level)
	content, linesRead, err := readFileLines(
		args.FilePath,
		startLine,
		endLine,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Apply depth filtering if depth parameter is provided
	filteredContent := content
	if args.Depth != nil {
		// Find the root section level
		rootLevel := findSectionLevel(entries, startLine)
		if rootLevel > 0 {
			filteredContent = filterContentByDepth(
				rootLevel,
				*args.Depth,
				content,
			)
		}
	}

	return MarkdownReadSectionResponse{
		Content:     filteredContent,
		SectionName: sectionName,
		StartLine:   startLine,
		EndLine:     endLine,
		LinesRead:   linesRead,
	}, nil
}

// filterContentByDepth filters markdown content to only include headings
// up to the specified depth relative to the root heading level.
//
// Parameters:
//   - rootLevel: The heading level of the root section (1-6 for H1-H6)
//   - depth: How many levels deep to include (0 = root only, 1 = root + children, etc.)
//   - content: The full markdown content to filter
//
// Returns filtered content with headings deeper than (rootLevel + depth) removed.
func filterContentByDepth(rootLevel int, depth int, content string) string {
	// Handle depth=0 case - return only content until first subsection
	if depth <= 0 {
		return filterDepthZero(rootLevel, content)
	}

	maxAllowedLevel := rootLevel + depth
	var result strings.Builder
	inSkipMode := false

	headingRegex := regexp.MustCompile(`^(#{1,6})\s+`)

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Check if this line is a heading
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			headingLevel := len(matches[1]) // Count the #'s

			if headingLevel > maxAllowedLevel {
				inSkipMode = true // Too deep, start skipping
			} else if headingLevel > rootLevel {
				inSkipMode = false // Allowed level, include
			}
			// If headingLevel <= rootLevel, it's a sibling/parent, shouldn't happen
			// in a properly bounded section, but if it does, include it
		}

		// Include line if not in skip mode
		if !inSkipMode {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimRight(result.String(), "\n")
}

// filterDepthZero handles the special case of depth=0 where we only want
// the root section content without any subsections.
func filterDepthZero(rootLevel int, content string) string {
	var result strings.Builder
	headingRegex := regexp.MustCompile(`^(#{1,6})\s+`)

	lines := strings.Split(content, "\n")
	firstLine := true

	for _, line := range lines {
		// Check if this line is a heading
		if matches := headingRegex.FindStringSubmatch(line); matches != nil {
			headingLevel := len(matches[1]) // Count the #'s

			// Include the root heading (first heading encountered)
			if firstLine {
				result.WriteString(line)
				result.WriteString("\n")
				firstLine = false
				continue
			}

			// Stop at any subsection (level > rootLevel)
			if headingLevel > rootLevel {
				break
			}
		} else if !firstLine {
			// Include non-heading lines
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimRight(result.String(), "\n")
}

// calculateEndLine determines the actual end line based on depth parameter.
// depth=nil: unlimited depth (read all subsections)
// depth=0: no subsections (only section content)
// depth=1: immediate children only (e.g., H2 + H3, skip H4)
// depth=2: children + grandchildren (e.g., H2 + H3 + H4, skip H5)
// Negative depth values are treated as 0.
func calculateEndLine(
	entries []*ctags.TagEntry,
	startLine, endLine int,
	depth *int,
) int {
	// Find the current section's level
	currentLevel := findSectionLevel(entries, startLine)
	if currentLevel == 0 {
		return endLine // No level found, return original
	}

	// Handle unlimited depth case
	if depth == nil {
		return endLine
	}

	// Handle depth=0 case - no subsections, only section content
	if *depth <= 0 {
		return calculateEndLineDepthZero(
			entries,
			startLine,
			endLine,
			currentLevel,
		)
	}

	// Depth >= 1: Include children up to specified depth
	return calculateEndLineWithDepth(
		entries,
		startLine,
		endLine,
		currentLevel,
		*depth,
	)
}

// calculateEndLineDepthZero handles depth=0 case.
func calculateEndLineDepthZero(
	entries []*ctags.TagEntry,
	startLine, endLine int,
	currentLevel int,
) int {
	// Find first child section (any entry with level > currentLevel)
	for _, entry := range entries {
		if entry.Line > startLine {
			// Stop at sibling or parent
			if entry.Level <= currentLevel {
				return entry.Line - 1
			}
			// Stop at any child (level > currentLevel)
			if entry.Level > currentLevel {
				return entry.Line - 1
			}
		}
	}
	// No child or sibling found, return original end line
	return endLine
}

// calculateEndLineWithDepth handles depth >= 1 case.
func calculateEndLineWithDepth(
	entries []*ctags.TagEntry,
	startLine, endLine int,
	currentLevel int,
	depth int,
) int {
	maxAllowedLevel := currentLevel + depth
	lastAllowedLine := startLine - 1
	foundAnyAllowed := false

	for i, entry := range entries {
		if entry.Line <= startLine {
			continue
		}

		// Stop at sibling or parent (same or lower level number)
		if entry.Level <= currentLevel {
			if foundAnyAllowed {
				return lastAllowedLine
			}
			return entry.Line - 1
		}

		// Entry is within allowed depth (child sections we want to include)
		if entry.Level <= maxAllowedLevel && entry.Level > currentLevel {
			foundAnyAllowed = true
			sectionEnd := findSectionEnd(entries, i, entry.Level, endLine)
			if sectionEnd > lastAllowedLine {
				lastAllowedLine = sectionEnd
			}
		}
		// If entry.Level > maxAllowedLevel, skip but continue scanning
	}

	// No sibling/parent found
	if foundAnyAllowed {
		return lastAllowedLine
	}
	return endLine
}

// findSectionEnd finds where a section ends by looking for the next
// entry at the same or lower level number.
func findSectionEnd(
	entries []*ctags.TagEntry,
	currentIndex int,
	currentLevel int,
	defaultEnd int,
) int {
	// Look for next entry at same or lower level
	for i := currentIndex + 1; i < len(entries); i++ {
		if entries[i].Level <= currentLevel {
			return entries[i].Line - 1
		}
	}
	// No next entry at same level, goes to default end
	return defaultEnd
}

// findSectionLevel finds the heading level of the section at the given line.
func findSectionLevel(entries []*ctags.TagEntry, line int) int {
	for _, entry := range entries {
		if entry.Line == line {
			return entry.Level
		}
	}
	return 0
}

// readFileLines reads lines from a file between startLine and endLine (inclusive)
// If endLine is 0, reads to EOF.
func readFileLines(
	filePath string,
	startLine, endLine int,
) (string, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	currentLine := 1
	linesRead := 0

	for scanner.Scan() {
		if currentLine >= startLine {
			if endLine > 0 && currentLine > endLine {
				break
			}
			lines = append(lines, scanner.Text())
			linesRead++
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", 0, fmt.Errorf("failed to scan file: %w", err)
	}

	return strings.Join(lines, "\n"), linesRead, nil
}
