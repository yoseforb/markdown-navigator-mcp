package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_tools.go <markdown-file> [tags-file]")
		fmt.Println(
			"Example: go run test_tools.go mcp-markdown-nav-server-prompt.md",
		)
		os.Exit(1)
	}

	markdownFile := os.Args[1]
	tagsFile := "tags"
	if len(os.Args) > 2 {
		tagsFile = os.Args[2]
	}

	flag.Parse()

	fmt.Println("=== Testing Markdown Navigation Tools ===")
	fmt.Printf("Markdown file: %s\n", markdownFile)
	fmt.Printf("Tags file: %s\n\n", tagsFile)

	// Test 1: Parse Tags File
	fmt.Println("1. Testing ParseTagsFile...")
	entries, err := ctags.ParseTagsFile(tagsFile, markdownFile)
	if err != nil {
		log.Fatalf("❌ ParseTagsFile failed: %v", err)
	}
	fmt.Printf("✓ Found %d entries\n\n", len(entries))

	if len(entries) == 0 {
		log.Fatal("❌ No entries found - cannot continue tests")
	}

	// Test 2: Build Tree Structure
	fmt.Println("2. Testing BuildTreeStructure...")
	tree := ctags.BuildTreeStructure(entries)
	if tree == "" {
		log.Fatal("❌ BuildTreeStructure returned empty tree")
	}
	fmt.Println("✓ Tree structure built successfully")
	fmt.Println("Tree preview:")
	lines := 0
	for _, line := range tree {
		if line == '\n' {
			lines++
			if lines > 10 {
				fmt.Println("  ... (truncated)")
				break
			}
		}
		if lines <= 10 {
			fmt.Printf("  %c", line)
		}
	}
	fmt.Println()

	// Test 3: Filter By Level (H2)
	fmt.Println("\n3. Testing FilterByLevel (H2)...")
	h2Entries := ctags.FilterByLevel(entries, 2)
	fmt.Printf("✓ Found %d H2 sections\n", len(h2Entries))
	if len(h2Entries) > 0 {
		fmt.Println("First H2 sections:")
		for i, entry := range h2Entries {
			if i >= 3 {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  - %s (line %d)\n", entry.Name, entry.Line)
		}
	}

	// Test 4: Filter By Pattern
	fmt.Println("\n4. Testing FilterByPattern...")
	if len(h2Entries) > 0 {
		firstSection := h2Entries[0].Name
		// Use just first word for pattern matching
		pattern := firstSection
		if len(pattern) > 10 {
			pattern = pattern[:10]
		}
		filtered := ctags.FilterByPattern(entries, pattern)
		fmt.Printf("✓ Pattern '%s' found %d matches\n", pattern, len(filtered))
	}

	// Test 5: Find Section Bounds
	fmt.Println("\n5. Testing FindSectionBounds...")
	if len(h2Entries) > 0 {
		firstSection := h2Entries[0].Name
		startLine, endLine, sectionName, found := ctags.FindSectionBounds(
			entries,
			firstSection,
		)
		if !found {
			log.Fatalf("❌ Failed to find section: %s", firstSection)
		}
		fmt.Printf("✓ Found section: %s\n", sectionName)
		fmt.Printf("  Start line: %d\n", startLine)
		if endLine > 0 {
			fmt.Printf("  End line: %d\n", endLine)
			fmt.Printf("  Total lines: %d\n", endLine-startLine+1)
		} else {
			fmt.Println("  End line: EOF")
		}
	}

	// Test 6: Read File Lines (via markdown_read_section logic)
	fmt.Println("\n6. Testing section reading...")
	if len(h2Entries) > 0 {
		firstSection := h2Entries[0].Name
		startLine, endLine, _, found := ctags.FindSectionBounds(
			entries,
			firstSection,
		)
		if found {
			file, err := os.Open(markdownFile)
			if err != nil {
				log.Fatalf("❌ Failed to open file: %v", err)
			}
			defer file.Close()

			// Read lines using bufio.Scanner (like the actual implementation)
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
				log.Fatalf("❌ Failed to scan file: %v", err)
			}

			content := ""
			for i, line := range lines {
				content += line
				if i < len(lines)-1 {
					content += "\n"
				}
			}

			if len(content) > 0 {
				fmt.Printf(
					"✓ Section content read successfully (%d lines)\n",
					linesRead,
				)
				// if len(content) > 200 {
				// fmt.Printf("  Content preview (first 200 chars): %s...\n", content[:200])
				// } else {
				fmt.Printf("  Content: %s\n", content)
				// }
			} else {
				fmt.Println("❌ No content read")
			}
		}
	}

	fmt.Println("\n=== All Tests Passed! ===")
}
