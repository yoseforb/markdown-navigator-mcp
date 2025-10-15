package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/localrivet/gomcp/server"
	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
	"github.com/yoseforb/markdown-nav-mcp/pkg/tools"
)

const (
	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	ShutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func run() error {
	// Create base context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		return fmt.Errorf("invalid ctags path: %w", err)
	}

	logger.Info("Configured ctags executable",
		"path", ctags.GetCtagsPath(),
	)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Launch signal handler goroutine
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal",
			"signal", sig.String(),
		)
		cancel()
	}()

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

	// Run server in goroutine with error channel
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Run(); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		logger.Info("Shutting down gracefully")
	case err := <-errChan:
		logger.Error("Server error", "error", err)
		return err
	}

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(
		ctx,
		ShutdownTimeout,
	)
	defer shutdownCancel()

	// Shutdown server (check if srv has Shutdown method)
	logger.Info("Shutting down server",
		"timeout", ShutdownTimeout.String(),
	)

	// Note: gomcp server may not have a Shutdown() method
	// If it doesn't, the server will stop when Run() returns
	// Context cancellation will propagate through tool handlers

	// Wait for shutdown context or early completion
	<-shutdownCtx.Done()

	// Get cache statistics and shutdown cache
	cache := ctags.GetGlobalCache()
	cache.Shutdown(logger)

	logger.Info("Server shutdown complete")
	return nil
}
