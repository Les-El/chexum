package console

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"

	"hashi/internal/config"
)

func TestOutputManager_OpenOutputFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hashi-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := config.DefaultConfig()
	
t.Run("creates new file", func(t *testing.T) {
		manager := NewOutputManager(cfg, nil)
		path := filepath.Join(tmpDir, "new.txt")
		
f, err := manager.OpenOutputFile(path, false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f == nil {
			t.Fatal("expected file, got nil")
		}
		f.Close()
		
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("file was not created")
		}
	})

	t.Run("fails if file exists without force", func(t *testing.T) {
		// Use a non-terminal reader to ensure isInteractive() is false
		manager := NewOutputManager(cfg, strings.NewReader(""))
		path := filepath.Join(tmpDir, "exists.txt")
		os.WriteFile(path, []byte("content"), 0644)
		
		_, err := manager.OpenOutputFile(path, false, false)
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "exists") {
			t.Errorf("expected exists error, got: %v", err)
		}
	})

	t.Run("overwrites if force is true", func(t *testing.T) {
		manager := NewOutputManager(cfg, nil)
		path := filepath.Join(tmpDir, "overwrite.txt")
		os.WriteFile(path, []byte("old content"), 0644)
		
f, err := manager.OpenOutputFile(path, false, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		f.Write([]byte("new content"))
		f.Close()
		
		content, _ := os.ReadFile(path)
		if string(content) != "new content" {
			t.Errorf("expected 'new content', got %q", string(content))
		}
	})

	t.Run("appends if append is true", func(t *testing.T) {
		manager := NewOutputManager(cfg, nil)
		path := filepath.Join(tmpDir, "append.txt")
		os.WriteFile(path, []byte("first "), 0644)
		
f, err := manager.OpenOutputFile(path, true, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		f.Write([]byte("second"))
		f.Close()
		
		content, _ := os.ReadFile(path)
		if string(content) != "first second" {
			t.Errorf("expected 'first second', got %q", string(content))
		}
	})

	t.Run("creates subdirectories", func(t *testing.T) {
		manager := NewOutputManager(cfg, nil)
		path := filepath.Join(tmpDir, "sub/dir/file.txt")
		
f, err := manager.OpenOutputFile(path, false, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		f.Close()
		
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("file in subdirectory was not created")
		}
	})

	t.Run("prompts for confirmation when interactive", func(t *testing.T) {
		// Mock stdin with "y\n"
		in := strings.NewReader("y\n")
		// We can't easily mock isInteractive() because it checks if it's an *os.File.
		// However, we can test the prompt logic directly if we export it or test it via a helper.
		// For now, let's just test that the logic works if we assume interactive.
		
		// Actually, let's test the prompt() method directly.
		manager := &OutputManager{cfg: cfg, in: in}
		if !manager.prompt("test") {
			t.Error("expected true for 'y' response")
		}
		
		manager.in = strings.NewReader("n\n")
		if manager.prompt("test") {
			t.Error("expected false for 'n' response")
		}
	})

	t.Run("handles JSON log append", func(t *testing.T) {
		manager := NewOutputManager(cfg, nil)
		path := filepath.Join(tmpDir, "log.json")
		
		// 1. First write
		f1, _ := manager.OpenJSONLog(path)
		f1.Write([]byte(`{"run": 1}`))
		f1.Close()
		
		content1, _ := os.ReadFile(path)
		expected1 := "[\n{\"run\": 1}\n]"
		if string(content1) != expected1 {
			t.Errorf("expected %q, got %q", expected1, string(content1))
		}
		
		// 2. Second write (append)
		f2, _ := manager.OpenJSONLog(path)
		f2.Write([]byte(`{"run": 2}`))
		f2.Close()
		
		content2, _ := os.ReadFile(path)
		// Current logic adds newline before comma if previous run ended with newline
		expected2 := "[\n{\"run\": 1}\n,\n{\"run\": 2}\n]"
		if string(content2) != expected2 {
			t.Errorf("expected %q, got %q", expected2, string(content2))
		}
	})
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
