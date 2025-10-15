package ctags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSONTags_ValidInput(t *testing.T) {
	jsonData := []byte(
		`{"_type":"tag","name":"Introduction","path":"test.md","pattern":"/^# Introduction$/","line":1,"kind":"chapter"}
{"_type":"tag","name":"Getting Started","path":"test.md","pattern":"/^## Getting Started$/","line":10,"kind":"section","scope":"Introduction","scopeKind":"chapter"}
{"_type":"tag","name":"Installation","path":"test.md","pattern":"/^### Installation$/","line":20,"kind":"subsection","scope":"Getting Started","scopeKind":"section"}
`,
	)

	entries, err := ParseJSONTags(jsonData, "test.md")
	require.NoError(t, err)
	require.Len(t, entries, 3)

	// Verify first entry (H1 - chapter)
	assert.Equal(t, "Introduction", entries[0].Name)
	assert.Equal(t, "test.md", entries[0].File)
	assert.Equal(t, 1, entries[0].Line)
	assert.Equal(t, "chapter", entries[0].Kind)
	assert.Equal(t, 1, entries[0].Level)
	assert.Equal(t, 0, entries[0].End) // No end specified in test data

	// Verify second entry (H2 - section)
	assert.Equal(t, "Getting Started", entries[1].Name)
	assert.Equal(t, 2, entries[1].Level)
	assert.Equal(t, "Introduction", entries[1].Scope)

	// Verify third entry (H3 - subsection)
	assert.Equal(t, "Installation", entries[2].Name)
	assert.Equal(t, 3, entries[2].Level)
	assert.Equal(t, "Getting Started", entries[2].Scope)
}

func TestParseJSONTags_EmptyInput(t *testing.T) {
	entries, err := ParseJSONTags([]byte{}, "test.md")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestParseJSONTags_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{"_type":"tag","name":"Test"`)

	entries, err := ParseJSONTags(jsonData, "test.md")
	require.NoError(t, err)
	assert.Empty(t, entries) // Invalid JSON is skipped
}

func TestParseJSONTags_NonTagEntries(t *testing.T) {
	jsonData := []byte(`{"_type":"metadata","version":"1.0"}
{"_type":"tag","name":"Test","path":"test.md","pattern":"/^# Test$/","line":1,"kind":"chapter"}
`)

	entries, err := ParseJSONTags(jsonData, "test.md")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Test", entries[0].Name)
}

func TestParseJSONTags_FileFiltering(t *testing.T) {
	jsonData := []byte(
		`{"_type":"tag","name":"Test1","path":"file1.md","pattern":"/^# Test1$/","line":1,"kind":"chapter"}
{"_type":"tag","name":"Test2","path":"file2.md","pattern":"/^# Test2$/","line":1,"kind":"chapter"}
{"_type":"tag","name":"Test3","path":"file1.md","pattern":"/^## Test3$/","line":10,"kind":"section"}
`,
	)

	entries, err := ParseJSONTags(jsonData, "file1.md")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, "Test1", entries[0].Name)
	assert.Equal(t, "Test3", entries[1].Name)
}

func TestParseJSONTags_AllHeadingLevels(t *testing.T) {
	jsonData := []byte(
		`{"_type":"tag","name":"H1","path":"test.md","pattern":"/^# H1$/","line":1,"kind":"chapter"}
{"_type":"tag","name":"H2","path":"test.md","pattern":"/^## H2$/","line":2,"kind":"section"}
{"_type":"tag","name":"H3","path":"test.md","pattern":"/^### H3$/","line":3,"kind":"subsection"}
{"_type":"tag","name":"H4","path":"test.md","pattern":"/^#### H4$/","line":4,"kind":"subsubsection"}
`,
	)

	entries, err := ParseJSONTags(jsonData, "test.md")
	require.NoError(t, err)
	require.Len(t, entries, 4)

	assert.Equal(t, 1, entries[0].Level)
	assert.Equal(t, 2, entries[1].Level)
	assert.Equal(t, 3, entries[2].Level)
	assert.Equal(t, 4, entries[3].Level)
}

func TestParseJSONTags_MissingOptionalFields(t *testing.T) {
	jsonData := []byte(
		`{"_type":"tag","name":"Test","path":"test.md","line":1,"kind":"chapter"}`,
	)

	entries, err := ParseJSONTags(jsonData, "test.md")
	require.NoError(t, err)
	require.Len(t, entries, 1)

	assert.Equal(t, "Test", entries[0].Name)
	assert.Empty(t, entries[0].Pattern)
	assert.Empty(t, entries[0].Scope)
}

func TestParseJSONTags_UnknownKind(t *testing.T) {
	jsonData := []byte(
		`{"_type":"tag","name":"Test","path":"test.md","line":1,"kind":"unknown"}
{"_type":"tag","name":"Valid","path":"test.md","line":2,"kind":"chapter"}
`,
	)

	entries, err := ParseJSONTags(jsonData, "test.md")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Valid", entries[0].Name)
}

func TestJsonEntryToTagEntry(t *testing.T) {
	tests := []struct {
		name     string
		input    *JSONEntry
		expected *TagEntry
	}{
		{
			name: "valid chapter",
			input: &JSONEntry{
				Type:    "tag",
				Name:    "Test",
				Path:    "test.md",
				Pattern: "/^# Test$/",
				Line:    1,
				Kind:    "chapter",
				End:     5,
			},
			expected: &TagEntry{
				Name:    "Test",
				File:    "test.md",
				Pattern: "/^# Test$/",
				Kind:    "chapter",
				Line:    1,
				End:     5,
				Level:   1,
			},
		},
		{
			name: "invalid kind",
			input: &JSONEntry{
				Type: "tag",
				Name: "Test",
				Path: "test.md",
				Kind: "invalid",
			},
			expected: nil,
		},
		{
			name: "with scope",
			input: &JSONEntry{
				Type:      "tag",
				Name:      "Subsection",
				Path:      "test.md",
				Line:      10,
				Kind:      "subsection",
				Scope:     "Parent",
				ScopeKind: "section",
				End:       15,
			},
			expected: &TagEntry{
				Name:  "Subsection",
				File:  "test.md",
				Kind:  "subsection",
				Line:  10,
				End:   15,
				Scope: "Parent",
				Level: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := jsonEntryToTagEntry(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Name, result.Name)
				assert.Equal(t, tt.expected.File, result.File)
				assert.Equal(t, tt.expected.Pattern, result.Pattern)
				assert.Equal(t, tt.expected.Kind, result.Kind)
				assert.Equal(t, tt.expected.Line, result.Line)
				assert.Equal(t, tt.expected.End, result.End)
				assert.Equal(t, tt.expected.Scope, result.Scope)
				assert.Equal(t, tt.expected.Level, result.Level)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected [][]byte
	}{
		{
			name:     "single line",
			input:    []byte("line1"),
			expected: [][]byte{[]byte("line1")},
		},
		{
			name:  "multiple lines",
			input: []byte("line1\nline2\nline3"),
			expected: [][]byte{
				[]byte("line1"),
				[]byte("line2"),
				[]byte("line3"),
			},
		},
		{
			name:     "trailing newline",
			input:    []byte("line1\nline2\n"),
			expected: [][]byte{[]byte("line1"), []byte("line2")},
		},
		{
			name:     "empty input",
			input:    []byte{},
			expected: [][]byte{},
		},
		{
			name:     "only newlines",
			input:    []byte("\n\n\n"),
			expected: [][]byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			assert.Len(t, result, len(tt.expected))
			for i := range tt.expected {
				assert.Equal(t, tt.expected[i], result[i])
			}
		})
	}
}
