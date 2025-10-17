package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// MarkdownTreeArgs defines the input arguments for the markdown_tree tool.
type MarkdownTreeArgs struct {
	FilePath           string  `json:"file_path"                      description:"Path to markdown file"                                                                                             required:"true"`
	Format             *string `json:"format,omitempty"               description:"Output format: 'json' for structured data or 'ascii' for visual tree. Default: 'json'"`
	SectionNamePattern *string `json:"section_name_pattern,omitempty" description:"Regex pattern to filter which sections appear in tree. Example: 'Task.*' shows only sections starting with 'Task'"`
	MaxDepth           *int    `json:"max_depth,omitempty"            description:"Maximum tree depth to display (1-6, 0=all). Default: 2 (H1+H2)"`
}

// MarkdownTreeResponse defines the response structure.
type MarkdownTreeResponse struct {
	TreeLines []string        `json:"tree_lines,omitempty"` // ASCII format as array of lines
	TreeJSON  *ctags.TreeNode `json:"tree_json,omitempty"`  // JSON format (default)
	Format    string          `json:"format"`               // "json" or "ascii"
}

// splitLines splits a string into lines for better JSON readability.
func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

// RegisterMarkdownTree registers the markdown_tree tool with the MCP server.
func RegisterMarkdownTree(srv server.Server) {
	srv.Tool(
		"markdown_tree",
		"Display hierarchical document structure as visual tree. Use for deeply nested documents when you need to visualize parent-child relationships. For simple section lists, use markdown_list_sections instead.",
		func(_ *server.Context, args MarkdownTreeArgs) (interface{}, error) {
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

			// Filter by pattern if provided
			if args.SectionNamePattern != nil &&
				*args.SectionNamePattern != "" {
				entries = ctags.FilterByPatternWithParents(
					entries,
					*args.SectionNamePattern,
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
				Format:    format,
				TreeLines: nil,
				TreeJSON:  nil,
			}

			switch format {
			case "json":
				response.TreeJSON = ctags.BuildTreeJSON(entries)
			case "ascii":
				treeString := ctags.BuildTreeStructure(entries)
				// Split into lines for better readability in JSON
				response.TreeLines = splitLines(treeString)
			}

			return response, nil
		},
	)
}
