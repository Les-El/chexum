package console

import (
	"os"
	"testing"
	"github.com/Les-El/hashi/internal/config"
)

func TestOutputManager_Prompt(t *testing.T) {
	r, w, _ := os.Pipe()
	m := NewOutputManager(config.DefaultConfig(), r)

	go func() {
		w.WriteString("y\n")
	}()

	if !m.prompt("test") {
		t.Error("Expected true for 'y'")
	}

	go func() {
		w.WriteString("n\n")
		w.Close()
	}()
	if m.prompt("test") {
		t.Error("Expected false for 'n'")
	}
}

func TestOutputManager_OpenOutputFile_Coverage(t *testing.T) {
	m := NewOutputManager(config.DefaultConfig(), nil)

	t.Run("EmptyPath", func(t *testing.T) {
		f, err := m.OpenOutputFile("", false, false)
		if f != nil || err != nil {
			t.Error("Expected nil, nil for empty path")
		}
	})

	t.Run("ExistingFileNoForce", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "existing")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		_, err := m.OpenOutputFile(tmpFile.Name(), false, false)
		if err == nil {
			t.Error("Expected error for existing file without force")
		}
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "console_test")
		defer os.RemoveAll(tmpDir)
		path := tmpDir + "/sub/dir/file.txt"

		f, err := m.OpenOutputFile(path, false, false)
		if err != nil {
			t.Fatalf("Failed to create file in new directory: %v", err)
		}
		f.Close()
	})
}

func TestOutputManager_OpenJSONLog_Coverage(t *testing.T) {
	m := NewOutputManager(config.DefaultConfig(), nil)

	t.Run("EmptyPath", func(t *testing.T) {
		f, err := m.OpenJSONLog("")
		if f != nil || err != nil {
			t.Error("Expected nil, nil for empty path")
		}
	})

	t.Run("NewAndAppend", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "test.json")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		w, err := m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"a":1}`))
		w.Close()

		// Append
		w, err = m.OpenJSONLog(tmpFile.Name())
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(`{"b":2}`))
		w.Close()

		content, _ := os.ReadFile(tmpFile.Name())
		expected := "[\n{\"a\":1}\n,\n{\"b\":2}\n]"
		if string(content) != expected {
			t.Errorf("Unexpected JSON content:\nGot:  %q\nWant: %q", string(content), expected)
		}
	})
}

func TestInitStreams_Coverage(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		cfg := &config.Config{}
		streams, cleanup, err := InitStreams(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		if streams.Out == nil {
			t.Error("Expected non-nil Out")
		}
	})

	t.Run("WithOutputFile", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "out")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cfg := &config.Config{OutputFile: tmpFile.Name(), Force: true}
		streams, cleanup, err := InitStreams(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
		streams.Out.Write([]byte("hello"))
	})

	t.Run("WithLogFile", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "log")
		defer os.Remove(tmpFile.Name())
		tmpFile.Close()

		cfg := &config.Config{LogFile: tmpFile.Name()}
		_, cleanup, err := InitStreams(cfg)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()
	})
}
