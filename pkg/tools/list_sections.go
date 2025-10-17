package tools

import (
	"context"
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownListSectionsArgs defines the input arguments.
type MarkdownListSectionsArgs struct {
	FilePath           string  `json:"file_path"                      description:"Path to markdown file to list sections from"                                                                            required:"true"`
	MaxDepth           *int    `json:"max_depth,omitempty"            description:"Maximum heading depth to show (1-6). Default: 2 (H1+H2). Use 0 for all levels. Example: 1=only H1, 2=H1+H2, 3=H1+H2+H3"`
	SectionNamePattern *string `json:"section_name_pattern,omitempty" description:"Regex pattern to filter section names. Example: 'Task.*' matches sections starting with 'Task'"`
}

// SectionInfo represents a single section in the list.
type SectionInfo struct {
	Name      string `json:"name"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Level     string `json:"level"`
}

// MarkdownListSectionsResponse defines the response structure.
type MarkdownListSectionsResponse struct {
	Sections []SectionInfo `json:"sections"`
	Count    int           `json:"count"`
}

// RegisterMarkdownListSections registers the markdown_list_sections tool.
func RegisterMarkdownListSections(srv server.Server) {
	srv.Tool(
		"markdown_list_sections",
		"List sections to explore document structure before reading content. Returns section names, levels, and line ranges. Most efficient way to navigate unfamiliar markdown files. Use before markdown_read_section to identify relevant sections.",
		func(_ *server.Context, args MarkdownListSectionsArgs) (interface{}, error) {
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

			// Filter by maximum depth (default: 2, meaning H1+H2)
			filteredEntries := entries
			maxDepth := 2 // Default to depth 2 (H1+H2)
			if args.MaxDepth != nil {
				maxDepth = *args.MaxDepth
			}

			// Validate max_depth range
			if maxDepth < 0 || maxDepth > 6 {
				return nil, fmt.Errorf(
					"%w: %d (must be 0-6, where 0 means all levels)",
					ErrInvalidLevel,
					maxDepth,
				)
			}

			// Apply depth filtering (0 means all levels, no filtering)
			filteredEntries = ctags.FilterByDepth(filteredEntries, maxDepth)

			// Filter by pattern if specified
			if args.SectionNamePattern != nil &&
				*args.SectionNamePattern != "" {
				filteredEntries = ctags.FilterByPattern(
					filteredEntries,
					*args.SectionNamePattern,
				)
			}

			// Convert to response format
			sections := make([]SectionInfo, 0, len(filteredEntries))
			for _, entry := range filteredEntries {
				sections = append(sections, SectionInfo{
					Name:      entry.Name,
					StartLine: entry.Line,
					EndLine:   entry.End,
					Level:     fmt.Sprintf("H%d", entry.Level),
				})
			}

			return MarkdownListSectionsResponse{
				Sections: sections,
				Count:    len(sections),
			}, nil
		},
	)
}
