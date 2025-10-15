package ctags

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

// TagEntry represents a single ctags entry.
type TagEntry struct {
	Name    string
	File    string
	Pattern string
	Kind    string
	Line    int
	Scope   string // Full scope with separators
	Level   int    // Heading level (1-4)
}

// kindLevelMap maps ctags kind to heading level.
var kindLevelMap = map[string]int{
	"chapter":       1, // H1: #
	"section":       2, // H2: ##
	"subsection":    3, // H3: ###
	"subsubsection": 4, // H4: ####
}

// NewTagEntry creates a new TagEntry with level determined from kind.
func NewTagEntry(
	name, file, pattern, kind string,
	line int,
	scope string,
) *TagEntry {
	level := kindLevelMap[kind]
	return &TagEntry{
		Name:    name,
		File:    file,
		Pattern: pattern,
		Kind:    kind,
		Line:    line,
		Scope:   scope,
		Level:   level,
	}
}

// ParseTagsFile parses a ctags file and extracts entries for the target file.
func ParseTagsFile(tagsPath, targetFile string) ([]*TagEntry, error) {
	file, err := os.Open(tagsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tags file: %w", err)
	}
	defer file.Close()

	var entries []*TagEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip meta lines starting with !_TAG
		if strings.HasPrefix(line, "!_TAG") {
			continue
		}

		// Parse tab-separated fields
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}

		name := parts[0]
		fileName := parts[1]
		pattern := parts[2]

		// Only process entries for target file
		if fileName != targetFile {
			continue
		}

		// Parse the rest of the fields
		var kind string
		var lineNum int
		var scope string

		for _, part := range parts[3:] {
			if strings.HasPrefix(part, ";\"") {
				// Skip extension marker
				continue
			} else if _, ok := kindLevelMap[part]; ok {
				kind = part
			} else if strings.HasPrefix(part, "line:") {
				numStr := strings.TrimPrefix(part, "line:")
				lineNum, _ = strconv.Atoi(numStr)
			} else if strings.HasPrefix(part, "chapter:") ||
				strings.HasPrefix(part, "section:") ||
				strings.HasPrefix(part, "subsection:") {
				// Extract scope (parent hierarchy)
				colonIdx := strings.Index(part, ":")
				if colonIdx != -1 {
					scope = part[colonIdx+1:]
				}
			}
		}

		if kind != "" {
			entry := NewTagEntry(name, fileName, pattern, kind, lineNum, scope)
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading tags file: %w", err)
	}

	// Sort by line number
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Line < entries[j].Line
	})

	return entries, nil
}

// FindSectionBounds finds the start and end line numbers for a section.
func FindSectionBounds(
	entries []*TagEntry,
	sectionQuery string,
) (startLine, endLine int, sectionName string, found bool) {
	// Find matching section (case-insensitive substring match)
	var targetIdx int
	var targetKind string
	lowerQuery := strings.ToLower(sectionQuery)

	for idx, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name), lowerQuery) {
			targetIdx = idx
			targetKind = entry.Kind
			sectionName = entry.Name
			found = true
			break
		}
	}

	if !found {
		return 0, 0, "", false
	}

	startLine = entries[targetIdx].Line
	targetLevel := kindLevelMap[targetKind]

	// Find end line (next section at same or higher level)
	endLine = 0 // 0 means EOF
	for idx := targetIdx + 1; idx < len(entries); idx++ {
		entryLevel := kindLevelMap[entries[idx].Kind]
		if entryLevel <= targetLevel {
			endLine = entries[idx].Line - 1
			break
		}
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
