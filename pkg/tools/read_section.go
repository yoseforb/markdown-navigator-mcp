package tools

import (
	"bufio"
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
	TagsFile           string `json:"tags_file,omitempty"           jsonschema:"description=Path to ctags file (default: tags)"`
	IncludeSubsections bool   `json:"include_subsections,omitempty" jsonschema:"description=Include child sections (default: true)"`
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
		func(_ *server.Context, args MarkdownReadSectionArgs) (interface{}, error) {
			// Default tags file
			tagsFile := args.TagsFile
			if tagsFile == "" {
				tagsFile = DefaultTagsFile
			}

			// Note: IncludeSubsections is available in args if needed for future functionality

			// Parse tags file
			entries, err := ctags.ParseTagsFile(tagsFile, args.FilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse tags file: %w", err)
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

			// Read the file content
			content, linesRead, err := readFileLines(
				args.FilePath,
				startLine,
				endLine,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}

			return MarkdownReadSectionResponse{
				Content:     content,
				SectionName: sectionName,
				StartLine:   startLine,
				EndLine:     endLine,
				LinesRead:   linesRead,
			}, nil
		},
	)
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
