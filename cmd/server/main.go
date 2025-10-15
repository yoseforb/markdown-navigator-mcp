package main

import (
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
	"github.com/yoseforb/markdown-nav-mcp/pkg/tools"
)

func main() {
	// Parse command-line flags
	ctagsPath := flag.String(
		"ctags-path",
		"ctags",
		"Path to the ctags executable (defaults to 'ctags' in PATH)",
	)
	flag.Parse()

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	}))

	// Configure ctags executable path
	if err := ctags.SetCtagsPath(*ctagsPath); err != nil {
		logger.Error("Failed to configure ctags path",
			"path", *ctagsPath,
			"error", err,
		)
		log.Fatalf("Invalid ctags path: %v", err)
	}

	logger.Info("Configured ctags executable",
		"path", ctags.GetCtagsPath(),
	)

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
