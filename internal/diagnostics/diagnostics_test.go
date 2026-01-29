package diagnostics

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Les-El/hashi/internal/config"
	"github.com/Les-El/hashi/internal/console"
)

func TestRunDiagnostics(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hashi-diag-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		cfg      *config.Config
		contains []string
	}{
		{
			name: "Basic info",
			cfg:  &config.Config{Algorithm: "sha256"},
			contains: []string{
				"System Information",
				"Algorithm 'sha256' sanity check passed",
			},
		},
		{
			name: "File inspection",
			cfg: &config.Config{
				Algorithm: "sha256",
				Files:     []string{testFile},
			},
			contains: []string{
				"Inspecting 1 input arguments",
				"Checking '" + testFile + "'",
				"Exists: YES",
				"Size: 5 bytes",
				"Readable: YES",
			},
		},
		{
			name: "Missing file",
			cfg: &config.Config{
				Algorithm: "sha256",
				Files:     []string{filepath.Join(tmpDir, "missing.txt")},
			},
			contains: []string{
				"Exists: NO",
				"Parent directory exists: YES",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outBuf bytes.Buffer
			streams := &console.Streams{
				Out: &outBuf,
				Err: &outBuf,
			}

			exitCode := RunDiagnostics(tt.cfg, streams)
			if exitCode != config.ExitSuccess {
				t.Errorf("Expected exit code %d, got %d", config.ExitSuccess, exitCode)
			}

			output := outBuf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("Output missing expected string: %q", s)
				}
			}
		})
	}
}
