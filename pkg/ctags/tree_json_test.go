package ctags

import (
	"testing"
)

func TestBuildTreeJSON_EmptyEntries(t *testing.T) {
	t.Parallel()

	result := BuildTreeJSON([]*TagEntry{})

	if result != nil {
		t.Errorf("BuildTreeJSON([]) = %v, want nil", result)
	}
}

func TestBuildTreeJSON_SingleEntry(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{
			Name:  "Introduction",
			File:  "/tmp/test.md",
			Level: 1,
			Line:  5,
			End:   20,
		},
	}

	result := BuildTreeJSON(entries)

	if result == nil {
		t.Fatal("BuildTreeJSON() returned nil, want non-nil")
	}

	// Check root node
	if result.Name != "test.md" {
		t.Errorf("Root name = %s, want test.md", result.Name)
	}

	if result.Level != "H0" {
		t.Errorf("Root level = %s, want H0", result.Level)
	}

	if len(result.Children) != 1 {
		t.Fatalf("Root children count = %d, want 1", len(result.Children))
	}

	// Check child node
	child := result.Children[0]
	if child.Name != "Introduction" {
		t.Errorf("Child name = %s, want Introduction", child.Name)
	}

	if child.Level != "H1" {
		t.Errorf("Child level = %s, want H1", child.Level)
	}

	if child.StartLine != 5 {
		t.Errorf("Child start_line = %d, want 5", child.StartLine)
	}

	if child.EndLine != 20 {
		t.Errorf("Child end_line = %d, want 20", child.EndLine)
	}

	if len(child.Children) != 0 {
		t.Errorf("Child children count = %d, want 0", len(child.Children))
	}
}

func TestBuildTreeJSON_FlatStructure(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Section 1", File: "/tmp/test.md", Level: 2, Line: 10, End: 20},
		{Name: "Section 2", File: "/tmp/test.md", Level: 2, Line: 21, End: 30},
		{Name: "Section 3", File: "/tmp/test.md", Level: 2, Line: 31, End: 40},
	}

	result := BuildTreeJSON(entries)

	if result == nil {
		t.Fatal("BuildTreeJSON() returned nil")
	}

	// All entries should be children of root (same level)
	if len(result.Children) != 3 {
		t.Fatalf("Root children count = %d, want 3", len(result.Children))
	}

	expectedNames := []string{"Section 1", "Section 2", "Section 3"}
	for i, child := range result.Children {
		if child.Name != expectedNames[i] {
			t.Errorf(
				"Child[%d] name = %s, want %s",
				i,
				child.Name,
				expectedNames[i],
			)
		}

		if child.Level != "H2" {
			t.Errorf("Child[%d] level = %s, want H2", i, child.Level)
		}

		if len(child.Children) != 0 {
			t.Errorf("Child[%d] should have no children", i)
		}
	}
}

func TestBuildTreeJSON_NestedStructure(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Chapter 1", File: "/tmp/test.md", Level: 1, Line: 1, End: 100},
		{Name: "Section 1.1", File: "/tmp/test.md", Level: 2, Line: 5, End: 50},
		{
			Name:  "Subsection 1.1.1",
			File:  "/tmp/test.md",
			Level: 3,
			Line:  10,
			End:   30,
		},
		{
			Name:  "Subsubsection 1.1.1.1",
			File:  "/tmp/test.md",
			Level: 4,
			Line:  15,
			End:   25,
		},
		{
			Name:  "Subsection 1.1.2",
			File:  "/tmp/test.md",
			Level: 3,
			Line:  35,
			End:   45,
		},
		{
			Name:  "Section 1.2",
			File:  "/tmp/test.md",
			Level: 2,
			Line:  55,
			End:   95,
		},
	}

	result := BuildTreeJSON(entries)

	if result == nil {
		t.Fatal("BuildTreeJSON() returned nil")
	}

	// Check root has 1 child (Chapter 1)
	if len(result.Children) != 1 {
		t.Fatalf("Root children count = %d, want 1", len(result.Children))
	}

	chapter := result.Children[0]
	if chapter.Name != "Chapter 1" {
		t.Errorf("Chapter name = %s, want Chapter 1", chapter.Name)
	}

	// Chapter should have 2 sections
	if len(chapter.Children) != 2 {
		t.Fatalf("Chapter children count = %d, want 2", len(chapter.Children))
	}

	// Check Section 1.1
	section11 := chapter.Children[0]
	if section11.Name != "Section 1.1" {
		t.Errorf("Section 1.1 name = %s, want Section 1.1", section11.Name)
	}

	// Section 1.1 should have 2 subsections
	if len(section11.Children) != 2 {
		t.Fatalf(
			"Section 1.1 children count = %d, want 2",
			len(section11.Children),
		)
	}

	// Check Subsection 1.1.1
	subsection111 := section11.Children[0]
	if subsection111.Name != "Subsection 1.1.1" {
		t.Errorf(
			"Subsection 1.1.1 name = %s, want Subsection 1.1.1",
			subsection111.Name,
		)
	}

	// Subsection 1.1.1 should have 1 child (Subsubsection 1.1.1.1)
	if len(subsection111.Children) != 1 {
		t.Fatalf(
			"Subsection 1.1.1 children count = %d, want 1",
			len(subsection111.Children),
		)
	}

	// Check Subsubsection 1.1.1.1
	subsubsection1111 := subsection111.Children[0]
	if subsubsection1111.Name != "Subsubsection 1.1.1.1" {
		t.Errorf(
			"Subsubsection name = %s, want Subsubsection 1.1.1.1",
			subsubsection1111.Name,
		)
	}

	if subsubsection1111.Level != "H4" {
		t.Errorf(
			"Subsubsection level = %s, want H4",
			subsubsection1111.Level,
		)
	}

	if len(subsubsection1111.Children) != 0 {
		t.Errorf("Subsubsection should have no children")
	}

	// Check Section 1.2
	section12 := chapter.Children[1]
	if section12.Name != "Section 1.2" {
		t.Errorf("Section 1.2 name = %s, want Section 1.2", section12.Name)
	}

	if len(section12.Children) != 0 {
		t.Errorf("Section 1.2 should have no children")
	}
}

func TestBuildTreeJSON_MultipleRoots(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Chapter 1", File: "/tmp/test.md", Level: 1, Line: 1, End: 50},
		{Name: "Section 1.1", File: "/tmp/test.md", Level: 2, Line: 5, End: 25},
		{
			Name:  "Section 1.2",
			File:  "/tmp/test.md",
			Level: 2,
			Line:  30,
			End:   45,
		},
		{Name: "Chapter 2", File: "/tmp/test.md", Level: 1, Line: 51, End: 100},
		{
			Name:  "Section 2.1",
			File:  "/tmp/test.md",
			Level: 2,
			Line:  55,
			End:   95,
		},
	}

	result := BuildTreeJSON(entries)

	if result == nil {
		t.Fatal("BuildTreeJSON() returned nil")
	}

	// Root should have 2 children (Chapter 1 and Chapter 2)
	if len(result.Children) != 2 {
		t.Fatalf("Root children count = %d, want 2", len(result.Children))
	}

	// Check Chapter 1
	chapter1 := result.Children[0]
	if chapter1.Name != "Chapter 1" {
		t.Errorf("Chapter 1 name = %s, want Chapter 1", chapter1.Name)
	}

	if len(chapter1.Children) != 2 {
		t.Errorf(
			"Chapter 1 children count = %d, want 2",
			len(chapter1.Children),
		)
	}

	// Check Chapter 2
	chapter2 := result.Children[1]
	if chapter2.Name != "Chapter 2" {
		t.Errorf("Chapter 2 name = %s, want Chapter 2", chapter2.Name)
	}

	if len(chapter2.Children) != 1 {
		t.Errorf(
			"Chapter 2 children count = %d, want 1",
			len(chapter2.Children),
		)
	}
}

func TestBuildTreeJSON_LineRanges(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Section 1", File: "/tmp/test.md", Level: 1, Line: 10, End: 50},
		{Name: "Section 2", File: "/tmp/test.md", Level: 1, Line: 51, End: 100},
	}

	result := BuildTreeJSON(entries)

	if result == nil {
		t.Fatal("BuildTreeJSON() returned nil")
	}

	if len(result.Children) != 2 {
		t.Fatalf("Root children count = %d, want 2", len(result.Children))
	}

	// Check line ranges are preserved
	section1 := result.Children[0]
	if section1.StartLine != 10 {
		t.Errorf("Section 1 start_line = %d, want 10", section1.StartLine)
	}

	if section1.EndLine != 50 {
		t.Errorf("Section 1 end_line = %d, want 50", section1.EndLine)
	}

	section2 := result.Children[1]
	if section2.StartLine != 51 {
		t.Errorf("Section 2 start_line = %d, want 51", section2.StartLine)
	}

	if section2.EndLine != 100 {
		t.Errorf("Section 2 end_line = %d, want 100", section2.EndLine)
	}
}

func TestGetLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		levelStr  string
		wantLevel int
	}{
		{name: "H0", levelStr: "H0", wantLevel: 0},
		{name: "H1", levelStr: "H1", wantLevel: 1},
		{name: "H2", levelStr: "H2", wantLevel: 2},
		{name: "H3", levelStr: "H3", wantLevel: 3},
		{name: "H4", levelStr: "H4", wantLevel: 4},
		{name: "Invalid format", levelStr: "invalid", wantLevel: 0},
		{name: "Empty string", levelStr: "", wantLevel: 0},
		{name: "Single char", levelStr: "H", wantLevel: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getLevel(tt.levelStr)
			if got != tt.wantLevel {
				t.Errorf(
					"getLevel(%s) = %d, want %d",
					tt.levelStr,
					got,
					tt.wantLevel,
				)
			}
		})
	}
}

func TestFilterByPatternWithParents_NoPattern(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Section 1", Level: 1, Line: 1, End: 10},
		{Name: "Section 2", Level: 1, Line: 11, End: 20},
	}

	result := FilterByPatternWithParents(entries, "")

	if len(result) != len(entries) {
		t.Errorf(
			"FilterByPatternWithParents('') count = %d, want %d",
			len(result),
			len(entries),
		)
	}
}

func TestFilterByPatternWithParents_DirectMatch(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Introduction", Level: 1, Line: 1, End: 10},
		{Name: "Testing Guidelines", Level: 1, Line: 11, End: 20},
		{Name: "Conclusion", Level: 1, Line: 21, End: 30},
	}

	result := FilterByPatternWithParents(entries, "Testing")

	if len(result) != 1 {
		t.Fatalf(
			"FilterByPatternWithParents('Testing') count = %d, want 1",
			len(result),
		)
	}

	if result[0].Name != "Testing Guidelines" {
		t.Errorf(
			"Matched entry name = %s, want Testing Guidelines",
			result[0].Name,
		)
	}
}

func TestFilterByPatternWithParents_ParentPreservation(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Chapter 1", Level: 1, Line: 1, End: 100},
		{Name: "Section 1.1", Level: 2, Line: 5, End: 50},
		{Name: "Subsection Testing", Level: 3, Line: 10, End: 40},
		{Name: "Section 1.2", Level: 2, Line: 55, End: 95},
		{Name: "Chapter 2", Level: 1, Line: 101, End: 200},
	}

	result := FilterByPatternWithParents(entries, "Testing")

	// Should include: Chapter 1, Section 1.1, Subsection Testing
	expectedCount := 3
	if len(result) != expectedCount {
		t.Fatalf(
			"FilterByPatternWithParents count = %d, want %d",
			len(result),
			expectedCount,
		)
	}

	expectedNames := []string{"Chapter 1", "Section 1.1", "Subsection Testing"}
	for i, entry := range result {
		if entry.Name != expectedNames[i] {
			t.Errorf(
				"Result[%d] name = %s, want %s",
				i,
				entry.Name,
				expectedNames[i],
			)
		}
	}
}

func TestFilterByPatternWithParents_MultipleMatches(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Testing Introduction", Level: 1, Line: 1, End: 50},
		{Name: "Unit Tests", Level: 2, Line: 10, End: 30},
		{Name: "Implementation", Level: 1, Line: 51, End: 100},
		{Name: "Testing Strategy", Level: 2, Line: 60, End: 90},
	}

	result := FilterByPatternWithParents(entries, "Testing")

	// Should include both H1 sections and the nested "Testing Strategy"
	expectedCount := 3
	if len(result) != expectedCount {
		t.Fatalf(
			"FilterByPatternWithParents count = %d, want %d",
			len(result),
			expectedCount,
		)
	}

	expectedNames := []string{
		"Testing Introduction",
		"Implementation",
		"Testing Strategy",
	}
	for i, entry := range result {
		if entry.Name != expectedNames[i] {
			t.Errorf(
				"Result[%d] name = %s, want %s",
				i,
				entry.Name,
				expectedNames[i],
			)
		}
	}
}

func TestFilterByPatternWithParents_NoMatches(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Introduction", Level: 1, Line: 1, End: 10},
		{Name: "Implementation", Level: 1, Line: 11, End: 20},
	}

	result := FilterByPatternWithParents(entries, "NonExistent")

	if len(result) != 0 {
		t.Errorf(
			"FilterByPatternWithParents('NonExistent') count = %d, want 0",
			len(result),
		)
	}
}

func TestFilterByPatternWithParents_CaseInsensitive(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Testing Guidelines", Level: 1, Line: 1, End: 10},
		{Name: "Implementation", Level: 1, Line: 11, End: 20},
	}

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

			result := FilterByPatternWithParents(entries, tt.pattern)

			if len(result) != 1 {
				t.Fatalf(
					"FilterByPatternWithParents('%s') count = %d, want 1",
					tt.pattern,
					len(result),
				)
			}

			if result[0].Name != "Testing Guidelines" {
				t.Errorf(
					"Matched entry name = %s, want Testing Guidelines",
					result[0].Name,
				)
			}
		})
	}
}

func TestFilterByPatternWithParents_PartialMatch(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Unit Testing", Level: 1, Line: 1, End: 10},
		{Name: "Integration Testing", Level: 1, Line: 11, End: 20},
		{Name: "Implementation", Level: 1, Line: 21, End: 30},
	}

	result := FilterByPatternWithParents(entries, "Test")

	if len(result) != 2 {
		t.Fatalf(
			"FilterByPatternWithParents('Test') count = %d, want 2",
			len(result),
		)
	}

	expectedNames := []string{"Unit Testing", "Integration Testing"}
	for i, entry := range result {
		if entry.Name != expectedNames[i] {
			t.Errorf(
				"Result[%d] name = %s, want %s",
				i,
				entry.Name,
				expectedNames[i],
			)
		}
	}
}

func TestFilterByPatternWithParents_DeepNesting(t *testing.T) {
	t.Parallel()

	entries := []*TagEntry{
		{Name: "Chapter 1", Level: 1, Line: 1, End: 100},
		{Name: "Section 1.1", Level: 2, Line: 5, End: 80},
		{Name: "Subsection 1.1.1", Level: 3, Line: 10, End: 50},
		{Name: "Deep Testing Topic", Level: 4, Line: 15, End: 40},
		{Name: "Chapter 2", Level: 1, Line: 101, End: 200},
	}

	result := FilterByPatternWithParents(entries, "Testing")

	// Should include full path: Chapter 1 -> Section 1.1 ->
	// Subsection 1.1.1 -> Deep Testing Topic
	expectedCount := 4
	if len(result) != expectedCount {
		t.Fatalf(
			"FilterByPatternWithParents count = %d, want %d",
			len(result),
			expectedCount,
		)
	}

	expectedNames := []string{
		"Chapter 1",
		"Section 1.1",
		"Subsection 1.1.1",
		"Deep Testing Topic",
	}
	for i, entry := range result {
		if entry.Name != expectedNames[i] {
			t.Errorf(
				"Result[%d] name = %s, want %s",
				i,
				entry.Name,
				expectedNames[i],
			)
		}
	}
}
