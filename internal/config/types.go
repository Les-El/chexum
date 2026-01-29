package config

import (
	"time"
)

// Exit codes for scripting support
const (
	ExitSuccess        = 0   // All files processed successfully
	ExitNoMatches      = 1   // No matches found (with --match-required)
	ExitPartialFailure = 2   // Some files failed to process
	ExitInvalidArgs    = 3   // Invalid arguments or flags
	ExitFileNotFound   = 4   // One or more files not found
	ExitPermissionErr  = 5   // Permission denied
	ExitInterrupted    = 130 // Interrupted by Ctrl-C (128 + SIGINT)
)

// ConfigCommandError is returned when a user tries to use a config subcommand.
type ConfigCommandError struct{}

// Error returns the formatted error message.
func (e *ConfigCommandError) Error() string {
	return `Error: hashi does not support config subcommands

Configuration must be done by manually editing config files.

Hashi auto-loads config from these standard locations:
  • .hashi.toml (project-specific)
  • hashi/config.toml (in XDG config directory)
  • .hashi/config.toml (traditional dotfile)

For configuration documentation and examples, see:
  https://github.com/[your-repo]/hashi#configuration`
}

// ExitCode returns the appropriate exit code for the error.
func (e *ConfigCommandError) ExitCode() int {
	return ExitInvalidArgs
}

// EnvConfig holds environment variable configuration.
type EnvConfig struct {
	NoColor     bool   // NO_COLOR environment variable
	Debug       bool   // DEBUG environment variable
	TmpDir      string // TMPDIR environment variable
	Home        string // HOME environment variable
	ConfigHome  string // XDG_CONFIG_HOME environment variable
	HashiConfig string // HASHI_CONFIG environment variable

	HashiAlgorithm      string // HASHI_ALGORITHM
	HashiOutputFormat   string // HASHI_OUTPUT_FORMAT
	HashiDryRun         bool   // HASHI_DRY_RUN
	HashiRecursive      bool   // HASHI_RECURSIVE
	HashiHidden         bool   // HASHI_HIDDEN
	HashiVerbose        bool   // HASHI_VERBOSE
	HashiQuiet          bool   // HASHI_QUIET
	HashiBool           bool   // HASHI_BOOL
	HashiPreserveOrder  bool   // HASHI_PRESERVE_ORDER
	HashiMatchRequired  bool   // HASHI_MATCH_REQUIRED
	HashiManifest       string // HASHI_MANIFEST
	HashiOnlyChanged    bool   // HASHI_ONLY_CHANGED
	HashiOutputManifest string // HASHI_OUTPUT_MANIFEST

	HashiOutputFile string // HASHI_OUTPUT_FILE
	HashiAppend     bool   // HASHI_APPEND
	HashiForce      bool   // HASHI_FORCE

	HashiLogFile string // HASHI_LOG_FILE
	HashiLogJSON string // HASHI_LOG_JSON

	HashiHelp    bool // HASHI_HELP
	HashiVersion bool // HASHI_VERSION

	HashiJobs int // HASHI_JOBS

	HashiBlacklistFiles string // HASHI_BLACKLIST_FILES
	HashiBlacklistDirs  string // HASHI_BLACKLIST_DIRS
	HashiWhitelistFiles string // HASHI_WHITELIST_FILES
	HashiWhitelistDirs  string // HASHI_WHITELIST_DIRS
}

// Config holds all configuration options for hashi.
type Config struct {
	Input       InputConfig
	Output      OutputConfig
	Processing  ProcessingConfig
	Incremental IncrementalConfig
	Security    SecurityConfig

	ConfigFile  string
	ShowHelp    bool
	ShowVersion bool

	// Deprecated: Moving to structured fields
	Files  []string
	Hashes []string

	Recursive     bool
	Hidden        bool
	Algorithm     string
	DryRun        bool
	Verbose       bool
	Quiet         bool
	Bool          bool
	PreserveOrder bool
	Jobs          int
	Test          bool

	MatchRequired bool

	OutputFormat string
	JSON         bool
	JSONL        bool
	Plain        bool
	OutputFile   string
	Append       bool
	Force        bool

	LogFile string
	LogJSON string

	Include        []string
	Exclude        []string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time

	Manifest       string
	OnlyChanged    bool
	OutputManifest string

	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string
}

// InputConfig holds file discovery and filtering options.
type InputConfig struct {
	Files          []string
	Hashes         []string
	Include        []string
	Exclude        []string
	MinSize        int64
	MaxSize        int64
	ModifiedAfter  time.Time
	ModifiedBefore time.Time
}

// OutputConfig holds output formatting and destination options.
type OutputConfig struct {
	Format     string
	JSON       bool
	JSONL      bool
	Plain      bool
	OutputFile string
	Append     bool
	Force      bool
	LogFile    string
	LogJSON    string
}

// ProcessingConfig holds core processing behavior options.
type ProcessingConfig struct {
	Recursive     bool
	Hidden        bool
	Algorithm     string
	DryRun        bool
	Verbose       bool
	Quiet         bool
	Bool          bool
	PreserveOrder bool
	MatchRequired bool
	Jobs          int
}

// IncrementalConfig holds options for incremental hashing.
type IncrementalConfig struct {
	Manifest       string
	OnlyChanged    bool
	OutputManifest string
}

// SecurityConfig holds security policy overrides.
type SecurityConfig struct {
	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string
}

// ValidatedConfig is a marker type for a configuration that has been validated.
type ValidatedConfig struct {
	*Config
}
