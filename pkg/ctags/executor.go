package ctags

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	// CtagsExecutionTimeout is the maximum time allowed for ctags execution.
	// This prevents hanging on very large files or ctags issues.
	CtagsExecutionTimeout = 5 * time.Second

	// CtagsBinary is the name of the ctags executable to search for in PATH.
	CtagsBinary = "ctags"
)

// Config holds the global configuration for ctags execution.
// This pattern is acceptable for configuration as it provides a single point of
// coordination for ctags operations throughout the application.
type Config struct {
	ctagsPath string
	mu        sync.RWMutex
}

// globalConfig is the singleton configuration instance.
var globalConfig = &Config{ //nolint:gochecknoglobals // singleton config pattern
	ctagsPath: CtagsBinary,
	mu:        sync.RWMutex{},
}

// SetCtagsPath sets the path to the ctags executable.
// This should be called once during application initialization.
// If path is empty, it defaults to "ctags" (searches in PATH).
func SetCtagsPath(path string) error {
	if path == "" {
		path = CtagsBinary
	}

	// Validate that the path points to an executable
	resolvedPath, err := exec.LookPath(path)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidCtagsPath, path)
	}

	globalConfig.mu.Lock()
	globalConfig.ctagsPath = resolvedPath
	globalConfig.mu.Unlock()

	return nil
}

// GetCtagsPath returns the currently configured ctags executable path.
func GetCtagsPath() string {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.ctagsPath
}

// ExecuteCtags executes Universal Ctags on a markdown file and returns JSON output.
// It includes timeout protection, validates that ctags is installed, and checks
// that the file exists before execution.
//
// The function executes:
//
//	ctags --output-format=json --fields=+KnSe --languages=markdown -f - <file>
//
// Returns the raw JSON output suitable for parsing with ParseJSONTags.
// Errors include: ErrFileNotFound, ErrCtagsNotFound, ErrCtagsTimeout, ErrCtagsExecution.
func ExecuteCtags(filePath string) ([]byte, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, filePath)
	}

	// Get configured ctags path
	ctagsPath := GetCtagsPath()

	// Validate ctags executable exists
	if _, err := exec.LookPath(ctagsPath); err != nil {
		return nil, fmt.Errorf(
			"%w: install universal-ctags (https://github.com/universal-ctags/ctags)",
			ErrCtagsNotFound,
		)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(
		context.Background(),
		CtagsExecutionTimeout,
	)
	defer cancel()

	// Build ctags command
	cmd := exec.CommandContext(
		ctx,
		ctagsPath,
		"--output-format=json", // JSON output
		"--fields=+KnSe",       // Include kind, line number, scope, end line
		"--languages=markdown", // Only markdown
		"-f", "-",              // Output to stdout
		filePath,
	)

	// Execute command
	output, err := cmd.Output()
	if err != nil {
		// Check for timeout
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf(
				"%w: exceeded %v for file %s",
				ErrCtagsTimeout,
				CtagsExecutionTimeout,
				filePath,
			)
		}

		// Check for execution error
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf(
				"%w: %w (stderr: %s)",
				ErrCtagsExecution,
				err,
				string(exitErr.Stderr),
			)
		}

		return nil, fmt.Errorf("%w: %w", ErrCtagsExecution, err)
	}

	return output, nil
}

// IsCtagsInstalled checks if Universal Ctags is available at the configured path.
// This can be used for pre-flight checks or diagnostics.
func IsCtagsInstalled() bool {
	ctagsPath := GetCtagsPath()
	_, err := exec.LookPath(ctagsPath)
	return err == nil
}
