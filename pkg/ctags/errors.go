package ctags

import "errors"

// Ctags execution errors.
var (
	ErrCtagsNotFound    = errors.New("ctags not found in PATH")
	ErrCtagsExecution   = errors.New("ctags execution failed")
	ErrCtagsTimeout     = errors.New("ctags execution timeout")
	ErrFileNotFound     = errors.New("file not found")
	ErrInvalidCtagsPath = errors.New("invalid ctags executable path")
)
