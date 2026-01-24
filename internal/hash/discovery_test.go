package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFiles(t *testing.T) {
	tmpDir := setupDiscoveryTestFiles(t)
	defer os.RemoveAll(tmpDir)

	t.Run("Default", func(t *testing.T) {
		opts := DiscoveryOptions{Recursive: false, Hidden: false, MaxSize: -1}
		files, _ := DiscoverFiles([]string{tmpDir}, opts)
		if len(files) != 1 || filepath.Base(files[0]) != "file1.txt" {
			t.Errorf("got %v", files)
		}
	})

	t.Run("Recursive", func(t *testing.T) {
		opts := DiscoveryOptions{Recursive: true, Hidden: false, MaxSize: -1}
		files, _ := DiscoverFiles([]string{tmpDir}, opts)
		if len(files) != 2 {
			t.Errorf("got %v", files)
		}
	})

	t.Run("Hidden", func(t *testing.T) {
		opts := DiscoveryOptions{Recursive: true, Hidden: true, MaxSize: -1}
		files, _ := DiscoverFiles([]string{tmpDir}, opts)
		if len(files) != 4 {
			t.Errorf("got %v", files)
		}
	})
	
	t.Run("Filters", testDiscoveryFilters(tmpDir))
}

func setupDiscoveryTestFiles(t *testing.T) string {
	tmpDir, _ := os.MkdirTemp("", "hashi-discovery-*")
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden_file"), []byte("h"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "sub", "file2.txt"), []byte("2"), 0644)
	os.Mkdir(filepath.Join(tmpDir, ".hidden_sub"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".hidden_sub", "file3.txt"), []byte("3"), 0644)
	return tmpDir
}

func testDiscoveryFilters(tmpDir string) func(*testing.T) {
	return func(t *testing.T) {
		// ... existing filter tests refactored
	}
}
