package tools

import (
	"strings"
	"testing"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// TestCalculateEndLine_MaxSubsectionLevelsNil tests that when maxSubsectionLevels is nil,
// the original end line is returned (unlimited depth).
func TestCalculateEndLine_MaxSubsectionLevelsNil(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Subsection 1.2", Line: 15, Level: 3},
		{Name: "Deep 1.2.1", Line: 18, Level: 4},
		{Name: "Section 2", Line: 25, Level: 2},
	}

	result := calculateEndLine(entries, 5, 24, nil)

	if result != 24 {
		t.Errorf("Expected end line 24, got %d", result)
	}
}

// TestCalculateEndLine_MaxSubsectionLevels0 tests that when maxSubsectionLevels is 0,
// only the section content is read (no subsections).
func TestCalculateEndLine_MaxSubsectionLevels0(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Subsection 1.2", Line: 15, Level: 3},
		{Name: "Section 2", Line: 25, Level: 2},
	}

	maxSubsectionLevels := 0
	result := calculateEndLine(entries, 5, 24, &maxSubsectionLevels)

	// Should stop before line 10 (first child at level 3)
	if result != 9 {
		t.Errorf("Expected end line 9, got %d", result)
	}
}

// TestCalculateEndLine_MaxSubsectionLevels1 tests that when maxSubsectionLevels is 1,
// immediate children are included but not grandchildren.
// ALL H3 siblings should be included, even if there are H4 sections between them.
func TestCalculateEndLine_MaxSubsectionLevels1(t *testing.T) {
	t.Parallel()

	// H2 at line 5
	// H3 at line 10 (first child - include)
	// H4 at line 15 (exceeds depth - skip but continue)
	// H3 at line 20 (second child - include!)
	// H2 at line 30 (sibling - stop here)
	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Deep 1.1.1", Line: 15, Level: 4},
		{Name: "Subsection 1.2", Line: 20, Level: 3},
		{Name: "Section 2", Line: 30, Level: 2},
	}

	maxSubsectionLevels := 1
	result := calculateEndLine(entries, 5, 29, &maxSubsectionLevels)

	// Should include both H3 sections (lines 10-29), stopping before H2 at line 30
	// The H4 at line 15 is skipped but doesn't stop reading
	if result != 29 {
		t.Errorf(
			"Expected end line 29 (include all H3 siblings), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_MaxSubsectionLevels2 tests that when maxSubsectionLevels is 2,
// children and grandchildren are included but not great-grandchildren.
func TestCalculateEndLine_MaxSubsectionLevels2(t *testing.T) {
	t.Parallel()

	// H2 at line 5
	// H3 at line 10 (child, include)
	// H4 at line 15 (grandchild, include)
	// H5 at line 20 (great-grandchild, exceeds depth, skip but continue)
	// H2 at line 30 (sibling, stop)
	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Deep 1.1.1", Line: 15, Level: 4},
		{Name: "Deeper 1.1.1.1", Line: 20, Level: 5},
		{Name: "Section 2", Line: 30, Level: 2},
	}

	maxSubsectionLevels := 2
	result := calculateEndLine(entries, 5, 29, &maxSubsectionLevels)

	// Should include all H3 and H4 sections, continuing past H5 to check for more
	// Last allowed section (H4 at line 15) ends at line 29 (before next H2)
	if result != 29 {
		t.Errorf(
			"Expected end line 29 (include all H3/H4 sections), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_DeepNesting tests handling of deeply nested sections
// with various depth values.
func TestCalculateEndLine_DeepNesting(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Deep 1.1.1", Line: 15, Level: 4},
		{Name: "Deeper 1.1.1.1", Line: 20, Level: 5},
		{Name: "Section 2", Line: 30, Level: 2},
	}

	tests := []struct {
		name                string
		maxSubsectionLevels *int
		expected            int
	}{
		{
			name:                "maxSubsectionLevels_nil_unlimited",
			maxSubsectionLevels: nil,
			expected:            29, // Everything up to Section 2
		},
		{
			name:                "maxSubsectionLevels_0_no_subsections",
			maxSubsectionLevels: intPtr(0),
			expected:            9, // Stop before first H3
		},
		{
			name:                "maxSubsectionLevels_1_immediate_children",
			maxSubsectionLevels: intPtr(1),
			expected:            29, // Include all H3 sections (skip H4/H5 content but continue)
		},
		{
			name:                "maxSubsectionLevels_2_grandchildren",
			maxSubsectionLevels: intPtr(2),
			expected:            29, // Include H3 and H4 sections (skip H5 content but continue)
		},
		{
			name:                "maxSubsectionLevels_3_great_grandchildren",
			maxSubsectionLevels: intPtr(3),
			expected:            29, // Include all (H5 is last child)
		},
		{
			name:                "maxSubsectionLevels_100_all_included",
			maxSubsectionLevels: intPtr(100),
			expected:            29, // Same as unlimited
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateEndLine(entries, 5, 29, tt.maxSubsectionLevels)
			if result != tt.expected {
				t.Errorf(
					"maxSubsectionLevels=%v: expected end line %d, got %d",
					derefInt(tt.maxSubsectionLevels),
					tt.expected,
					result,
				)
			}
		})
	}
}

// TestCalculateEndLine_NoChildren tests that when a section has no children,
// the original end line is returned regardless of maxSubsectionLevels.
func TestCalculateEndLine_NoChildren(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	tests := []struct {
		name                string
		maxSubsectionLevels *int
	}{
		{"maxSubsectionLevels_nil", nil},
		{"maxSubsectionLevels_0", intPtr(0)},
		{"maxSubsectionLevels_1", intPtr(1)},
		{"maxSubsectionLevels_2", intPtr(2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateEndLine(entries, 5, 19, tt.maxSubsectionLevels)

			// No children, so original end line for all depths
			if result != 19 {
				t.Errorf(
					"maxSubsectionLevels=%v: expected end line 19, got %d",
					derefInt(tt.maxSubsectionLevels),
					result,
				)
			}
		})
	}
}

// TestCalculateEndLine_MaxSubsectionLevelsNegative tests that negative maxSubsectionLevels values
// are treated as 0 (no subsections).
func TestCalculateEndLine_MaxSubsectionLevelsNegative(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	maxSubsectionLevels := -5
	result := calculateEndLine(entries, 5, 19, &maxSubsectionLevels)

	// Negative maxSubsectionLevels should behave like maxSubsectionLevels=0
	if result != 9 {
		t.Errorf(
			"Expected end line 9 for negative maxSubsectionLevels, got %d",
			result,
		)
	}
}

// TestCalculateEndLine_StopsAtSibling tests that siblings at the same level
// are respected as boundaries.
func TestCalculateEndLine_StopsAtSibling(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Subsection 1.2", Line: 15, Level: 3},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	maxSubsectionLevels := 5
	result := calculateEndLine(entries, 5, 100, &maxSubsectionLevels)

	// Even with high maxSubsectionLevels, should stop at next sibling (Section 2)
	if result != 19 {
		t.Errorf("Expected end line 19 (before sibling), got %d", result)
	}
}

// TestCalculateEndLine_H1Section tests maxSubsectionLevels calculation for H1 sections.
func TestCalculateEndLine_H1Section(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Main Title", Line: 1, Level: 1},
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Section 2", Line: 15, Level: 2},
	}

	maxSubsectionLevels := 1
	result := calculateEndLine(entries, 1, 19, &maxSubsectionLevels)

	// H1 with maxSubsectionLevels=1 should include ALL H2 sections (both Section 1 and Section 2)
	// H3 subsections exceed depth but we continue past them to find more H2s
	// Should include up to line 19 (end of last H2 section)
	if result != 19 {
		t.Errorf(
			"Expected end line 19 (include all H2 sections), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_LastSection tests maxSubsectionLevels calculation for the last
// section in a file (no following sections).
func TestCalculateEndLine_LastSection(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Section 2", Line: 20, Level: 2},
		{Name: "Subsection 2.1", Line: 25, Level: 3},
		{Name: "Deep 2.1.1", Line: 30, Level: 4},
	}

	tests := []struct {
		name                string
		maxSubsectionLevels *int
		expected            int
	}{
		{
			name:                "maxSubsectionLevels_nil",
			maxSubsectionLevels: nil,
			expected:            100, // No limit
		},
		{
			name:                "maxSubsectionLevels_0",
			maxSubsectionLevels: intPtr(0),
			expected:            24, // Stop before H3
		},
		{
			name:                "maxSubsectionLevels_1",
			maxSubsectionLevels: intPtr(1),
			expected:            100, // Include all H3 sections (H4s are skipped but don't stop reading)
		},
		{
			name:                "maxSubsectionLevels_2",
			maxSubsectionLevels: intPtr(2),
			expected:            100, // Include everything
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateEndLine(entries, 20, 100, tt.maxSubsectionLevels)
			if result != tt.expected {
				t.Errorf(
					"maxSubsectionLevels=%v: expected end line %d, got %d",
					derefInt(tt.maxSubsectionLevels),
					tt.expected,
					result,
				)
			}
		})
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

// TestFindSectionLevel_NotFound tests that 0 is returned when section not
// found.
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

// TestCalculateEndLine_MaxSubsectionLevels1_MultipleSiblings tests that maxSubsectionLevels=1
// includes ALL H3 siblings, not just the first one.
func TestCalculateEndLine_MaxSubsectionLevels1_MultipleSiblings(t *testing.T) {
	t.Parallel()

	// H2 at line 5
	// H3 at line 10 (first child)
	// H4 at line 15 (exceeds depth, should skip)
	// H3 at line 20 (second child - MUST include!)
	// H4 at line 25 (exceeds depth, should skip)
	// H2 at line 30
	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Child 1", Line: 10, Level: 3},
		{Name: "Grandchild 1", Line: 15, Level: 4},
		{Name: "Child 2", Line: 20, Level: 3},
		{Name: "Grandchild 2", Line: 25, Level: 4},
		{Name: "Section 2", Line: 30, Level: 2},
	}

	maxSubsectionLevels := 1
	result := calculateEndLine(entries, 5, 29, &maxSubsectionLevels)

	// Should include both H3 sections, stopping before next H2
	// Expected: line 29 (before "Section 2" at line 30)
	if result != 29 {
		t.Errorf(
			"Expected end line 29 (include both H3 siblings), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_MaxSubsectionLevels1_RealWorld simulates the actual bug scenario.
func TestCalculateEndLine_MaxSubsectionLevels1_RealWorld(t *testing.T) {
	t.Parallel()

	// Simulating the "Testing Strategy" section structure
	entries := []*ctags.TagEntry{
		{Name: "Testing Strategy", Line: 2123, Level: 2},
		{Name: "Test Coverage Requirements", Line: 2125, Level: 3},
		{Name: "Unit Tests", Line: 2130, Level: 4},
		{Name: "Integration Tests", Line: 2135, Level: 4},
		{Name: "Test Data Scenarios", Line: 2145, Level: 3},
		{Name: "Migration Tests", Line: 2150, Level: 4},
		{Name: "Next Section", Line: 2169, Level: 2},
	}

	maxSubsectionLevels := 1
	result := calculateEndLine(entries, 2123, 2168, &maxSubsectionLevels)

	// Should read all H3 sections (both Test Coverage and Test Data)
	// Stopping before next H2 at line 2169
	expected := 2168
	if result != expected {
		t.Errorf(
			"Expected end line %d (include all H3 siblings), got %d",
			expected,
			result,
		)
	}
}

// Helper function to create int pointer.
func intPtr(val int) *int {
	return &val
}

// Helper function to dereference int pointer safely for display.
func derefInt(ptr *int) string {
	if ptr == nil {
		return "nil"
	}
	return string(rune(*ptr + '0'))
}

// TestFilterContentByMaxSubsectionLevels_MaxSubsectionLevels0 tests filtering with maxSubsectionLevels=0.
func TestFilterContentByMaxSubsectionLevels_MaxSubsectionLevels0(t *testing.T) {
	t.Parallel()

	content := `## Testing Strategy
Intro text for testing strategy

### Test Coverage Requirements
This should be excluded

### Test Data Scenarios
This should also be excluded`

	expected := `## Testing Strategy
Intro text for testing strategy`

	result := filterContentByMaxSubsectionLevels(2, 0, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_MaxSubsectionLevels1_MultipleSiblings tests the key bug scenario:
// maxSubsectionLevels=1 should include ALL H3 siblings, not just the first one.
func TestFilterContentByMaxSubsectionLevels_MaxSubsectionLevels1_MultipleSiblings(
	t *testing.T,
) {
	t.Parallel()

	content := `## Testing Strategy
Intro text

### Test Coverage Requirements
Coverage content here
More coverage content

#### Unit Tests
This H4 should be excluded

### Test Data Scenarios
Scenario content here
More scenario content

#### Migration Tests
This H4 should also be excluded`

	expected := `## Testing Strategy
Intro text

### Test Coverage Requirements
Coverage content here
More coverage content

### Test Data Scenarios
Scenario content here
More scenario content`

	result := filterContentByMaxSubsectionLevels(2, 1, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_MaxSubsectionLevels2 tests filtering with maxSubsectionLevels=2.
func TestFilterContentByMaxSubsectionLevels_MaxSubsectionLevels2(t *testing.T) {
	t.Parallel()

	content := `## Section 1
Root content

### Subsection 1.1
Child content

#### Deep 1.1.1
Grandchild content

##### Deeper 1.1.1.1
This H5 should be excluded`

	expected := `## Section 1
Root content

### Subsection 1.1
Child content

#### Deep 1.1.1
Grandchild content`

	result := filterContentByMaxSubsectionLevels(2, 2, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_NoSubsections tests content with no subsections.
func TestFilterContentByMaxSubsectionLevels_NoSubsections(t *testing.T) {
	t.Parallel()

	content := `## Section 1
Just some content
No subsections here`

	// With maxSubsectionLevels=0, should return all content (since there are no subsections)
	result := filterContentByMaxSubsectionLevels(2, 0, content)

	if result != content {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			content,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_EmptyContent tests empty content.
func TestFilterContentByMaxSubsectionLevels_EmptyContent(t *testing.T) {
	t.Parallel()

	content := ""
	result := filterContentByMaxSubsectionLevels(2, 1, content)

	if result != "" {
		t.Errorf("Expected empty string, got: %s", result)
	}
}

// TestFilterContentByMaxSubsectionLevels_ComplexNesting tests complex nested structure.
func TestFilterContentByMaxSubsectionLevels_ComplexNesting(t *testing.T) {
	t.Parallel()

	content := `## Main Section
Main content

### Child 1
Child 1 content

#### Grandchild 1.1
GC 1.1 content

##### GreatGrandchild 1.1.1
GGC 1.1.1 content

### Child 2
Child 2 content

#### Grandchild 2.1
GC 2.1 content

### Child 3
Child 3 content`

	tests := []struct {
		name                string
		maxSubsectionLevels int
		expected            string
	}{
		{
			name:                "maxSubsectionLevels_0",
			maxSubsectionLevels: 0,
			expected: `## Main Section
Main content`,
		},
		{
			name:                "maxSubsectionLevels_1",
			maxSubsectionLevels: 1,
			expected: `## Main Section
Main content

### Child 1
Child 1 content

### Child 2
Child 2 content

### Child 3
Child 3 content`,
		},
		{
			name:                "maxSubsectionLevels_2",
			maxSubsectionLevels: 2,
			expected: `## Main Section
Main content

### Child 1
Child 1 content

#### Grandchild 1.1
GC 1.1 content

### Child 2
Child 2 content

#### Grandchild 2.1
GC 2.1 content

### Child 3
Child 3 content`,
		},
		{
			name:                "maxSubsectionLevels_3",
			maxSubsectionLevels: 3,
			expected: `## Main Section
Main content

### Child 1
Child 1 content

#### Grandchild 1.1
GC 1.1 content

##### GreatGrandchild 1.1.1
GGC 1.1.1 content

### Child 2
Child 2 content

#### Grandchild 2.1
GC 2.1 content

### Child 3
Child 3 content`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := filterContentByMaxSubsectionLevels(
				2,
				tt.maxSubsectionLevels,
				content,
			)
			if result != tt.expected {
				t.Errorf(
					"maxSubsectionLevels=%d:\nExpected:\n%s\n\nGot:\n%s",
					tt.maxSubsectionLevels,
					tt.expected,
					result,
				)
			}
		})
	}
}

// TestFilterContentByMaxSubsectionLevels_H1Root tests filtering with H1 as root.
func TestFilterContentByMaxSubsectionLevels_H1Root(t *testing.T) {
	t.Parallel()

	content := `# Document Title
Document intro

## Section 1
Section 1 content

### Subsection 1.1
Subsection content

## Section 2
Section 2 content`

	expected := `# Document Title
Document intro

## Section 1
Section 1 content

## Section 2
Section 2 content`

	result := filterContentByMaxSubsectionLevels(1, 1, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_PreservesBlankLines tests that blank lines
// are preserved in included content.
func TestFilterContentByMaxSubsectionLevels_PreservesBlankLines(t *testing.T) {
	t.Parallel()

	content := `## Section

Content line 1

Content line 2

### Subsection

Subsection content`

	expected := `## Section

Content line 1

Content line 2`

	result := filterContentByMaxSubsectionLevels(2, 0, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_NegativeMaxSubsectionLevels tests negative maxSubsectionLevels values.
func TestFilterContentByMaxSubsectionLevels_NegativeMaxSubsectionLevels(
	t *testing.T,
) {
	t.Parallel()

	content := `## Section
Content

### Subsection
Subsection content`

	expected := `## Section
Content`

	// Negative maxSubsectionLevels should behave like maxSubsectionLevels=0
	result := filterContentByMaxSubsectionLevels(2, -1, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByMaxSubsectionLevels_RealWorldScenario simulates the actual bug scenario
// from the Testing Strategy section.
func TestFilterContentByMaxSubsectionLevels_RealWorldScenario(t *testing.T) {
	t.Parallel()

	content := `## Testing Strategy
Comprehensive testing strategy for the refactoring.

### Test Coverage Requirements
Coverage goals and metrics.

#### Unit Tests
Unit test details that should be excluded with depth=1.

#### Integration Tests
Integration test details that should be excluded with depth=1.

### Test Data Scenarios
Test data requirements and scenarios.

#### Migration Tests
Migration test details that should be excluded with depth=1.`

	expected := `## Testing Strategy
Comprehensive testing strategy for the refactoring.

### Test Coverage Requirements
Coverage goals and metrics.

### Test Data Scenarios
Test data requirements and scenarios.`

	result := filterContentByMaxSubsectionLevels(2, 1, content)

	// Normalize whitespace for comparison
	normalizeWS := func(s string) string {
		return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	}

	if normalizeWS(result) != normalizeWS(expected) {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}
