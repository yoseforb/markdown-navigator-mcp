package ctags

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildTreeStructure(t *testing.T) {
	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	cache := GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	tree := BuildTreeStructure(entries)

	if tree == "" {
		t.Fatal("Expected non-empty tree")
	}

	// Check that tree starts with filename
	if !strings.HasPrefix(tree, "sample.md") {
		t.Errorf("Expected tree to start with 'sample.md', got: %s", tree[:20])
	}

	// Check that tree contains expected sections
	expectedSections := []string{
		"Test Document H1:1",
		"Section 1: Introduction H2:5",
		"Section 2: Implementation H2:21",
		"Section 3: Conclusion H2:44",
		"Subsection 1.1: Background H3:10",
		"Subsection 2.2: Testing H3:32",
		"Deep Section 2.2.1: Unit Tests H4:36",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(tree, expected) {
			t.Errorf("Expected tree to contain %q, but it doesn't", expected)
		}
	}

	// Check that tree uses proper indentation characters
	if !strings.Contains(tree, "└") {
		t.Error("Expected tree to contain '└' character")
	}

	if !strings.Contains(tree, "│") {
		t.Error("Expected tree to contain '│' character")
	}

	// Verify the tree has proper structure with indentation
	lines := strings.Split(tree, "\n")
	foundIndented := false
	for _, line := range lines {
		if strings.HasPrefix(line, "  ") {
			foundIndented = true
			break
		}
	}

	if !foundIndented {
		t.Error("Expected tree to have indented lines for hierarchy")
	}
}

func TestBuildTreeStructureEmptyEntries(t *testing.T) {
	entries := []*TagEntry{}
	tree := BuildTreeStructure(entries)

	if tree != "" {
		t.Errorf("Expected empty string for empty entries, got %q", tree)
	}
}

func TestBuildTreeStructureSingleEntry(t *testing.T) {
	entries := []*TagEntry{
		NewTagEntry(
			"Single Section",
			"test.md",
			"/pattern/",
			"section",
			10,
			20,
			"",
		),
	}

	tree := BuildTreeStructure(entries)

	if !strings.Contains(tree, "test.md") {
		t.Error("Expected tree to contain filename")
	}

	if !strings.Contains(tree, "Single Section H2:10") {
		t.Error("Expected tree to contain section name with line number")
	}
}

func TestBuildTreeStructureHierarchy(t *testing.T) {
	// Create a simple hierarchy: H1 > H2 > H3
	entries := []*TagEntry{
		NewTagEntry("Chapter", "test.md", "/pattern/", "chapter", 1, 4, ""),
		NewTagEntry(
			"Section",
			"test.md",
			"/pattern/",
			"section",
			5,
			9,
			"Chapter",
		),
		NewTagEntry(
			"Subsection",
			"test.md",
			"/pattern/",
			"subsection",
			10,
			15,
			"Chapter\"\"Section",
		),
	}

	tree := BuildTreeStructure(entries)

	lines := strings.Split(tree, "\n")

	// Find the lines with our sections
	var chapterLine, sectionLine, subsectionLine string
	for _, line := range lines {
		if strings.Contains(line, "Chapter H1:1") {
			chapterLine = line
		}
		if strings.Contains(line, "Section H2:5") {
			sectionLine = line
		}
		if strings.Contains(line, "Subsection H3:10") {
			subsectionLine = line
		}
	}

	// Check that Section is indented more than Chapter
	chapterIndent := len(chapterLine) - len(strings.TrimLeft(chapterLine, " "))
	sectionIndent := len(sectionLine) - len(strings.TrimLeft(sectionLine, " "))
	subsectionIndent := len(
		subsectionLine,
	) - len(
		strings.TrimLeft(subsectionLine, " "),
	)

	if sectionIndent <= chapterIndent {
		t.Errorf(
			"Expected Section to be more indented than Chapter (section: %d, chapter: %d)",
			sectionIndent,
			chapterIndent,
		)
	}

	if subsectionIndent <= sectionIndent {
		t.Errorf(
			"Expected Subsection to be more indented than Section (subsection: %d, section: %d)",
			subsectionIndent,
			sectionIndent,
		)
	}
}
