// Package config handles configuration and argument parsing for hashi.
//
// It supports multiple configuration sources with the following precedence:
// flags > environment variables > project config > user config > system config
//
// The package uses spf13/pflag for POSIX-compliant flag parsing, supporting
// both short (-v) and long (--verbose) flag formats.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"
	"github.com/Les-El/hashi/internal/conflict"
	"github.com/Les-El/hashi/internal/security"
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
	
	HashiAlgorithm     string // HASHI_ALGORITHM
	HashiOutputFormat  string // HASHI_OUTPUT_FORMAT
	HashiRecursive     bool   // HASHI_RECURSIVE
	HashiHidden        bool   // HASHI_HIDDEN
	HashiVerbose       bool   // HASHI_VERBOSE
	HashiQuiet         bool   // HASHI_QUIET
	HashiBool          bool   // HASHI_BOOL
	HashiPreserveOrder bool   // HASHI_PRESERVE_ORDER
	HashiMatchRequired bool   // HASHI_MATCH_REQUIRED
	
	HashiOutputFile string // HASHI_OUTPUT_FILE
	HashiAppend     bool   // HASHI_APPEND
	HashiForce      bool   // HASHI_FORCE
	
	HashiLogFile string // HASHI_LOG_FILE
	HashiLogJSON string // HASHI_LOG_JSON
	
	HashiHelp    bool // HASHI_HELP
	HashiVersion bool // HASHI_VERSION
	
	HashiBlacklistFiles string // HASHI_BLACKLIST_FILES
	HashiBlacklistDirs  string // HASHI_BLACKLIST_DIRS
	HashiWhitelistFiles string // HASHI_WHITELIST_FILES
	HashiWhitelistDirs  string // HASHI_WHITELIST_DIRS
}

// Config holds all configuration options for hashi.
type Config struct {
	Files  []string
	Hashes []string

	Recursive     bool
	Hidden        bool
	Algorithm     string
	Verbose       bool
	Quiet         bool
	Bool          bool
	PreserveOrder bool

	MatchRequired bool

	OutputFormat string
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

	ConfigFile string

	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string

	ShowHelp    bool
	ShowVersion bool
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Algorithm:    "sha256",
		OutputFormat: "default",
		MinSize:      0,
		MaxSize:      -1, // No limit
	}
}

// WriteError returns a generic error message for security-sensitive write failures.
func WriteError() error {
	return fmt.Errorf("Unknown write/append error")
}

// WriteErrorWithVerbose returns either a generic or detailed error message.
func WriteErrorWithVerbose(verbose bool, verboseDetails string) error {
	if verbose {
		return fmt.Errorf("%s", verboseDetails)
	}
	return WriteError()
}

// FileSystemError returns a generic error for file system operations.
func FileSystemError(verbose bool, verboseDetails string) error {
	if verbose {
		return fmt.Errorf("%s", verboseDetails)
	}
	return WriteError()
}

// HandleFileWriteError processes file writing errors.
func HandleFileWriteError(err error, verbose bool, path string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	
	if strings.Contains(errStr, "permission denied") || 
	   strings.Contains(errStr, "access is denied") {
		return FileSystemError(verbose, fmt.Sprintf("permission denied writing to %s", path))
	}
	
	if strings.Contains(errStr, "no space left") || 
	   strings.Contains(errStr, "disk full") {
		return FileSystemError(verbose, fmt.Sprintf("insufficient disk space for %s", path))
	}
	
	if strings.Contains(errStr, "network") || 
	   strings.Contains(errStr, "connection") ||
	   strings.Contains(errStr, "timeout") {
		return FileSystemError(verbose, fmt.Sprintf("network error writing to %s", path))
	}
	
	if strings.Contains(errStr, "file name too long") || 
	   strings.Contains(errStr, "path too long") {
		return FileSystemError(verbose, fmt.Sprintf("path too long: %s", path))
	}
	
	return err
}

// validateOutputPath validates that an output path is safe.
func validateOutputPath(path string, cfg *Config) error {
	opts := security.Options{
		Verbose:        cfg.Verbose,
		BlacklistFiles: cfg.BlacklistFiles,
		BlacklistDirs:  cfg.BlacklistDirs,
		WhitelistFiles: cfg.WhitelistFiles,
		WhitelistDirs:  cfg.WhitelistDirs,
	}
	return security.ValidateOutputPath(path, opts)
}

// LoadEnvConfig reads environment variables.
func LoadEnvConfig() *EnvConfig {
	env := &EnvConfig{
		NoColor:    os.Getenv("NO_COLOR") != "",
		Debug:      parseBoolEnv("DEBUG"),
		TmpDir:     os.Getenv("TMPDIR"),
		Home:       os.Getenv("HOME"),
		ConfigHome: os.Getenv("XDG_CONFIG_HOME"),
		
		HashiConfig:        os.Getenv("HASHI_CONFIG"),
		HashiAlgorithm:     os.Getenv("HASHI_ALGORITHM"),
		HashiOutputFormat:  os.Getenv("HASHI_OUTPUT_FORMAT"),
		HashiRecursive:     parseBoolEnv("HASHI_RECURSIVE"),
		HashiHidden:        parseBoolEnv("HASHI_HIDDEN"),
		HashiVerbose:       parseBoolEnv("HASHI_VERBOSE"),
		HashiQuiet:         parseBoolEnv("HASHI_QUIET"),
		HashiBool:          parseBoolEnv("HASHI_BOOL"),
		HashiPreserveOrder: parseBoolEnv("HASHI_PRESERVE_ORDER"),
		HashiMatchRequired: parseBoolEnv("HASHI_MATCH_REQUIRED"),
		
		HashiOutputFile: os.Getenv("HASHI_OUTPUT_FILE"),
		HashiAppend:     parseBoolEnv("HASHI_APPEND"),
		HashiForce:      parseBoolEnv("HASHI_FORCE"),
		
		HashiLogFile: os.Getenv("HASHI_LOG_FILE"),
		HashiLogJSON: os.Getenv("HASHI_LOG_JSON"),
		
		HashiHelp:    parseBoolEnv("HASHI_HELP"),
		HashiVersion: parseBoolEnv("HASHI_VERSION"),
		
		HashiBlacklistFiles: os.Getenv("HASHI_BLACKLIST_FILES"),
		HashiBlacklistDirs:  os.Getenv("HASHI_BLACKLIST_DIRS"),
		HashiWhitelistFiles: os.Getenv("HASHI_WHITELIST_FILES"),
		HashiWhitelistDirs:  os.Getenv("HASHI_WHITELIST_DIRS"),
	}
	
	return env
}

func parseBoolEnv(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ApplyEnvConfig applies environment variable configuration to a Config.
func (env *EnvConfig) ApplyEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
	env.applyBasicEnvConfig(cfg, flagSet)
	env.applyOutputEnvConfig(cfg, flagSet)
	env.applyBlacklistEnvConfig(cfg)
	env.applyWhitelistEnvConfig(cfg)
}

func (env *EnvConfig) applyBasicEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
	if !flagSet.Changed("algorithm") && env.HashiAlgorithm != "" {
		cfg.Algorithm = env.HashiAlgorithm
	}
	if !flagSet.Changed("recursive") && env.HashiRecursive {
		cfg.Recursive = env.HashiRecursive
	}
	if !flagSet.Changed("hidden") && env.HashiHidden {
		cfg.Hidden = env.HashiHidden
	}
	if !flagSet.Changed("verbose") && env.HashiVerbose {
		cfg.Verbose = env.HashiVerbose
	}
	if !flagSet.Changed("quiet") && env.HashiQuiet {
		cfg.Quiet = env.HashiQuiet
	}
	if !flagSet.Changed("bool") && env.HashiBool {
		cfg.Bool = env.HashiBool
	}
	if !flagSet.Changed("preserve-order") && env.HashiPreserveOrder {
		cfg.PreserveOrder = env.HashiPreserveOrder
	}
	if !flagSet.Changed("match-required") && env.HashiMatchRequired {
		cfg.MatchRequired = env.HashiMatchRequired
	}
	if !flagSet.Changed("help") && env.HashiHelp {
		cfg.ShowHelp = env.HashiHelp
	}
	if !flagSet.Changed("version") && env.HashiVersion {
		cfg.ShowVersion = env.HashiVersion
	}
}

func (env *EnvConfig) applyOutputEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
	if !flagSet.Changed("format") && env.HashiOutputFormat != "" {
		cfg.OutputFormat = env.HashiOutputFormat
	}
	if !flagSet.Changed("output") && env.HashiOutputFile != "" {
		cfg.OutputFile = env.HashiOutputFile
	}
	if !flagSet.Changed("append") && env.HashiAppend {
		cfg.Append = env.HashiAppend
	}
	if !flagSet.Changed("force") && env.HashiForce {
		cfg.Force = env.HashiForce
	}
	if !flagSet.Changed("log-file") && env.HashiLogFile != "" {
		cfg.LogFile = env.HashiLogFile
	}
	if !flagSet.Changed("log-json") && env.HashiLogJSON != "" {
		cfg.LogJSON = env.HashiLogJSON
	}
}

func (env *EnvConfig) applyBlacklistEnvConfig(cfg *Config) {
	if env.HashiBlacklistFiles != "" {
		patterns := parseCommaSeparated(env.HashiBlacklistFiles)
		cfg.BlacklistFiles = append(cfg.BlacklistFiles, patterns...)
	}
	if env.HashiBlacklistDirs != "" {
		patterns := parseCommaSeparated(env.HashiBlacklistDirs)
		cfg.BlacklistDirs = append(cfg.BlacklistDirs, patterns...)
	}
}

func (env *EnvConfig) applyWhitelistEnvConfig(cfg *Config) {
	if env.HashiWhitelistFiles != "" {
		patterns := parseCommaSeparated(env.HashiWhitelistFiles)
		cfg.WhitelistFiles = append(cfg.WhitelistFiles, patterns...)
	}
	if env.HashiWhitelistDirs != "" {
		patterns := parseCommaSeparated(env.HashiWhitelistDirs)
		cfg.WhitelistDirs = append(cfg.WhitelistDirs, patterns...)
	}
}

// LoadDotEnv loads environment variables from a .env file.
func LoadDotEnv(path string) error {
	if path == "" {
		path = ".env"
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf(".env line %d: invalid format (expected KEY=VALUE): %s", lineNum, line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}
	return nil
}

type ConfigFile struct {
	Defaults struct {
		Recursive     *bool   `toml:"recursive,omitempty"`
		Hidden        *bool   `toml:"hidden,omitempty"`
		Algorithm     *string `toml:"algorithm,omitempty"`
		Verbose       *bool   `toml:"verbose,omitempty"`
		Quiet         *bool   `toml:"quiet,omitempty"`
		Bool          *bool   `toml:"bool,omitempty"`
		PreserveOrder *bool   `toml:"preserve_order,omitempty"`
		MatchRequired *bool   `toml:"match_required,omitempty"`
		OutputFormat  *string `toml:"output_format,omitempty"`
		OutputFile    *string `toml:"output_file,omitempty"`
		Append        *bool   `toml:"append,omitempty"`
		Force         *bool   `toml:"force,omitempty"`
		LogFile       *string `toml:"log_file,omitempty"`
		LogJSON       *string `toml:"log_json,omitempty"`
		Include       []string `toml:"include,omitempty"`
		Exclude       []string `toml:"exclude,omitempty"`
		MinSize       *string  `toml:"min_size,omitempty"`
		MaxSize       *string  `toml:"max_size,omitempty"`
	} `toml:"defaults"`
	Security struct {
		BlacklistFiles []string `toml:"blacklist_files,omitempty"`
		BlacklistDirs  []string `toml:"blacklist_dirs,omitempty"`
		WhitelistFiles []string `toml:"whitelist_files,omitempty"`
		WhitelistDirs  []string `toml:"whitelist_dirs,omitempty"`
	} `toml:"security"`
	Files []string `toml:"files,omitempty"`
}

// LoadConfigFile reads and parses the configuration file at the given path.
func LoadConfigFile(path string) (*ConfigFile, error) {
	if path == "" {
		return &ConfigFile{}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{}, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	if strings.HasSuffix(strings.ToLower(path), ".toml") {
		return loadTOMLConfig(file)
	}
	return loadTextConfig(file)
}

func loadTOMLConfig(file *os.File) (*ConfigFile, error) {
	var cfg ConfigFile
	if _, err := toml.DecodeReader(file, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse TOML config: %w", err)
	}
	return &cfg, nil
}

func loadTextConfig(file *os.File) (*ConfigFile, error) {
	cfg := &ConfigFile{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cfg.Files = append(cfg.Files, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading text config: %w", err)
	}
	return cfg, nil
}

// ApplyConfigFile merges the file configuration into the main Config, respecting flag precedence.
func (cf *ConfigFile) ApplyConfigFile(cfg *Config, flagSet *pflag.FlagSet) error {
	cf.applyBoolDefaults(cfg, flagSet)
	cf.applyStringDefaults(cfg, flagSet)
	
	if err := cf.applySizeDefaults(cfg, flagSet); err != nil {
		return err
	}
	
	cf.applyListDefaults(cfg, flagSet)
	cf.applySecurityDefaults(cfg)
	
	if len(cf.Files) > 0 && len(cfg.Files) == 0 {
		cfg.Files = cf.Files
	}
	
	return nil
}

func (cf *ConfigFile) applyBoolDefaults(cfg *Config, flagSet *pflag.FlagSet) {
	d := cf.Defaults
	boolFlags := []struct {
		val  *bool
		name string
		ptr  *bool
	}{
		{d.Recursive, "recursive", &cfg.Recursive},
		{d.Hidden, "hidden", &cfg.Hidden},
		{d.Verbose, "verbose", &cfg.Verbose},
		{d.Quiet, "quiet", &cfg.Quiet},
		{d.Bool, "bool", &cfg.Bool},
		{d.PreserveOrder, "preserve-order", &cfg.PreserveOrder},
		{d.MatchRequired, "match-required", &cfg.MatchRequired},
		{d.Append, "append", &cfg.Append},
		{d.Force, "force", &cfg.Force},
	}

	for _, f := range boolFlags {
		if f.val != nil && !flagSet.Changed(f.name) {
			*f.ptr = *f.val
		}
	}
}

func (cf *ConfigFile) applyStringDefaults(cfg *Config, flagSet *pflag.FlagSet) {
	d := cf.Defaults
	stringFlags := []struct {
		val  *string
		name string
		ptr  *string
	}{
		{d.Algorithm, "algorithm", &cfg.Algorithm},
		{d.OutputFormat, "format", &cfg.OutputFormat},
		{d.OutputFile, "output", &cfg.OutputFile},
		{d.LogFile, "log-file", &cfg.LogFile},
		{d.LogJSON, "log-json", &cfg.LogJSON},
	}

	for _, f := range stringFlags {
		if f.val != nil && !flagSet.Changed(f.name) {
			*f.ptr = *f.val
		}
	}
}

func (cf *ConfigFile) applySizeDefaults(cfg *Config, flagSet *pflag.FlagSet) error {
	d := cf.Defaults
	if d.MinSize != nil && !flagSet.Changed("min-size") {
		size, err := parseSize(*d.MinSize)
		if err != nil {
			return fmt.Errorf("invalid min_size in config: %w", err)
		}
		cfg.MinSize = size
	}
	if d.MaxSize != nil && !flagSet.Changed("max-size") {
		size, err := parseSize(*d.MaxSize)
		if err != nil {
			return fmt.Errorf("invalid max_size in config: %w", err)
		}
		cfg.MaxSize = size
	}
	return nil
}

func (cf *ConfigFile) applyListDefaults(cfg *Config, flagSet *pflag.FlagSet) {
	d := cf.Defaults
	if len(d.Include) > 0 && !flagSet.Changed("include") {
		cfg.Include = d.Include
	}
	if len(d.Exclude) > 0 && !flagSet.Changed("exclude") {
		cfg.Exclude = d.Exclude
	}
}

func (cf *ConfigFile) applySecurityDefaults(cfg *Config) {
	s := cf.Security
	cfg.BlacklistFiles = append(cfg.BlacklistFiles, s.BlacklistFiles...)
	cfg.BlacklistDirs = append(cfg.BlacklistDirs, s.BlacklistDirs...)
	cfg.WhitelistFiles = append(cfg.WhitelistFiles, s.WhitelistFiles...)
	cfg.WhitelistDirs = append(cfg.WhitelistDirs, s.WhitelistDirs...)
}

var ValidOutputFormats = []string{"default", "verbose", "json", "plain"}
var ValidAlgorithms = []string{"sha256", "md5", "sha1", "sha512", "blake2b"}

// ValidateOutputFormat checks if the provided format string is supported.
func ValidateOutputFormat(format string) error {
	for _, valid := range ValidOutputFormats {
		if format == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid output format %q: must be one of %s", format, strings.Join(ValidOutputFormats, ", "))
}

// ValidateAlgorithm checks if the provided algorithm string is supported.
func ValidateAlgorithm(algorithm string) error {
	for _, valid := range ValidAlgorithms {
		if algorithm == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid algorithm %q: must be one of %s", algorithm, strings.Join(ValidAlgorithms, ", "))
}

// ValidateConfig validates the configuration and returns an error if invalid.
func ValidateConfig(cfg *Config) ([]conflict.Warning, error) {
	warnings := make([]conflict.Warning, 0)
	
	opts := security.Options{
		Verbose:        cfg.Verbose,
		BlacklistFiles: cfg.BlacklistFiles,
		BlacklistDirs:  cfg.BlacklistDirs,
		WhitelistFiles: cfg.WhitelistFiles,
		WhitelistDirs:  cfg.WhitelistDirs,
	}

	// 1. Security validation of inputs
	if err := security.ValidateInputs(cfg.Files, cfg.Hashes, opts); err != nil {
		return warnings, err
	}

	// 2. Format and algorithm validation
	if err := ValidateOutputFormat(cfg.OutputFormat); err != nil {
		return warnings, err
	}

	if err := ValidateAlgorithm(cfg.Algorithm); err != nil {
		return warnings, err
	}

	if cfg.MinSize < 0 {
		return warnings, fmt.Errorf("min-size must be non-negative, got %d", cfg.MinSize)
	}
	if cfg.MaxSize != -1 && cfg.MaxSize < 0 {
		return warnings, fmt.Errorf("max-size must be non-negative or -1 (no limit), got %d", cfg.MaxSize)
	}
	if cfg.MaxSize != -1 && cfg.MinSize > cfg.MaxSize {
		return warnings, fmt.Errorf("min-size (%d) cannot be greater than max-size (%d)", cfg.MinSize, cfg.MaxSize)
	}

	if err := validateOutputPath(cfg.OutputFile, cfg); err != nil {
		return warnings, fmt.Errorf("output file: %w", err)
	}
	
	if err := validateOutputPath(cfg.LogFile, cfg); err != nil {
		return warnings, fmt.Errorf("log file: %w", err)
	}
	
	if err := validateOutputPath(cfg.LogJSON, cfg); err != nil {
		return warnings, fmt.Errorf("JSON log file: %w", err)
	}

	return warnings, nil
}

func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" || s == "-1" {
		return -1, nil
	}

	suffixes := []struct {
		suffix string
		mult   int64
	}{
		{"TB", 1024 * 1024 * 1024 * 1024},
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"T", 1024 * 1024 * 1024 * 1024},
		{"G", 1024 * 1024 * 1024},
		{"M", 1024 * 1024},
		{"K", 1024},
		{"B", 1},
	}

	for _, s2 := range suffixes {
		if strings.HasSuffix(s, s2.suffix) {
			numStr := strings.TrimSuffix(s, s2.suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid size %q: %w", s, err)
			}
			return int64(num * float64(s2.mult)), nil
		}
	}

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: must be a number or include unit (KB, MB, GB)", s)
	}
	return num, nil
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date %q: use format YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS", s)
}

// ParseArgs parses command-line arguments and returns a Config.
func ParseArgs(args []string) (*Config, []conflict.Warning, error) {
	cfg := DefaultConfig()
	fs := pflag.NewFlagSet("hashi", pflag.ContinueOnError)

	// 1. Define and parse flags
	defineFlags(fs, cfg)
	if err := parseFlags(fs, args); err != nil {
		return nil, nil, err
	}

	// 2. Validate basic flag values
	if err := validateBasicFlags(cfg, fs); err != nil {
		return nil, nil, err
	}

	// 3. Handle remaining positional arguments
	if err := handleArguments(cfg, fs); err != nil {
		return nil, nil, err
	}

	// 4. Apply external configuration
	if err := applyExternalConfig(cfg, fs); err != nil {
		return nil, nil, err
	}

	// 5. Resolve final state and validate
	return finalizeConfig(cfg, args, fs)
}

func defineFlags(flagSet *pflag.FlagSet, cfg *Config) {
	flagSet.BoolVarP(&cfg.Recursive, "recursive", "r", false, "Process directories recursively")
	flagSet.BoolVar(&cfg.Hidden, "hidden", false, "Include hidden files")
	flagSet.StringVarP(&cfg.Algorithm, "algorithm", "a", "sha256", "Hash algorithm")
	flagSet.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	flagSet.BoolVarP(&cfg.Quiet, "quiet", "q", false, "Suppress stdout")
	flagSet.BoolVarP(&cfg.Bool, "bool", "b", false, "Boolean output mode")
	flagSet.BoolVar(&cfg.PreserveOrder, "preserve-order", false, "Keep input order")
	flagSet.BoolVar(&cfg.MatchRequired, "match-required", false, "Exit 0 only if matches found")
	flagSet.StringVarP(&cfg.OutputFormat, "format", "f", "default", "Output format")
	flagSet.StringVarP(&cfg.OutputFile, "output", "o", "", "Write output to file")
	flagSet.BoolVar(&cfg.Append, "append", false, "Append to output file")
	flagSet.BoolVar(&cfg.Force, "force", false, "Overwrite without prompting")

	jsonOutput := new(bool)
	plainOutput := new(bool)
	flagSet.BoolVar(jsonOutput, "json", false, "Output in JSON format")
	flagSet.BoolVar(plainOutput, "plain", false, "Output in plain format")

	flagSet.StringVar(&cfg.LogFile, "log-file", "", "File for logging")
	flagSet.StringVar(&cfg.LogJSON, "log-json", "", "File for JSON logging")

	flagSet.StringSliceVarP(&cfg.Include, "include", "i", nil, "Glob patterns to include")
	flagSet.StringSliceVarP(&cfg.Exclude, "exclude", "e", nil, "Glob patterns to exclude")
	
	// Add placeholders for string-based filters that need parsing
	flagSet.String("min-size", "0", "Minimum file size")
	flagSet.String("max-size", "-1", "Maximum file size")
	flagSet.String("modified-after", "", "Date")
	flagSet.String("modified-before", "", "Date")

	flagSet.StringVarP(&cfg.ConfigFile, "config", "c", "", "Path to config file")
	flagSet.BoolVarP(&cfg.ShowHelp, "help", "h", false, "Show help")
	flagSet.BoolVarP(&cfg.ShowVersion, "version", "V", false, "Show version")
}

func parseFlags(fs *pflag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "unknown flag: ") {
			unknown := strings.TrimPrefix(errMsg, "unknown flag: ")
			if suggestion := SuggestFlag(unknown); suggestion != "" {
				return fmt.Errorf("%s (Did you mean %s?)", errMsg, suggestion)
			}
		}
		return err
	}
	return nil
}

func validateBasicFlags(cfg *Config, fs *pflag.FlagSet) error {
	var err error
	if minStr, _ := fs.GetString("min-size"); minStr != "" {
		if cfg.MinSize, err = parseSize(minStr); err != nil {
			return fmt.Errorf("invalid --min-size: %w", err)
		}
	}
	if maxStr, _ := fs.GetString("max-size"); maxStr != "" {
		if cfg.MaxSize, err = parseSize(maxStr); err != nil {
			return fmt.Errorf("invalid --max-size: %w", err)
		}
	}
	if afterStr, _ := fs.GetString("modified-after"); afterStr != "" {
		if cfg.ModifiedAfter, err = parseDate(afterStr); err != nil {
			return fmt.Errorf("invalid --modified-after: %w", err)
		}
	}
	if beforeStr, _ := fs.GetString("modified-before"); beforeStr != "" {
		if cfg.ModifiedBefore, err = parseDate(beforeStr); err != nil {
			return fmt.Errorf("invalid --modified-before: %w", err)
		}
	}
	return nil
}

func handleArguments(cfg *Config, fs *pflag.FlagSet) error {
	remainingArgs := fs.Args()
	for _, arg := range remainingArgs {
		if arg == "config" {
			return &ConfigCommandError{}
		}
	}

	if len(remainingArgs) > 0 && allArgsAreNonExistentFiles(remainingArgs) {
		hasHashLikeArgs := false
		for _, arg := range remainingArgs {
			if looksLikeHashString(arg) {
				hasHashLikeArgs = true
				break
			}
		}
		if hasHashLikeArgs {
			cfg.Files = []string{}
			cfg.Hashes = remainingArgs
			return nil
		}
	}

	files, hashes, err := ClassifyArguments(remainingArgs, cfg.Algorithm)
	if err != nil {
		return err
	}
	cfg.Files = files
	cfg.Hashes = hashes
	return nil
}

func applyExternalConfig(cfg *Config, flagSet *pflag.FlagSet) error {
	envCfg := LoadEnvConfig()
	configPath := cfg.ConfigFile

	// Priority: Flag > Environment > Auto-discovery
	if !flagSet.Changed("config") && envCfg.HashiConfig != "" {
		configPath = envCfg.HashiConfig
	}

	if configPath == "" {
		configPath = FindConfigFile()
	}
	if configPath != "" {
		configFile, err := LoadConfigFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config file %s: %w", configPath, err)
		}
		if err := configFile.ApplyConfigFile(cfg, flagSet); err != nil {
			return fmt.Errorf("failed to apply config file %s: %w", configPath, err)
		}
	}

	envCfg.ApplyEnvConfig(cfg, flagSet)
	return nil
}

func finalizeConfig(cfg *Config, args []string, flagSet *pflag.FlagSet) (*Config, []conflict.Warning, error) {
	flagSetMap := map[string]bool{
		"json":    flagSet.Changed("json"),
		"plain":   flagSet.Changed("plain"),
		"quiet":   cfg.Quiet,
		"verbose": cfg.Verbose,
		"bool":    cfg.Bool,
	}

	state, resolveWarnings, err := conflict.ResolveState(args, flagSetMap, cfg.OutputFormat)
	if err != nil {
		return nil, nil, err
	}

	cfg.OutputFormat = string(state.Format)
	cfg.Quiet = (state.Verbosity == conflict.VerbosityQuiet)
	cfg.Verbose = (state.Verbosity == conflict.VerbosityVerbose)
	if state.Mode == conflict.ModeBool {
		cfg.Bool = true
		cfg.Quiet = true 
	}

	validationWarnings, err := ValidateConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	return cfg, append(resolveWarnings, validationWarnings...), nil
}

// HelpText returns the formatted help text.
func HelpText() string {
	var sb strings.Builder
	sb.WriteString(helpHeader)
	sb.WriteString(helpUsage)
	sb.WriteString(helpBooleanMode)
	sb.WriteString(helpOutputFormats)
	sb.WriteString(helpFiltering)
	sb.WriteString(helpConfiguration)
	sb.WriteString(helpEnvironment)
	sb.WriteString(helpFooter)
	return sb.String()
}

const helpHeader = `hashi - A command-line hash comparison tool

EXAMPLES
  hashi                          Hash all files in current directory
  hashi file1.txt file2.txt      Compare hashes of two files
  hashi -b file1.txt file2.txt   Boolean check: do files match? (outputs true/false)
  hashi -r /path/to/dir          Recursively hash directory
  hashi --json *.txt             Output results as JSON
  hashi -                        Read file list from stdin
`

const helpUsage = `
USAGE
  hashi [flags] [files...]

FLAGS
  -h, --help                Show this help
  -V, --version             Show version
  -v, --verbose             Enable verbose output
  -q, --quiet               Suppress stdout, only return exit code
  -b, --bool                Boolean output mode (true/false)
  -r, --recursive           Process directories recursively
      --hidden              Include hidden files
  -a, --algorithm string    Hash algorithm: sha256, md5, sha1, sha512, blake2b (default: sha256)
      --preserve-order      Keep input order instead of grouping by hash
`

const helpBooleanMode = `
BOOLEAN MODE (-b / --bool)
  Boolean mode outputs just "true" or "false" for scripting use cases.
  It overrides other output formats and implies quiet behavior.

  Default behavior (no match flags):
    hashi -b file1 file2 file3     # true if ALL files match

  With --match-required:
    hashi -b --match-required *.txt    # true if ANY matches found
`

const helpOutputFormats = `
OUTPUT FORMATS
  -f, --format string       Output format: default, verbose, json, plain
      --json                Shorthand for --format=json
      --plain               Shorthand for --format=plain
  -o, --output string       Write output to file
      --append              Append to output file
      --force               Overwrite without prompting
      --log-file string     File for logging
      --log-json string     File for JSON logging
`

const helpFiltering = `
EXIT CODE CONTROL
      --match-required      Exit 0 only if matches found

FILTERING
  -i, --include strings     Glob patterns to include
  -e, --exclude strings     Glob patterns to exclude
      --min-size string     Minimum file size (e.g., 10KB, 1MB, 1GB)
      --max-size string     Maximum file size (-1 for no limit)
      --modified-after      Only files modified after date (YYYY-MM-DD)
      --modified-before     Only files modified before date (YYYY-MM-DD)
`

const helpConfiguration = `
CONFIGURATION
  -c, --config string       Path to config file

  Config File Auto-Discovery (searched in order):
    ./.hashi.toml                        Project-specific (highest priority)
    $XDG_CONFIG_HOME/hashi/config.toml   XDG standard location
    ~/.config/hashi/config.toml          XDG fallback location
    ~/.hashi/config.toml                 Traditional dotfile location
`

const helpEnvironment = `
ENVIRONMENT VARIABLES
  HASHI_* Variables (override config file settings):
    HASHI_CONFIG            Default config file path
    HASHI_ALGORITHM         Hash algorithm (sha256, md5, sha1, sha512, blake2b)
    HASHI_OUTPUT_FORMAT     Output format (default, verbose, json, plain)
    HASHI_RECURSIVE         Process directories recursively (true/false)
`

const helpFooter = `
EXIT CODES
  0   Success
  1   No matches found (with --match-required)
  2   Some files failed to process
  3   Invalid arguments
  4   File not found
  5   Permission denied
  130 Interrupted (Ctrl-C)

For more information, visit: https://github.com/example/hashi
`

// VersionText returns the current version string.
func VersionText() string {
	return "hashi version 0.0.19"
}

// HasStdinMarker checks if the special "-" argument is present in the file list.
func (c *Config) HasStdinMarker() bool {
	for _, file := range c.Files {
		if file == "-" {
			return true
		}
	}
	return false
}

// FilesWithoutStdin returns the list of files excluding the stdin marker "-".
func (c *Config) FilesWithoutStdin() []string {
	result := make([]string, 0, len(c.Files))
	for _, file := range c.Files {
		if file != "-" {
			result = append(result, file)
		}
	}
	return result
}

// ClassifyArguments separates arguments into file paths and hash strings.
func ClassifyArguments(args []string, algorithm string) (files []string, hashes []string, err error) {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if arg == "-" {
			files = append(files, arg)
			continue
		}
		if _, err := os.Stat(arg); err == nil {
			files = append(files, arg)
			continue
		}
		detectedAlgorithms := detectHashAlgorithm(arg)
		if len(detectedAlgorithms) == 0 {
			if isValidHexString(arg) {
				return nil, nil, fmt.Errorf("argument %q looks like a hash but has an unknown length", arg)
			}
			files = append(files, arg)
			continue
		}
		currentAlgorithmFound := false
		for _, detected := range detectedAlgorithms {
			if detected == algorithm {
				currentAlgorithmFound = true
				break
			}
		}
		if currentAlgorithmFound {
			hashes = append(hashes, strings.ToLower(arg))
		} else {
			if len(detectedAlgorithms) == 1 {
				return nil, nil, fmt.Errorf("hash length doesn't match %s (expected %d characters, got %d).\nThis looks like %s. Try: hashi --algo %s [files...] %s",
					algorithm, getExpectedLength(algorithm), len(arg), 
					strings.ToUpper(detectedAlgorithms[0]), detectedAlgorithms[0], arg)
			} else {
				algorithmList := make([]string, len(detectedAlgorithms))
				for i, alg := range detectedAlgorithms {
					algorithmList[i] = strings.ToUpper(alg)
				}
				return nil, nil, fmt.Errorf("hash length doesn't match %s (expected %d characters, got %d).\nCould be: %s\nSpecify algorithm with: hashi --algo [algorithm] [files...] %s",
					algorithm, getExpectedLength(algorithm), len(arg),
					strings.Join(algorithmList, ", "), arg)
			}
		}
	}
	return files, hashes, nil
}

func detectHashAlgorithm(hashStr string) []string {
	if !isValidHexString(hashStr) {
		return []string{}
	}
	switch len(hashStr) {
	case 32:
		return []string{"md5"}
	case 40:
		return []string{"sha1"}
	case 64:
		return []string{"sha256"}
	case 128:
		return []string{"sha512", "blake2b"}
	default:
		return []string{}
	}
}

func isValidHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func getExpectedLength(algorithm string) int {
	switch algorithm {
	case "md5":
		return 32
	case "sha1":
		return 40
	case "sha256":
		return 64
	case "sha512", "blake2b":
		return 128
	default:
		return 64
	}
}

// FindConfigFile searches standard locations for a configuration file.
func FindConfigFile() string {
	locations := []string{
		"./.hashi.toml",
	}
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		locations = append(locations, filepath.Join(xdgConfigHome, "hashi", "config.toml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, filepath.Join(home, ".config", "hashi", "config.toml"))
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, filepath.Join(home, ".hashi", "config.toml"))
	}
	for _, location := range locations {
		if _, err := os.Stat(location); err == nil {
			return location
		}
	}
	return ""
}

func allArgsAreNonExistentFiles(args []string) bool {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		if arg == "-" {
			return false
		}
		if _, err := os.Stat(arg); err == nil {
			return false
		}
	}
	return true
}

func looksLikeHashString(s string) bool {
	if len(s) == 0 {
		return false
	}
	if len(s) < 4 || len(s) > 256 {
		return false
	}
	if strings.Contains(s, ".") {
		return false
	}
	if strings.Contains(s, "/") || strings.Contains(s, "\\") {
		return false
	}
	if strings.Contains(s, " ") {
		return false
	}
	return true
}
