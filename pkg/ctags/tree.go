package ctags

import (
	"fmt"
	"path/filepath"
	"strings"
)

// BuildTreeStructure builds a vim-vista-like tree structure from tag entries.
func BuildTreeStructure(entries []*TagEntry) string {
	if len(entries) == 0 {
		return ""
	}

	var lines []string
	var stack []stackEntry // Track parent entries at each level

	for _, entry := range entries {
		level := entry.Level

		// Pop stack to current level
		for len(stack) > 0 && stack[len(stack)-1].Level >= level {
			stack = stack[:len(stack)-1]
		}

		// Calculate indentation
		indent := strings.Repeat("  ", len(stack))

		// Determine tree character
		var treeChar string
		if len(stack) == 0 {
			treeChar = "└"
		} else {
			treeChar = "│"
		}

		// Format line
		lineInfo := fmt.Sprintf("H%d:%d", level, entry.Line)
		formatted := fmt.Sprintf(
			"%s%s %s %s",
			indent,
			treeChar,
			entry.Name,
			lineInfo,
		)
		lines = append(lines, formatted)

		// Push to stack for children
		stack = append(stack, stackEntry{Level: level, Entry: entry})
	}

	// Add filename as root
	if len(entries) > 0 {
		filename := filepath.Base(entries[0].File)
		result := fmt.Sprintf("%s\n\n%s", filename, strings.Join(lines, "\n"))
		return result
	}

	return strings.Join(lines, "\n")
}

// stackEntry is used for tracking parent entries while building the tree.
type stackEntry struct {
	Level int
	Entry *TagEntry
}
