package console

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"

	"github.com/Les-El/hashi/internal/config"
)

func TestOpenOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	
	t.Run("creates new file", func(t *testing.T) {
		testCreateNewFile(t, tmpDir, cfg)
	})

	t.Run("fails if file exists without force", func(t *testing.T) {
		testFailsIfExistsWithoutForce(t, tmpDir, cfg)
	})

	t.Run("overwrites if force is true", func(t *testing.T) {
		testOverwritesIfForceIsTrue(t, tmpDir, cfg)
	})

	t.Run("appends if append is true", func(t *testing.T) {
		testAppendsIfAppendIsTrue(t, tmpDir, cfg)
	})
}

func testCreateNewFile(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "new.txt")
	if f, err := manager.OpenOutputFile(path, false, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	} else {
		f.Close()
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func testFailsIfExistsWithoutForce(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, strings.NewReader(""))
	path := filepath.Join(tmpDir, "exists.txt")
	os.WriteFile(path, []byte("content"), 0644)
	if _, err := manager.OpenOutputFile(path, false, false); err == nil {
		t.Error("expected error")
	}
}

func testOverwritesIfForceIsTrue(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "overwrite.txt")
	os.WriteFile(path, []byte("old"), 0644)
	if f, err := manager.OpenOutputFile(path, false, true); err == nil {
		f.Write([]byte("new"))
		f.Close()
	}
	content, _ := os.ReadFile(path)
	if string(content) != "new" {
		t.Errorf("got %q", string(content))
	}
}

func testAppendsIfAppendIsTrue(t *testing.T, tmpDir string, cfg *config.Config) {
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "append.txt")
	os.WriteFile(path, []byte("first "), 0644)
	if f, err := manager.OpenOutputFile(path, true, false); err == nil {
		f.Write([]byte("second"))
		f.Close()
	}
	content, _ := os.ReadFile(path)
	if string(content) != "first second" {
		t.Errorf("got %q", string(content))
	}
}

func TestNewOutputManager(t *testing.T) {
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	if manager == nil {
		t.Fatal("NewOutputManager returned nil")
	}
}

func TestOpenJSONLog(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	path := filepath.Join(tmpDir, "log.json")
	f, err := manager.OpenJSONLog(path)
	if err != nil {
		t.Fatalf("OpenJSONLog failed: %v", err)
	}
	f.Close()
}

func TestWrite(t *testing.T) {
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	n, err := manager.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}
}

func TestClose(t *testing.T) {
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)
	if err := manager.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestProperty_AppendModePreservesContent verifies that append mode preserves existing content.
// Property 12: Append mode preserves existing content
func TestProperty_AppendModePreservesContent(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "hashi-prop-*")
	defer os.RemoveAll(tmpDir)
	
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)

	f := func(initial, addition string) bool {
		path := filepath.Join(tmpDir, "prop_append.txt")
		os.Remove(path)
		
		// Write initial
		os.WriteFile(path, []byte(initial), 0644)
		
		// Append
		f, err := manager.OpenOutputFile(path, true, false)
		if err != nil {
			return false
		}
		f.Write([]byte(addition))
		f.Close()
		
		content, _ := os.ReadFile(path)
		return string(content) == initial+addition
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestProperty_JSONLogValidity verifies that JSON log append maintains array validity.
// Property 13: JSON log append maintains validity
func TestProperty_JSONLogValidity(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "hashi-prop-json-*")
	defer os.RemoveAll(tmpDir)
	
	cfg := config.DefaultConfig()
	manager := NewOutputManager(cfg, nil)

	// Simplified check: can it be parsed as an array?
	f := func(entries []string) bool {
		if len(entries) == 0 {
			return true
		}
		
		path := filepath.Join(tmpDir, "prop_log.json")
		os.Remove(path)
		
		for _, entry := range entries {
			// Use json.Marshal to ensure valid JSON string escaping
			entryData, _ := json.Marshal(entry)
			f, err := manager.OpenJSONLog(path)
			if err != nil {
				return false
			}
			f.Write(entryData)
			f.Close()
		}
		
		content, err := os.ReadFile(path)
		if err != nil {
			return false
		}
		
		var result []string
		err = json.Unmarshal(content, &result)
		if err != nil {
			fmt.Printf("Unmarshal error: %v\nContent:\n%s\n", err, string(content))
			return false
		}
		return len(result) == len(entries)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
