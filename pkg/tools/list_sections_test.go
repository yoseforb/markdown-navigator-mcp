package tools

import (
	"path/filepath"
	"testing"

	"github.com/yoseforb/markdown-nav-mcp/pkg/ctags"
)

// validateSectionFields checks that section has valid field values.
func validateSectionFields(t *testing.T, section SectionInfo) {
	t.Helper()

	if section.Name == "" {
		t.Error("Section name should not be empty")
	}
	if section.StartLine <= 0 {
		t.Errorf(
			"Section %s start_line should be positive, got %d",
			section.Name,
			section.StartLine,
		)
	}
	if section.EndLine <= 0 {
		t.Errorf(
			"Section %s end_line should be positive, got %d",
			section.Name,
			section.EndLine,
		)
	}
	if section.EndLine < section.StartLine {
		t.Errorf(
			"Section %s end_line (%d) should be >= start_line (%d)",
			section.Name,
			section.EndLine,
			section.StartLine,
		)
	}
}

// assertSectionExists checks if expected section exists in sections.
func assertSectionExists(
	t *testing.T,
	sections []SectionInfo,
	expected string,
) {
	t.Helper()

	for _, section := range sections {
		if section.Name == expected {
			return
		}
	}
	t.Errorf("Expected to find section %q", expected)
}

func TestMarkdownListSections_NoPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to H2 sections only
	filteredEntries := ctags.FilterByLevel(entries, 2)

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H2",
		})
	}

	// Validate results
	if len(sections) == 0 {
		t.Fatal("Expected at least one H2 section")
	}

	// Check that all sections have required fields
	for _, section := range sections {
		validateSectionFields(t, section)
		if section.Level != "H2" {
			t.Errorf("Expected level H2, got %s", section.Level)
		}
	}

	// Check for expected sections in sample.md
	expectedSections := []string{
		"Section 1: Introduction",
		"Section 2: Implementation",
		"Section 3: Conclusion",
	}

	for _, expected := range expectedSections {
		assertSectionExists(t, sections, expected)
	}
}

func TestMarkdownListSections_WithPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by level and pattern
	filteredEntries := ctags.FilterByLevel(entries, 2)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "Section 1")

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H2",
		})
	}

	// Validate results
	if len(sections) != 1 {
		t.Fatalf(
			"Expected 1 section matching 'Section 1', got %d",
			len(sections),
		)
	}

	section := sections[0]
	if section.Name != "Section 1: Introduction" {
		t.Errorf("Expected 'Section 1: Introduction', got %q", section.Name)
	}

	// Verify line numbers
	if section.StartLine != 5 {
		t.Errorf("Expected start_line 5, got %d", section.StartLine)
	}
	if section.EndLine <= section.StartLine {
		t.Errorf(
			"Expected end_line > start_line, got %d <= %d",
			section.EndLine,
			section.StartLine,
		)
	}
}

func TestMarkdownListSections_EmptyPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by level with empty pattern (should return all)
	filteredEntries := ctags.FilterByLevel(entries, 2)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "")

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H2",
		})
	}

	// Validate that empty pattern returns all sections
	if len(sections) != 3 {
		t.Errorf(
			"Expected 3 H2 sections with empty pattern, got %d",
			len(sections),
		)
	}
}

func TestMarkdownListSections_LineRanges(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Test line ranges for all entries
	for _, entry := range entries {
		section := SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H" + string(rune('0'+entry.Level)),
		}

		// Basic validation
		if section.StartLine <= 0 {
			t.Errorf(
				"Section %s: start_line must be positive, got %d",
				section.Name,
				section.StartLine,
			)
		}

		if section.EndLine <= 0 {
			t.Errorf(
				"Section %s: end_line must be positive, got %d",
				section.Name,
				section.EndLine,
			)
		}

		if section.EndLine < section.StartLine {
			t.Errorf(
				"Section %s: end_line (%d) must be >= start_line (%d)",
				section.Name,
				section.EndLine,
				section.StartLine,
			)
		}
	}
}

func TestMarkdownListSections_AllLevels(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache (no filtering by level)
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Convert to response format without level filtering
	sections := make([]SectionInfo, 0, len(entries))
	for _, entry := range entries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H" + string(rune('0'+entry.Level)),
		})
	}

	// Validate we have sections at multiple levels
	levelCounts := make(map[string]int)
	for _, section := range sections {
		levelCounts[section.Level]++
	}

	if len(levelCounts) < 2 {
		t.Error("Expected sections at multiple heading levels")
	}

	// Should have H1, H2, H3, and H4 in sample.md
	expectedLevels := []string{"H1", "H2", "H3", "H4"}
	for _, level := range expectedLevels {
		if levelCounts[level] == 0 {
			t.Errorf("Expected at least one section at level %s", level)
		}
	}
}

func TestMarkdownListSections_H3WithPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to H3 sections matching "Testing"
	filteredEntries := ctags.FilterByLevel(entries, 3)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "Testing")

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H3",
		})
	}

	// Validate results
	if len(sections) != 1 {
		t.Fatalf(
			"Expected 1 H3 section matching 'Testing', got %d",
			len(sections),
		)
	}

	section := sections[0]
	if section.Name != "Subsection 2.2: Testing" {
		t.Errorf("Expected 'Subsection 2.2: Testing', got %q", section.Name)
	}

	// Verify it has valid line ranges
	if section.StartLine <= 0 || section.EndLine <= 0 {
		t.Error("Section should have valid line ranges")
	}
}

func TestMarkdownListSections_NoMatchingPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter with pattern that won't match
	filteredEntries := ctags.FilterByPattern(entries, "NonExistentSection")

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H" + string(rune('0'+entry.Level)),
		})
	}

	// Should return empty list, not error
	if len(sections) != 0 {
		t.Errorf(
			"Expected 0 sections for non-matching pattern, got %d",
			len(sections),
		)
	}
}

func TestMarkdownListSections_H4Sections(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to H4 sections only
	filteredEntries := ctags.FilterByLevel(entries, 4)

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H4",
		})
	}

	// Validate we have H4 sections in sample.md
	if len(sections) == 0 {
		t.Fatal("Expected at least one H4 section")
	}

	// Check expected H4 sections
	expectedH4Sections := []string{
		"Deep Section 2.2.1: Unit Tests",
		"Deep Section 2.2.2: Integration Tests",
	}

	for _, expected := range expectedH4Sections {
		found := false
		for _, section := range sections {
			if section.Name == expected {
				found = true
				// Verify it has proper line ranges
				if section.StartLine <= 0 || section.EndLine <= 0 {
					t.Errorf(
						"Section %s should have valid line ranges",
						section.Name,
					)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected to find H4 section %q", expected)
		}
	}
}

func TestMarkdownListSections_CaseInsensitivePattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Test case-insensitive matching
	testCases := []struct {
		pattern       string
		expectedCount int
	}{
		{
			"section",
			9,
		}, // Should match "Section", "Subsection", "Deep Section" (case-insensitive)
		{
			"SECTION",
			9,
		}, // Should match "Section", "Subsection", "Deep Section" (case-insensitive)
		{
			"SeCTion",
			9,
		}, // Should match "Section", "Subsection", "Deep Section" (case-insensitive)
		{"testing", 1},      // Should match "Testing"
		{"TESTING", 1},      // Should match "Testing"
		{"introduction", 1}, // Should match "Introduction"
	}

	for _, tc := range testCases {
		filteredEntries := ctags.FilterByPattern(entries, tc.pattern)
		if len(filteredEntries) != tc.expectedCount {
			t.Errorf(
				"Pattern %q: expected %d matches, got %d",
				tc.pattern,
				tc.expectedCount,
				len(filteredEntries),
			)
		}
	}
}

func TestMarkdownListSections_HeadingLevelALL(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// When heading_level is "ALL", should return all levels
	// (no level filtering applied)
	allSections := make([]SectionInfo, 0, len(entries))
	for _, entry := range entries {
		allSections = append(allSections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H" + string(rune('0'+entry.Level)),
		})
	}

	// Validate we have sections at multiple levels
	levelCounts := make(map[string]int)
	for _, section := range allSections {
		levelCounts[section.Level]++
	}

	if len(levelCounts) < 2 {
		t.Error(
			"Expected sections at multiple heading levels when using ALL",
		)
	}

	// Should have H1, H2, H3, and H4 in sample.md
	expectedLevels := []string{"H1", "H2", "H3", "H4"}
	for _, level := range expectedLevels {
		if levelCounts[level] == 0 {
			t.Errorf(
				"Expected at least one section at level %s when using ALL",
				level,
			)
		}
	}
}

func TestMarkdownListSections_ALLWithPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by pattern only (ALL means no level filtering)
	filteredEntries := ctags.FilterByPattern(entries, "Section")

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H" + string(rune('0'+entry.Level)),
		})
	}

	// Validate we have sections at multiple levels matching "Section"
	if len(sections) == 0 {
		t.Fatal(
			"Expected at least one section matching 'Section' with ALL",
		)
	}

	// Should have sections at different levels
	levelCounts := make(map[string]int)
	for _, section := range sections {
		levelCounts[section.Level]++
	}

	if len(levelCounts) < 2 {
		t.Error(
			"Expected sections at multiple levels when filtering with ALL",
		)
	}
}
