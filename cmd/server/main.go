package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/tools"
)

func main() {
	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	}))

	// Create a new MCP server
	srv := server.NewServer("markdown-nav",
		server.WithLogger(logger),
	).AsStdio()

	// Register all markdown navigation tools
	tools.RegisterMarkdownTree(srv)
	tools.RegisterMarkdownSectionBounds(srv)
	tools.RegisterMarkdownReadSection(srv)
	tools.RegisterMarkdownListSections(srv)

	logger.Info("Starting markdown-nav MCP server",
		"tools", []string{
			"markdown_tree",
			"markdown_section_bounds",
			"markdown_read_section",
			"markdown_list_sections",
		},
	)

	// Start the server
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
