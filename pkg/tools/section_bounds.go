package tools

import (
	"context"
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownSectionBoundsArgs defines the input arguments.
type MarkdownSectionBoundsArgs struct {
	FilePath       string `json:"file_path"       description:"Path to markdown file"                                                                                                   required:"true"`
	SectionHeading string `json:"section_heading" description:"Exact heading text to find (case-sensitive, without # symbols). Example: 'Executive Summary' not '## Executive Summary'" required:"true"`
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
		"Get line number boundaries for a section without reading content. Use when you only need to know WHERE a section is located. If you need the actual content, use markdown_read_section instead.",
		func(_ *server.Context, args MarkdownSectionBoundsArgs) (interface{}, error) {
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
				args.SectionHeading,
			)
			if !found {
				return nil, fmt.Errorf(
					"%w: '%s'",
					ErrSectionNotFound,
					args.SectionHeading,
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
