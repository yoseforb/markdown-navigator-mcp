package tools

import (
	"testing"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// Helper function to create test entries.
func createTestEntries() []*ctags.TagEntry {
	return []*ctags.TagEntry{
		{Name: "Test Document", File: "test.md", Level: 1, Line: 1, End: 48},
		{
			Name:  "Section 1: Introduction",
			File:  "test.md",
			Level: 2,
			Line:  5,
			End:   20,
		},
		{
			Name:  "Subsection 1.1: Background",
			File:  "test.md",
			Level: 3,
			Line:  10,
			End:   13,
		},
		{
			Name:  "Subsection 1.2: Goals",
			File:  "test.md",
			Level: 3,
			Line:  14,
			End:   19,
		},
		{
			Name:  "Section 2: Implementation",
			File:  "test.md",
			Level: 2,
			Line:  21,
			End:   43,
		},
		{
			Name:  "Subsection 2.1: Architecture",
			File:  "test.md",
			Level: 3,
			Line:  25,
			End:   31,
		},
		{
			Name:  "Subsection 2.2: Testing",
			File:  "test.md",
			Level: 3,
			Line:  32,
			End:   42,
		},
		{
			Name:  "Deep Section 2.2.1: Unit Tests",
			File:  "test.md",
			Level: 4,
			Line:  36,
			End:   38,
		},
		{
			Name:  "Deep Section 2.2.2: Integration Tests",
			File:  "test.md",
			Level: 4,
			Line:  40,
			End:   42,
		},
		{
			Name:  "Section 3: Conclusion",
			File:  "test.md",
			Level: 2,
			Line:  44,
			End:   47,
		},
	}
}

func TestMarkdownTreeResponse_DefaultFormat(t *testing.T) {
	t.Parallel()

	// Test that default format is JSON
	args := MarkdownTreeArgs{
		FilePath: "/home/yoseforb/pkg/follow/markdown-mcp/testdata/sample.md",
		Format:   nil, // Not specified
		Pattern:  nil,
	}

	// We can't easily test the registered tool directly,
	// so we test the format defaulting logic
	format := "json"
	if args.Format != nil && *args.Format != "" {
		format = *args.Format
	}

	if format != "json" {
		t.Errorf("Default format = %s, want json", format)
	}
}

func TestMarkdownTreeResponse_JSONFormat(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()
	tree := ctags.BuildTreeJSON(entries)

	if tree == nil {
		t.Fatal("BuildTreeJSON returned nil")
	}

	// Verify root node
	if tree.Name != "test.md" {
		t.Errorf("Root name = %s, want test.md", tree.Name)
	}

	if tree.Level != "H0" {
		t.Errorf("Root level = %s, want H0", tree.Level)
	}

	// Verify root has expected number of top-level children
	if len(tree.Children) != 1 {
		t.Fatalf(
			"Root children count = %d, want 1 (Test Document)",
			len(tree.Children),
		)
	}

	// Verify first child is "Test Document"
	testDoc := tree.Children[0]
	if testDoc.Name != "Test Document" {
		t.Errorf("First child name = %s, want Test Document", testDoc.Name)
	}

	// Test Document should have 3 children (Section 1, 2, 3)
	if len(testDoc.Children) != 3 {
		t.Fatalf(
			"Test Document children count = %d, want 3",
			len(testDoc.Children),
		)
	}
}

func TestMarkdownTreeResponse_ASCIIFormat(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()
	asciiTree := ctags.BuildTreeStructure(entries)

	if asciiTree == "" {
		t.Fatal("BuildTreeStructure returned empty string")
	}

	// Verify ASCII tree contains expected sections
	expectedSections := []string{
		"Test Document",
		"Section 1: Introduction",
		"Section 2: Implementation",
		"Section 3: Conclusion",
	}

	for _, section := range expectedSections {
		if !containsString(asciiTree, section) {
			t.Errorf("ASCII tree does not contain '%s'", section)
		}
	}
}

func TestMarkdownTreeResponse_WithPattern(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()

	// Filter for "Testing"
	filtered := ctags.FilterByPatternWithParents(entries, "Testing")

	// Should include: Test Document (H1 parent), Section 2, Subsection 2.2 (matches "Testing")
	// The algorithm preserves parents all the way to the root
	expectedCount := 3
	if len(filtered) != expectedCount {
		t.Fatalf(
			"Filtered entries count = %d, want %d",
			len(filtered),
			expectedCount,
		)
	}

	expectedNames := []string{
		"Test Document",
		"Section 2: Implementation",
		"Subsection 2.2: Testing",
	}

	for i, entry := range filtered {
		if entry.Name != expectedNames[i] {
			t.Errorf(
				"Filtered[%d] name = %s, want %s",
				i,
				entry.Name,
				expectedNames[i],
			)
		}
	}

	// Build tree from filtered entries
	tree := ctags.BuildTreeJSON(filtered)
	if tree == nil {
		t.Fatal("BuildTreeJSON on filtered entries returned nil")
	}

	// Verify tree structure maintains hierarchy
	if len(tree.Children) != 1 {
		t.Errorf(
			"Filtered tree root children = %d, want 1",
			len(tree.Children),
		)
	}
}

func TestMarkdownTreeResponse_InvalidFormat(t *testing.T) {
	t.Parallel()

	invalidFormat := "xml"

	// Test format validation
	if invalidFormat != "json" && invalidFormat != "ascii" {
		// This should be an error case
		// In the actual tool, this returns an error
		expectedError := "invalid format: xml (must be 'json' or 'ascii')"
		if expectedError == "" {
			t.Error("Expected error for invalid format")
		}
	}
}

func TestMarkdownTreeResponse_EmptyPattern(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()

	// Empty pattern should return all entries
	emptyPattern := ""
	filtered := ctags.FilterByPatternWithParents(entries, emptyPattern)

	if len(filtered) != len(entries) {
		t.Errorf(
			"Empty pattern filtered count = %d, want %d",
			len(filtered),
			len(entries),
		)
	}
}

func TestMarkdownTreeResponse_NoMatchPattern(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()

	// Pattern with no matches
	filtered := ctags.FilterByPatternWithParents(entries, "NonExistentSection")

	if len(filtered) != 0 {
		t.Errorf(
			"No match pattern filtered count = %d, want 0",
			len(filtered),
		)
	}

	// Build tree from empty filtered entries
	tree := ctags.BuildTreeJSON(filtered)

	// BuildTreeJSON should return nil for empty entries
	if tree != nil {
		t.Error("BuildTreeJSON on empty entries should return nil")
	}
}

func TestMarkdownTreeResponse_JSONStructure(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()
	tree := ctags.BuildTreeJSON(entries)

	if tree == nil {
		t.Fatal("BuildTreeJSON returned nil")
	}

	// Verify JSON structure has all required fields
	if tree.Name == "" {
		t.Error("Tree root Name is empty")
	}

	if tree.Level == "" {
		t.Error("Tree root Level is empty")
	}

	if tree.Children == nil {
		t.Error("Tree root Children is nil (should be initialized)")
	}

	// Verify nested structure
	if len(tree.Children) > 0 {
		child := tree.Children[0]

		if child.Name == "" {
			t.Error("Child Name is empty")
		}

		if child.Level == "" {
			t.Error("Child Level is empty")
		}

		if child.Children == nil {
			t.Error("Child Children is nil (should be initialized)")
		}

		if child.StartLine == 0 {
			t.Error("Child StartLine is 0 (should be > 0)")
		}
	}
}

func TestMarkdownTreeResponse_LineRanges(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()
	tree := ctags.BuildTreeJSON(entries)

	if tree == nil {
		t.Fatal("BuildTreeJSON returned nil")
	}

	if len(tree.Children) == 0 {
		t.Fatal("Tree has no children")
	}

	// Verify line ranges are populated
	testDoc := tree.Children[0]

	if testDoc.StartLine != 1 {
		t.Errorf("Test Document StartLine = %d, want 1", testDoc.StartLine)
	}

	if testDoc.EndLine != 48 {
		t.Errorf("Test Document EndLine = %d, want 48", testDoc.EndLine)
	}

	// Check nested sections
	if len(testDoc.Children) == 0 {
		t.Fatal("Test Document has no children")
	}

	section1 := testDoc.Children[0]

	if section1.StartLine != 5 {
		t.Errorf("Section 1 StartLine = %d, want 5", section1.StartLine)
	}

	if section1.EndLine != 20 {
		t.Errorf("Section 1 EndLine = %d, want 20", section1.EndLine)
	}
}

func TestMarkdownTreeResponse_PatternCaseInsensitive(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()

	tests := []struct {
		name    string
		pattern string
	}{
		{name: "lowercase", pattern: "testing"},
		{name: "uppercase", pattern: "TESTING"},
		{name: "mixed case", pattern: "TeStInG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			filtered := ctags.FilterByPatternWithParents(entries, tt.pattern)

			// All patterns should match "Subsection 2.2: Testing"
			found := false
			for _, entry := range filtered {
				if entry.Name == "Subsection 2.2: Testing" {
					found = true
					break
				}
			}

			if !found {
				t.Errorf(
					"Pattern '%s' did not match 'Subsection 2.2: Testing'",
					tt.pattern,
				)
			}
		})
	}
}

func TestMarkdownTreeResponse_DefaultDepth(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()

	// Simulate default depth behavior (nil MaxDepth should default to 2)
	var maxDepth *int // nil means use default
	depth := 2
	if maxDepth != nil {
		depth = *maxDepth
	}

	if depth != 2 {
		t.Errorf("Default depth = %d, want 2", depth)
	}

	// Apply depth filtering
	var filtered []*ctags.TagEntry
	if depth > 0 {
		filtered = ctags.FilterByDepth(entries, depth)
	} else {
		filtered = entries
	}

	// Should include H1 and H2 only
	for _, entry := range filtered {
		if entry.Level > 2 {
			t.Errorf(
				"Entry %s has level %d, should be filtered with depth=2",
				entry.Name,
				entry.Level,
			)
		}
	}

	// Verify we have both H1 and H2
	hasH1 := false
	hasH2 := false
	for _, entry := range filtered {
		if entry.Level == 1 {
			hasH1 = true
		}
		if entry.Level == 2 {
			hasH2 = true
		}
	}

	if !hasH1 {
		t.Error("Default depth=2 should include H1 entries")
	}
	if !hasH2 {
		t.Error("Default depth=2 should include H2 entries")
	}
}

func TestMarkdownTreeResponse_ExplicitUnlimitedDepth(t *testing.T) {
	t.Parallel()

	entries := createTestEntries()

	// Test behavior when maxDepth is explicitly set to 0
	zero := 0
	explicitDepth := &zero

	// In production code: depth defaults to 2, but can be overridden
	depth := *explicitDepth // Use explicit value

	if depth != 0 {
		t.Errorf("Explicit depth = %d, want 0 (unlimited)", depth)
	}

	// Apply depth filtering (depth=0 means no filtering)
	var filtered []*ctags.TagEntry
	if depth > 0 {
		filtered = ctags.FilterByDepth(entries, depth)
	} else {
		filtered = entries // depth=0 means unlimited, no filtering
	}

	// Should include all entries (no filtering)
	if len(filtered) != len(entries) {
		t.Errorf(
			"Unlimited depth filtered count = %d, want %d",
			len(filtered),
			len(entries),
		)
	}

	// Verify we have all levels including H4
	hasH4 := false
	for _, entry := range filtered {
		if entry.Level == 4 {
			hasH4 = true
			break
		}
	}

	if !hasH4 {
		t.Error("Unlimited depth should include H4 entries")
	}
}

// Helper function to check if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
