package tools

import (
	"context"
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownListSectionsArgs defines the input arguments.
type MarkdownListSectionsArgs struct {
	FilePath     string  `json:"file_path"               jsonschema:"required,description=Path to markdown file"`
	HeadingLevel *string `json:"heading_level,omitempty" jsonschema:"description=Filter by level (H1, H2, H3, H4, or ALL)"`
	Pattern      *string `json:"pattern,omitempty"       jsonschema:"description=Search pattern (fuzzy match)"`
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
		"List all top-level sections (or sections matching a pattern)",
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

			// Filter by heading level if specified (ALL means no filtering)
			filteredEntries := entries
			if args.HeadingLevel != nil && *args.HeadingLevel != "" &&
				*args.HeadingLevel != "ALL" {
				var level int
				switch *args.HeadingLevel {
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
						*args.HeadingLevel,
					)
				}
				filteredEntries = ctags.FilterByLevel(filteredEntries, level)
			}

			// Filter by pattern if specified
			if args.Pattern != nil && *args.Pattern != "" {
				filteredEntries = ctags.FilterByPattern(
					filteredEntries,
					*args.Pattern,
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
