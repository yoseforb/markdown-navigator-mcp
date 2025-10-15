package ctags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCtags_ValidMarkdown(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	// Create temporary markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")

	content := `# Chapter One
## Section One
### Subsection
#### Subsubsection
`

	err := os.WriteFile(mdFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Execute ctags
	output, err := ExecuteCtags(mdFile)
	require.NoError(t, err)
	require.NotEmpty(t, output)

	// Verify output is valid JSON
	entries, err := ParseJSONTags(output, mdFile)
	require.NoError(t, err)
	assert.Len(t, entries, 4)

	// Verify entries
	assert.Equal(t, "Chapter One", entries[0].Name)
	assert.Equal(t, 1, entries[0].Level)
	assert.Equal(t, "Section One", entries[1].Name)
	assert.Equal(t, 2, entries[1].Level)
}

func TestExecuteCtags_EmptyFile(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	// Create empty markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "empty.md")

	err := os.WriteFile(mdFile, []byte(""), 0o644)
	require.NoError(t, err)

	// Execute ctags
	output, err := ExecuteCtags(mdFile)
	require.NoError(t, err)

	// Empty file should return empty JSON
	entries, err := ParseJSONTags(output, mdFile)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestExecuteCtags_NoHeadings(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	// Create markdown file without headings
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "no_headings.md")

	content := `This is just regular text.
No headings here.
Just paragraphs.
`

	err := os.WriteFile(mdFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Execute ctags
	output, err := ExecuteCtags(mdFile)
	require.NoError(t, err)

	// No headings should return empty entries
	entries, err := ParseJSONTags(output, mdFile)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestExecuteCtags_FileNotFound(t *testing.T) {
	output, err := ExecuteCtags("/nonexistent/file.md")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrFileNotFound)
	assert.Nil(t, output)
}

func TestExecuteCtags_LargeFile(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	// Create large markdown file (should still complete within timeout)
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "large.md")

	var content string
	for i := 1; i <= 100; i++ {
		content += "# Heading " + string(rune(i)) + "\n"
		content += "Some content here.\n\n"
	}

	err := os.WriteFile(mdFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Execute ctags (should complete within 5 seconds)
	output, err := ExecuteCtags(mdFile)
	require.NoError(t, err)
	require.NotEmpty(t, output)

	// Verify we got tags
	entries, err := ParseJSONTags(output, mdFile)
	require.NoError(t, err)
	assert.NotEmpty(t, entries)
}

func TestExecuteCtags_SpecialCharacters(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	// Create markdown file with special characters
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "special.md")

	content := `# Chapter: Setup & Configuration
## Section (Part 1)
### Subsection - Details
`

	err := os.WriteFile(mdFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Execute ctags
	output, err := ExecuteCtags(mdFile)
	require.NoError(t, err)

	// Verify entries
	entries, err := ParseJSONTags(output, mdFile)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.Contains(t, entries[0].Name, "Setup & Configuration")
}

func TestIsCtagsInstalled(t *testing.T) {
	result := IsCtagsInstalled()
	// This test just verifies the function runs without panic
	// The actual result depends on the system
	t.Logf("IsCtagsInstalled: %v", result)
}

func TestExecuteCtags_Integration(t *testing.T) {
	if !IsCtagsInstalled() {
		t.Skip("ctags not installed, skipping test")
	}

	// Create a realistic markdown structure
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "document.md")

	content := `# Introduction
This is the introduction section.

## Background
Some background information.

## Motivation
Why we're doing this.

# Implementation
The main implementation section.

## Architecture
System architecture details.

### Frontend
Frontend components.

### Backend
Backend services.

## Testing
Testing strategies.

# Conclusion
Final thoughts.
`

	err := os.WriteFile(mdFile, []byte(content), 0o644)
	require.NoError(t, err)

	// Execute ctags
	output, err := ExecuteCtags(mdFile)
	require.NoError(t, err)

	// Parse and verify structure
	entries, err := ParseJSONTags(output, mdFile)
	require.NoError(t, err)

	// Should have headings at multiple levels
	assert.GreaterOrEqual(t, len(entries), 8)

	// Verify some key entries
	var h1Count, h2Count, h3Count int
	for _, entry := range entries {
		switch entry.Level {
		case 1:
			h1Count++
		case 2:
			h2Count++
		case 3:
			h3Count++
		}
	}

	assert.Equal(t, 3, h1Count, "Should have 3 H1 headings")
	assert.GreaterOrEqual(t, h2Count, 4, "Should have at least 4 H2 headings")
	assert.Equal(t, 2, h3Count, "Should have 2 H3 headings")
}
