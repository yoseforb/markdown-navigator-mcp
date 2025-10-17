package tools

import (
	"context"
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

// validateH4Section validates a single H4 section has correct properties.
func validateH4Section(t *testing.T, section SectionInfo) {
	t.Helper()

	if section.StartLine <= 0 || section.EndLine <= 0 {
		t.Errorf(
			"Section %s should have valid line ranges",
			section.Name,
		)
	}
	if section.Level != "H4" {
		t.Errorf(
			"Section %s should be H4, got %s",
			section.Name,
			section.Level,
		)
	}
}

// findSection finds a section by name in the sections slice.
func findSection(sections []SectionInfo, name string) *SectionInfo {
	for i, section := range sections {
		if section.Name == name {
			return &sections[i]
		}
	}
	return nil
}

func TestMarkdownListSections_DefaultDepth(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to depth 2 (H1+H2) - default behavior
	filteredEntries := ctags.FilterByDepth(entries, 2)

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

	// Validate results
	if len(sections) == 0 {
		t.Fatal("Expected at least one section at depth <= 2")
	}

	// Should have H1 and H2 sections
	hasH1 := false
	hasH2 := false
	for _, section := range sections {
		validateSectionFields(t, section)
		if section.Level == "H1" {
			hasH1 = true
		}
		if section.Level == "H2" {
			hasH2 = true
		}
		// Should not have H3 or deeper
		if section.Level != "H1" && section.Level != "H2" {
			t.Errorf(
				"Expected only H1 and H2 with depth=2, got %s",
				section.Level,
			)
		}
	}

	if !hasH1 || !hasH2 {
		t.Error("Expected both H1 and H2 sections with depth=2")
	}

	// Check for expected H2 sections in sample.md
	expectedSections := []string{
		"Section 1: Introduction",
		"Section 2: Implementation",
		"Section 3: Conclusion",
	}

	for _, expected := range expectedSections {
		assertSectionExists(t, sections, expected)
	}
}

func TestMarkdownListSections_Depth1(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to depth 1 (H1 only)
	filteredEntries := ctags.FilterByDepth(entries, 1)

	// Convert to response format
	sections := make([]SectionInfo, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		sections = append(sections, SectionInfo{
			Name:      entry.Name,
			StartLine: entry.Line,
			EndLine:   entry.End,
			Level:     "H1",
		})
	}

	// Validate results
	if len(sections) != 1 {
		t.Fatalf("Expected exactly 1 H1 section, got %d", len(sections))
	}

	section := sections[0]
	if section.Name != "Test Document" {
		t.Errorf("Expected 'Test Document', got %q", section.Name)
	}
	if section.Level != "H1" {
		t.Errorf("Expected level H1, got %s", section.Level)
	}
}

func TestMarkdownListSections_Depth3(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to depth 3 (H1+H2+H3)
	filteredEntries := ctags.FilterByDepth(entries, 3)

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

	// Validate results
	if len(sections) == 0 {
		t.Fatal("Expected at least one section at depth <= 3")
	}

	// Should have H1, H2, and H3 sections
	levelCounts := make(map[string]int)
	for _, section := range sections {
		validateSectionFields(t, section)
		levelCounts[section.Level]++

		// Should not have H4 or deeper
		if section.Level == "H4" || section.Level == "H5" ||
			section.Level == "H6" {
			t.Errorf("Expected only H1-H3 with depth=3, got %s", section.Level)
		}
	}

	if levelCounts["H1"] == 0 || levelCounts["H2"] == 0 ||
		levelCounts["H3"] == 0 {
		t.Error("Expected H1, H2, and H3 sections with depth=3")
	}
}

func TestMarkdownListSections_AllLevels(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache (no filtering by depth)
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by depth 0 (all levels)
	filteredEntries := ctags.FilterByDepth(entries, 0)

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

func TestMarkdownListSections_WithPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by depth and pattern
	filteredEntries := ctags.FilterByDepth(entries, 2)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "Section 1")

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
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by depth with empty pattern (should return all at that depth)
	filteredEntries := ctags.FilterByDepth(entries, 2)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "")

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

	// Validate that empty pattern returns all sections at depth <= 2
	// Sample.md has 1 H1 and 3 H2 sections = 4 total
	if len(sections) != 4 {
		t.Errorf(
			"Expected 4 sections (H1+H2) with empty pattern, got %d",
			len(sections),
		)
	}
}

func TestMarkdownListSections_LineRanges(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
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

func TestMarkdownListSections_Depth3WithPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to depth 3 (H1+H2+H3) sections matching "Testing"
	filteredEntries := ctags.FilterByDepth(entries, 3)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "Testing")

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

	// Validate results
	if len(sections) != 1 {
		t.Fatalf(
			"Expected 1 section matching 'Testing' at depth <= 3, got %d",
			len(sections),
		)
	}

	section := sections[0]
	if section.Name != "Subsection 2.2: Testing" {
		t.Errorf("Expected 'Subsection 2.2: Testing', got %q", section.Name)
	}
	if section.Level != "H3" {
		t.Errorf("Expected level H3, got %s", section.Level)
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
	entries, err := cache.GetTags(context.Background(), targetFile)
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

func TestMarkdownListSections_Depth4Sections(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter to depth 4 (H1+H2+H3+H4)
	filteredEntries := ctags.FilterByDepth(entries, 4)

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

	// Validate we have H4 sections in sample.md
	hasH4 := false
	for _, section := range sections {
		if section.Level == "H4" {
			hasH4 = true
			break
		}
	}

	if !hasH4 {
		t.Fatal("Expected at least one H4 section at depth 4")
	}

	// Check expected H4 sections
	expectedH4Sections := []string{
		"Deep Section 2.2.1: Unit Tests",
		"Deep Section 2.2.2: Integration Tests",
	}

	for _, expected := range expectedH4Sections {
		section := findSection(sections, expected)
		if section == nil {
			t.Errorf("Expected to find H4 section %q", expected)
			continue
		}
		validateH4Section(t, *section)
	}
}

func TestMarkdownListSections_CaseInsensitivePattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
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

func TestMarkdownListSections_AllLevelsWithPattern(t *testing.T) {
	t.Parallel()

	targetFile := filepath.Join("..", "..", "testdata", "sample.md")

	// Get tags from cache
	cache := ctags.GetGlobalCache()
	entries, err := cache.GetTags(context.Background(), targetFile)
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	// Filter by pattern only (depth 0 means all levels)
	filteredEntries := ctags.FilterByDepth(entries, 0)
	filteredEntries = ctags.FilterByPattern(filteredEntries, "Section")

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
			"Expected at least one section matching 'Section' with all levels",
		)
	}

	// Should have sections at different levels
	levelCounts := make(map[string]int)
	for _, section := range sections {
		levelCounts[section.Level]++
	}

	if len(levelCounts) < 2 {
		t.Error(
			"Expected sections at multiple levels when filtering with all levels",
		)
	}
}
