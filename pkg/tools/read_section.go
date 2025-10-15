package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownReadSectionArgs defines the input arguments.
type MarkdownReadSectionArgs struct {
	FilePath           string `json:"file_path"                     jsonschema:"required,description=Path to markdown file"`
	SectionQuery       string `json:"section_query"                 jsonschema:"required,description=Section name or search query"`
	IncludeSubsections *bool  `json:"include_subsections,omitempty" jsonschema:"description=Include child sections (default: true)"`
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

	// Determine actual end line based on include_subsections parameter
	actualEndLine := calculateEndLine(
		entries,
		startLine,
		endLine,
		args.IncludeSubsections,
	)

	// Read the file content
	content, linesRead, err := readFileLines(
		args.FilePath,
		startLine,
		actualEndLine,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return MarkdownReadSectionResponse{
		Content:     content,
		SectionName: sectionName,
		StartLine:   startLine,
		EndLine:     actualEndLine,
		LinesRead:   linesRead,
	}, nil
}

// calculateEndLine determines the actual end line based on include_subsections.
// When include_subsections is false, it finds the first child section and
// returns the line before it. Otherwise, returns the original end line.
func calculateEndLine(
	entries []*ctags.TagEntry,
	startLine, endLine int,
	includeSubsections *bool,
) int {
	// Determine if subsections should be included (default: true)
	shouldIncludeSubsections := true
	if includeSubsections != nil {
		shouldIncludeSubsections = *includeSubsections
	}

	// If including subsections, return original end line
	if shouldIncludeSubsections {
		return endLine
	}

	// Find the current section's level
	currentLevel := findSectionLevel(entries, startLine)
	if currentLevel == 0 {
		return endLine // No level found, return original
	}

	// Find first child section (higher level number)
	for _, entry := range entries {
		if entry.Line > startLine && entry.Level > currentLevel {
			// Found first child - stop reading before it
			return entry.Line - 1
		}
	}

	// No children found, return original end line
	return endLine
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
