// Package ctags provides markdown document structure parsing using
// Universal Ctags with intelligent mtime-based caching.
//
// The package automatically executes ctags on markdown files and caches
// the parsed results in memory. Cache invalidation is based on file
// modification time (mtime), providing fast cache validation with
// automatic invalidation when files change.
//
// Performance characteristics:
//   - Cache hit: ~528ns (mtime check only)
//   - Cache miss: ~13ms (ctags execution + JSON parsing)
//   - Typical usage: 90%+ cache hit rate
//
// Example usage:
//
//	cache := ctags.GetGlobalCache()
//	tags, err := cache.GetTags("/path/to/file.md")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, tag := range tags {
//	    fmt.Printf("%s (line %d)\n", tag.Name, tag.Line)
//	}
package ctags

import (
	"encoding/json"
	"fmt"
	"path/filepath"
)

// JSONEntry represents a single ctags entry in JSON format.
// This structure maps directly to the JSON output from Universal Ctags
// with --output-format=json flag.
type JSONEntry struct {
	Type      string `json:"_type"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Pattern   string `json:"pattern"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`
	Scope     string `json:"scope,omitempty"`
	ScopeKind string `json:"scopeKind,omitempty"`
	End       int    `json:"end,omitempty"`
}

// ParseJSONTags parses ctags JSON output and converts it to TagEntry structs.
// It filters entries to only include those from the target file and converts
// ctags "kind" fields (chapter, section, subsection, subsubsection) to heading
// levels (1, 2, 3, 4).
//
// The function handles NDJSON (newline-delimited JSON) format where each line
// is a separate JSON object. Invalid JSON lines and non-tag entries are skipped.
//
// Returns an empty slice (not an error) if no valid entries are found.
func ParseJSONTags(jsonData []byte, targetFile string) ([]*TagEntry, error) {
	if len(jsonData) == 0 {
		return []*TagEntry{}, nil
	}

	// Get absolute path for comparison
	targetAbs, err := filepath.Abs(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target file path: %w", err)
	}

	var entries []*TagEntry

	// Parse JSON line by line (ctags outputs NDJSON - newline delimited JSON)
	lines := splitLines(jsonData)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		var jsonEntry JSONEntry
		if err := json.Unmarshal(line, &jsonEntry); err != nil {
			// Skip invalid JSON lines (metadata or malformed entries)
			continue
		}

		// Skip non-tag entries
		if jsonEntry.Type != "tag" {
			continue
		}

		// Resolve entry path to absolute
		entryAbs, err := filepath.Abs(jsonEntry.Path)
		if err != nil {
			// If we can't resolve, try simple comparison
			entryAbs = jsonEntry.Path
		}

		// Filter by target file
		if entryAbs != targetAbs && jsonEntry.Path != targetFile {
			continue
		}

		// Map JSON entry to TagEntry
		entry := jsonEntryToTagEntry(&jsonEntry)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// jsonEntryToTagEntry converts a JSONEntry to a TagEntry.
// Returns nil if the entry has an unknown or invalid kind.
func jsonEntryToTagEntry(jsonEntry *JSONEntry) *TagEntry {
	// Skip entries without a valid kind
	level, exists := kindLevelMap[jsonEntry.Kind]
	if !exists {
		return nil
	}

	return &TagEntry{
		Name:    jsonEntry.Name,
		File:    jsonEntry.Path,
		Pattern: jsonEntry.Pattern,
		Kind:    jsonEntry.Kind,
		Line:    jsonEntry.Line,
		End:     jsonEntry.End,
		Scope:   jsonEntry.Scope,
		Level:   level,
	}
}

// splitLines splits byte data into lines for NDJSON parsing.
// Each line becomes a separate byte slice for individual JSON parsing.
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	var line []byte

	for _, b := range data {
		if b == '\n' {
			if len(line) > 0 {
				// Make a copy to avoid data sharing
				lineCopy := make([]byte, len(line))
				copy(lineCopy, line)
				lines = append(lines, lineCopy)
				line = line[:0]
			}
		} else {
			line = append(line, b)
		}
	}

	// Add last line if it doesn't end with newline
	if len(line) > 0 {
		lines = append(lines, line)
	}

	return lines
}
