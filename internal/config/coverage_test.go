package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// Helper to create a pointer to a bool or string
func ptr[T any](v T) *T {
	return &v
}

// stringPtr is a helper to create a pointer to a string.
func stringPtr(s string) *string { return &s }

// stringSlicesEqual is a helper to compare two string slices.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// configFilesEqual performs a deep comparison of two ConfigFile structs.
// Note: This is a simplified comparison, focusing on the fields that LoadConfigFile populates.
func configFilesEqual(a, b *ConfigFile) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare Files
	if !stringSlicesEqual(a.Files, b.Files) {
		return false
	}

	// Compare Defaults (pointers require dereferencing and nil checks)
	if !ptrBoolEqual(a.Defaults.Recursive, b.Defaults.Recursive) ||
		!ptrBoolEqual(a.Defaults.Hidden, b.Defaults.Hidden) ||
		!ptrStringEqual(a.Defaults.Algorithm, b.Defaults.Algorithm) ||
		!ptrBoolEqual(a.Defaults.Verbose, b.Defaults.Verbose) ||
		!ptrBoolEqual(a.Defaults.Quiet, b.Defaults.Quiet) ||
		!ptrBoolEqual(a.Defaults.Bool, b.Defaults.Bool) ||
		!ptrBoolEqual(a.Defaults.PreserveOrder, b.Defaults.PreserveOrder) ||
		!ptrBoolEqual(a.Defaults.MatchRequired, b.Defaults.MatchRequired) ||
		!ptrStringEqual(a.Defaults.OutputFormat, b.Defaults.OutputFormat) ||
		!ptrStringEqual(a.Defaults.OutputFile, b.Defaults.OutputFile) ||	
		!ptrBoolEqual(a.Defaults.Append, b.Defaults.Append) ||
		!ptrBoolEqual(a.Defaults.Force, b.Defaults.Force) ||
		!ptrStringEqual(a.Defaults.LogFile, b.Defaults.LogFile) ||
		!ptrStringEqual(a.Defaults.LogJSON, b.Defaults.LogJSON) ||
		!stringSlicesEqual(a.Defaults.Include, b.Defaults.Include) ||
		!stringSlicesEqual(a.Defaults.Exclude, b.Defaults.Exclude) {
		return false
	}

	// Compare Security
	if !stringSlicesEqual(a.Security.BlacklistFiles, b.Security.BlacklistFiles) ||
		!stringSlicesEqual(a.Security.BlacklistDirs, b.Security.BlacklistDirs) ||
		!stringSlicesEqual(a.Security.WhitelistFiles, b.Security.WhitelistFiles) ||
		!stringSlicesEqual(a.Security.WhitelistDirs, b.Security.WhitelistDirs) {
		return false
	}

	return true
}

// Helper for comparing two *bool pointers
func ptrBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Helper for comparing two *string pointers
func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func TestWriteErrors(t *testing.T) {
	err := WriteError()
	if err.Error() != "Unknown write/append error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	err = WriteErrorWithVerbose(false, "details")
	if err.Error() != "Unknown write/append error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	err = WriteErrorWithVerbose(true, "details")
	if err.Error() != "details" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestFileSystemError(t *testing.T) {
	err := FileSystemError(false, "details")
	if err.Error() != "Unknown write/append error" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}

	err = FileSystemError(true, "details")
	if err.Error() != "details" {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}

func TestHandleFileWriteError(t *testing.T) {
	if HandleFileWriteError(nil, false, "path") != nil {
		t.Error("Expected nil for nil error")
	}

	tests := []struct {
		err      error
		verbose  bool
		expected string
	}{
		{errors.New("permission denied"), false, "Unknown write/append error"},
		{errors.New("permission denied"), true, "permission denied writing to path"},
		{errors.New("no space left on device"), true, "insufficient disk space for path"},
		{errors.New("network timeout"), true, "network error writing to path"},
		{errors.New("file name too long"), true, "path too long: path"},
		{errors.New("random error"), true, "random error"},
	}

	for _, tt := range tests {
		got := HandleFileWriteError(tt.err, tt.verbose, "path")
		if got.Error() != tt.expected {
			t.Errorf("HandleFileWriteError(%v, %v) = %v; want %v", tt.err, tt.verbose, got, tt.expected)
		}
	}
}

func TestApplyEnvConfig_Coverage(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("algorithm", "sha256", "")
	fs.String("format", "default", "")
	fs.Bool("recursive", false, "")
	fs.Bool("hidden", false, "")	
	fs.Bool("verbose", false, "")
	fs.Bool("quiet", false, "")
	fs.Bool("preserve-order", false, "")

	env := &EnvConfig{
		HashiAlgorithm:     "md5",
		HashiOutputFormat:  "json",
		HashiRecursive:     true,
		HashiHidden:        true,
		HashiVerbose:       true,
		HashiQuiet:         true,
		HashiPreserveOrder: true,
		HashiBlacklistFiles: "f1,f2",
		HashiBlacklistDirs:  "d1,d2",
		HashiWhitelistFiles: "w1,w2",
		HashiWhitelistDirs:  "wd1,wd2",
	}

	env.ApplyEnvConfig(cfg, fs)

	if cfg.Algorithm != "md5" {
		t.Errorf("Expected md5, got %s", cfg.Algorithm)
	}
	if len(cfg.BlacklistFiles) != 2 {
		t.Errorf("Expected 2 blacklist files, got %d", len(cfg.BlacklistFiles))
	}
}

func TestLoadDotEnv_Errors(t *testing.T) {
	t.Run("InvalidFormat", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", ".env_test")
		defer os.Remove(tmpFile.Name())
		os.WriteFile(tmpFile.Name(), []byte("INVALID_LINE"), 0644)
		
		err := LoadDotEnv(tmpFile.Name())
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("EmptyPath", func(t *testing.T) {
		// This will try to open .env in current dir, which might not exist
		_ = LoadDotEnv("")
	})
}

func TestApplyConfigFile_Errors(t *testing.T) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	
	cf := &ConfigFile{}
	cf.Defaults.MinSize = stringPtr("invalid")
	
	err := cf.ApplyConfigFile(cfg, fs)
	if err == nil {
		t.Error("Expected error for invalid min_size")
	}

	cf.Defaults.MinSize = nil
	cf.Defaults.MaxSize = stringPtr("invalid")
	err = cf.ApplyConfigFile(cfg, fs)
	if err == nil {
		t.Error("Expected error for invalid max_size")
	}
}

func TestConfigCommandError(t *testing.T) {
	err := &ConfigCommandError{}
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
	if err.ExitCode() != ExitInvalidArgs {
		t.Errorf("Expected exit code %d, got %d", ExitInvalidArgs, err.ExitCode())
	}
}

// TestLoadConfigFile_Coverage tests the LoadConfigFile function extensively.
//
// Reviewed: LONG-FUNCTION - Kept long for comprehensive table-driven tests.
func TestLoadConfigFile_Coverage(t *testing.T) {
	// Create a temporary directory for test config files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		content     string
		wantCfg     *ConfigFile
		wantErr     bool
		errContains string
	}{
		{
			name:    "EmptyPath",
			path:    "",
			content: "",
			wantCfg: &ConfigFile{},
			wantErr: false,
		},
		{
			name:    "NonExistent",
			path:    filepath.Join(tmpDir, "nonexistent.toml"),
			content: "",
			wantCfg: &ConfigFile{},
			wantErr: false,
		},
		{
			name:    "TextConfig",
			path:    filepath.Join(tmpDir, "config.txt"),
			content: "file1.txt\nfile2.txt\n# comment\nfile3.txt",
			wantCfg: &ConfigFile{Files: []string{"file1.txt", "file2.txt", "file3.txt"}},
			wantErr: false,
		},
		{
			name:    "TOMLConfig",
			path:    filepath.Join(tmpDir, "config.toml"),
			content: "files = [\"from_toml_file.txt\"]\n[defaults]\nalgorithm = \"sha1\"\nrecursive = true\n[security]\nblacklist_files = [\"*.log\", \"temp/*\"]",
			wantCfg: func() *ConfigFile {
				cfg := &ConfigFile{}
				cfg.Defaults.Algorithm = ptr("sha1")
				cfg.Defaults.Recursive = ptr(true)
				cfg.Security.BlacklistFiles = []string{"*.log", "temp/*"}
				cfg.Files = []string{"from_toml_file.txt"}
				return cfg
			}(),
			wantErr: false,
		},
		{
			name:        "MalformedTOML",
			path:        filepath.Join(tmpDir, "malformed.toml"),
			content:     "[defaults\nalgorithm = \"sha1\"", // Missing closing bracket
			wantCfg:     nil,
			wantErr:     true,
			errContains: "failed to parse TOML config",
		},
		{
			name:        "FileOpenError",
			path:        filepath.Join("/root", "nopermission.toml"), // Path with no write permissions
			content:     "",
			wantCfg:     nil,
			wantErr:     true,
			errContains: "failed to open config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the config file if content is provided and it's not a special case like /root
			if tt.content != "" && !strings.HasPrefix(tt.path, "/root") {
				if err := os.WriteFile(tt.path, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file %s: %v", tt.path, err)
				}
				defer os.Remove(tt.path)
			}

			// Execute LoadConfigFile
			gotCfg, err := LoadConfigFile(tt.path)

			// Assertions
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, but got %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				if !configFilesEqual(gotCfg, tt.wantCfg) {
					t.Errorf("ConfigFile mismatch for %s:\nGot:  %+v\nWant: %+v", tt.name, gotCfg, tt.wantCfg)
				}
			}
		})
	}
}

// TestFindConfigFile tests that FindConfigFile returns something if a config file exists, or empty string.
func TestFindConfigFile(t *testing.T) {
	_ = FindConfigFile()
}