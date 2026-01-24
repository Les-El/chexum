package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/quick"

	"github.com/spf13/pflag"
)

// Helper for contains check
func containsSlice(slice []string, item string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, item) {
			return true
		}
	}
	return false
}

// TestParseArgs_Bool tests the --bool flag.
func TestParseArgs_Bool(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short bool", []string{"-b"}, true},
		{"long bool", []string{"--bool"}, true},
		{"no bool", []string{"file.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.Bool != tt.want {
				t.Errorf("Bool = %v, want %v", cfg.Bool, tt.want)
			}

			// Bool now overrides and implies Quiet behavior
			// Bool should automatically set Quiet=true
			if tt.want && !cfg.Quiet {
				t.Errorf("Bool=true should automatically set Quiet=true, got Quiet=%v", cfg.Quiet)
			}
		})
	}
}

// TestParseArgs_BoolWithMatchFlags tests bool combined with match requirement flags.
func TestParseArgs_BoolWithMatchFlags(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		wantMatchRequired bool
	}{
		{
			name:              "bool alone (no match flags)",
			args:              []string{"-b"},
			wantMatchRequired: false,
		},
		{
			name:              "bool with match-required",
			args:              []string{"-b", "--match-required"},
			wantMatchRequired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.MatchRequired != tt.wantMatchRequired {
				t.Errorf("MatchRequired = %v, want %v", cfg.MatchRequired, tt.wantMatchRequired)
			}
		})
	}
}

// TestParseArgs_Help tests that help flags are recognized.
func TestParseArgs_Help(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short help", []string{"-h"}, true},
		{"long help", []string{"--help"}, true},
		{"no help", []string{"file.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.ShowHelp != tt.want {
				t.Errorf("ShowHelp = %v, want %v", cfg.ShowHelp, tt.want)
			}
		})
	}
}

// TestParseArgs_Verbose tests verbose flag parsing.
func TestParseArgs_Verbose(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"short verbose", []string{"-v"}, true},
		{"long verbose", []string{"--verbose"}, true},
		{"no verbose", []string{"file.txt"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.Verbose != tt.want {
				t.Errorf("Verbose = %v, want %v", cfg.Verbose, tt.want)
			}
		})
	}
}

// TestParseArgs_OutputFormat tests output format flag parsing.
func TestParseArgs_OutputFormat(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"default format", []string{}, "default"},
		{"json shorthand", []string{"--json"}, "json"},
		{"plain shorthand", []string{"--plain"}, "plain"},
		{"format flag json", []string{"--format=json"}, "json"},
		{"format flag plain", []string{"--format=plain"}, "plain"},
		{"format flag verbose", []string{"--format=verbose"}, "verbose"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.OutputFormat != tt.want {
				t.Errorf("OutputFormat = %v, want %v", cfg.OutputFormat, tt.want)
			}
		})
	}
}

// TestParseArgs_Files tests that positional arguments are collected as files.
func TestParseArgs_Files(t *testing.T) {
	args := []string{"file1.txt", "file2.txt", "file3.txt"}
	cfg, _, err := ParseArgs(args)
	if err != nil {
		t.Fatalf("ParseArgs() error = %v", err)
	}

	if len(cfg.Files) != 3 {
		t.Errorf("Files count = %d, want 3", len(cfg.Files))
	}

	for i, want := range args {
		if cfg.Files[i] != want {
			t.Errorf("Files[%d] = %v, want %v", i, cfg.Files[i], want)
		}
	}
}

// TestParseArgs_Algorithm tests algorithm flag parsing.
func TestParseArgs_Algorithm(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"default algorithm", []string{}, "sha256"},
		{"short md5", []string{"-a", "md5"}, "md5"},
		{"long sha1", []string{"--algorithm=sha1"}, "sha1"},
		{"long sha512", []string{"--algorithm", "sha512"}, "sha512"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs() error = %v", err)
			}
			if cfg.Algorithm != tt.want {
				t.Errorf("Algorithm = %v, want %v", cfg.Algorithm, tt.want)
			}
		})
	}
}

// TestParseArgs_FlagOrderIndependence is a property-based test that verifies
// flags can be provided in any order and produce the same result.
// Property 6: Flags accept any order
// Validates: Requirements 4.4
func TestParseArgs_FlagOrderIndependence(t *testing.T) {
	// Test that specific flag combinations work in different orders
	// Note: Avoid mutually exclusive combinations like --json with --verbose
	testCases := [][]string{
		{"-v", "file.txt"},
		{"--verbose", "file.txt"},
		{"file.txt", "-v"},
		{"file.txt", "--verbose"},
	}

	var expected *Config
	for i, args := range testCases {
		cfg, _, err := ParseArgs(args)
		if err != nil {
			t.Fatalf("ParseArgs(%v) error = %v", args, err)
		}

		if i == 0 {
			expected = cfg
		} else {
			// Compare relevant fields
			if cfg.Verbose != expected.Verbose {
				t.Errorf("Order %d: Verbose = %v, want %v", i, cfg.Verbose, expected.Verbose)
			}
			if cfg.OutputFormat != expected.OutputFormat {
				t.Errorf("Order %d: OutputFormat = %v, want %v", i, cfg.OutputFormat, expected.OutputFormat)
			}
			if len(cfg.Files) != len(expected.Files) {
				t.Errorf("Order %d: Files count = %d, want %d", i, len(cfg.Files), len(expected.Files))
			}
		}
	}
}

// TestParseArgs_FlagOrderIndependence_Property is a property-based test using testing/quick.
// It generates random permutations of flags and verifies they produce equivalent configs.
// Property 6: Flags accept any order
// Validates: Requirements 4.4
func TestParseArgs_FlagOrderIndependence_Property(t *testing.T) {
	// Define test flag sets that should produce the same result regardless of order
	// Use --flag=value syntax to keep flag-value pairs together during permutation
	// Note: We avoid invalid combinations like --quiet with --verbose or --json with --verbose
	flagSets := [][]string{
		{"-v", "-r"},
		{"--plain", "--hidden"},
		{"--json", "--recursive"},
		{"--algorithm=md5", "--preserve-order"},
		{"-r", "--hidden", "--preserve-order"},
	}

	for _, flags := range flagSets {
		// Generate all permutations and verify they produce equivalent configs
		permutations := generatePermutations(flags)

		var baseConfig *Config
		for i, perm := range permutations {
			cfg, _, err := ParseArgs(perm)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", perm, err)
			}

			if i == 0 {
				baseConfig = cfg
			} else {
				if !configsEquivalent(baseConfig, cfg) {
					t.Errorf("Flag order affected result:\n  Order 0: %v\n  Order %d: %v", flags, i, perm)
				}
			}
		}
	}
}

// generatePermutations generates all permutations of a string slice.
// Limited to small slices to avoid combinatorial explosion.
func generatePermutations(arr []string) [][]string {
	if len(arr) <= 1 {
		return [][]string{arr}
	}

	// Limit to first 24 permutations (4! = 24) to keep tests fast
	var result [][]string
	permute(arr, 0, &result, 24)
	return result
}

func permute(arr []string, start int, result *[][]string, limit int) {
	if len(*result) >= limit {
		return
	}
	if start == len(arr) {
		perm := make([]string, len(arr))
		copy(perm, arr)
		*result = append(*result, perm)
		return
	}
	for i := start; i < len(arr); i++ {
		arr[start], arr[i] = arr[i], arr[start]
		permute(arr, start+1, result, limit)
		arr[start], arr[i] = arr[i], arr[start]
	}
}

// configsEquivalent checks if two configs have equivalent flag values.
// It ignores the Files field since file order may differ.
func configsEquivalent(a, b *Config) bool {
	return a.Recursive == b.Recursive &&
		a.Hidden == b.Hidden &&
		a.Algorithm == b.Algorithm &&
		a.Verbose == b.Verbose &&
		a.Quiet == b.Quiet &&
		a.PreserveOrder == b.PreserveOrder &&
		a.MatchRequired == b.MatchRequired &&
		a.OutputFormat == b.OutputFormat &&
		a.OutputFile == b.OutputFile &&
		a.Append == b.Append &&
		a.Force == b.Force
}

// TestDefaultConfig tests that DefaultConfig returns sensible defaults.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Algorithm != "sha256" {
		t.Errorf("Algorithm = %v, want sha256", cfg.Algorithm)
	}
	if cfg.OutputFormat != "default" {
		t.Errorf("OutputFormat = %v, want default", cfg.OutputFormat)
	}
	if cfg.MaxSize != -1 {
		t.Errorf("MaxSize = %v, want -1", cfg.MaxSize)
	}
}

// TestHelpText tests that help text is non-empty and contains key sections.
func TestHelpText(t *testing.T) {
	help := HelpText()

	if len(help) == 0 {
		t.Error("HelpText() returned empty string")
	}

	// Check for key sections
	sections := []string{"EXAMPLES", "USAGE", "FLAGS", "EXIT CODES"}
	for _, section := range sections {
		if !contains(help, section) {
			t.Errorf("HelpText() missing section: %s", section)
		}
	}
}

// TestVersionText tests that version text is non-empty.
func TestVersionText(t *testing.T) {
	version := VersionText()

	if len(version) == 0 {
		t.Error("VersionText() returned empty string")
	}

	if !contains(version, "hashi") {
		t.Error("VersionText() should contain 'hashi'")
	}
}

// Property-based test: parsing should not panic on random input
func TestParseArgs_NoPanic(t *testing.T) {
	f := func(args []string) bool {
		// Filter out nil strings
		filtered := make([]string, 0, len(args))
		for _, arg := range args {
			if arg != "" {
				filtered = append(filtered, arg)
			}
		}

		// This should not panic
		_, _, _ = ParseArgs(filtered)
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestParseArgs_AbbreviationRejection verifies that arbitrary abbreviations are rejected.
// Property 15: Abbreviated flags are rejected
func TestParseArgs_AbbreviationRejection(t *testing.T) {
	tests := []struct {
		arg     string
		wantErr bool
	}{
		{"--verb", true},  // Abbreviation of --verbose
		{"--help", false}, // Exact match
		{"-v", false},     // Exact short flag
		{"--vers", true},  // Abbreviation of --version
	}

	for _, tt := range tests {
		_, _, err := ParseArgs([]string{tt.arg})
		if tt.wantErr && err == nil {
			t.Errorf("ParseArgs(%q) expected error for abbreviation, got nil", tt.arg)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ParseArgs(%q) unexpected error: %v", tt.arg, err)
		}
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseArgs_ShortAndLongFlags tests that both short and long flag variants work.
func TestParseArgs_ShortAndLongFlags(t *testing.T) {
	t.Run("Boolean Flags", testBooleanFlags)
	t.Run("Value Flags", testValueFlags)
	t.Run("Slice Flags", testSliceFlags)
}

func testBooleanFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(*Config) bool
	}{
		{"recursive_short", []string{"-r"}, func(c *Config) bool { return c.Recursive }},
		{"recursive_long", []string{"--recursive"}, func(c *Config) bool { return c.Recursive }},
		{"verbose_short", []string{"-v"}, func(c *Config) bool { return c.Verbose }},
		{"quiet_short", []string{"-q"}, func(c *Config) bool { return c.Quiet }},
	}
	runParseSubtests(t, tests)
}

func testValueFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(*Config) bool
	}{
		{"algorithm", []string{"-a", "md5"}, func(c *Config) bool { return c.Algorithm == "md5" }},
		{"output", []string{"-o", "out.txt"}, func(c *Config) bool { return c.OutputFile == "out.txt" }},
	}
	runParseSubtests(t, tests)
}

func testSliceFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(*Config) bool
	}{
		{"include", []string{"-i", "*.txt"}, func(c *Config) bool { return len(c.Include) == 1 }},
		{"exclude", []string{"-e", "*.log"}, func(c *Config) bool { return len(c.Exclude) == 1 }},
	}
	runParseSubtests(t, tests)
}

func runParseSubtests(t *testing.T, tests []struct {
	name  string
	args  []string
	check func(*Config) bool
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if !tt.check(cfg) {
				t.Errorf("Flag %v did not set expected value", tt.args)
			}
		})
	}
}

// TestParseArgs_StdinSupport tests that "-" is recognized as stdin marker.
// Validates: Requirements 4.3
func TestParseArgs_StdinSupport(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		hasStdin bool
	}{
		{"stdin only", []string{"-"}, true},
		{"stdin with files", []string{"file1.txt", "-", "file2.txt"}, true},
		{"stdin with flags", []string{"-v", "-"}, true},
		{"no stdin", []string{"file1.txt", "file2.txt"}, false},
		{"empty args", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.HasStdinMarker() != tt.hasStdin {
				t.Errorf("HasStdinMarker() = %v, want %v", cfg.HasStdinMarker(), tt.hasStdin)
			}
		})
	}
}

// TestParseArgs_FilesWithoutStdin tests the FilesWithoutStdin method.
//
// Reviewed: LONG-FUNCTION - Kept long for comprehensive tests.
func TestFilesWithoutStdin(t *testing.T) {
	cfg := &Config{
		Files: []string{"file1.txt", "-", "file2.txt"},
	}
	got := cfg.FilesWithoutStdin()
	want := []string{"file1.txt", "file2.txt"}
	if !stringSlicesEqual(got, want) {
		t.Errorf("FilesWithoutStdin() = %v, want %v", got, want)
	}
}

// TestHasStdinMarker tests the HasStdinMarker method.
func TestHasStdinMarker(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  bool
	}{
		{"with stdin", []string{"file.txt", "-"}, true},
		{"without stdin", []string{"file.txt", "other.txt"}, false},
		{"empty", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Files: tt.files}
			if cfg.HasStdinMarker() != tt.want {
				t.Errorf("HasStdinMarker() = %v, want %v", cfg.HasStdinMarker(), tt.want)
			}
		})
	}
}

// Reviewed: LONG-FUNCTION - Kept long for comprehensive table-driven tests.
func TestClassifyArguments(t *testing.T) {
	// Create dummy files for testing file existence
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "existing_file1.txt")
	file2 := filepath.Join(tempDir, "existing_file2.txt")
	os.WriteFile(file1, []byte("content"), 0644)
	os.WriteFile(file2, []byte("content"), 0644)

	tests := []struct {
		name        string
		args        []string
		algorithm   string
		wantFiles   []string
		wantHashes  []string
		wantErr     bool
		errContains string
	}{
		{
			name:       "OnlyFiles",
			args:       []string{file1, file2},
			algorithm:  "sha256",
			wantFiles:  []string{file1, file2},
			wantHashes: nil,
			wantErr:    false,
		},
		{
			name:       "OnlyHashesSHA256",
			args:       []string{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "d290c05a0ce8e84a22ae80a22de4c0c1b09b0b4b8a4d4e0e4b8a4d4e0e4b8a4d"},
			algorithm:  "sha256",
			wantFiles:  nil,
			wantHashes: []string{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "d290c05a0ce8e84a22ae80a22de4c0c1b09b0b4b8a4d4e0e4b8a4d4e0e4b8a4d"},
			wantErr:    false,
		},
		{
			name:       "OnlyHashesMD5",
			args:       []string{"d41d8cd98f00b204e9800998ecf8427e"},
			algorithm:  "md5",
			wantFiles:  nil,
			wantHashes: []string{"d41d8cd98f00b204e9800998ecf8427e"},
			wantErr:    false,
		},
		{
			name:       "MixedFilesAndHashes",
			args:       []string{file1, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", file2},
			algorithm:  "sha256",
			wantFiles:  []string{file1, file2},
			wantHashes: []string{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
			wantErr:    false,
		},
		{
			name:       "StdinMarker",
			args:       []string{file1, "-", file2},
			algorithm:  "sha256",
			wantFiles:  []string{file1, "-", file2},
			wantHashes: nil,
			wantErr:    false,
		},
		{
			name:       "NonExistentButNotHash",
			args:       []string{"non_existent_file.txt"},
			algorithm:  "sha256",
			wantFiles:  []string{"non_existent_file.txt"},
			wantHashes: nil,
			wantErr:    false,
		},
		{
			name:        "LooksLikeHashWrongLength",
			args:        []string{"abcde12345"}, // too short for any known hash
			algorithm:   "sha256",
			wantFiles:   nil,
			wantHashes:  nil,
			wantErr:     true,
			errContains: "unknown length",
		},
		{
			name:        "LooksLikeHashWrongAlgorithm",
			args:        []string{"d41d8cd98f00b204e9800998ecf8427e"}, // valid MD5
			algorithm:   "sha256",
			wantFiles:   nil,
			wantHashes:  nil,
			wantErr:     true,
			errContains: "hash length doesn't match sha256 (expected 64 characters, got 32)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, gotHashes, err := ClassifyArguments(tt.args, tt.algorithm)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ClassifyArguments() expected error, got nil")
				}
				if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ClassifyArguments() expected error to contain %q, got %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("ClassifyArguments() unexpected error = %v", err)
				}
				if !stringSlicesEqual(gotFiles, tt.wantFiles) {
					t.Errorf("ClassifyArguments() gotFiles = %v, want %v", gotFiles, tt.wantFiles)
				}
				if !stringSlicesEqual(gotHashes, tt.wantHashes) {
					t.Errorf("ClassifyArguments() gotHashes = %v, want %v", gotHashes, tt.wantHashes)
				}
			}
		})
	}
}

// TestApplyEnvConfig tests the ApplyEnvConfig function.
//
// Reviewed: LONG-FUNCTION - Kept long for comprehensive table-driven tests.
func TestApplyEnvConfig(t *testing.T) {
	t.Run("ApplyDefaults", func(t *testing.T) {
		// Ensure environment variables are applied when flags are not set
		os.Setenv("HASHI_ALGORITHM", "md5")
		os.Setenv("HASHI_RECURSIVE", "true")
		os.Setenv("HASHI_BLACKLIST_FILES", "file1.log,file2.tmp")
		defer os.Clearenv()

		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		// Define flags so their 'Changed' status can be checked
		fs.String("algorithm", "", "")
		fs.Bool("recursive", false, "")
		fs.StringSlice("blacklist-files", []string{}, "")

		envCfg := LoadEnvConfig()
		envCfg.ApplyEnvConfig(cfg, fs)

		if cfg.Algorithm != "md5" {
			t.Errorf("Expected algorithm to be md5, got %s", cfg.Algorithm)
		}
		if !cfg.Recursive {
			t.Errorf("Expected recursive to be true, got %v", cfg.Recursive)
		}
		if !stringSlicesEqual(cfg.BlacklistFiles, []string{"file1.log", "file2.tmp"}) {
			t.Errorf("Expected blacklist files [file1.log, file2.tmp], got %v", cfg.BlacklistFiles)
		}
	})

	t.Run("FlagsOverrideEnv", func(t *testing.T) {
		// Ensure flags override environment variables
		os.Setenv("HASHI_ALGORITHM", "md5")
		defer os.Clearenv()

		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("algorithm", "sha1", "")
		fs.Set("algorithm", "sha1") // Mark flag as changed
		cfg.Algorithm = "sha1"      // Manually set config value to simulate flag parsing

		envCfg := LoadEnvConfig()
		envCfg.ApplyEnvConfig(cfg, fs)

		if cfg.Algorithm != "sha1" {
			t.Errorf("Expected algorithm to be sha1 (from flag), got %s", cfg.Algorithm)
		}
	})

	t.Run("NoEnvNoFlagChange", func(t *testing.T) {
		// No environment variable set, no flag changed, should use default
		os.Clearenv()

		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("algorithm", "", "")

		envCfg := LoadEnvConfig()
		envCfg.ApplyEnvConfig(cfg, fs)

		if cfg.Algorithm != "sha256" {
			t.Errorf("Expected algorithm to be default sha256, got %s", cfg.Algorithm)
		}
	})

	t.Run("WhitelistEnv", func(t *testing.T) {
		os.Setenv("HASHI_WHITELIST_FILES", "white1.txt, white2.txt")
		defer os.Clearenv()

		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringSlice("whitelist-files", []string{}, "")

		envCfg := LoadEnvConfig()
		envCfg.ApplyEnvConfig(cfg, fs)

		if !stringSlicesEqual(cfg.WhitelistFiles, []string{"white1.txt", "white2.txt"}) {
			t.Errorf("Expected whitelist files [white1.txt, white2.txt], got %v", cfg.WhitelistFiles)
		}
	})
}

// TestWriteErrorWithVerbose tests the WriteErrorWithVerbose function.
func TestWriteErrorWithVerbose(t *testing.T) {
	tests := []struct {
		name           string
		verbose        bool
		verboseDetails string
		expectedError  string
	}{
		{
			name:           "VerboseTrue",
			verbose:        true,
			verboseDetails: "detailed error message",
			expectedError:  "detailed error message",
		},
		{
			name:           "VerboseFalse",
			verbose:        false,
			verboseDetails: "detailed error message", // Should be ignored when verbose is false
			expectedError:  "Unknown write/append error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteErrorWithVerbose(tt.verbose, tt.verboseDetails)
			if err.Error() != tt.expectedError {
				t.Errorf("WriteErrorWithVerbose(%v, %q) got error %q, want %q", tt.verbose, tt.verboseDetails, err.Error(), tt.expectedError)
			}
		})
	}
}

// TestApplyConfigFile tests the ApplyConfigFile function.
//
// Reviewed: LONG-FUNCTION - Kept long for comprehensive table-driven tests.
func TestApplyConfigFile(t *testing.T) {
	t.Run("ApplyBoolDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Bool("recursive", false, "")
		cf := &ConfigFile{}
		cf.Defaults.Recursive = ptr(true)

		_ = cf.ApplyConfigFile(cfg, fs)
		if !cfg.Recursive {
			t.Errorf("Expected Recursive to be true, got %v", cfg.Recursive)
		}

		// Flag should override
		cfg = DefaultConfig()
		fs = pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.Bool("recursive", false, "")
		fs.Set("recursive", "false")
		cf = &ConfigFile{}
		cf.Defaults.Recursive = ptr(true)

		_ = cf.ApplyConfigFile(cfg, fs)
		if cfg.Recursive {
			t.Errorf("Expected Recursive to be false due to flag, got %v", cfg.Recursive)
		}
	})

	t.Run("ApplyStringDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("algorithm", "", "")
		cf := &ConfigFile{}
		cf.Defaults.Algorithm = ptr("md5")

		_ = cf.ApplyConfigFile(cfg, fs)
		if cfg.Algorithm != "md5" {
			t.Errorf("Expected Algorithm to be md5, got %s", cfg.Algorithm)
		}
	})

	t.Run("ApplySizeDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.String("min-size", "", "")
		cf := &ConfigFile{}
		cf.Defaults.MinSize = ptr("100KB")

		_ = cf.ApplyConfigFile(cfg, fs)
		if cfg.MinSize != 102400 {
			t.Errorf("Expected MinSize to be 102400, got %d", cfg.MinSize)
		}

		// Error case
		cfg = DefaultConfig()
		cf.Defaults.MinSize = ptr("invalid-size")
		err := cf.ApplyConfigFile(cfg, fs)
		if err == nil || !strings.Contains(err.Error(), "invalid min_size") {
			t.Errorf("Expected invalid min_size error, got %v", err)
		}
	})

	t.Run("ApplyListDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		fs.StringSlice("include", []string{}, "")
		cf := &ConfigFile{}
		cf.Defaults.Include = []string{"*.go", "*.mod"}

		_ = cf.ApplyConfigFile(cfg, fs)
		if !stringSlicesEqual(cfg.Include, []string{"*.go", "*.mod"}) {
			t.Errorf("Expected Include to be [*.go, *.mod], got %v", cfg.Include)
		}
	})

	t.Run("ApplySecurityDefaults", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		cf := &ConfigFile{}
		cf.Security.BlacklistFiles = []string{"secret.txt"}

		_ = cf.ApplyConfigFile(cfg, fs)
		if !stringSlicesEqual(cfg.BlacklistFiles, []string{"secret.txt"}) {
			t.Errorf("Expected BlacklistFiles to be [secret.txt], got %v", cfg.BlacklistFiles)
		}
	})

	t.Run("ApplyFiles", func(t *testing.T) {
		cfg := DefaultConfig()
		fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
		cf := &ConfigFile{}
		cf.Files = []string{"file_from_config.txt"}

		_ = cf.ApplyConfigFile(cfg, fs)
		if !stringSlicesEqual(cfg.Files, []string{"file_from_config.txt"}) {
			t.Errorf("Expected Files to be [file_from_config.txt], got %v", cfg.Files)
		}

		// Should not override existing files from args/flags
		cfg = DefaultConfig()
		cfg.Files = []string{"file_from_args.txt"}
		cf = &ConfigFile{}
		cf.Files = []string{"file_from_config.txt"}

		_ = cf.ApplyConfigFile(cfg, fs)
		// The existing files from args/flags should be preserved, not overwritten by config file files.
		if !stringSlicesEqual(cfg.Files, []string{"file_from_args.txt"}) {
			t.Errorf("Expected Files to remain [file_from_args.txt], got %v", cfg.Files)
		}
	})
}

// TestParseArgs_FlagValidation tests that invalid flag values are rejected.
func TestParseArgs_FlagValidation(t *testing.T) {
	t.Run("Invalid Values", testInvalidFlagValues)
	t.Run("Override Warnings", testFlagOverrideWarnings)
	t.Run("Valid Flags", testValidFlags)
}

func testInvalidFlagValues(t *testing.T) {
	tests := []struct {
		args []string
		msg  string
	}{
		{[]string{"--format=invalid"}, "invalid output format"},
		{[]string{"--algorithm=invalid"}, "invalid algorithm"},
		{[]string{"--min-size=abc"}, "invalid"},
		{[]string{"--modified-after=not-a-date"}, "invalid"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			_, _, err := ParseArgs(tt.args)
			if err == nil || !contains(err.Error(), tt.msg) {
				t.Errorf("Expected error containing %q, got %v", tt.msg, err)
			}
		})
	}
}

func testFlagOverrideWarnings(t *testing.T) {
	cfg, warnings, _ := ParseArgs([]string{"--quiet", "--verbose"})
	if len(warnings) == 0 || !contains(warnings[0].Message, "--quiet overrides --verbose") {
		t.Errorf("Expected quiet override warning, got %v", warnings)
	}
	if !cfg.Quiet || cfg.Verbose {
		t.Error("Override logic failed")
	}
}

func testValidFlags(t *testing.T) {
	if _, _, err := ParseArgs([]string{"-v", "-r"}); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestParseArgs_ErrorCases tests various error scenarios.
func TestParseArgs_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"unknown flag", []string{"--unknown-flag"}, true},
		{"missing flag value", []string{"--algorithm"}, true},
		{"invalid size with unit", []string{"--min-size=10XB"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseArgs(tt.args)
			if tt.wantErr && err == nil {
				t.Errorf("ParseArgs(%v) expected error, got nil", tt.args)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ParseArgs(%v) unexpected error = %v", tt.args, err)
			}
		})
	}
}

// TestParseArgs_SizeUnits tests human-readable size parsing.
func TestParseArgs_SizeUnits(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantSize int64
	}{
		{"bytes", []string{"--min-size=100"}, 100},
		{"kilobytes", []string{"--min-size=1KB"}, 1024},
		{"megabytes", []string{"--min-size=1MB"}, 1024 * 1024},
		{"gigabytes", []string{"--min-size=1GB"}, 1024 * 1024 * 1024},
		{"short K", []string{"--min-size=10K"}, 10 * 1024},
		{"short M", []string{"--min-size=10M"}, 10 * 1024 * 1024},
		{"decimal", []string{"--min-size=1.5MB"}, int64(1.5 * 1024 * 1024)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.MinSize != tt.wantSize {
				t.Errorf("MinSize = %d, want %d", cfg.MinSize, tt.wantSize)
			}
		})
	}
}

// TestParseArgs_DateParsing tests date flag parsing.
func TestParseArgs_DateParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantYear int
	}{
		{"simple date", []string{"--modified-after=2024-01-15"}, 2024},
		{"with time", []string{"--modified-after=2023-06-01T12:00:00"}, 2023},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.ModifiedAfter.Year() != tt.wantYear {
				t.Errorf("ModifiedAfter.Year() = %d, want %d", cfg.ModifiedAfter.Year(), tt.wantYear)
			}
		})
	}
}

// TestParseArgs_FlagValueSyntax tests both --flag=value and --flag value syntax.
// Validates: Requirements 4.4
func TestParseArgs_FlagValueSyntax(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"equals syntax", []string{"--algorithm=md5"}, "md5"},
		{"space syntax", []string{"--algorithm", "md5"}, "md5"},
		{"short with space", []string{"-a", "sha1"}, "sha1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _, err := ParseArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseArgs(%v) error = %v", tt.args, err)
			}
			if cfg.Algorithm != tt.want {
				t.Errorf("Algorithm = %v, want %v", cfg.Algorithm, tt.want)
			}
		})
	}
}

// TestValidateOutputFormat tests output format validation.
func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		format  string
		wantErr bool
	}{
		{"default", false},
		{"verbose", false},
		{"json", false},
		{"plain", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			err := ValidateOutputFormat(tt.format)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateOutputFormat(%q) expected error", tt.format)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateOutputFormat(%q) unexpected error = %v", tt.format, err)
			}
		})
	}
}

// TestValidateAlgorithm tests algorithm validation.
func TestValidateAlgorithm(t *testing.T) {
	tests := []struct {
		algorithm string
		wantErr   bool
	}{
		{"sha256", false},
		{"md5", false},
		{"sha1", false},
		{"sha512", false},
		{"invalid", true},
		{"SHA256", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			err := ValidateAlgorithm(tt.algorithm)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateAlgorithm(%q) expected error", tt.algorithm)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateAlgorithm(%q) unexpected error = %v", tt.algorithm, err)
			}
		})
	}
}

// TestParseArgs_BoolOverridesBehavior tests that --bool overrides other output flags.
func TestParseArgs_BoolOverridesBehavior(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		cfg, _, _ := ParseArgs([]string{"--bool"})
		if !cfg.Bool || !cfg.Quiet {
			t.Error("Bool should imply Quiet")
		}
	})
	t.Run("Overrides Verbose", func(t *testing.T) {
		cfg, _, _ := ParseArgs([]string{"--bool", "--verbose"})
		if !cfg.Bool || !cfg.Quiet || cfg.Verbose {
			t.Error("Bool should override Verbose")
		}
	})
	t.Run("Overrides Format", func(t *testing.T) {
		cfg, _, _ := ParseArgs([]string{"--bool", "--json"})
		if !cfg.Bool || cfg.OutputFormat != "default" {
			t.Errorf("Bool should reset format, got %s", cfg.OutputFormat)
		}
	})
}

// TestError tests the ConfigCommandError error type.
func TestError(t *testing.T) {
	e := &ConfigCommandError{}
	if e.Error() == "" {
		t.Error("ConfigCommandError.Error() returned empty string")
	}
}

// TestExitCode tests the ExitCode method of ConfigCommandError.
func TestExitCode(t *testing.T) {
	e := &ConfigCommandError{}
	if e.ExitCode() != ExitInvalidArgs {
		t.Errorf("expected ExitInvalidArgs, got %d", e.ExitCode())
	}
}

// TestWriteError tests the WriteError function.
func TestWriteError(t *testing.T) {
	if WriteError() == nil {
		t.Error("WriteError() returned nil")
	}
}

// TestLoadEnvConfig tests that LoadEnvConfig loads environment variables.
func TestLoadEnvConfig(t *testing.T) {
	env := LoadEnvConfig()
	if env == nil {
		t.Error("LoadEnvConfig() returned nil")
	}
}

// TestLoadDotEnv tests the LoadDotEnv function.
//
// Reviewed: LONG-FUNCTION - Kept long for comprehensive table-driven tests.
func TestLoadDotEnv(t *testing.T) {
	// Create a temporary .env file for testing
	tempDir := t.TempDir()
	dotEnvPath := filepath.Join(tempDir, ".env")

	// Test case 1: Valid .env file
	t.Run("ValidDotEnvFile", func(t *testing.T) {
		content := []byte("KEY1=value1\nKEY2=\"value 2\"\n#comment\nKEY3='value3'")
		err := os.WriteFile(dotEnvPath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}

		os.Clearenv() // Clear all env vars to ensure a clean state
		err = LoadDotEnv(dotEnvPath)
		if err != nil {
			t.Errorf("LoadDotEnv() error = %v, wantErr %v", err, nil)
		}

		if os.Getenv("KEY1") != "value1" {
			t.Errorf("KEY1 = %s, want value1", os.Getenv("KEY1"))
		}
		if os.Getenv("KEY2") != "value 2" {
			t.Errorf("KEY2 = %s, want value 2", os.Getenv("KEY2"))
		}
		if os.Getenv("KEY3") != "value3" {
			t.Errorf("KEY3 = %s, want value3", os.Getenv("KEY3"))
		}
		os.Remove(dotEnvPath)
	})

	// Test case 2: Non-existent .env file
	t.Run("NonExistentDotEnvFile", func(t *testing.T) {
		os.Clearenv()
		err := LoadDotEnv("non_existent.env")
		if err != nil {
			t.Errorf("LoadDotEnv() error for non-existent file = %v, wantErr %v", err, nil)
		}
	})

	// Test case 3: Invalid format in .env file
	t.Run("InvalidFormatDotEnvFile", func(t *testing.T) {
		content := []byte("KEY1=value1\nINVALID_LINE\nKEY2=value2")
		err := os.WriteFile(dotEnvPath, content, 0644)
		if err != nil {
			t.Fatalf("Failed to create .env file: %v", err)
		}

		os.Clearenv()
		err = LoadDotEnv(dotEnvPath)
		if err == nil || !strings.Contains(err.Error(), "invalid format") {
			t.Errorf("LoadDotEnv() expected error for invalid format, got %v", err)
		}
		os.Remove(dotEnvPath)
	})

	// Test case 4: Existing environment variables should not be overwritten

	t.Run("ExistingEnvVarNotOverwritten", func(t *testing.T) {

		content := []byte("EXISTING_KEY=new_value")

		err := os.WriteFile(dotEnvPath, content, 0644)

		if err != nil {

			t.Fatalf("Failed to create .env file: %v", err)

		}

		os.Setenv("EXISTING_KEY", "original_value")

		defer os.Unsetenv("EXISTING_KEY")

		err = LoadDotEnv(dotEnvPath)

		if err != nil {

			t.Errorf("LoadDotEnv() error = %v, wantErr %v", err, nil)

		}

		if os.Getenv("EXISTING_KEY") != "original_value" {

			t.Errorf("Existing env var overwritten: EXISTING_KEY = %s, want original_value", os.Getenv("EXISTING_KEY"))

		}

		os.Remove(dotEnvPath)

	})

}

// TestParseArgs tests the ParseArgs function.

func TestParseArgs(t *testing.T) {

	args := []string{"--algorithm", "md5", "file.txt"}

	// Create test file

	os.WriteFile("file.txt", []byte("hello"), 0644)

	defer os.Remove("file.txt")

	cfg, warnings, err := ParseArgs(args)

	if err != nil {

		t.Errorf("ParseArgs() error = %v", err)

	}

	if cfg.Algorithm != "md5" {

		t.Errorf("Algorithm = %s, want md5", cfg.Algorithm)

	}

	if len(warnings) > 0 {

		t.Logf("Warnings: %v", warnings)

	}

}

// TestValidateConfig tests the ValidateConfig function.

func TestValidateConfig(t *testing.T) {

	cfg := DefaultConfig()

	cfg.Files = []string{"test.txt"}

	// Create test file

	os.WriteFile("test.txt", []byte("hello"), 0644)

	defer os.Remove("test.txt")

	warnings, err := ValidateConfig(cfg)

	if err != nil {

		t.Errorf("ValidateConfig() error = %v", err)

	}

	if len(warnings) > 0 {

		t.Logf("Warnings: %v", warnings)

	}

}

// TestLoadConfigFile tests the LoadConfigFile function.

func TestLoadConfigFile(t *testing.T) {

	path := "test_config_simple.toml"

	content := "[defaults]\nalgorithm = \"md5\""

	os.WriteFile(path, []byte(content), 0644)

	defer os.Remove(path)

	cfg, err := LoadConfigFile(path)

	if err != nil {

		t.Errorf("LoadConfigFile() error = %v", err)

	}

	if *cfg.Defaults.Algorithm != "md5" {

		t.Errorf("Algorithm = %s, want md5", *cfg.Defaults.Algorithm)

	}

}
