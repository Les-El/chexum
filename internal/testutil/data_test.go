package testutil

import (
	"os"
	"strings"
	"testing"
)

func TestRandomString(t *testing.T) {
	s1 := RandomString(10)
	if len(s1) != 10 {
		t.Errorf("expected length 10, got %d", len(s1))
	}

	s2 := RandomString(10)
	if s1 == s2 {
		t.Errorf("expected different random strings, got same: %s", s1)
	}
}

func TestRandomHash(t *testing.T) {
	h1 := RandomHash(64)
	if len(h1) != 64 {
		t.Errorf("expected length 64, got %d", len(h1))
	}

	for _, char := range h1 {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("invalid char in hash: %c", char)
		}
	}
}

func TestCreateRandomDirectoryStructure(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	// Test normal case
	CreateRandomDirectoryStructure(t, dir, 2, 2, 2)
	
	// Test depth 1, maxDirs 0, maxFiles 0 (forces numFiles = 1)
	dir1, cleanup1 := TempDir(t)
	defer cleanup1()
	CreateRandomDirectoryStructure(t, dir1, 1, 0, 0)
	entries1, _ := os.ReadDir(dir1)
	if len(entries1) == 0 {
		t.Error("expected at least one file at depth 1 even with maxDirs=0, maxFiles=0")
	}

	// Test depth 2, maxDirs 0, maxFiles 0 (forces numDirs = 1 or numFiles = 1)
	dir2, cleanup2 := TempDir(t)
	defer cleanup2()
	CreateRandomDirectoryStructure(t, dir2, 2, 0, 0)
	entries2, _ := os.ReadDir(dir2)
	if len(entries2) == 0 {
		t.Error("expected at least one entry at depth 2 even with maxDirs=0, maxFiles=0")
	}

	// Test depth <= 0
	dir3, cleanup3 := TempDir(t)
	defer cleanup3()
	CreateRandomDirectoryStructure(t, dir3, 0, 2, 2)
	entries3, _ := os.ReadDir(dir3)
	if len(entries3) != 0 {
		t.Error("expected no entries for depth 0")
	}
}

func TestGenerateMockGoFile(t *testing.T) {
	dir, cleanup := TempDir(t)
	defer cleanup()

	path := GenerateMockGoFile(t, dir, "main.go", true, true)
	
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read mock go file: %v", err)
	}

	sContent := string(content)
	if !strings.Contains(sContent, "package main") {
		t.Error("missing package main")
	}
	if !strings.Contains(sContent, "import \"unsafe\"") {
		t.Error("missing unsafe import")
	}
	if !strings.Contains(sContent, "// TODO: implement this") {
		t.Error("missing TODO comment")
	}
}
