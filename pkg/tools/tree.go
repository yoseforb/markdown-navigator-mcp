package tools

import (
	"fmt"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownTreeArgs defines the input arguments for the markdown_tree tool.
type MarkdownTreeArgs struct {
	FilePath string  `json:"file_path"           jsonschema:"required,description=Path to markdown file"`
	Format   *string `json:"format,omitempty"    jsonschema:"description=Output format: json (default) or ascii"`
	Pattern  *string `json:"pattern,omitempty"   jsonschema:"description=Filter to sections matching pattern"`
	MaxDepth *int    `json:"max_depth,omitempty" jsonschema:"description=Maximum depth to display (default: 2, 0 = unlimited)"`
}

// MarkdownTreeResponse defines the response structure.
type MarkdownTreeResponse struct {
	Tree     string          `json:"tree,omitempty"`      // ASCII format (deprecated)
	TreeJSON *ctags.TreeNode `json:"tree_json,omitempty"` // JSON format (default)
	Format   string          `json:"format"`              // "json" or "ascii"
}

// RegisterMarkdownTree registers the markdown_tree tool with the MCP server.
func RegisterMarkdownTree(srv server.Server) {
	srv.Tool(
		"markdown_tree",
		"Display hierarchical document structure (JSON or ASCII format)",
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

			// Filter by pattern if provided
			if args.Pattern != nil && *args.Pattern != "" {
				entries = ctags.FilterByPatternWithParents(
					entries,
					*args.Pattern,
				)
			}

			// Filter by depth (default: 2, use 0 for unlimited)
			depth := 2
			if args.MaxDepth != nil {
				depth = *args.MaxDepth
			}
			if depth > 0 {
				entries = ctags.FilterByDepth(entries, depth)
			}

			// Default format to json
			format := "json"
			if args.Format != nil && *args.Format != "" {
				format = *args.Format
			}

			// Validate format
			if format != "json" && format != "ascii" {
				return nil, fmt.Errorf(
					"%w: %s (must be 'json' or 'ascii')",
					ErrInvalidFormat,
					format,
				)
			}

			// Build response based on format
			response := MarkdownTreeResponse{
				Format:   format,
				Tree:     "",
				TreeJSON: nil,
			}

			switch format {
			case "json":
				response.TreeJSON = ctags.BuildTreeJSON(entries)
			case "ascii":
				response.Tree = ctags.BuildTreeStructure(entries)
			}

			return response, nil
		},
	)
}
