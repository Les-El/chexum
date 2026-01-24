// Package console handles the output streams and file management.
package console

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Les-El/hashi/internal/config"
)

// OutputManager handles safe file operations for output and logging.
type OutputManager struct {
	cfg *config.Config
	in  io.Reader // For confirmation prompts
}

// NewOutputManager creates a new OutputManager.
func NewOutputManager(cfg *config.Config, in io.Reader) *OutputManager {
	if in == nil {
		in = os.Stdin
	}
	return &OutputManager{
		cfg: cfg,
		in:  in,
	}
}

// OpenOutputFile opens a file for output with safety checks.
// It handles overwrite protection, append mode, and directory creation.
func (m *OutputManager) OpenOutputFile(path string, appendMode bool, force bool) (io.WriteCloser, error) {
	if path == "" {
		return nil, nil
	}

	// 1. Check if file exists for overwrite protection
	if _, err := os.Stat(path); err == nil {
		if !appendMode && !force {
			// Prompt for confirmation if interactive
			if m.isInteractive() {
				if !m.prompt(fmt.Sprintf("File %s exists. Overwrite?", path)) {
					return nil, fmt.Errorf("operation cancelled by user")
				}
			} else {
				return nil, fmt.Errorf("output file %s exists (use --force to overwrite or --append to append)", path)
			}
		}
	}

	// 2. Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, config.FileSystemError(m.cfg.Verbose, fmt.Sprintf("failed to create directory for %s: %v", path, err))
	}

	// 3. Determine flags
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	// 4. Open file
	// Note: Atomic writes (temp file + rename) will be enhanced in Task 42.
	// For now, we use standard file opening but wrapped in a helper.
	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, config.HandleFileWriteError(err, m.cfg.Verbose, path)
	}

	return f, nil
}

// isInteractive checks if the input is a terminal.
func (m *OutputManager) isInteractive() bool {
	if f, ok := m.in.(*os.File); ok {
		stat, err := f.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// prompt asks the user for confirmation.
func (m *OutputManager) prompt(msg string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", msg)
	scanner := bufio.NewScanner(m.in)
	if scanner.Scan() {
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		return response == "y" || response == "yes"
	}
	return false
}

// OpenJSONLog opens a JSON file for logging, maintaining array validity.
func (m *OutputManager) OpenJSONLog(path string) (io.WriteCloser, error) {
	if path == "" {
		return nil, nil
	}

	// 1. Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, config.FileSystemError(m.cfg.Verbose, fmt.Sprintf("failed to create directory for %s: %v", path, err))
	}

	// 2. Open file for read/write
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, config.HandleFileWriteError(err, m.cfg.Verbose, path)
	}

	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	isNew := stat.Size() == 0
	return &jsonLogWriter{f: f, isNew: isNew}, nil
}

type jsonLogWriter struct {
	f     *os.File
	isNew bool
}

// Write appends data to the JSON log, handling array formatting.
func (w *jsonLogWriter) Write(p []byte) (n int, err error) {
	if w.isNew {
		if _, err := w.f.Write([]byte("[\n")); err != nil {
			return 0, err
		}
		w.isNew = false
	} else {
		// Seek to find the closing bracket
		// For simplicity in this version, we seek to the end and look back
		// A more robust version would handle trailing whitespace
		pos, err := w.f.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}
		if pos > 0 {
			// Back up to remove ']'
			if _, err := w.f.Seek(-1, io.SeekEnd); err != nil {
				return 0, err
			}
			// Read last char
			buf := make([]byte, 1)
			if _, err := w.f.Read(buf); err != nil {
				return 0, err
			}
			if buf[0] == ']' {
				if _, err := w.f.Seek(-1, io.SeekEnd); err != nil {
					return 0, err
				}
				if _, err := w.f.Write([]byte(",\n")); err != nil {
					return 0, err
				}
			} else {
				// If not ']', just append at the end (might be invalid JSON initially)
				if _, err := w.f.Seek(0, io.SeekEnd); err != nil {
					return 0, err
				}
			}
		}
	}
	return w.f.Write(p)
}

// Close finishes the JSON array and closes the file.
func (w *jsonLogWriter) Close() error {
	if _, err := w.f.Write([]byte("\n]")); err != nil {
		w.f.Close()
		return err
	}
	return w.f.Close()
}
