package tools

import (
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownTreeArgs defines the input arguments for the markdown_tree tool.
type MarkdownTreeArgs struct {
	FilePath string `json:"file_path"           jsonschema:"required,description=Path to markdown file"`
	TagsFile string `json:"tags_file,omitempty" jsonschema:"description=Path to ctags file (default: tags)"`
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
			// Default tags file
			tagsFile := args.TagsFile
			if tagsFile == "" {
				tagsFile = DefaultTagsFile
			}

			// Parse tags file
			entries, err := ctags.ParseTagsFile(tagsFile, args.FilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse tags file: %w", err)
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
