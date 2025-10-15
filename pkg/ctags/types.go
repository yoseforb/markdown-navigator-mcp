package ctags

import (
	"sort"
	"strings"
)

// TagEntry represents a single ctags entry.
type TagEntry struct {
	Name    string
	File    string
	Pattern string
	Kind    string
	Line    int
	End     int    // End line of section (from ctags JSON output)
	Scope   string // Full scope with separators
	Level   int    // Heading level (1-4)
}

// kindLevelMap maps ctags kind to heading level.
// This is package-level and immutable to ensure consistent heading level
// mapping across all parsers and tag operations. It's equivalent to a
// constant map and never modified after initialization.
var kindLevelMap = map[string]int{ //nolint:gochecknoglobals // immutable lookup map
	"chapter":       1, // H1: #
	"section":       2, // H2: ##
	"subsection":    3, // H3: ###
	"subsubsection": 4, // H4: ####
}

// NewTagEntry creates a new TagEntry with level determined from kind.
func NewTagEntry(
	name, file, pattern, kind string,
	line, end int,
	scope string,
) *TagEntry {
	level := kindLevelMap[kind]
	return &TagEntry{
		Name:    name,
		File:    file,
		Pattern: pattern,
		Kind:    kind,
		Line:    line,
		End:     end,
		Scope:   scope,
		Level:   level,
	}
}

// FindSectionBounds finds the start and end line numbers for a section.
// Uses the End field from ctags JSON output for accurate section boundaries.
func FindSectionBounds(
	entries []*TagEntry,
	sectionQuery string,
) (startLine, endLine int, sectionName string, found bool) {
	// Find matching section (case-insensitive substring match)
	lowerQuery := strings.ToLower(sectionQuery)

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lowerQuery) {
			startLine = entry.Line
			endLine = entry.End // Use End field from ctags JSON
			sectionName = entry.Name
			found = true
			break
		}
	}

	if !found {
		return 0, 0, "", false
	}

	return startLine, endLine, sectionName, true
}

// FilterByLevel filters entries by heading level.
func FilterByLevel(entries []*TagEntry, level int) []*TagEntry {
	var filtered []*TagEntry
	for _, entry := range entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// FilterByPattern filters entries by a pattern (case-insensitive substring match).
func FilterByPattern(entries []*TagEntry, pattern string) []*TagEntry {
	if pattern == "" {
		return entries
	}

	var filtered []*TagEntry
	lowerPattern := strings.ToLower(pattern)
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lowerPattern) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// SortByLine sorts entries by line number in ascending order.
// This is useful after parsing to ensure entries are in document order.
func SortByLine(entries []*TagEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Line < entries[j].Line
	})
}

// FilterByPatternWithParents filters entries by pattern but preserves parent
// sections to maintain tree hierarchy. This ensures that matching sections are
// shown in context.
//
// Example: If searching for "Testing" matches "Section 2.1: Testing", the
// result will include "Section 2" (parent) even if it doesn't match the pattern.
func FilterByPatternWithParents(
	entries []*TagEntry,
	pattern string,
) []*TagEntry {
	if pattern == "" {
		return entries
	}

	// First pass: identify all matching entries and their descendants
	matchingIndices := make(map[int]bool)
	lowerPattern := strings.ToLower(pattern)

	for i, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lowerPattern) {
			matchingIndices[i] = true
		}
	}

	if len(matchingIndices) == 0 {
		return []*TagEntry{}
	}

	// Second pass: mark entries that should be included (matches + their parents)
	shouldInclude := make(map[int]bool)

	// Mark all matches
	for i := range matchingIndices {
		shouldInclude[i] = true
	}

	// For each match, mark all its parents
	for matchIdx := range matchingIndices {
		matchLevel := entries[matchIdx].Level

		// Look backwards to find parents (entries with lower level)
		for i := matchIdx - 1; i >= 0; i-- {
			if entries[i].Level < matchLevel {
				shouldInclude[i] = true
				matchLevel = entries[i].Level // Update to find higher-level parents
			}
		}
	}

	// Third pass: build result from marked entries
	var result []*TagEntry
	for i, entry := range entries {
		if shouldInclude[i] {
			result = append(result, entry)
		}
	}

	return result
}

// FilterByDepth filters entries to maximum heading level depth.
// depth=1 shows only H1, depth=2 shows H1+H2, depth=0 shows all.
func FilterByDepth(entries []*TagEntry, depth int) []*TagEntry {
	if depth <= 0 {
		return entries
	}

	var filtered []*TagEntry
	for _, entry := range entries {
		if entry.Level <= depth {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
