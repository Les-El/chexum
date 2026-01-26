package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain runs cleanup after all tests in the checkpoint package
// This removes temporary files created during testing to prevent disk space buildup
func TestMain(m *testing.M) {
	code := m.Run()
	cleanupTemporaryFiles()
	os.Exit(code)
}

func cleanupTemporaryFiles() {
	// Clean up only temporary files created by tests, leaving Go build artifacts alone
	tmpPatterns := []string{
		"/tmp/hashi-*",
		"/tmp/checkpoint-*",
		"/tmp/test-*",
	}

	for _, pattern := range tmpPatterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			os.RemoveAll(match)
		}
	}
}

