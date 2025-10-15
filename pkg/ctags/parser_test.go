package ctags

import (
	"path/filepath"
	"testing"
)

func TestParseTagsFile(t *testing.T) {
	tagsPath := filepath.Join("..", "..", "testdata", "tags")
	targetFile := "sample.md"

	entries, err := ParseTagsFile(tagsPath, targetFile)
	if err != nil {
		t.Fatalf("ParseTagsFile failed: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("Expected entries, got none")
	}

	// Should have entries sorted by line number
	for i := 1; i < len(entries); i++ {
		if entries[i].Line < entries[i-1].Line {
			t.Errorf(
				"Entries not sorted: entry[%d].Line=%d < entry[%d].Line=%d",
				i,
				entries[i].Line,
				i-1,
				entries[i-1].Line,
			)
		}
	}

	// Check specific entries
	foundTestDoc := false
	foundSection1 := false
	for _, entry := range entries {
		if entry.Name == "Test Document" && entry.Level == 1 {
			foundTestDoc = true
		}
		if entry.Name == "Section 1: Introduction" && entry.Level == 2 {
			foundSection1 = true
		}
	}

	if !foundTestDoc {
		t.Error("Expected to find 'Test Document' H1 entry")
	}
	if !foundSection1 {
		t.Error("Expected to find 'Section 1: Introduction' H2 entry")
	}
}

func TestParseTagsFileNonExistentFile(t *testing.T) {
	_, err := ParseTagsFile("nonexistent.tags", "sample.md")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseTagsFileNonExistentTarget(t *testing.T) {
	tagsPath := filepath.Join("..", "..", "testdata", "tags")
	entries, err := ParseTagsFile(tagsPath, "nonexistent.md")
	if err != nil {
		t.Fatalf("ParseTagsFile failed: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf(
			"Expected 0 entries for non-existent target file, got %d",
			len(entries),
		)
	}
}

func TestFindSectionBounds(t *testing.T) {
	tagsPath := filepath.Join("..", "..", "testdata", "tags")
	targetFile := "sample.md"

	entries, err := ParseTagsFile(tagsPath, targetFile)
	if err != nil {
		t.Fatalf("ParseTagsFile failed: %v", err)
	}

	tests := []struct {
		name             string
		query            string
		wantFound        bool
		wantName         string
		wantStart        int
		expectEndNonZero bool
	}{
		{
			name:             "Find Section 1",
			query:            "Section 1",
			wantFound:        true,
			wantName:         "Section 1: Introduction",
			wantStart:        5,
			expectEndNonZero: true,
		},
		{
			name:             "Find Section 2",
			query:            "Section 2",
			wantFound:        true,
			wantName:         "Section 2: Implementation",
			wantStart:        21,
			expectEndNonZero: true,
		},
		{
			name:             "Find Section 3 (last section)",
			query:            "Section 3",
			wantFound:        true,
			wantName:         "Section 3: Conclusion",
			wantStart:        44,
			expectEndNonZero: false, // Last section goes to EOF
		},
		{
			name:      "Non-existent section",
			query:     "Nonexistent Section",
			wantFound: false,
		},
		{
			name:             "Fuzzy match",
			query:            "implement",
			wantFound:        true,
			wantName:         "Section 2: Implementation",
			wantStart:        21,
			expectEndNonZero: true,
		},
		{
			name:             "Subsection",
			query:            "Background",
			wantFound:        true,
			wantName:         "Subsection 1.1: Background",
			wantStart:        10,
			expectEndNonZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startLine, endLine, sectionName, found := FindSectionBounds(
				entries,
				tt.query,
			)

			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}

			if !found {
				return
			}

			if sectionName != tt.wantName {
				t.Errorf("sectionName = %q, want %q", sectionName, tt.wantName)
			}

			if startLine != tt.wantStart {
				t.Errorf("startLine = %d, want %d", startLine, tt.wantStart)
			}

			if tt.expectEndNonZero && endLine == 0 {
				t.Errorf("endLine = 0, expected non-zero")
			}

			if !tt.expectEndNonZero && endLine != 0 {
				t.Errorf("endLine = %d, expected 0 (EOF)", endLine)
			}
		})
	}
}

func TestFilterByLevel(t *testing.T) {
	tagsPath := filepath.Join("..", "..", "testdata", "tags")
	targetFile := "sample.md"

	entries, err := ParseTagsFile(tagsPath, targetFile)
	if err != nil {
		t.Fatalf("ParseTagsFile failed: %v", err)
	}

	tests := []struct {
		name  string
		level int
		want  int
	}{
		{"Filter H1", 1, 1}, // Only "Test Document"
		{"Filter H2", 2, 3}, // Section 1, 2, 3
		{"Filter H3", 3, 4}, // 4 subsections
		{"Filter H4", 4, 2}, // 2 deep sections
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterByLevel(entries, tt.level)
			if len(filtered) != tt.want {
				t.Errorf(
					"FilterByLevel(%d) = %d entries, want %d",
					tt.level,
					len(filtered),
					tt.want,
				)
			}

			// Verify all filtered entries have the correct level
			for _, entry := range filtered {
				if entry.Level != tt.level {
					t.Errorf(
						"Entry %q has level %d, want %d",
						entry.Name,
						entry.Level,
						tt.level,
					)
				}
			}
		})
	}
}

func TestFilterByPattern(t *testing.T) {
	tagsPath := filepath.Join("..", "..", "testdata", "tags")
	targetFile := "sample.md"

	entries, err := ParseTagsFile(tagsPath, targetFile)
	if err != nil {
		t.Fatalf("ParseTagsFile failed: %v", err)
	}

	tests := []struct {
		name    string
		pattern string
		want    int
	}{
		{"Empty pattern", "", len(entries)},
		{
			"Pattern 'Section'",
			"Section",
			9,
		}, // All entries with "Section" or "Subsection"
		{"Pattern 'Subsection'", "Subsection", 4},
		{"Pattern 'Deep'", "Deep", 2},
		{
			"Pattern 'Test'",
			"Test",
			4,
		}, // "Test Document" + "Unit Tests" + "Integration Tests" + "Testing"
		{"Pattern 'xyz' (no match)", "xyz", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterByPattern(entries, tt.pattern)
			if len(filtered) != tt.want {
				t.Errorf(
					"FilterByPattern(%q) = %d entries, want %d",
					tt.pattern,
					len(filtered),
					tt.want,
				)
			}
		})
	}
}

func TestNewTagEntry(t *testing.T) {
	entry := NewTagEntry(
		"Test",
		"file.md",
		"/pattern/",
		"section",
		10,
		"parent",
	)

	if entry.Name != "Test" {
		t.Errorf("Name = %q, want %q", entry.Name, "Test")
	}

	if entry.Level != 2 { // section = H2
		t.Errorf("Level = %d, want 2", entry.Level)
	}

	if entry.Line != 10 {
		t.Errorf("Line = %d, want 10", entry.Line)
	}
}
