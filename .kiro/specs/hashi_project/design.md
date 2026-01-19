# Design Document: Hashi CLI Tool

## Overview

This design document describes the architecture and implementation of `hashi`, a command-line hash comparison tool built from the ground up to follow industry-standard CLI design guidelines. This is a **complete rewrite** implementing a modular architecture with human-first design principles.

**Project Status**: Tasks 1-8 complete (core infrastructure and error handling). See `tasks.md` for detailed progress.

## Guiding Principles

### 1. Developer Continuity
Every function explains *why* it exists, not just *what* it does. Architecture decisions are documented with rationale. The codebase serves as a teaching tool, not just a product.

### 2. User-First Design
Everything is designed with the question: "What functionality does the user need, and what behavior does the user expect?" Features solve real user problems, defaults match user expectations, and error messages guide users to solutions.

### 3. No Lock-Out
Users must never be locked out of functionality due to design choices. Every default behavior has an escape hatch:
- `--raw` bypasses ZIP auto-verification when users want to hash the file itself
- `--preserve-order` bypasses default grouping when users need input order
- When adding new "smart" defaults, always provide a flag to override them

### 4. CLI Guidelines Compliance
All design decisions follow the [CLI Guidelines](https://clig.dev/) standard, prioritizing:
- Human-first design with machine-readable alternatives
- Simple parts that work together (composability)
- Consistency across the interface
- Appropriate verbosity (saying just enough)
- Ease of discovery
- Conversational interaction
- Robustness and reliability

## Architecture Overview

Hashi follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│  Entry point, argument parsing, help system                 │
│  Location: cmd/hashi/main.go                                │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    Configuration Layer                      │
│  internal/config/    - Argument parsing, flag handling      │
│  internal/conflict/  - Flag conflict detection              │
│  internal/filter/    - File filtering (patterns, size, date)│
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     Processing Layer                        │
│  internal/hash/      - Hash computation                     │
│  internal/archive/   - ZIP verification (CRC32)             │
│  internal/dryrun/    - Preview mode (no hashing)            │
│  internal/manifest/  - Incremental operations               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                       Output Layer                          │
│  internal/output/    - Formatters (default, JSON, plain)    │
│  internal/fileout/   - Safe file writing (atomic ops)       │
│  internal/color/     - TTY detection, color handling        │
│  internal/progress/  - Progress indicators                  │
│  internal/errors/    - Error formatting, suggestions        │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Control Layer                          │
│  internal/signals/   - Signal handling (Ctrl-C)             │
└─────────────────────────────────────────────────────────────┘
```


## Core Components

### 1. Configuration System (`internal/config/`)

**Purpose**: Parse command-line arguments, load configuration from multiple sources, and merge with proper precedence.

**Key Design Decisions**:
- **TOML format**: Human-readable configuration files (not JSON)
- **No bookmarks/templates**: Keep simple, avoid complexity that violates hashi principles
- **Safe config display**: `--show-config` reveals minimal information, no sensitive paths
- **Per-directory config**: Support `.hashi.toml` in current directory
- **No validation commands**: Config errors are reported naturally during execution

**Key Types**:
```go
type Config struct {
    // Input sources
    Files []string
    Hashes []string
    
    // Behavior flags
    Recursive bool
    Hidden bool
    Algorithm string
    Verbose bool
    Quiet bool          // Suppress stdout, return exit code only
    PreserveOrder bool  // Maintain input order vs grouping by hash
    DryRun bool         // Preview mode - enumerate files without hashing
    
    // Exit code control
    MatchRequired bool    // Exit 0 only if matches found
    // Note: --no-match-required flag ELIMINATED - use shell negation instead
    
    // Output configuration
    OutputFormat string  // "default", "verbose", "json", "plain"
    OutputFile string
    Append bool
    Force bool
    
    // Filtering
    Include []string     // Glob patterns for files to include
    Exclude []string     // Glob patterns for files to exclude
    MinSize int64        // Minimum file size in bytes
    MaxSize int64        // Maximum file size in bytes
    ModifiedAfter time.Time   // Only process files modified after this date
    ModifiedBefore time.Time  // Only process files modified before this date
    
    // Incremental operations
    ManifestFile string       // Previous manifest for comparison
    OnlyChanged bool          // Process only changed files
    OutputManifest string     // Save manifest for future incremental runs
    
    // Special handling
    Raw bool  // Treat special files (ZIP) as raw bytes
    
    // Boolean output
    BoolOutput bool  // Output only "true"/"false" for scripting
    
    // Source tracking for configuration display
    Sources []ConfigSource
}
```

**Configuration File Format**: TOML (human-readable, well-supported)

**Configuration Precedence** (highest to lowest):
1. Command-line flags
2. Environment variables (HASHI_* prefix)
3. Project-level config (`.hashi.toml` in working directory)
4. User-level config (`~/.config/hashi/config.toml`)
5. Built-in defaults

**TOML Configuration Example**:
```toml
# ~/.config/hashi/config.toml
[defaults]
algorithm = "sha256"
output_format = "plain"
recursive = false
quiet = false
parallel = 0  # auto-detect

[colors]
enabled = true
matches = "green"
mismatches = "red"
warnings = "yellow"
```

**Config File Auto-Discovery**:
When no `--config` flag is provided, hashi searches for config files in this order:
1. `./.hashi.toml` - Project-specific (highest priority)
2. `$XDG_CONFIG_HOME/hashi/config.toml` - XDG standard
3. `~/.config/hashi/config.toml` - XDG fallback
4. `~/.hashi/config.toml` - Traditional dotfile

**Safe Configuration Display**:
```bash
$ hashi --show-config
Current configuration:
  Algorithm: sha256 (default)
  Output format: plain (user-config)
  Recursive: true (environment)
  Quiet: false (default)
  Parallel: auto (default)

Configuration loaded from 3 sources.
Environment variables: 2 found
Config files: 1 loaded

For manual configuration help:
  https://github.com/[repo]/hashi#configuration
```

**Security-First Design**:
- No paths, no values, no structure exposed
- Only shows effective settings and source types
- Perfect for automation/scripting
- Secure by default
- Links to documentation for advanced needs

**Config Command Handling**:
Hashi does NOT support a `config` subcommand. Users must manually edit config files.

**Error Message** (when user tries `hashi config` anywhere in arguments):
```
Error: hashi does not support config subcommands

Configuration must be done by manually editing config files.

Hashi auto-loads config from these standard locations:
  • .hashi.toml (project-specific)
  • hashi/config.toml (in XDG config directory)
  • .hashi/config.toml (traditional dotfile)

For configuration documentation and examples, see:
  https://github.com/[your-repo]/hashi#configuration
```


### 2. Color and TTY Detection (`internal/color/`)

**Purpose**: Detect terminal capabilities and manage color output according to CLI guidelines.

**Key Features**:
- Automatic TTY detection using `term.IsTerminal()`
- Respects `NO_COLOR` environment variable
- Respects `TERM=dumb`
- Individual checking of stdout and stderr (colors on stderr even when stdout is piped)
- Provides semantic color methods (Success, Error, Warning, Info, Path)

**Color Scheme**:
- **Green**: Matches, success messages
- **Red**: Errors, mismatches
- **Yellow**: Warnings
- **Blue**: Informational messages
- **Cyan**: File paths
- **Gray**: Secondary information

**Interface**:
```go
type ColorHandler struct {
    stdoutEnabled bool
    stderrEnabled bool
}

func NewColorHandler() *ColorHandler
func (c *ColorHandler) Success(text string) string
func (c *ColorHandler) Error(text string) string
func (c *ColorHandler) Warning(text string) string
func (c *ColorHandler) Info(text string) string
func (c *ColorHandler) Path(text string) string
```


### 3. Progress Indicators (`internal/progress/`)

**Purpose**: Show progress for operations taking longer than 100ms, following CLI guidelines for responsive feedback.

**Key Features**:
- Only displays when stdout is a TTY
- Shows percentage, count, and ETA
- Supports both file-level and batch-level progress
- Automatically hides when operation completes quickly (<100ms)

**Display Format**:
```
Processing files... [████████████░░░░░░░░] 60% (6/10) ETA: 2s
```

**Interface**:
```go
type ProgressBar struct {
    total int64
    current int64
    startTime time.Time
    enabled bool
}

func NewProgressBar(total int64) *ProgressBar
func (p *ProgressBar) Update(n int64)
func (p *ProgressBar) Finish()
```


### 4. Error Handling (`internal/errors/`)

**Purpose**: Transform technical errors into user-friendly messages with actionable suggestions.

**Design Philosophy**:
- **Be Specific**: Tell the user exactly what went wrong
- **Be Actionable**: Suggest how to fix the problem
- **Be Concise**: Keep messages short and to the point
- **Be Friendly**: Use conversational language, not technical jargon
- **Group Similar Errors**: Reduce noise by grouping repeated error types

**Error Categories**:
1. **User Input Errors**: Invalid arguments, missing files, bad hash format
2. **File System Errors**: Permission denied, file not found, disk full
3. **Processing Errors**: Hash computation failures, memory issues
4. **Configuration Errors**: Invalid config file, bad environment variables

**Example Transformations**:

Before (technical):
```
Error: failed to open file "document.pdf": no such file or directory
```

After (user-friendly):
```
✗ Cannot find file: document.pdf

  The file doesn't exist in the current directory.
  
  Try:
    • Check the file name for typos
    • Use 'ls' to see available files
    • Provide the full path: /path/to/document.pdf
```

**Interface**:
```go
type ErrorHandler struct {
    verbose bool
}

func (e *ErrorHandler) FormatError(err error) string
func (e *ErrorHandler) SuggestFix(err error) string
func (e *ErrorHandler) GroupErrors(errors []error) []ErrorGroup
```


### 5. Hash Computation (`internal/hash/`)

**Purpose**: Compute cryptographic hashes efficiently using streaming for memory efficiency.

**Key Features**:
- Streaming file processing (never loads entire file into memory)
- Support for multiple algorithms (SHA-256, SHA-512, MD5, SHA-1)
- Efficient handling of large files
- Progress tracking integration

**Interface**:
```go
type HashEntry struct {
    Original string    // Original argument (file path or hash string)
    Hash string        // Computed or provided hash
    IsFile bool        // True if this is a file path
    Error error        // Processing error if any
    Size int64         // File size in bytes
    ModTime time.Time  // File modification time
    Algorithm string   // Hash algorithm used
}

func ComputeFileHash(filepath string, algorithm string) (string, error)
func ProcessEntries(entries []string, config *Config) ([]HashEntry, error)
```

**Streaming Implementation**:
```go
func ComputeFileHash(filepath string, algorithm string) (string, error) {
    file, err := os.Open(filepath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    hasher := createHasher(algorithm)
    
    // Stream in chunks - memory efficient for any file size
    _, err = io.Copy(hasher, file)
    if err != nil {
        return "", err
    }
    
    return hex.EncodeToString(hasher.Sum(nil)), nil
}
```

#### Hash Algorithm Detection

**Purpose**: Detect possible hash algorithms from a hash string based on length and character validation.

**Hash Length Reference**:
- CRC32: 8 characters
- MD5: 32 characters
- SHA-1: 40 characters
- SHA-256: 64 characters (default)
- SHA-512: 128 characters
- BLAKE2b-512: 128 characters

**Interface**:
```go
// DetectHashAlgorithm returns possible algorithms for a hash string.
// Returns empty slice if not a valid hash format.
// Returns multiple algorithms if ambiguous (e.g., SHA-512 and BLAKE2b-512 both = 128 chars).
func DetectHashAlgorithm(hashStr string) []string
```

**Logic**:
1. Check if string contains only hex characters (0-9, a-f, A-F)
2. Check length against known algorithms
3. Return all matching algorithms
4. Return empty slice if no match

**Example outputs**:
- `"a1b2c3d4..."` (32 chars) → `["md5"]`
- `"abc123..."` (64 chars) → `["sha256"]`
- `"def456..."` (128 chars) → `["sha512", "blake2b"]`
- `"xyz"` (3 chars) → `[]`
- `"gggggg..."` (64 chars, invalid hex) → `[]`


#### Argument Classification

**Purpose**: Separate command-line arguments into files and hash strings with clear precedence rules.

**Interface**:
```go
// ClassifyArguments separates command-line arguments into files and hash strings.
// Files take precedence: if an argument exists as a file, it's treated as a file
// even if it looks like a valid hash string.
//
// Parameters:
//   args - raw command-line arguments (after flag parsing)
//   algorithm - the current hash algorithm (from --algo flag or default)
//
// Returns:
//   files - list of file paths
//   hashes - list of hash strings
//   error - helpful error if hash doesn't match current algorithm
func ClassifyArguments(args []string, algorithm string) (files []string, hashes []string, err error)
```

**Classification Logic** (for each argument):
1. Check if file exists with `os.Stat(arg)` → add to files list
2. If not a file, call `DetectHashAlgorithm(arg)`
3. If no algorithms detected → treat as file (will error later if not found)
4. If algorithms detected:
   - If current algorithm in list → add to hashes list
   - If current algorithm NOT in list and only 1 detected → return error with suggestion
   - If current algorithm NOT in list and multiple detected → return error listing options

**Error Messages**:
```
Hash length doesn't match sha256 (expected 64 characters, got 32).
This looks like MD5. Try: hashi --algo md5 file.txt [hash]
```

```
Hash length doesn't match sha256 (expected 64 characters, got 128).
Could be: sha512, blake2b
Specify algorithm with: hashi --algo sha512 file.txt [hash]
```


### 6. Output Formatters (`internal/output/`)

**Purpose**: Format results in multiple output formats for both human and machine consumption.

**Formatter Types**:

#### Default Formatter (Human-First)
Groups files by matching hash with blank lines between groups. This makes it easy to visually identify duplicates.

```
file_name_1.txt    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
file_name_4.pdf    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855

file_name_2.doc    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
file_name_3.zip    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
file_name_7.tar    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890

file_name_5.jpg    9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef
```

Key features:
- Files grouped by matching hash
- Blank line between groups
- Groups with 2+ files are matches
- Single files are unmatched
- Still pipeable (one file per line, consistent format)

#### Preserve Order Formatter
Maintains input order when `--preserve-order` flag is used. Same format but no grouping.

#### Verbose Formatter
Provides detailed output with summaries:

```
Processed 7 files in 0.234s

Match Groups:
  Group 1 (2 files):
    file_name_1.txt    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
    file_name_4.pdf    e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
  
  Group 2 (3 files):
    file_name_2.doc    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
    file_name_3.zip    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
    file_name_7.tar    a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890

Unmatched Files:
  file_name_5.jpg    9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef

Summary: 2 match groups, 1 unmatched file
```

#### JSON Formatter
Machine-readable structured output:

```json
{
  "processed": 7,
  "duration_ms": 234,
  "match_groups": [
    {
      "hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "count": 2,
      "files": ["file_name_1.txt", "file_name_4.pdf"]
    }
  ],
  "unmatched": [
    {
      "file": "file_name_5.jpg",
      "hash": "9876543210abcdef9876543210abcdef9876543210abcdef9876543210abcdef"
    }
  ],
  "errors": []
}
```

#### Plain Formatter
Tab-separated output optimized for grep/awk/cut:

```
file_name_1.txt	e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
file_name_2.doc	a1b2c3d4e5f6789012345678901234567890123456789012345678901234567890
```

No grouping, no blank lines, consistent format for scripting.


### 7. Archive Verification (`internal/archive/`)

**Purpose**: Verify internal checksums of archive files (currently ZIP CRC32).

**Security Model**:
- Verifies **integrity** (bits not corrupted) NOT **authenticity** (not tampered with)
- Always uses CRC32 for ZIP files regardless of metadata (prevents algorithm substitution attacks)
- A malicious actor can craft files with valid CRC32 - this is clearly documented

**Default Behavior**:
```bash
# Default: verify ZIP integrity, return boolean (exit code only)
hashi file.zip                    # Exit 0 if CRC32 passes, 6 if fails
hashi file1.zip file2.zip         # Exit 0 only if ALL pass
hashi /path/to/zips/              # Verify all ZIPs in directory

# Override: treat ZIP as raw bytes for standard hashing
hashi --raw file.zip              # Compute SHA-256 of the ZIP file itself
```

**Output Modes**:
- **Default (boolean)**: No stdout, exit code only - perfect for scripts
- **Verbose**: Show detailed verification results per entry
- **JSON**: Machine-readable details including failed entries

**Interface**:
```go
type ArchiveVerifier struct {
    verbose bool
}

type VerificationResult struct {
    FilePath string
    Passed bool
    FailedEntries []string  // Names of entries that failed CRC32
    Error error
}

func (v *ArchiveVerifier) VerifyZIP(path string) (*VerificationResult, error)
func (v *ArchiveVerifier) IsArchiveFile(path string) bool
```


### 8. Signal Handling (`internal/signals/`)

**Purpose**: Handle Ctrl-C gracefully according to CLI guidelines.

**Behavior**:
- **First Ctrl-C**: Stop processing immediately, show status, run cleanup with timeout
- **Second Ctrl-C**: Skip remaining cleanup, exit immediately
- **Always**: Display what cleanup was skipped (if any)

**Design Philosophy**:
- Exit as soon as possible (responsive)
- Leave system in recoverable state (crash-only design)
- Expect to be started when cleanup hasn't run
- Never hang on cleanup (5-second timeout)

**Interface**:
```go
type SignalHandler struct {
    interrupted bool
    cleanup func()
    cleanupTimeout time.Duration
}

func NewSignalHandler(cleanup func()) *SignalHandler
func (s *SignalHandler) Start()
func (s *SignalHandler) IsInterrupted() bool
```

**Example Output**:
```
^C
Interrupted. Cleaning up...
(Press Ctrl-C again to force quit)

Cleanup complete. Exiting.
```

Or if forced:
```
^C
Interrupted. Cleaning up...
^C
Forced exit. Skipped: temp file cleanup
```


### 9. Flag Conflict Detection (`internal/conflict/`)

**Purpose**: Detect and handle conflicting flag combinations using a matrix-based system.

**Design Philosophy**: Matrix-based conflict resolution with clear rules for different conflict types.

**Conflict Types**:
```go
type ConflictType int
const (
    Override ConflictType = iota      // Higher precedence wins
    MutuallyExclusive                 // Fatal error
    Implies                           // Flag A automatically sets flag B
)
```

**Conflict Resolution Matrix**:
```go
var ConflictMatrix = []ConflictRule{
    // Output format hierarchy (Override type)
    {
        Flags: []string{"--bool", "--json"},
        Type: Override,
        Winner: "--bool",
        Message: "Boolean output overrides JSON format",
        SuppressWarning: true, // Bool mode is quiet
    },
    {
        Flags: []string{"--bool", "--quiet"},
        Type: Override,
        Winner: "--bool", 
        Message: "Boolean output includes quiet behavior",
        SuppressWarning: true,
    },
    {
        Flags: []string{"--quiet", "--verbose"},
        Type: Override,
        Winner: "--quiet", 
        Message: "Quiet mode overrides verbose output",
        SuppressWarning: true, // Quiet suppresses warnings
    },
    {
        Flags: []string{"--json", "--plain"},
        Type: Override,
        Winner: "last-specified", // Special case
        Message: "Using last specified format",
    },
    
    // Mutually exclusive conflicts (Fatal)
    {
        Flags: []string{"--raw", "--verify"},
        Type: MutuallyExclusive,
        Message: "Choose: treat as raw bytes OR verify internal checksums",
    },
}
```

**Key Design Decisions**:
- **--bool beats --quiet**: Bool output implies quiet behavior, no conflict
- **No --no-match-required**: Eliminated confusing flag, use shell negation instead
- **Matrix-based**: Scalable system for adding new conflict rules
- **Override vs Fatal**: Override conflicts are resolved with warnings, fatal conflicts stop execution

**Interface**:
```go
type ConflictResolver struct {
    rules []ConflictRule
    logger io.Writer
}

type ConflictRule struct {
    Flags []string
    Type ConflictType
    Winner string  // For Override type
    Message string
    SuppressWarning bool
}

func NewResolver() *ConflictResolver
func (c *ConflictResolver) Check(config *Config) error
func (c *ConflictResolver) ResolveOverrides(config *Config) []string // Returns warnings
```


### 10. Exit Code System

**Purpose**: Provide meaningful exit codes for scripting and automation.

**Exit Codes**:
```go
const (
    ExitSuccess         = 0   // All files processed successfully
    ExitNoMatches       = 1   // No matches found (with --match-required)
    ExitMatchesFound    = 1   // Matches found (with --no-match-required)
    ExitPartialFailure  = 2   // Some files failed to process
    ExitInvalidArgs     = 3   // Invalid arguments or flags
    ExitFileNotFound    = 4   // One or more files not found
    ExitPermissionError = 5   // Permission denied
    ExitIntegrityFail   = 6   // Archive integrity verification failed
    ExitInterrupted     = 130 // Interrupted by Ctrl-C (128 + SIGINT)
)
```

**Behavior**:
- **Default**: Exit 0 if all files processed, non-zero for errors
- **With `--match-required`**: Exit 0 only if matches found, 1 if no matches
- **With `--no-match-required`**: Exit 0 only if no matches found, 1 if matches exist
- **Always**: Use specific error codes for invalid args, file not found, etc.

**Scripting Examples**:
```bash
# Check if files match (boolean check)
if hashi file1.txt file2.txt --match-required --quiet; then
    echo "Files are identical"
else
    echo "Files differ"
fi

# Verify no duplicates exist
if hashi *.jpg --no-match-required --quiet; then
    echo "No duplicate images"
else
    echo "Duplicates found!"
fi

# Complex check with filtering
if hashi --recursive --include "*.pdf" --min-size 1MB --match-required -q; then
    echo "Found duplicate large PDFs"
    # Now run again without --quiet to see details
    hashi --recursive --include "*.pdf" --min-size 1MB
fi
```

**Quiet Mode**:
- With `-q` or `--quiet`: Suppress all stdout, only return exit code
- Errors still go to stderr (so failures can be diagnosed)
- Perfect for boolean checks in scripts


### 11. Filter Engine (`internal/filter/`)

**Purpose**: Apply file filtering based on patterns, size, and modification dates.

**Key Features**:
- Glob pattern matching for include/exclude
- Size-based filtering (min/max)
- Date-based filtering (modified before/after)
- Multiple pattern support
- Exclude takes precedence over include

**Interface**:
```go
type FilterEngine struct {
    include []string
    exclude []string
    minSize int64
    maxSize int64
    modifiedAfter time.Time
    modifiedBefore time.Time
}

func NewFilterEngine(config *Config) *FilterEngine
func (f *FilterEngine) ShouldProcess(path string, info os.FileInfo) bool
func (f *FilterEngine) MatchesPattern(path string, patterns []string) bool
```

**Pattern Matching**:
- Supports glob patterns: `*.pdf`, `**/*.jpg`, `test_*.txt`
- Multiple patterns can be specified (comma-separated or multiple flags)
- Case-sensitive by default (platform-dependent)

**Examples**:
```bash
# Include only PDFs
hashi --include "*.pdf" -a

# Exclude temporary files
hashi --exclude "*.tmp,*.log" -a

# Large files only
hashi --min-size 1MB --max-size 100MB -a

# Recent files
hashi --modified-after 2026-01-01 -a

# Complex filtering
hashi --include "*.jpg,*.png" --exclude "*_thumb.*" --min-size 100KB -a
```


### 12. Dry Run System (`internal/dryrun/`)

**Purpose**: Preview operations without actually computing hashes.

**Key Features**:
- File enumeration with filters applied
- Size calculation and aggregation
- Time estimation based on file sizes
- No actual hash computation

**Interface**:
```go
type DryRunResult struct {
    FileCount int
    TotalSize int64
    EstimatedTime time.Duration
    FilesListed []string
}

func PerformDryRun(config *Config) (*DryRunResult, error)
func EstimateProcessingTime(totalSize int64) time.Duration
```

**Output Format**:
```
Dry Run Preview
===============
Files to process: 1,234
Total size: 45.2 GB
Estimated time: 2m 34s

Filters applied:
  Include: *.pdf
  Min size: 1 MB
  Modified after: 2026-01-01

Run without --dry-run to process these files.
```

**Time Estimation**:
- Based on average throughput (e.g., 200 MB/s for SHA-256)
- Accounts for file system overhead
- Provides conservative estimates


### 13. Manifest System (`internal/manifest/`)

**Purpose**: Track file state for incremental operations.

**Key Features**:
- JSON-based manifest format
- File metadata tracking (path, hash, size, mtime)
- Change detection logic
- Manifest validation

**Manifest Format**:
```json
{
  "version": "1.0",
  "algorithm": "sha256",
  "created": "2026-01-17T10:30:00Z",
  "files": [
    {
      "path": "src/main.go",
      "hash": "a1b2c3d4...",
      "size": 1024,
      "modified": "2026-01-17T09:15:00Z"
    }
  ]
}
```

**Interface**:
```go
type Manifest struct {
    Version string
    Algorithm string
    Created time.Time
    Files []ManifestEntry
}

type ManifestEntry struct {
    Path string
    Hash string
    Size int64
    Modified time.Time
}

func LoadManifest(path string) (*Manifest, error)
func (m *Manifest) SaveTo(path string) error
func (m *Manifest) HasChanged(path string, info os.FileInfo) bool
func (m *Manifest) GetChangedFiles(currentFiles []string) ([]string, error)
```

**Change Detection Logic**:
1. Check if file exists in manifest
2. If not in manifest → changed (new file)
3. If in manifest but missing → changed (deleted file)
4. If size differs → changed
5. If mtime differs → changed
6. Otherwise → unchanged

**Usage Examples**:
```bash
# First run - create manifest
hashi -a --output-manifest baseline.json

# Incremental run - process only changed files
hashi -a --manifest baseline.json --only-changed --output-manifest new.json

# CI/CD workflow
hashi --manifest previous-build.json --only-changed ./artifacts
```


### 14. File Output Manager (`internal/fileout/`)

**Purpose**: Safely write output to files with atomic operations and safety checks.

**Key Features**:
- Atomic writes (temp file + rename)
- Overwrite protection with confirmation
- Append mode support
- JSON log append with validity maintenance
- Path validation

**Interface**:
```go
type FileOutputManager struct {
    force bool
    append bool
}

func NewFileOutputManager(force, append bool) *FileOutputManager
func (f *FileOutputManager) WriteToFile(path string, content []byte) error
func (f *FileOutputManager) AppendToFile(path string, content []byte) error
func (f *FileOutputManager) AppendJSONEntry(path string, entry interface{}) error
func (f *FileOutputManager) PromptOverwrite(path string) (bool, error)
```

**Atomic Write Process**:
1. Write content to temporary file (`.hashi-tmp-XXXXX`)
2. Sync to disk (ensure data is written)
3. Rename temp file to target path (atomic operation)
4. If any step fails, clean up temp file

**JSON Append Logic**:
For maintaining valid JSON array structure:
1. Read existing file
2. Parse as JSON array
3. Append new entry to array
4. Write back atomically

**Safety Features**:
- Check if file exists before writing
- Prompt for confirmation unless `--force`
- Validate path is writable
- Preserve original file on write failure
- Clear error messages for permission issues


## Data Flow

The complete data flow through the system:

```
1. INPUT PHASE
   ├─ Parse command-line arguments (internal/config)
   ├─ Auto-discover config file if --config not specified
   ├─ Load environment variables
   ├─ Load configuration files (.env, config.json)
   ├─ Merge with precedence: flags > env > config
   ├─ Classify arguments into files and hashes
   └─ Detect flag conflicts (internal/conflict)

2. VALIDATION PHASE
   ├─ Validate all inputs (paths, hashes, flags)
   ├─ Check file existence and permissions
   ├─ Validate hash strings match current algorithm
   ├─ Resolve file type ambiguities (ZIP vs raw)
   ├─ Validate filter patterns and date formats
   ├─ Load and validate manifest file if specified
   └─ Fail fast with user-friendly errors

3. MODE SELECTION
   ├─ If --dry-run → Dry Run Mode
   ├─ If no files and only hashes → Hash Validation Mode
   ├─ If one file and one hash → File+Hash Comparison Mode
   ├─ If multiple files and hashes → Error (not supported)
   ├─ If --only-changed with manifest → Incremental Mode
   └─ Otherwise → Standard Hash Computation Mode

4. FILE ENUMERATION PHASE
   ├─ Traverse directories (recursive if -a flag)
   ├─ Apply filter engine (internal/filter)
   │  ├─ Pattern matching (include/exclude)
   │  ├─ Size filtering (min/max)
   │  └─ Date filtering (modified before/after)
   ├─ If incremental mode: compare against manifest
   └─ Build list of files to process

5. DRY RUN MODE (if --dry-run)
   ├─ Count files and calculate total size
   ├─ Estimate processing time
   ├─ Display preview with filters applied
   └─ Exit with code 0

6. PROCESSING PHASE
   ├─ Initialize progress tracking (internal/progress)
   ├─ For each file:
   │  ├─ Check if archive (internal/archive)
   │  ├─ If archive and not --raw: verify CRC32
   │  └─ If not archive or --raw: compute hash (internal/hash)
   ├─ Update progress indicators
   ├─ Handle interruptions (internal/signals)
   └─ Build manifest entries if --output-manifest

7. MATCHING PHASE
   ├─ Group entries by hash
   ├─ Identify match groups (2+ files with same hash)
   └─ Identify unmatched files

8. OUTPUT PHASE
   ├─ Select formatter based on flags (internal/output)
   ├─ Apply color if TTY (internal/color)
   ├─ Format errors if any (internal/errors)
   ├─ Write to stdout or file (internal/fileout)
   ├─ Save manifest if --output-manifest
   └─ Determine exit code based on results and flags
```


## Operation Modes

### Mode 1: Hash String Validation
**Trigger**: `hashi [hash_string]` (no files, only hash strings)

**Purpose**: Validate hash format and identify possible algorithms.

**Output**:
```
✅ Valid SHA-256 hash format (64 hex characters)
```

For ambiguous lengths:
```
✅ Valid hash format (128 hex characters)
   Could be: SHA-512, BLAKE2b-512
   Specify algorithm with --algo to verify a file
```

For invalid:
```
❌ Invalid hash format
   Expected 64 hex characters for SHA-256, got 32
   This looks like MD5. Use: hashi --algo md5 [hash]
```

**Exit Codes**: 0 for valid, 3 for invalid format

### Mode 2: File + Hash Comparison
**Trigger**: `hashi file.txt [hash_string]` (one file, one hash)

**Purpose**: Compute file hash and compare to provided hash string.

**Output** (default):
```
[SHA-256] Verifying file.txt...
✅ PASS: Hash matches provided string.
```

```
[SHA-256] Verifying file.txt...
🔴 FAIL: Hash mismatch!
   Expected: a1b2c3d4e5f6...
   Computed: f0e1d2c3b4a5...
```

**Output** (with `--bool`):
```
true
```
or
```
false
```

**Exit Codes**: 0 for match, 1 for mismatch, 4 for file not found, 5 for permission denied

### Mode 3: Standard Hash Computation
**Trigger**: Default mode when files are provided without hash strings

**Purpose**: Compute hashes for files and group by matches.

**Output**: See Output Formatters section.


### Mode 4: Dry Run Preview
**Trigger**: `hashi --dry-run [options]`

**Purpose**: Preview what files would be processed without computing hashes.

**Output**:
```
Dry Run Preview
===============
Files to process: 1,234
Total size: 45.2 GB
Estimated time: 2m 34s

Filters applied:
  Include: *.pdf
  Exclude: *.tmp
  Min size: 1 MB
  Modified after: 2026-01-01

Run without --dry-run to process these files.
```

**Exit Code**: Always 0 (no processing errors can occur)

**Use Cases**:
- Verify filters before expensive operations
- Estimate processing time for large datasets
- Preview what incremental mode would process


### Mode 5: Incremental Processing
**Trigger**: `hashi --manifest previous.json --only-changed [options]`

**Purpose**: Process only files that have changed since the last run.

**Change Detection**:
- File not in manifest → changed (new file)
- File in manifest but missing → changed (deleted)
- Size differs → changed
- Modification time differs → changed
- Otherwise → unchanged

**Output**: Same as standard mode, but only for changed files

**Manifest Creation**:
```bash
# First run - create baseline
hashi -a --output-manifest baseline.json

# Subsequent runs - process only changes
hashi -a --manifest baseline.json --only-changed --output-manifest new.json
```

**Use Cases**:
- CI/CD pipelines (only verify changed files)
- Large codebases (dramatically faster incremental runs)
- Backup verification (detect what changed)


## Correctness Properties

Properties are formal statements about system behavior that should hold true across all valid executions. These serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.

### Core Properties

**Property 1: Default behavior processes current directory**
- *For any* directory, when `hashi` is run with no arguments, it should process all non-hidden files in that directory.
- Validates: Requirements 1.1

**Property 2: Color output respects TTY detection**
- *For any* output destination, when output is sent to a non-TTY or `NO_COLOR` is set, the output should contain no ANSI color codes.
- Validates: Requirements 2.2, 2.3

**Property 3: Progress indicators appear for long operations**
- *For any* operation taking longer than 100ms, some progress indication should be displayed before completion.
- Validates: Requirements 2.4, 6.2

**Property 4: Error messages are human-readable**
- *For any* error condition, the error message should not contain raw stack traces or technical jargon in non-verbose mode.
- Validates: Requirements 3.1, 3.6

**Property 5: Similar errors are grouped**
- *For any* set of errors of the same type, they should be grouped together in the output rather than repeated individually.
- Validates: Requirements 3.5

**Property 6: Flags accept any order**
- *For any* valid combination of flags and arguments, changing their order should produce the same result.
- Validates: Requirements 4.4

**Property 7: Configuration precedence is correct**
- *For any* configuration parameter set at multiple levels, the value from the highest precedence source (flags > env > config) should be used.
- Validates: Requirements 5.3

**Property 8: Input validation occurs before processing**
- *For any* invalid input, it should be rejected with an error message before any file processing begins.
- Validates: Requirements 6.1

**Property 9: Operations are idempotent**
- *For any* set of inputs, running the same command twice should produce the same results.
- Validates: Requirements 6.6

**Property 10: JSON output is valid**
- *For any* result set, when `--json` is specified, the output should be parseable as valid JSON.
- Validates: Requirements 7.1

**Property 11: Plain output is line-based**
- *For any* result set, when `--plain` is specified, each result should be on a single line with consistent field separation.
- Validates: Requirements 7.2

**Property 12: Append mode preserves existing content**
- *For any* existing file, when `--append` is used with `--output-file`, the original content should be preserved and new content added at the end.
- Validates: Requirements 8.3

**Property 13: JSON log append maintains validity**
- *For any* existing JSON log file, appending new entries should result in a valid JSON structure.
- Validates: Requirements 8.5

**Property 14: Interrupted operations leave recoverable state**
- *For any* operation interrupted by SIGINT, the system should be in a state where the operation can be safely retried.
- Validates: Requirements 12.4, 12.5

**Property 15: Abbreviated flags are rejected**
- *For any* flag abbreviation that is not explicitly defined, it should be rejected with an error.
- Validates: Requirements 11.5

**Property 16: Exit codes reflect processing status**
- *For any* execution, the exit code should accurately reflect whether processing succeeded, failed, or was interrupted.
- Validates: Requirements 13.1, 13.2, 13.3, 13.4, 13.9

**Property 17: Match-required flag controls exit code**
- *For any* execution with `--match-required`, the exit code should be 0 if and only if matches were found.
- Validates: Requirements 13.5, 13.6

**Property 18: No-match-required flag controls exit code**
- *REMOVED*: The `--no-match-required` flag has been eliminated. Use shell negation instead: `! hashi --match-required`

**Property 19: Default output groups by matches**
- *For any* set of files with matching hashes, the default output should group them together with blank lines between groups.
- Validates: Requirements 2.5

**Property 20: Preserve-order flag maintains input order**
- *For any* set of inputs, when `--preserve-order` is specified, the output order should match the input order exactly.
- Validates: Requirements 2.5

**Property 21: Quiet mode suppresses stdout**
- *For any* execution with `--quiet` flag, no output should be written to stdout (only stderr for errors).
- Validates: Requirements 2.8

**Property 22: ZIP verification uses CRC32 only**
- *For any* ZIP file verification, the algorithm used should always be CRC32 regardless of any metadata in the file suggesting alternatives.
- Validates: Requirements 15.4, 20.3

**Property 23: ZIP verification returns boolean by default**
- *For any* ZIP file passed to hashi without flags, the output should be boolean (exit code only, no stdout) unless `--verbose` is specified.
- Validates: Requirements 15.1, 16.1

**Property 24: Raw flag bypasses special file handling**
- *For any* file with special handling (e.g., ZIP), when `--raw` flag is specified, the file should be treated as raw bytes and hashed normally.
- Validates: Requirements 15.5, 17.2

**Property 25: Mutually exclusive flags are rejected**
- *For any* combination of mutually exclusive flags (e.g., `--json` and `--plain`), the command should be rejected with a clear error message.
- Validates: Requirements 17.1

**Property 26: Multiple ZIP verification returns single boolean**
- *For any* set of ZIP files passed to hashi, the exit code should be 0 only if ALL files pass CRC32 verification.
- Validates: Requirements 15.2

**Property 27: Hash algorithm detection validates hex characters**
- *For any* string, if it contains non-hex characters, DetectHashAlgorithm should return an empty slice.
- Validates: Requirements 21.1

**Property 28: Hash algorithm detection identifies correct algorithms by length**
- *For any* valid hex string, DetectHashAlgorithm should return algorithms matching that length (8→CRC32, 32→MD5, 40→SHA-1, 64→SHA-256, 128→SHA-512/BLAKE2b).
- Validates: Requirements 21.2

**Property 29: Argument classification prioritizes files over hash strings**
- *For any* argument that exists as a file, ClassifyArguments should treat it as a file even if it matches a valid hash format.
- Validates: Requirements 22.1, 22.2

**Property 30: Hash strings are normalized to lowercase**
- *For any* hash string comparison, the comparison should be case-insensitive (normalized to lowercase).
- Validates: Requirements 22.6

**Property 31: Config auto-discovery finds first available config**
- *For any* set of config file locations, FindConfigFile should return the first one that exists in priority order.
- Validates: Requirements 23.1, 23.2

**Property 32: Hash validation mode reports correct algorithms**
- *For any* valid hash string, hash validation mode should report all possible algorithms for that length.
- Validates: Requirements 24.2, 24.4

**Property 33: File+hash comparison returns correct exit codes**
- *For any* file and hash string, comparison mode should exit 0 on match and 1 on mismatch.
- Validates: Requirements 25.2, 25.3

**Property 34: Bool output produces only true/false**
- *For any* comparison with --bool flag, the output should be exactly "true" or "false" with no other text.
- Validates: Requirements 26.1

**Property 35: Config command is rejected with helpful error**
- *For any* invocation of `hashi config` or `hashi config [subcommand]`, the command should be rejected with exit code 3 and a helpful error message explaining manual config file editing.
- Validates: Requirements 27.1, 27.5

**Property 36: Include patterns filter correctly**
- *For any* set of files and include patterns, only files matching at least one include pattern should be processed.
- Validates: Requirements 28.1

**Property 37: Exclude patterns filter correctly**
- *For any* set of files and exclude patterns, files matching any exclude pattern should be skipped.
- Validates: Requirements 28.2

**Property 38: Exclude takes precedence over include**
- *For any* file matching both include and exclude patterns, the file should be excluded.
- Validates: Requirements 28.8

**Property 39: Size filters work correctly**
- *For any* file and size limits (min/max), the file should be processed only if its size is within the specified range.
- Validates: Requirements 28.3, 28.4

**Property 40: Date filters work correctly**
- *For any* file and date limits (before/after), the file should be processed only if its modification time is within the specified range.
- Validates: Requirements 28.5, 28.6

**Property 41: Multiple filters combine with AND logic**
- *For any* file and multiple filter criteria, the file should be processed only if it passes ALL filters.
- Validates: Requirements 28.7

**Property 42: Dry run enumerates without hashing**
- *For any* set of files with --dry-run flag, all files should be enumerated but no hashes should be computed.
- Validates: Requirements 29.1

**Property 43: Dry run shows accurate counts**
- *For any* set of files with --dry-run flag, the displayed file count and total size should match the files that would be processed.
- Validates: Requirements 29.2, 29.3

**Property 44: Dry run applies filters**
- *For any* set of files with --dry-run and filter flags, the preview should show only files that pass the filters.
- Validates: Requirements 29.7

**Property 45: Manifest detects changed files**
- *For any* file in a manifest, if the file's size or modification time differs from the manifest, it should be detected as changed.
- Validates: Requirements 30.3

**Property 46: Only-changed processes changed files only**
- *For any* set of files with --only-changed flag and a manifest, only files that are new, modified, or missing should be processed.
- Validates: Requirements 30.2

**Property 47: Manifest format is valid JSON**
- *For any* manifest file created by hashi, the file should be valid JSON parseable by standard JSON parsers.
- Validates: Requirements 30.7

**Property 48: Atomic writes preserve original on failure**
- *For any* file write operation that fails, the original file (if it existed) should remain unchanged.
- Validates: Requirements 31.7

**Property 49: Append mode preserves existing content**
- *For any* existing file with --append flag, the original content should be preserved and new content added at the end.
- Validates: Requirements 31.3

**Property 50: JSON log append maintains validity**
- *For any* existing JSON log file, appending new entries should result in a valid JSON array structure.
- Validates: Requirements 31.6


## Testing Strategy

### Unit Tests

Unit tests verify specific behaviors and edge cases for each component:

**internal/config/**
- Flag parsing (short and long versions)
- Argument validation
- Configuration precedence
- Environment variable loading
- .env file parsing

**internal/color/**
- TTY detection logic
- NO_COLOR environment variable
- Color code generation
- Individual stdout/stderr checking

**internal/progress/**
- Progress calculation
- ETA calculation
- TTY detection for display
- Update and finish behavior

**internal/errors/**
- Error message formatting
- Suggestion generation
- Error grouping logic
- Path sanitization

**internal/hash/**
- Hash computation for various algorithms
- Streaming file processing
- Large file handling
- Error cases (permission denied, file not found)

**internal/output/**
- Each formatter with sample data
- Edge cases (empty results, single file, many matches)
- Output format correctness
- Color integration

**internal/archive/**
- ZIP CRC32 verification
- Failed entry detection
- --raw flag behavior
- Multiple file verification

**internal/conflict/**
- Each known conflict pair
- Valid flag combinations
- Error message quality

**internal/filter/**
- Pattern matching (glob patterns)
- Size filtering (min/max)
- Date filtering (before/after)
- Multiple filter combination (AND logic)
- Exclude precedence over include

**internal/dryrun/**
- File enumeration without hashing
- Size calculation and aggregation
- Time estimation accuracy
- Filter application in preview

**internal/manifest/**
- Manifest loading and saving
- JSON format validation
- Change detection logic
- File metadata tracking

**internal/fileout/**
- Atomic write operations
- Overwrite protection
- Append mode
- JSON log append with validity
- Error handling and rollback

**internal/signals/**
- SIGINT handling
- Cleanup timeout
- Double Ctrl-C behavior

### Property-Based Tests

Property-based tests verify universal properties across many randomly generated inputs (minimum 100 iterations each):

1. **Default Behavior**: Generate random directory contents, verify all non-hidden files processed
2. **Color Output**: Generate random outputs, verify color codes present/absent based on TTY
3. **Progress Indicators**: Generate operations of varying duration, verify progress shown for >100ms
4. **Error Grouping**: Generate multiple similar errors, verify grouping
5. **Flag Order Independence**: Generate random flag orderings, verify same results
6. **Config Precedence**: Generate conflicting configs at different levels, verify correct precedence
7. **Idempotence**: Generate random inputs, verify running twice gives same results
8. **JSON Validity**: Generate random results, verify JSON output is valid
9. **Plain Output Format**: Generate random results, verify line-based format
10. **Append Mode**: Generate random content, verify append preserves original
11. **Signal Handling**: Interrupt at random points, verify recoverable state
12. **Exit Codes**: Generate various scenarios, verify correct exit codes
13. **Match-Required**: Generate random file sets, verify exit code logic
14. **No-Match-Required**: Generate random file sets, verify exit code logic
15. **Abbreviated Flags**: Generate random abbreviations, verify rejection
16. **Default Output Grouping**: Generate random file sets, verify grouping by hash
17. **Preserve Order**: Generate random inputs, verify order maintained
18. **Quiet Mode**: Generate random executions, verify stdout suppression
19. **ZIP CRC32**: Generate random ZIP files, verify CRC32 usage
20. **Raw Flag**: Generate random special files, verify raw byte handling
21. **Mutually Exclusive Flags**: Generate random flag combinations, verify conflict detection
22. **Include Patterns**: Generate random file sets and patterns, verify correct filtering
23. **Exclude Patterns**: Generate random file sets and patterns, verify correct exclusion
24. **Size Filters**: Generate files of various sizes, verify min/max filtering
25. **Date Filters**: Generate files with various modification times, verify date filtering
26. **Filter Combination**: Generate random filter combinations, verify AND logic
27. **Dry Run Accuracy**: Generate random file sets, verify count and size calculations
28. **Manifest Change Detection**: Generate files with various changes, verify detection
29. **Incremental Processing**: Generate file sets with changes, verify only changed files processed
30. **Atomic Writes**: Simulate write failures, verify original file preservation
31. **JSON Log Append**: Generate random JSON entries, verify array validity after append

### Integration Tests

Integration tests verify end-to-end workflows:

1. **Basic Hash Comparison**: Hash files and compare results
2. **Recursive Directory Processing**: Process directory trees with various structures
3. **Output to File**: Write results to files with various formats
4. **Config File Usage**: Load and apply configuration from files
5. **Error Recovery**: Handle various error conditions gracefully
6. **Signal Interruption**: Interrupt and resume operations
7. **ZIP Verification**: Verify ZIP files with various states (valid, corrupted)
8. **Quiet Mode Scripting**: Use in scripts with exit code checking
9. **Multiple Output Formats**: Test all formatters in realistic scenarios
10. **Flag Conflicts**: Test all known conflict scenarios
11. **Advanced Filtering**: Test complex filter combinations with real file trees
12. **Dry Run Mode**: Test preview accuracy with various file sets
13. **Incremental Operations**: Test full incremental workflow (baseline → changes → update)
14. **File Output Safety**: Test atomic writes, overwrite protection, append modes
15. **Manifest Workflows**: Test CI/CD-style incremental verification workflows


## Security Considerations

### Security Philosophy

Hashi is designed as a **read-only information discovery tool**. This fundamental principle guides all security decisions:

1. **No File Modification**: Hashi never modifies input files
2. **No File Deletion**: Hashi never deletes files
3. **No System State Changes**: Hashi doesn't modify system configuration
4. **Limited Write Operations**: Only writes to explicitly specified output/log files
5. **No Network Operations**: No external network calls (if URL support is added, it must be explicit)

### Threat Model

**What we protect against**:
1. **Accidental Data Loss**: Prevent overwriting important files without confirmation
2. **Path Traversal**: Validate file paths to prevent directory traversal attacks
3. **Resource Exhaustion**: Limit memory usage, handle large files via streaming
4. **Information Disclosure**: Don't leak sensitive paths in error messages
5. **Command Injection**: Sanitize all inputs, no shell execution

**What we don't protect against** (out of scope):
1. Malicious input files (we read them, but don't execute)
2. Compromised hash algorithms (we use standard libraries)
3. Physical access to the machine
4. Side-channel attacks on hash computation

### Security Features

#### 1. File Output Safety
- Check if file exists before writing
- Prompt for confirmation unless `--force`
- Support append mode with `--append`
- Use atomic writes (write to temp file, then rename)

#### 2. Path Validation
- Validate all file paths before processing
- Resolve to absolute paths safely
- Prevent directory traversal attacks
- Check for suspicious patterns

#### 3. Resource Limits
- Stream large files (never load entire file into memory)
- Set reasonable buffer sizes (100MB max for in-memory operations)
- Handle files of any size efficiently

#### 4. Error Message Sanitization
- Replace absolute paths with relative paths where appropriate
- Replace home directory with `~` in error messages
- Don't leak sensitive information in error output
- Provide debug mode for detailed errors when needed

#### 5. Input Validation
- Validate all inputs before processing
- Check hash format validity
- Validate size limits
- Validate date formats
- Fail fast with clear error messages

### Integrity vs Authenticity

**Critical User Education**:

When verifying ZIP files, hashi checks **integrity** (are the bits correct?) using CRC32. This does NOT verify **authenticity** (was the file tampered with?).

- **Integrity**: Confirms data wasn't corrupted during storage/transmission
- **Authenticity**: Confirms data wasn't tampered with by a malicious actor

A malicious actor can create a file with valid CRC32 checksums. CRC32 verification only confirms the file wasn't corrupted - not that it's safe.

For authenticity verification, use cryptographic signatures (GPG, etc.) in addition to hash verification.

### Security Hardening

**ZIP Verification**:
- Always use CRC32 regardless of metadata (prevents algorithm substitution attacks)
- Malicious ZIP metadata could suggest weaker algorithms - we ignore it
- Document clearly that CRC32 is for integrity, not authenticity

**Future Considerations**:
If features are added:
- **URL Support**: Validate URLs, implement timeouts, limit download sizes, HTTPS only
- **Config File Execution**: Never execute code from config files, only parse data
- **Plugin System**: Sandbox plugins, verify signatures, limit capabilities


## Implementation Status

### Completed (Tasks 1-10)

✅ **Task 1: Project Structure**
- Fresh start with modular architecture
- Package structure: cmd/hashi, internal/{config,hash,output,color,progress,errors,signals,archive,conflict}
- Dependencies added: fatih/color, schollz/progressbar, spf13/pflag, joho/godotenv, golang.org/x/term
- Testing framework with property-based testing support

✅ **Task 2: TTY Detection and Color Output**
- ColorHandler with TTY detection
- NO_COLOR environment variable support
- Semantic color methods (Success, Error, Warning, Info, Path)
- Property and unit tests complete

✅ **Task 3: Enhanced Argument Parser**
- Config struct with all new fields
- ParseArgs function with short and long flags
- Support for `-` stdin
- Support for `--flag=value` and `--flag value` syntax
- Property and unit tests complete

✅ **Task 4: Environment Variable and Configuration System**
- EnvConfig struct and LoadEnvConfig function
- .env file support
- Config file loading (JSON/text)
- Configuration precedence (flags > env > config)
- Property and unit tests complete

✅ **Task 5: Checkpoint - All Tests Pass**
- All tests passing for tasks 1-4

✅ **Task 6: Output Formatters**
- OutputFormatter interface
- DefaultFormatter (grouped by matches)
- PreserveOrderFormatter (input order maintained)
- VerboseFormatter (detailed with summaries)
- JSONFormatter (machine-readable)
- PlainFormatter (tab-separated for scripting)
- Property and unit tests complete

✅ **Task 7: Progress Indicators**
- ProgressBar component
- Shows progress for operations >100ms
- Displays percentage, count, and ETA
- Hides when output is not a TTY
- Property and unit tests complete

✅ **Task 8: Enhanced Error Handling**
- ErrorHandler component
- User-friendly error messages with suggestions
- Error grouping to reduce noise
- Path sanitization
- Property and unit tests complete

✅ **Task 9: Checkpoint - All Tests Pass**
- All tests passing for tasks 1-8

✅ **Task 10: Signal Handling**
- SignalHandler component
- Graceful Ctrl-C handling
- Double Ctrl-C for force quit
- Property and unit tests complete

### Next Up (Tasks 11-18) - New Features from Checkpoint Update

The following tasks implement hash detection, argument classification, config auto-discovery, and new operation modes:

- **Task 11**: Hash algorithm detection (DetectHashAlgorithm function)
- **Task 12**: Argument classification (files vs hash strings)
- **Task 13**: Config file auto-discovery
- **Task 14**: Checkpoint
- **Task 15**: Hash string validation mode (`hashi [hash]`)
- **Task 16**: File + hash comparison mode (`hashi file.txt [hash]`)
- **Task 17**: Boolean output flag (`--bool`)
- **Task 18**: Checkpoint

### Remaining (Tasks 19-38)

The following tasks are documented in `tasks.md` and ready for implementation:

- Task 19: File output manager with safety features
- Task 20: Exit code handler with scripting support
- Task 21: Quiet mode for boolean scripting
- Task 22: Checkpoint
- Task 23: Security features
- Task 24: Default behavior (no arguments = current directory)
- Task 25: Enhanced help system
- Task 26: Command suggestion for common mistakes
- Task 27: Checkpoint
- Task 28: Idempotence and recovery
- Task 29: Flag abbreviation rejection
- Task 30: Integration and wiring
- Task 31: Final checkpoint
- Task 32: Documentation and polish
- Task 33: Archive integrity verification (ZIP CRC32)
- Task 34: Flag conflict detection
- Task 35: Checkpoint
- Task 36: Educational code quality (ongoing)
- Task 37: Conflict testing infrastructure (moonshot)
- Task 38: Final checkpoint - All features complete
- Task 26: Flag conflict detection
- Task 27: Checkpoint
- Task 28: Educational code quality (ongoing)
- Task 29: Conflict testing infrastructure (moonshot)
- Task 30: Final checkpoint


## Educational Code Quality (Moonshot Goal)

### Philosophy

This project serves dual purposes:
1. A functional, production-quality CLI tool
2. A teaching resource for learning Go and CLI design

### Current Standards

**Every exported function has**:
- Doc comment explaining purpose and usage
- Parameter descriptions
- Return value descriptions
- Example usage where helpful

**Complex algorithms have**:
- Step-by-step explanations
- Comments explaining "why" not just "what"
- References to relevant documentation

**Go idioms are explained**:
- When first introduced in the codebase
- With links to Go documentation
- With rationale for the pattern choice

### Example: Well-Commented Code

```go
// ComputeFileHash reads a file and computes its cryptographic hash.
//
// This function uses streaming to handle files of any size without
// loading the entire file into memory. The io.Copy function reads
// chunks from the file and writes them to the hasher incrementally.
//
// Parameters:
//   - filepath: Path to the file to hash (can be relative or absolute)
//   - algorithm: Hash algorithm to use ("sha256", "md5", etc.)
//
// Returns:
//   - The hash as a lowercase hexadecimal string
//   - An error if the file cannot be read or doesn't exist
//
// Example:
//   hash, err := ComputeFileHash("document.pdf", "sha256")
//   // hash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
func ComputeFileHash(filepath string, algorithm string) (string, error) {
    // Open the file for reading
    // os.Open returns a file handle and an error
    // The file handle implements io.Reader, which we'll use for streaming
    file, err := os.Open(filepath)
    if err != nil {
        // Wrap the error with context so the caller knows what failed
        return "", fmt.Errorf("cannot open file %s: %w", filepath, err)
    }
    // defer ensures this runs when the function exits, even if we return early
    // This prevents resource leaks (forgetting to close files)
    defer file.Close()
    
    // Create a hasher based on the requested algorithm
    hasher := createHasher(algorithm)
    
    // Stream the file contents through the hasher
    // io.Copy reads from file in chunks and writes to hasher
    // This is memory-efficient: we never hold the whole file in RAM
    _, err = io.Copy(hasher, file)
    if err != nil {
        return "", fmt.Errorf("error reading file %s: %w", filepath, err)
    }
    
    // Get the final hash value as bytes, then convert to hex string
    // Sum(nil) finalizes the hash and returns the digest
    hashBytes := hasher.Sum(nil)
    return hex.EncodeToString(hashBytes), nil
}
```

### Moonshot: Annotated Edition

**Future Goal**: Maintain a separate branch or documentation with exhaustive comments suitable for complete beginners:
- Comments explaining every Go keyword and construct
- Cross-references to Go documentation and tutorials
- Explanations of design patterns and why they're used
- Step-by-step walkthroughs of complex logic


## Conflict Testing Infrastructure (Moonshot Goal)

### Pre-Implementation Conflict Documentation

Before coding begins, document all known flag conflicts in plain English:

**Output Format Conflicts**:
- `--json` vs `--plain`: Mutually exclusive, both specify output format
- `--json` vs `--verbose`: Mutually exclusive, JSON includes all details
- `--quiet` vs `--verbose`: Mutually exclusive, opposite verbosity levels

**File Handling Conflicts**:
- `--raw` vs default ZIP behavior: `--raw` overrides automatic CRC32 verification
- `--config file.json` vs `hashi file.json`: Explicit flag needed for config interpretation

**Match Requirement Conflicts**:
- `--match-required` vs `--no-match-required`: Mutually exclusive, opposite requirements

### Conflict Review Process

When adding or changing a flag:
1. List all existing flags
2. For each existing flag, ask: "Does this new flag interact with it?"
3. Document any interactions (conflicts, dependencies, overrides)
4. Add test cases for the interactions
5. Update help text if needed

### Conflict Testing Strategy

**Phase 1: Manual Documentation** (before coding)
- Document all anticipated conflicts
- Make design decisions for each
- Record in plain English in this spec

**Phase 2: Targeted Testing** (during development)
- Write tests for documented conflicts
- Test common flag combinations
- Test edge cases identified during review

**Phase 3: Fuzzing Tool** (moonshot, post-release)
- Build a tool to fuzz hashi
- Generate random flag combinations
- Generate random file inputs (various types, sizes, permissions)
- Record all outputs and exit codes
- Flag unexpected behaviors for review

### Fuzzing Tool Design (Moonshot)

```go
// FuzzConfig defines what to randomize
type FuzzConfig struct {
    MaxFlags int           // Max flags per invocation
    FileTypes []string     // File extensions to generate
    FileSizes []int64      // File sizes to test
    Iterations int         // Number of random tests
}

// FuzzResult captures one test run
type FuzzResult struct {
    Args []string          // Command line arguments used
    Files []string         // Test files created
    ExitCode int           // Actual exit code
    Stdout string          // Captured stdout
    Stderr string          // Captured stderr
    Duration time.Duration // How long it took
    Unexpected bool        // Did something unexpected happen?
    Notes string           // Why it was flagged
}

// Fuzzer generates and runs random test cases
type Fuzzer struct {
    config FuzzConfig
    results []FuzzResult
}

func (f *Fuzzer) Run() error {
    for i := 0; i < f.config.Iterations; i++ {
        // Generate random flags
        flags := f.randomFlags()
        
        // Generate random test files
        files := f.generateTestFiles()
        
        // Run hashi with these inputs
        result := f.execute(flags, files)
        
        // Check for unexpected behavior
        f.analyze(result)
        
        // Clean up test files
        f.cleanup(files)
        
        f.results = append(f.results, result)
    }
    return f.report()
}
```

**What "unexpected" means**:
- Crash or panic
- Hang (timeout exceeded)
- Exit code doesn't match documented behavior
- Output format doesn't match requested format
- Conflicting flags accepted without error
- Same inputs produce different outputs (non-determinism)


## Design Decisions and Rationale

### Why Modular Architecture?

**Decision**: Split functionality into focused internal packages rather than a monolithic main.go.

**Rationale**:
- **Testability**: Each component can be tested in isolation
- **Maintainability**: Clear boundaries make it easier to understand and modify
- **Reusability**: Components can be reused in other contexts
- **Developer Continuity**: New developers can understand one component at a time
- **Educational Value**: Each package demonstrates a specific design pattern

### Why Fresh Start Instead of Refactor?

**Decision**: Complete rewrite rather than refactoring existing code.

**Rationale**:
- New architecture is fundamentally different (modular vs monolithic)
- New components don't exist in old code (color, progress, conflict detection)
- Behavior changes are significant (default grouping, boolean output for ZIP)
- Attempting to refactor would create more confusion than starting clean
- Old code backed up in separate repository for reference

### Why Default Grouping by Hash?

**Decision**: Default output groups files by matching hash with blank lines between groups.

**Rationale**:
- **Human-First**: Makes duplicates immediately visible
- **Still Pipeable**: One file per line, consistent format
- **Escape Hatch**: `--preserve-order` flag for users who need input order
- **Follows Guidelines**: CLI guidelines prioritize human readability with machine-readable alternatives

### Why Boolean Output for ZIP Verification?

**Decision**: ZIP verification returns exit code only (no stdout) by default.

**Rationale**:
- **Script-Friendly**: Perfect for boolean checks in automation
- **Follows Convention**: Similar to tools like `test`, `grep -q`
- **Escape Hatches**: `--verbose` and `--json` for detailed output
- **Security**: Reduces attack surface compared to complex default output

### Why CRC32 Only for ZIP?

**Decision**: Always use CRC32 for ZIP verification regardless of metadata.

**Rationale**:
- **Security Hardening**: Prevents algorithm substitution attacks
- **Predictability**: Users know exactly what algorithm is used
- **Standard Compliance**: CRC32 is the ZIP standard
- **Clear Documentation**: Integrity vs authenticity distinction is explicit

### Why Quiet Mode?

**Decision**: Add `--quiet` flag that suppresses all stdout.

**Rationale**:
- **Scripting**: Enables clean boolean checks with exit codes
- **Follows Guidelines**: Standard pattern in CLI tools
- **Errors Still Visible**: stderr still works for debugging
- **Composability**: Works with all other flags and features

### Why Configuration Precedence?

**Decision**: Flags > Environment > Project Config > User Config > System Config

**Rationale**:
- **Follows Standards**: XDG spec and common CLI conventions
- **Predictability**: Users can override at any level
- **Flexibility**: Different contexts (system, user, project, invocation)
- **Explicit Wins**: Most explicit (flags) takes precedence

### Why Property-Based Testing?

**Decision**: Include property-based tests alongside unit tests.

**Rationale**:
- **Broader Coverage**: Tests properties across many inputs
- **Edge Case Discovery**: Finds cases developers didn't think of
- **Specification**: Properties serve as formal specifications
- **Confidence**: Higher confidence in correctness

### Why Educational Comments?

**Decision**: Extensive comments explaining "why" not just "what".

**Rationale**:
- **Developer Continuity**: Any developer can pick up the project
- **Teaching Tool**: Serves as learning resource for Go and CLI design
- **Maintenance**: Future maintainers understand design decisions
- **Onboarding**: New contributors can learn from the code itself

### Why Advanced Filtering?

**Decision**: Support pattern matching, size filtering, and date filtering.

**Rationale**:
- **User Demand**: Common request from use case analysis
- **Composability**: Filters combine naturally with other features
- **Performance**: Reduces unnecessary processing
- **Flexibility**: Users can target specific file subsets without external tools

### Why Dry Run Mode?

**Decision**: Add --dry-run flag for preview without processing.

**Rationale**:
- **Risk Reduction**: Users can verify filters before expensive operations
- **Time Estimation**: Helps users plan for long-running operations
- **Confidence**: Users know exactly what will be processed
- **Follows Guidelines**: Common pattern in CLI tools (rsync, rm, etc.)

### Why Incremental Operations?

**Decision**: Support manifest-based incremental processing.

**Rationale**:
- **CI/CD Value**: Dramatically reduces processing time in automated pipelines
- **Large Codebases**: Makes hashi practical for huge repositories
- **User Demand**: High-priority request from use case analysis
- **Efficiency**: Only process what changed, not everything

### Why Manifest Format?

**Decision**: Use JSON for manifest files with metadata.

**Rationale**:
- **Human-Readable**: Easy to inspect and debug
- **Machine-Parseable**: Standard format, many tools support it
- **Extensible**: Can add fields without breaking compatibility
- **Portable**: Works across platforms and languages

### Why Atomic File Writes?

**Decision**: Use temp file + rename pattern for all file writes.

**Rationale**:
- **Data Safety**: Never corrupt existing files on failure
- **Atomic Operation**: Rename is atomic on most file systems
- **Follows Best Practice**: Standard pattern for safe file writing
- **User Trust**: Users can rely on hashi not losing data


## Future Considerations

### Potential Features (Not Currently Planned)

**URL Support**:
- Download files from URLs and hash them
- Security considerations: HTTPS only, timeouts, size limits
- Would require explicit flag (e.g., `--url`)

**Additional Archive Formats**:
- TAR (no built-in checksums, would need external verification)
- RAR (proprietary format, licensing issues)
- 7z (CRC32 or SHA-256 depending on settings)

**Parallel Processing**:
- Process multiple files concurrently
- Challenges: Progress bar complexity, output ordering
- Benefits: Faster processing for many small files

**Watch Mode**:
- Monitor directory for changes and re-hash
- Use case: Continuous verification
- Challenges: Signal handling, long-running process

**Plugin System**:
- Allow custom hash algorithms or output formats
- Security: Sandboxing, signature verification
- Complexity: API stability, documentation

### Deprecation Strategy

If breaking changes are needed:

1. **Announce Early**: Warn in release notes and documentation
2. **Deprecation Period**: Minimum 6 months with warnings
3. **Migration Path**: Provide clear upgrade instructions
4. **Compatibility Mode**: Consider compatibility flags if feasible
5. **Version Bump**: Follow semantic versioning (major version for breaking changes)

### Maintenance Priorities

**High Priority**:
- Security updates for dependencies
- Bug fixes affecting correctness
- Performance improvements for common use cases
- Documentation improvements

**Medium Priority**:
- New output formats (if requested by users)
- Additional filtering options
- Enhanced error messages

**Low Priority**:
- Additional archive formats
- Cosmetic improvements
- Experimental features

**Features to be Evaluated**
- `hashishi`, a customizable wrapper used to enable advanced operations like --diff, which was removed from the feature list due to complexity
- Custom JSONs (via env vars, config, or hashishi wrapper)


## References

### CLI Guidelines
- [CLI Guidelines](https://clig.dev/) - Industry-standard CLI design guidelines
- [12 Factor CLI Apps](https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46) - Jeff Dickey
- [POSIX Utility Conventions](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap12.html)
- [GNU Coding Standards](https://www.gnu.org/prep/standards/html_node/Program-Behavior.html)

### Go Programming
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

### Security
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/)

### Testing
- [Property-Based Testing](https://hypothesis.works/articles/what-is-property-based-testing/)
- [Go Testing Best Practices](https://github.com/golang/go/wiki/TestComments)

### Project Documentation
- `requirements.md` - Detailed requirements specification
- `tasks.md` - Implementation task breakdown
- `CLI_guidelines.md` - Full CLI guidelines reference
- `CONTRIBUTING.md` - Contribution guidelines and principles
- `README.md` - Project overview and quick start

## Conclusion

This design document describes a complete rewrite of hashi following industry-standard CLI design guidelines. The modular architecture, human-first design, and comprehensive testing strategy ensure a robust, maintainable, and delightful command-line tool.

**Key Achievements**:
- ✅ Modular architecture with clear separation of concerns
- ✅ Human-first design with machine-readable alternatives
- ✅ Comprehensive error handling with actionable suggestions
- ✅ Multiple output formats (default, verbose, JSON, plain)
- ✅ TTY detection and color handling
- ✅ Progress indicators for long operations
- ✅ Configuration system with proper precedence
- ✅ Property-based testing for correctness guarantees

**Current Status**: Tasks 1-8 complete (core infrastructure and error handling)

**Next Steps**: Continue with signal handling, file output safety, and exit code system (tasks 9-13)

**Long-Term Vision**: A production-quality CLI tool that also serves as an educational resource for learning Go and CLI design principles.

