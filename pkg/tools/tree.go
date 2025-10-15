package tools

import (
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownTreeArgs defines the input arguments for the markdown_tree tool.
type MarkdownTreeArgs struct {
	FilePath string `json:"file_path" jsonschema:"required,description=Path to markdown file"`
}

// MarkdownTreeResponse defines the response structure.
type MarkdownTreeResponse struct {
	Tree string `json:"tree"`
}

// RegisterMarkdownTree registers the markdown_tree tool with the MCP server.
func RegisterMarkdownTree(srv server.Server) {
	srv.Tool(
		"markdown_tree",
		"Display hierarchical document structure (vim-vista style)",
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

			// Build tree structure
			tree := ctags.BuildTreeStructure(entries)

			return MarkdownTreeResponse{Tree: tree}, nil
		},
	)
}
