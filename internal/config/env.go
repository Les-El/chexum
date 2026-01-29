package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// EnvParser implements the Parser interface for environment variables.
type EnvParser struct {
	ExtraVars map[string]string
	FlagSet   *pflag.FlagSet
}

// NewEnvParser creates a new EnvParser.
func NewEnvParser(extra map[string]string, fs *pflag.FlagSet) *EnvParser {
	return &EnvParser{
		ExtraVars: extra,
		FlagSet:   fs,
	}
}

// Parse implements the Parser interface.
func (p *EnvParser) Parse(cfg *Config) error {
	env := LoadEnvConfig(p.ExtraVars)
	env.ApplyEnvConfig(cfg, p.FlagSet)
	return nil
}

// LoadEnvConfig reads environment variables into an EnvConfig struct.
// It accepts an optional map of extra variables (e.g. from a .env file)
// which take precedence over actual environment variables.
func LoadEnvConfig(extra map[string]string) *EnvConfig {
	get := func(key string) string {
		if val, ok := extra[key]; ok {
			return val
		}
		return os.Getenv(key)
	}

	env := &EnvConfig{
		NoColor:    get("NO_COLOR") != "",
		Debug:      parseBool(get("DEBUG")),
		TmpDir:     get("TMPDIR"),
		Home:       get("HOME"),
		ConfigHome: get("XDG_CONFIG_HOME"),

		HashiConfig:         get("HASHI_CONFIG"),
		HashiAlgorithm:      get("HASHI_ALGORITHM"),
		HashiOutputFormat:   get("HASHI_OUTPUT_FORMAT"),
		HashiDryRun:         parseBool(get("HASHI_DRY_RUN")),
		HashiRecursive:      parseBool(get("HASHI_RECURSIVE")),
		HashiHidden:         parseBool(get("HASHI_HIDDEN")),
		HashiVerbose:        parseBool(get("HASHI_VERBOSE")),
		HashiQuiet:          parseBool(get("HASHI_QUIET")),
		HashiBool:           parseBool(get("HASHI_BOOL")),
		HashiPreserveOrder:  parseBool(get("HASHI_PRESERVE_ORDER")),
		HashiMatchRequired:  parseBool(get("HASHI_MATCH_REQUIRED")),
		HashiManifest:       get("HASHI_MANIFEST"),
		HashiOnlyChanged:    parseBool(get("HASHI_ONLY_CHANGED")),
		HashiOutputManifest: get("HASHI_OUTPUT_MANIFEST"),

		HashiOutputFile: get("HASHI_OUTPUT_FILE"),
		HashiAppend:     parseBool(get("HASHI_APPEND")),
		HashiForce:      parseBool(get("HASHI_FORCE")),

		HashiLogFile: get("HASHI_LOG_FILE"),
		HashiLogJSON: get("HASHI_LOG_JSON"),

		HashiHelp:    parseBool(get("HASHI_HELP")),
		HashiVersion: parseBool(get("HASHI_VERSION")),

		HashiJobs:           parseInt(get("HASHI_JOBS")),
		HashiBlacklistFiles: get("HASHI_BLACKLIST_FILES"),
		HashiBlacklistDirs:  get("HASHI_BLACKLIST_DIRS"),
		HashiWhitelistFiles: get("HASHI_WHITELIST_FILES"),
		HashiWhitelistDirs:  get("HASHI_WHITELIST_DIRS"),
	}

	return env
}

func parseBool(val string) bool {
	val = strings.ToLower(val)
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func parseInt(val string) int {
	if val == "" {
		return 0
	}
	var res int
	fmt.Sscanf(val, "%d", &res)
	return res
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
	if !flagSet.Changed("dry-run") && env.HashiDryRun {
		cfg.DryRun = env.HashiDryRun
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
	if !flagSet.Changed("manifest") && env.HashiManifest != "" {
		cfg.Manifest = env.HashiManifest
	}
	if !flagSet.Changed("only-changed") && env.HashiOnlyChanged {
		cfg.OnlyChanged = env.HashiOnlyChanged
	}
	if !flagSet.Changed("help") && env.HashiHelp {
		cfg.ShowHelp = env.HashiHelp
	}
	if !flagSet.Changed("version") && env.HashiVersion {
		cfg.ShowVersion = env.HashiVersion
	}
	if !flagSet.Changed("jobs") && env.HashiJobs != 0 {
		cfg.Jobs = env.HashiJobs
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
	if !flagSet.Changed("output-manifest") && env.HashiOutputManifest != "" {
		cfg.OutputManifest = env.HashiOutputManifest
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

// LoadDotEnv loads environment variables from a .env file into a map.
func LoadDotEnv(path string) (map[string]string, error) {
	if path == "" {
		path = ".env"
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	envVars := make(map[string]string)
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
			return nil, fmt.Errorf(".env line %d: invalid format (expected KEY=VALUE): %s", lineNum, line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		envVars[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}
	return envVars, nil
}
