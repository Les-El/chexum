package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureOutput(t *testing.T) {
	stdout, stderr, err := CaptureOutput(func() {
		fmt.Print("hello stdout")
		fmt.Fprint(os.Stderr, "hello stderr")
	})

	if err != nil {
		t.Fatalf("CaptureOutput failed: %v", err)
	}

	if stdout != "hello stdout" {
		t.Errorf("expected stdout %q, got %q", "hello stdout", stdout)
	}

	if stderr != "hello stderr" {
		t.Errorf("expected stderr %q, got %q", "hello stderr", stderr)
	}
}

func TestTempDir(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected temp dir %s to exist", dir)
	}

	if !filepath.HasPrefix(filepath.Base(dir), "hashi-test-") {
		t.Errorf("expected temp dir to have prefix hashi-test-, got %s", filepath.Base(dir))
	}

	cleanup()
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected temp dir %s to be removed after cleanup", dir)
	}
}

func TestCreateFile(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	path := CreateFile(t, dir, "test.txt", "hello world")
	
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected content %q, got %q", "hello world", string(content))
	}
}

func TestAssertExitCode(t *testing.T) {
	// This just checks if it doesn't fail when codes match
	// Since it uses t.Errorf, we can't easily test the failure case without a mock testing.T
	AssertExitCode(t, 0, 0)
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "hello world", "world")
}

func TestAutoCleanupTmpfs(t *testing.T) {
	// Create a file that matches the pattern
	path := filepath.Join(os.TempDir(), "hashi-test-autocleanup")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file in /tmp: %v", err)
	}

	cleaned := AutoCleanupTmpfs(t)
	if !cleaned {
		t.Errorf("expected AutoCleanupTmpfs to return true (cleaned something)")
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file %s to be removed", path)
	}
}

func TestRequireCleanTmpfs(t *testing.T) {
	// Smoke test
	RequireCleanTmpfs(t)
}
