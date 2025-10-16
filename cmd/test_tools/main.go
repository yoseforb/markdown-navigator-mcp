// Package main provides a test tool for manual testing of markdown navigation.
// This is a development/testing utility, not production code.
//
//nolint:cyclop,gocognit,nestif,sloglint,gocritic,funlen // test tool complexity
package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

func main() {
	if len(os.Args) < 2 {
		slog.Info(
			"Usage: go run test_tools.go <markdown-file>",
		)
		slog.Info(
			"Example: go run test_tools.go mcp-mdnav-server-prompt.md",
		)
		os.Exit(1)
	}

	markdownFile := os.Args[1]

	flag.Parse()

	slog.Info("=== Testing Markdown Navigation Tools ===")
	slog.Info("Markdown file", "file", markdownFile)

	// Test 1: Get Tags from Cache (replaces ParseTagsFile)
	slog.Info("1. Testing GetTags from cache...")
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), markdownFile)
	if err != nil {
		log.Fatalf("GetTags failed: %v", err)
	}
	slog.Info("Found entries", "count", len(entries))

	if len(entries) == 0 {
		log.Fatal("No entries found - cannot continue tests")
	}

	// Test 2: Build Tree Structure
	slog.Info("2. Testing BuildTreeStructure...")
	tree := ctags.BuildTreeStructure(entries)
	if tree == "" {
		log.Fatal("BuildTreeStructure returned empty tree")
	}
	slog.Info("Tree structure built successfully")
	slog.Info("Tree preview: ")

	slog.Info(tree)

	// Test 3: Filter By Level (H2)
	slog.Info("3. Testing FilterByLevel (H2)...")
	h2Entries := ctags.FilterByLevel(entries, 2)
	slog.Info("Found H2 sections", "count", len(h2Entries))
	if len(h2Entries) > 0 {
		slog.Info("First H2 sections:")
		for i, entry := range h2Entries {
			if i >= 3 {
				slog.Info("  ...")
				break
			}
			slog.Info("", "name", entry.Name, "line", entry.Line)
		}
	}

	// Test 4: Filter By Pattern
	slog.Info("4. Testing FilterByPattern...")
	if len(h2Entries) > 0 {
		firstSection := h2Entries[0].Name
		// Use just first word for pattern matching
		pattern := firstSection
		if len(pattern) > 10 {
			pattern = pattern[:10]
		}
		filtered := ctags.FilterByPattern(entries, pattern)
		slog.Info(
			"Pattern matches found",
			"pattern",
			pattern,
			"matches",
			len(filtered),
		)
	}

	// Test 5: Find Section Bounds
	slog.Info("5. Testing FindSectionBounds...")
	if len(h2Entries) > 0 {
		firstSection := h2Entries[0].Name
		startLine, endLine, sectionName, found := ctags.FindSectionBounds(
			entries,
			firstSection,
		)
		if !found {
			log.Fatalf("Failed to find section: %s", firstSection)
		}
		slog.Info("Found section", "name", sectionName)
		slog.Info("Start line", "line", startLine)
		if endLine > 0 {
			slog.Info("End line", "line", endLine)
			slog.Info("Total lines", "count", endLine-startLine+1)
		} else {
			slog.Info("End line: EOF")
		}
	}

	// Test 6: Read File Lines
	slog.Info("6. Testing section reading...")
	if len(h2Entries) > 0 {
		firstSection := h2Entries[0].Name
		startLine, endLine, _, found := ctags.FindSectionBounds(
			entries,
			firstSection,
		)
		if found {
			file, err := os.Open(markdownFile)
			if err != nil {
				log.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			// Read lines using bufio.Scanner
			var lines []string
			scanner := bufio.NewScanner(file)
			currentLine := 1
			linesRead := 0

			for scanner.Scan() {
				if currentLine >= startLine {
					if endLine > 0 && currentLine > endLine {
						break
					}
					lines = append(lines, scanner.Text())
					linesRead++
				}
				currentLine++
			}

			if err := scanner.Err(); err != nil {
				log.Fatalf("Failed to scan file: %v", err)
			}

			content := ""
			for i, line := range lines {
				content += line
				if i < len(lines)-1 {
					content += "\n"
				}
			}

			if len(content) > 0 {
				slog.Info(
					"Section content read successfully",
					"lines",
					linesRead,
				)
				slog.Info("Content", "text", content)
			} else {
				slog.Error("No content read")
			}
		}
	}

	slog.Info("=== All Tests Passed! ===")
}
