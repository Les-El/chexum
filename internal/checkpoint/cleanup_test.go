package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCleanupManager(t *testing.T) {
	cleanup := NewCleanupManager(true)
	if cleanup == nil {
		t.Error("Expected non-nil cleanup manager")
	}
	if !cleanup.verbose {
		t.Error("Expected verbose mode to be enabled")
	}
}

func TestCheckTmpfsUsage(t *testing.T) {
	cleanup := NewCleanupManager(false)
	
	needsCleanup, usage := cleanup.CheckTmpfsUsage(100.0) // Set threshold to 100% so it shouldn't trigger
	if needsCleanup {
		t.Errorf("Expected needsCleanup to be false with 100%% threshold, got true (usage: %.1f%%)", usage)
	}
	
	if usage < 0 || usage > 100 {
		t.Errorf("Expected usage to be between 0-100%%, got %.1f%%", usage)
	}
}

func TestCleanupManager_FormatBytes(t *testing.T) {
	cleanup := NewCleanupManager(false)
	
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	
	for _, test := range tests {
		result := cleanup.formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestCleanupManager_GetDirSize(t *testing.T) {
	cleanup := NewCleanupManager(false)
	
	// Create a temporary directory with some files
	tmpDir := t.TempDir()
	
	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.txt")
	testFile2 := filepath.Join(tmpDir, "test2.txt")
	
	content1 := "Hello, World!"
	content2 := "This is a test file."
	
	if err := os.WriteFile(testFile1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	if err := os.WriteFile(testFile2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	size, err := cleanup.getDirSize(tmpDir)
	if err != nil {
		t.Fatalf("getDirSize failed: %v", err)
	}
	
	expectedSize := int64(len(content1) + len(content2))
	if size != expectedSize {
		t.Errorf("Expected directory size %d, got %d", expectedSize, size)
	}
}

func TestCleanupTemporaryFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cleanup test in short mode")
	}
	
	cleanup := NewCleanupManager(false) // Non-verbose for test
	
	// This test doesn't actually create/remove files in /tmp to avoid side effects
	// Instead, it just verifies the function runs without error
	result, err := cleanup.CleanupTemporaryFiles()
	if err != nil {
		t.Fatalf("CleanupTemporaryFiles failed: %v", err)
	}
	
	if result == nil {
		t.Error("Expected non-nil result")
	}
	
	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}
	
	if result.TmpfsUsageBefore < 0 || result.TmpfsUsageBefore > 100 {
		t.Errorf("Expected tmpfs usage before to be 0-100%%, got %.1f%%", result.TmpfsUsageBefore)
	}
	
	if result.TmpfsUsageAfter < 0 || result.TmpfsUsageAfter > 100 {
		t.Errorf("Expected tmpfs usage after to be 0-100%%, got %.1f%%", result.TmpfsUsageAfter)
	}
}

func TestCleanupOnExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	
	cleanup := NewCleanupManager(false)
	
	// Test that cleanup can be called without error
	err := cleanup.CleanupOnExit()
	if err != nil {
		t.Errorf("CleanupOnExit failed: %v", err)
	}
}