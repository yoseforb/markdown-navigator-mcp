package tools

import (
	"testing"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// TestCalculateEndLine_IncludeSubsectionsTrue tests that when
// include_subsections is true, the original end line is returned.
func TestCalculateEndLine_IncludeSubsectionsTrue(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Subsection 1.2", Line: 15, Level: 3},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	includeSubsections := true
	result := calculateEndLine(entries, 5, 19, &includeSubsections)

	if result != 19 {
		t.Errorf("Expected end line 19, got %d", result)
	}
}

// TestCalculateEndLine_IncludeSubsectionsFalse tests that when
// include_subsections is false, the end line stops before first child.
func TestCalculateEndLine_IncludeSubsectionsFalse(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Subsection 1.2", Line: 15, Level: 3},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	includeSubsections := false
	result := calculateEndLine(entries, 5, 19, &includeSubsections)

	// Should stop before line 10 (first child at level 3)
	if result != 9 {
		t.Errorf("Expected end line 9, got %d", result)
	}
}

// TestCalculateEndLine_DefaultTrue tests that when include_subsections is nil,
// it defaults to true (includes subsections).
func TestCalculateEndLine_DefaultTrue(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	result := calculateEndLine(entries, 5, 19, nil)

	// Default behavior: include subsections
	if result != 19 {
		t.Errorf("Expected end line 19, got %d", result)
	}
}

// TestCalculateEndLine_NoChildren tests that when a section has no children
// and include_subsections is false, the original end line is returned.
func TestCalculateEndLine_NoChildren(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	includeSubsections := false
	result := calculateEndLine(entries, 5, 19, &includeSubsections)

	// No children, so original end line
	if result != 19 {
		t.Errorf("Expected end line 19, got %d", result)
	}
}

// TestCalculateEndLine_DeepNesting tests handling of deeply nested sections.
func TestCalculateEndLine_DeepNesting(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Deep Section 1.1.1", Line: 15, Level: 4},
		{Name: "Section 2", Line: 30, Level: 2},
	}

	includeSubsections := false
	result := calculateEndLine(entries, 5, 29, &includeSubsections)

	// Should stop before line 10 (first child at level 3)
	if result != 9 {
		t.Errorf("Expected end line 9, got %d", result)
	}
}

// TestCalculateEndLine_SiblingOnly tests that siblings (same level) are not
// considered children when include_subsections is false.
func TestCalculateEndLine_SiblingOnly(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Section 2", Line: 10, Level: 2},
		{Name: "Section 3", Line: 15, Level: 2},
	}

	includeSubsections := false
	result := calculateEndLine(entries, 5, 9, &includeSubsections)

	// No children (only siblings), so original end line
	if result != 9 {
		t.Errorf("Expected end line 9, got %d", result)
	}
}

// TestFindSectionLevel_Found tests finding a section's level.
func TestFindSectionLevel_Found(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Section 2", Line: 10, Level: 3},
	}

	level := findSectionLevel(entries, 5)

	if level != 2 {
		t.Errorf("Expected level 2, got %d", level)
	}
}

// TestFindSectionLevel_NotFound tests that 0 is returned when section not found.
func TestFindSectionLevel_NotFound(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
	}

	level := findSectionLevel(entries, 999)

	if level != 0 {
		t.Errorf("Expected level 0 for not found, got %d", level)
	}
}

// TestFindSectionLevel_EmptyEntries tests handling of empty entries.
func TestFindSectionLevel_EmptyEntries(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{}

	level := findSectionLevel(entries, 5)

	if level != 0 {
		t.Errorf("Expected level 0 for empty entries, got %d", level)
	}
}
