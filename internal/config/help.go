package config

import (
	"strings"
)

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
      --dry-run             Preview files without hashing
  -a, --algorithm string    Hash algorithm: sha256, md5, sha1, sha512, blake2b (default: sha256)
      --test                Run system diagnostics for troubleshooting
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
  -f, --format string       Output format: default, verbose, json, jsonl, plain
      --json                Shorthand for --format=json
      --jsonl               Shorthand for --format=jsonl
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

INCREMENTAL OPERATIONS
      --manifest string     Path to baseline manifest file
      --only-changed        Process only new or modified files
      --output-manifest string  Path to save result as a manifest
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
