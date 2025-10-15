package tools

import (
	"strings"
	"testing"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// TestCalculateEndLine_DepthNil tests that when depth is nil,
// the original end line is returned (unlimited depth).
func TestCalculateEndLine_DepthNil(t *testing.T) {
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

// TestCalculateEndLine_Depth0 tests that when depth is 0,
// only the section content is read (no subsections).
func TestCalculateEndLine_Depth0(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Subsection 1.2", Line: 15, Level: 3},
		{Name: "Section 2", Line: 25, Level: 2},
	}

	depth := 0
	result := calculateEndLine(entries, 5, 24, &depth)

	// Should stop before line 10 (first child at level 3)
	if result != 9 {
		t.Errorf("Expected end line 9, got %d", result)
	}
}

// TestCalculateEndLine_Depth1 tests that when depth is 1,
// immediate children are included but not grandchildren.
// ALL H3 siblings should be included, even if there are H4 sections between them.
func TestCalculateEndLine_Depth1(t *testing.T) {
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

	depth := 1
	result := calculateEndLine(entries, 5, 29, &depth)

	// Should include both H3 sections (lines 10-29), stopping before H2 at line 30
	// The H4 at line 15 is skipped but doesn't stop reading
	if result != 29 {
		t.Errorf(
			"Expected end line 29 (include all H3 siblings), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_Depth2 tests that when depth is 2,
// children and grandchildren are included but not great-grandchildren.
func TestCalculateEndLine_Depth2(t *testing.T) {
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

	depth := 2
	result := calculateEndLine(entries, 5, 29, &depth)

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
		name     string
		depth    *int
		expected int
	}{
		{
			name:     "depth_nil_unlimited",
			depth:    nil,
			expected: 29, // Everything up to Section 2
		},
		{
			name:     "depth_0_no_subsections",
			depth:    intPtr(0),
			expected: 9, // Stop before first H3
		},
		{
			name:     "depth_1_immediate_children",
			depth:    intPtr(1),
			expected: 29, // Include all H3 sections (skip H4/H5 content but continue)
		},
		{
			name:     "depth_2_grandchildren",
			depth:    intPtr(2),
			expected: 29, // Include H3 and H4 sections (skip H5 content but continue)
		},
		{
			name:     "depth_3_great_grandchildren",
			depth:    intPtr(3),
			expected: 29, // Include all (H5 is last child)
		},
		{
			name:     "depth_100_all_included",
			depth:    intPtr(100),
			expected: 29, // Same as unlimited
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateEndLine(entries, 5, 29, tt.depth)
			if result != tt.expected {
				t.Errorf(
					"depth=%v: expected end line %d, got %d",
					derefInt(tt.depth),
					tt.expected,
					result,
				)
			}
		})
	}
}

// TestCalculateEndLine_NoChildren tests that when a section has no children,
// the original end line is returned regardless of depth.
func TestCalculateEndLine_NoChildren(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	tests := []struct {
		name  string
		depth *int
	}{
		{"depth_nil", nil},
		{"depth_0", intPtr(0)},
		{"depth_1", intPtr(1)},
		{"depth_2", intPtr(2)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateEndLine(entries, 5, 19, tt.depth)

			// No children, so original end line for all depths
			if result != 19 {
				t.Errorf(
					"depth=%v: expected end line 19, got %d",
					derefInt(tt.depth),
					result,
				)
			}
		})
	}
}

// TestCalculateEndLine_DepthNegative tests that negative depth values
// are treated as 0 (no subsections).
func TestCalculateEndLine_DepthNegative(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Section 2", Line: 20, Level: 2},
	}

	depth := -5
	result := calculateEndLine(entries, 5, 19, &depth)

	// Negative depth should behave like depth=0
	if result != 9 {
		t.Errorf("Expected end line 9 for negative depth, got %d", result)
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

	depth := 5
	result := calculateEndLine(entries, 5, 100, &depth)

	// Even with high depth, should stop at next sibling (Section 2)
	if result != 19 {
		t.Errorf("Expected end line 19 (before sibling), got %d", result)
	}
}

// TestCalculateEndLine_H1Section tests depth calculation for H1 sections.
func TestCalculateEndLine_H1Section(t *testing.T) {
	t.Parallel()

	entries := []*ctags.TagEntry{
		{Name: "Main Title", Line: 1, Level: 1},
		{Name: "Section 1", Line: 5, Level: 2},
		{Name: "Subsection 1.1", Line: 10, Level: 3},
		{Name: "Section 2", Line: 15, Level: 2},
	}

	depth := 1
	result := calculateEndLine(entries, 1, 19, &depth)

	// H1 with depth=1 should include ALL H2 sections (both Section 1 and Section 2)
	// H3 subsections exceed depth but we continue past them to find more H2s
	// Should include up to line 19 (end of last H2 section)
	if result != 19 {
		t.Errorf(
			"Expected end line 19 (include all H2 sections), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_LastSection tests depth calculation for the last
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
		name     string
		depth    *int
		expected int
	}{
		{
			name:     "depth_nil",
			depth:    nil,
			expected: 100, // No limit
		},
		{
			name:     "depth_0",
			depth:    intPtr(0),
			expected: 24, // Stop before H3
		},
		{
			name:     "depth_1",
			depth:    intPtr(1),
			expected: 100, // Include all H3 sections (H4s are skipped but don't stop reading)
		},
		{
			name:     "depth_2",
			depth:    intPtr(2),
			expected: 100, // Include everything
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateEndLine(entries, 20, 100, tt.depth)
			if result != tt.expected {
				t.Errorf(
					"depth=%v: expected end line %d, got %d",
					derefInt(tt.depth),
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

// TestCalculateEndLine_Depth1_MultipleSiblings tests that depth=1
// includes ALL H3 siblings, not just the first one.
func TestCalculateEndLine_Depth1_MultipleSiblings(t *testing.T) {
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

	depth := 1
	result := calculateEndLine(entries, 5, 29, &depth)

	// Should include both H3 sections, stopping before next H2
	// Expected: line 29 (before "Section 2" at line 30)
	if result != 29 {
		t.Errorf(
			"Expected end line 29 (include both H3 siblings), got %d",
			result,
		)
	}
}

// TestCalculateEndLine_Depth1_RealWorld simulates the actual bug scenario.
func TestCalculateEndLine_Depth1_RealWorld(t *testing.T) {
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

	depth := 1
	result := calculateEndLine(entries, 2123, 2168, &depth)

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

// TestFilterContentByDepth_Depth0 tests filtering with depth=0.
func TestFilterContentByDepth_Depth0(t *testing.T) {
	t.Parallel()

	content := `## Testing Strategy
Intro text for testing strategy

### Test Coverage Requirements
This should be excluded

### Test Data Scenarios
This should also be excluded`

	expected := `## Testing Strategy
Intro text for testing strategy`

	result := filterContentByDepth(2, 0, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByDepth_Depth1_MultipleSiblings tests the key bug scenario:
// depth=1 should include ALL H3 siblings, not just the first one.
func TestFilterContentByDepth_Depth1_MultipleSiblings(t *testing.T) {
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

	result := filterContentByDepth(2, 1, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByDepth_Depth2 tests filtering with depth=2.
func TestFilterContentByDepth_Depth2(t *testing.T) {
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

	result := filterContentByDepth(2, 2, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByDepth_NoSubsections tests content with no subsections.
func TestFilterContentByDepth_NoSubsections(t *testing.T) {
	t.Parallel()

	content := `## Section 1
Just some content
No subsections here`

	// With depth=0, should return all content (since there are no subsections)
	result := filterContentByDepth(2, 0, content)

	if result != content {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			content,
			result,
		)
	}
}

// TestFilterContentByDepth_EmptyContent tests empty content.
func TestFilterContentByDepth_EmptyContent(t *testing.T) {
	t.Parallel()

	content := ""
	result := filterContentByDepth(2, 1, content)

	if result != "" {
		t.Errorf("Expected empty string, got: %s", result)
	}
}

// TestFilterContentByDepth_ComplexNesting tests complex nested structure.
func TestFilterContentByDepth_ComplexNesting(t *testing.T) {
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
		name     string
		depth    int
		expected string
	}{
		{
			name:  "depth_0",
			depth: 0,
			expected: `## Main Section
Main content`,
		},
		{
			name:  "depth_1",
			depth: 1,
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
			name:  "depth_2",
			depth: 2,
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
			name:  "depth_3",
			depth: 3,
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
			result := filterContentByDepth(2, tt.depth, content)
			if result != tt.expected {
				t.Errorf(
					"depth=%d:\nExpected:\n%s\n\nGot:\n%s",
					tt.depth,
					tt.expected,
					result,
				)
			}
		})
	}
}

// TestFilterContentByDepth_H1Root tests filtering with H1 as root.
func TestFilterContentByDepth_H1Root(t *testing.T) {
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

	result := filterContentByDepth(1, 1, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByDepth_PreservesBlankLines tests that blank lines
// are preserved in included content.
func TestFilterContentByDepth_PreservesBlankLines(t *testing.T) {
	t.Parallel()

	content := `## Section

Content line 1

Content line 2

### Subsection

Subsection content`

	expected := `## Section

Content line 1

Content line 2`

	result := filterContentByDepth(2, 0, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByDepth_NegativeDepth tests negative depth values.
func TestFilterContentByDepth_NegativeDepth(t *testing.T) {
	t.Parallel()

	content := `## Section
Content

### Subsection
Subsection content`

	expected := `## Section
Content`

	// Negative depth should behave like depth=0
	result := filterContentByDepth(2, -1, content)

	if result != expected {
		t.Errorf(
			"Expected:\n%s\n\nGot:\n%s",
			expected,
			result,
		)
	}
}

// TestFilterContentByDepth_RealWorldScenario simulates the actual bug scenario
// from the Testing Strategy section.
func TestFilterContentByDepth_RealWorldScenario(t *testing.T) {
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

	result := filterContentByDepth(2, 1, content)

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
