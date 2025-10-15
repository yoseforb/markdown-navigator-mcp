package tools

import "errors"

// Static errors for tool operations.
var (
	ErrNoEntries       = errors.New("no entries found")
	ErrSectionNotFound = errors.New("section not found")
	ErrInvalidLevel    = errors.New("invalid heading level")
	ErrInvalidFormat   = errors.New("invalid format")
)
