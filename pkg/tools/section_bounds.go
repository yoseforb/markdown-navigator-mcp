package tools

import (
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownSectionBoundsArgs defines the input arguments.
type MarkdownSectionBoundsArgs struct {
	FilePath     string `json:"file_path"     jsonschema:"required,description=Path to markdown file"`
	SectionQuery string `json:"section_query" jsonschema:"required,description=Section name or search query (fuzzy match)"`
}

// MarkdownSectionBoundsResponse defines the response structure.
type MarkdownSectionBoundsResponse struct {
	SectionName  string `json:"section_name"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	HeadingLevel string `json:"heading_level"`
	TotalLines   int    `json:"total_lines"`
}

// RegisterMarkdownSectionBounds registers the markdown_section_bounds tool.
func RegisterMarkdownSectionBounds(srv server.Server) {
	srv.Tool(
		"markdown_section_bounds",
		"Find line number boundaries for a specific section",
		func(_ *server.Context, args MarkdownSectionBoundsArgs) (interface{}, error) {
			// Get tags from cache
			cache := ctags.GetGlobalCache()
			entries, err := cache.GetTags(args.FilePath)
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

			// Calculate total lines
			var totalLines int
			if endLine > 0 {
				totalLines = endLine - startLine + 1
			} else {
				// endLine == 0 means EOF, we can't calculate exact total without reading the file
				// For now, indicate it goes to EOF
				totalLines = -1 // Special value indicating "to EOF"
			}

			// Find the entry to get the heading level
			var headingLevel string
			for _, entry := range entries {
				if entry.Line == startLine {
					headingLevel = fmt.Sprintf("H%d", entry.Level)
					break
				}
			}

			return MarkdownSectionBoundsResponse{
				SectionName:  sectionName,
				StartLine:    startLine,
				EndLine:      endLine,
				HeadingLevel: headingLevel,
				TotalLines:   totalLines,
			}, nil
		},
	)
}
