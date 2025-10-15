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

		// Format line with end line if available
		var lineInfo string
		if entry.End > 0 {
			lineInfo = fmt.Sprintf("H%d:%d:%d", level, entry.Line, entry.End)
		} else {
			lineInfo = fmt.Sprintf("H%d:%d", level, entry.Line)
		}
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

// TreeNode represents a node in the hierarchical JSON tree structure.
type TreeNode struct {
	Name      string      `json:"name"`
	Level     string      `json:"level"`
	StartLine int         `json:"start_line"`
	EndLine   int         `json:"end_line"`
	Children  []*TreeNode `json:"children"`
}

// BuildTreeJSON builds a hierarchical JSON tree structure from tag entries.
// Returns the root node containing all sections and their children.
func BuildTreeJSON(entries []*TagEntry) *TreeNode {
	if len(entries) == 0 {
		return nil
	}

	// Create root node
	root := &TreeNode{
		Name:      filepath.Base(entries[0].File),
		Level:     "H0",
		StartLine: 0,
		EndLine:   0,
		Children:  []*TreeNode{},
	}

	// Stack to track parent nodes at each level
	stack := []*TreeNode{root}

	for _, entry := range entries {
		node := &TreeNode{
			Name:      entry.Name,
			Level:     fmt.Sprintf("H%d", entry.Level),
			StartLine: entry.Line,
			EndLine:   entry.End,
			Children:  []*TreeNode{},
		}

		// Pop stack to find correct parent (level-based)
		for len(stack) > 1 && getLevel(stack[len(stack)-1].Level) >= entry.Level {
			stack = stack[:len(stack)-1]
		}

		// Add as child of current parent
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)

		// Push current node for potential children
		stack = append(stack, node)
	}

	return root
}

// getLevel extracts numeric level from "H1", "H2", etc.
func getLevel(levelStr string) int {
	if len(levelStr) < 2 || levelStr[0] != 'H' {
		return 0
	}

	level := 0
	_, _ = fmt.Sscanf(levelStr[1:], "%d", &level)

	return level
}
