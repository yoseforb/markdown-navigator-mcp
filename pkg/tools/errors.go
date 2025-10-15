package tools

import "errors"

// Static errors for tool operations.
var (
	ErrNoEntries       = errors.New("no entries found in tags file")
	ErrSectionNotFound = errors.New("section not found")
	ErrInvalidLevel    = errors.New("invalid heading level")
)

// DefaultTagsFile is the default name for the ctags file.
const DefaultTagsFile = "tags"
