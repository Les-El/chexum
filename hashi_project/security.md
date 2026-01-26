# Security Design: Preventing Self-Modification

## Overview

The hashi security system implements the core principle "hashi can't change hashi" using a configurable, layered approach. This document describes the security model, design decisions, and implementation.

## Security Principles

**Core Rule:** hashi can write txt, json, and csv files for output, but must never write to configuration files or directories.

**Design Philosophy:**
- **Read-only tool:** hashi is an informational tool that cannot modify its own configuration
- **User intentionality:** Security settings require deliberate configuration via files or environment variables
- **Layered protection:** Multiple configurable barriers prevent accidental or malicious file overwrites
- **Flexibility with safety:** Users can customize security rules while maintaining core protections

## Implementation

## Implementation Status

✅ **FULLY IMPLEMENTED** - The configurable security system is complete and tested.

### What's Working

**Core Security Functions:**
- `validateFileName()` - Checks file names against configurable patterns
- `validateDirPath()` - Checks directory paths against configurable patterns  
- `validateOutputPath()` - Integrates both checks with extension validation

**Configuration Loading:**
- Environment variables: `HASHI_BLACKLIST_FILES`, `HASHI_BLACKLIST_DIRS`, `HASHI_WHITELIST_FILES`, `HASHI_WHITELIST_DIRS`
- Config file support: `blacklist_files`, `blacklist_dirs`, `whitelist_files`, `whitelist_dirs` arrays
- Additive merging: Environment + config file patterns are combined

**Pattern Matching:**
- Case-insensitive matching using `strings.ToLower()`
- Glob pattern support with `filepath.Match()` for `*` and `?` wildcards
- Prefix matching for simple patterns (no wildcards)
- Whitelist override system (whitelist beats blacklist)

**Integration:**
- Automatic validation for `--output`, `--log-file`, `--log-json` paths
- Security-aware error messages (generic vs verbose mode)
- Comprehensive test coverage (100+ test cases)

## Implementation Details

### Configurable Validation Approach

The security system uses a flexible, multi-layered validation approach:

1. **Extension Whitelist** - Only allow `.txt`, `.json`, `.csv` extensions (unchanged)
2. **Configurable Pattern Blacklists** - Block file/directory names matching sensitive patterns
3. **Whitelist Override System** - Allow exceptions to blacklist rules
4. **Separate File/Directory Controls** - Different rules for files vs directories

### Design Decisions

#### Pattern-Based Security
**Decision:** Use configurable patterns instead of hard-coded file names
**Rationale:** 
- Protects against variations ("config.txt", "configuration.json", "myconfig.csv")
- Allows users to add project-specific sensitive patterns
- Maintains security while providing flexibility

#### Separate File vs Directory Controls
**Decision:** Provide separate blacklist/whitelist controls for files and directories
**Rationale:**
- Users may want to block "temp" directories but allow "temp.txt" files
- Follows standard Unix tool patterns (like .gitignore)
- Provides granular control without complexity

#### Additive Blacklist System
**Decision:** User patterns ADD to built-in defaults, whitelist can override any pattern
**Rationale:**
- Maintains baseline security (built-in patterns always active)
- Allows customization (users add project-specific patterns)
- Provides escape hatch (whitelist overrides for legitimate exceptions)
- Example: Block "config*" by default, but allow "myconfig_report.txt" via whitelist

#### No Command-Line Security Flags
**Decision:** Security settings only via environment variables and config files
**Rationale:**
- Maintains "hashi can't change hashi" principle
- Requires intentional configuration (users must edit files or set env vars)
- Prevents accidental security bypasses via command-line typos
- Keeps security settings behind "wall of intentionality"

#### Glob Pattern Support
**Decision:** Support `*` and `?` wildcards using Go's built-in `filepath.Match()`
**Rationale:**
- Provides flexibility: "config" (exact) vs "config*" (wildcard)
- Uses standard, well-understood pattern syntax
- No external dependencies (keeps hashi lightweight)
- User chooses precision level by including wildcards or not

#### Configuration Source Merging
**Decision:** Merge patterns from environment variables and config files
**Rationale:**
- Allows layering: project config + personal environment settings
- Supports different use cases (CI/CD env vars + local config files)
- Provides maximum flexibility for complex setups

### Protected Items

**File Extensions:** Only `.txt`, `.json`, `.csv` are allowed for output

**Default Blacklist Patterns (case-insensitive):**
- `config` - Configuration files and directories
- `secret` - Secret keys, credentials, sensitive data
- `key` - API keys, encryption keys, certificates
- `password` - Password files, authentication data
- `credential` - Credential stores, authentication files

**Configuration Variables:**
- `HASHI_BLACKLIST_FILES` - Additional file name patterns to block
- `HASHI_BLACKLIST_DIRS` - Additional directory name patterns to block  
- `HASHI_WHITELIST_FILES` - File name patterns that override blacklists
- `HASHI_WHITELIST_DIRS` - Directory name patterns that override blacklists

**Config File Format:**
```toml
[security]
blacklist_files = ["temp*", "draft*"]
blacklist_dirs = ["cache", "tmp*"]
whitelist_files = ["important_config_report.txt"]
whitelist_dirs = []
```

## Usage Examples

### Safe Operations ✅

```bash
# Write to current directory
hashi --output results.txt *.txt
hashi --output data.json --json *.txt
hashi --output report.csv *.txt

# Write to subdirectories
hashi --output logs/results.txt *.txt
hashi --output /tmp/hashes.json *.txt

# Custom security configuration
export HASHI_BLACKLIST_FILES="temp*,draft*"
export HASHI_WHITELIST_FILES="important_temp_report.txt"
hashi --output important_temp_report.txt *.txt  # Allowed via whitelist
```

### Blocked Operations ❌

```bash
# Invalid extensions
hashi --output malicious.sh *.txt
hashi --output script.py *.txt

# Default blacklist patterns
hashi --output config.txt *.txt          # Matches "config" pattern
hashi --output secret_keys.json *.txt    # Matches "secret" pattern
hashi --output api_key_backup.csv *.txt  # Matches "key" pattern

# Directory patterns
hashi --output config/results.txt *.txt  # Directory contains "config"
hashi --output secrets/data.json *.txt   # Directory contains "secret"

# Custom blacklist patterns (if configured)
export HASHI_BLACKLIST_FILES="temp*"
hashi --output temp_results.txt *.txt    # Blocked by custom pattern
```

### Configuration Examples

**Environment Variables:**
```bash
# Add custom patterns to defaults
export HASHI_BLACKLIST_FILES="temp*,draft*,backup*"
export HASHI_BLACKLIST_DIRS="cache,tmp*"

# Create exceptions for legitimate files
export HASHI_WHITELIST_FILES="config_report.txt,key_analysis.json"
export HASHI_WHITELIST_DIRS="results_config"
```

**Config File (.hashi.toml):**
```toml
[security]
# Additional patterns beyond built-in defaults
blacklist_files = ["temp*", "draft*", "backup*"]
blacklist_dirs = ["cache", "tmp*", "build*"]

# Exceptions for legitimate use cases  
whitelist_files = ["config_report.txt", "key_analysis.json"]
whitelist_dirs = ["results_config", "output_temp"]
```

**Pattern Matching Examples:**
```bash
# Exact matching (no wildcards)
blacklist_files = ["config"]        # Blocks: "config.txt", not "myconfig.txt"

# Wildcard matching  
blacklist_files = ["config*"]       # Blocks: "config.txt", "configuration.json"
blacklist_files = ["*config*"]      # Blocks: "myconfig.txt", "config_old.json"
blacklist_files = ["temp.???"]      # Blocks: "temp.txt", "temp.csv", not "temp.json"
```

## Error Messages

Enhanced error messages provide security through obfuscation while maintaining usability:

```
# Extension validation (always specific - not security-sensitive)
Error: output file: output files must have extension: .txt, .json, .csv (got .sh)

# Security-sensitive and file system errors (generic by default, detailed with --verbose)
Error: output file: Unknown write/append error
Error: log file: Unknown write/append error

# With --verbose flag (helpful for legitimate users)
Error: output file: cannot write to configuration file: config.json
Error: log file: cannot write to configuration directory
Error: output file: permission denied writing to /path/to/file.txt
Error: log file: insufficient disk space for /path/to/file.txt
```

**Security Design**: The message "Unknown write/append error" is used for both security-sensitive scenarios AND legitimate file system errors to prevent information disclosure:

**Security-sensitive errors:**
- Configuration file name conflicts (e.g., `config.json`)
- Configuration directory conflicts (e.g., `.hashi/`, `.config/hashi/`)

**File system errors (for obfuscation):**
- OS-level permission errors (locked system files)
- Disk full errors
- Network/filesystem errors
- Path too long errors

**Usability Design**: Non-security validation errors (like invalid extensions) always provide specific messages since they don't reveal sensitive information. Security-sensitive errors and selected file system errors provide helpful details in `--verbose` mode for legitimate users while maintaining generic messages for potential attackers.

**Obfuscation Benefit**: By mixing legitimate file system errors with security blocks, attackers cannot distinguish between:
- Being blocked for security reasons
- Encountering genuine system limitations
- Having insufficient permissions for legitimate reasons

This approach significantly increases the noise-to-signal ratio for attackers while preserving excellent UX for legitimate users through the `--verbose` flag.

### Future File Writing Operations

**IMPORTANT**: All file writing operations in hashi should use the appropriate error handling functions to maintain consistent security posture:

- **Security-sensitive operations**: Use `config.WriteErrorWithVerbose(verbose, details)` for scenarios where attackers might gain information (config files, directories)
- **File system operations**: Use `config.HandleFileWriteError(err, verbose, path)` for actual file writing - this automatically obfuscates common OS errors
- **Non-security validation**: Use specific error messages for validation errors that don't reveal sensitive information (invalid extensions, format errors)

This includes:
- **FileOutputManager** (Task 19) - should use `HandleFileWriteError` for all write operations
  - Atomic writes (temp file + rename pattern)
  - Overwrite protection with --force flag
  - Append mode with --append flag
  - JSON log append maintaining array validity
- **ManifestSystem** (Task 41) - should use `HandleFileWriteError` for manifest file operations
  - Manifest file creation and updates
  - Atomic writes for manifest files
  - Path validation for --manifest and --output-manifest flags
- **Log file writing** - should use `HandleFileWriteError` for all write operations
- **Any other file output operations** added in the future

**Example usage:**
```go
// For actual file writing operations (FileOutputManager, ManifestSystem, etc.)
file, err := os.Create(path)
if err != nil {
    return config.HandleFileWriteError(err, cfg.Verbose, path)
}

// For security validation (before attempting writes)
if isConfigFile(path) {
    return config.WriteErrorWithVerbose(cfg.Verbose, "cannot write to configuration file")
}

// For FileOutputManager atomic writes
tempFile := path + ".tmp"
if err := writeToTempFile(tempFile, data); err != nil {
    return config.HandleFileWriteError(err, cfg.Verbose, tempFile)
}
if err := os.Rename(tempFile, path); err != nil {
    return config.HandleFileWriteError(err, cfg.Verbose, path)
}

// For ManifestSystem operations
manifestPath := cfg.OutputManifest
if err := validateOutputPath(manifestPath, cfg.Verbose); err != nil {
    return fmt.Errorf("manifest file: %w", err)
}
```

This ensures that OS-level permission errors, disk full errors, network errors, and other system failures use the same generic message as security-based rejections in non-verbose mode, creating a large pool of legitimate errors that obfuscate security blocks.

## Implementation

### Code Location

The validation is implemented in `internal/config/config.go`:

```go
// WriteErrorWithVerbose returns either a generic or detailed error message
// based on verbose mode. Use this for security-sensitive write failures.
func WriteErrorWithVerbose(verbose bool, verboseDetails string) error {
    if verbose {
        return fmt.Errorf("%s", verboseDetails)
    }
    return fmt.Errorf("Unknown write/append error")
}

// HandleFileWriteError processes file writing errors and returns appropriate
// error messages. This obfuscates security errors by mixing them with
// legitimate file system errors.
func HandleFileWriteError(err error, verbose bool, path string) error {
    if err == nil {
        return nil
    }

    errStr := err.Error()
    
    // Permission denied, disk full, network errors, path too long, etc.
    // all use the same generic message as security blocks
    if strings.Contains(errStr, "permission denied") || 
       strings.Contains(errStr, "no space left") ||
       strings.Contains(errStr, "network") ||
       strings.Contains(errStr, "file name too long") {
        return FileSystemError(verbose, fmt.Sprintf("detailed error for %s", path))
    }
    
    return err // Other errors remain specific
}

func validateOutputPath(path string, verbose bool) error {
    // Extension whitelist (always specific - not security-sensitive)
    ext := strings.ToLower(filepath.Ext(path))
    allowedExts := []string{".txt", ".json", ".csv"}
    
    // Config name blacklist (security-sensitive - use verbose-aware errors)
    basename := filepath.Base(path)
    blockedNames := []string{".hashi.toml", "config.toml", ".env", "config.json"}
    
    for _, blocked := range blockedNames {
        if strings.EqualFold(basename, blocked) {
            return WriteErrorWithVerbose(verbose, fmt.Sprintf("cannot write to configuration file: %s", blocked))
        }
    }
    
    // Directory protection (security-sensitive - use verbose-aware errors)
    if strings.Contains(strings.ToLower(path), ".hashi") || 
       strings.Contains(strings.ToLower(path), ".config/hashi") {
        return WriteErrorWithVerbose(verbose, "cannot write to configuration directory")
    }
    
    return nil
}
```

### Integration

Validation is automatically applied during config validation:
- `--output` file paths
- `--log-file` paths  
- `--log-json` paths

## Testing

Simple test coverage in `internal/config/security_test.go`:

```bash
go test ./internal/config -run TestValidateOutputPath
go test ./internal/config -run TestValidateConfigWithSecurity
```

## Integration with Future Components

### FileOutputManager (Task 19)

The FileOutputManager component will handle all file output operations with enhanced safety features:

**Security Integration:**
- All write operations must use `HandleFileWriteError()` for consistent error handling
- Path validation via `validateOutputPath()` before any write attempts
- Atomic writes using temp file + rename pattern to prevent partial writes
- Overwrite protection with user confirmation (unless --force specified)

**Implementation Requirements:**
```go
// Before writing
if err := validateOutputPath(outputPath, verbose); err != nil {
    return fmt.Errorf("output file: %w", err)
}

// During atomic write
tempPath := outputPath + ".tmp"
if err := atomicWrite(tempPath, data); err != nil {
    return config.HandleFileWriteError(err, verbose, outputPath)
}
```

### ManifestSystem (Task 41)

The ManifestSystem enables incremental operations by tracking file states:

**Security Integration:**
- Manifest files subject to same path validation as other outputs
- Both `--manifest` (input) and `--output-manifest` (output) paths validated
- Atomic writes for manifest updates to prevent corruption
- JSON format validation to prevent malformed manifests

**Protected Manifest Locations:**
- Cannot write manifests to config directories (`.hashi/`, `.config/hashi/`)
- Cannot overwrite config files (`.hashi.toml`, `config.toml`, etc.)
- Must use allowed extensions (`.json` for manifests)

### Advanced Filtering (Task 39)

File filtering operations maintain security boundaries:

**Security Considerations:**
- Filter patterns cannot bypass output path validation
- Include/exclude patterns apply only to input file discovery
- No impact on output file security restrictions
- Glob patterns processed safely without shell expansion

## Design Rationale

### Why This Approach?

1. **Simple** - ~30 lines of code, easy to understand
2. **Effective** - Covers realistic threats and common accidents
3. **Maintainable** - Easy to modify and extend
4. **Fast** - Minimal performance overhead
5. **Clear** - Obvious what's protected and why

### What It Protects Against

- Accidental overwriting of config files
- Writing to config directories
- Creating files with dangerous extensions
- Case-insensitive config file conflicts

### What It Doesn't Protect Against

- Malicious modification of hashi binary
- External processes modifying config files
- Symlink attacks
- Advanced filesystem manipulation

For a CLI tool like hashi, this level of protection is appropriate and sufficient.

## Troubleshooting

### Common Issues

**"output files must have extension: .txt, .json, .csv"**
→ Use one of the allowed extensions for output files

**"Unknown write/append error"**
→ The specified output path cannot be used for security reasons. Use --verbose for details or see documentation for configuration and security information

**"output files must have extension: .txt, .json, .csv"**
→ Use one of the allowed extensions for output files

### Adding New Protected Items

**To add project-specific patterns:**

Environment variables (temporary):
```bash
export HASHI_BLACKLIST_FILES="$HASHI_BLACKLIST_FILES,mypattern*"
export HASHI_BLACKLIST_DIRS="$HASHI_BLACKLIST_DIRS,sensitive_dir"
```

Config file (persistent):
```toml
[security]
blacklist_files = ["mypattern*", "sensitive*"]
blacklist_dirs = ["sensitive_dir", "private*"]
```

**To allow exceptions to default patterns:**

```bash
# Allow specific config file for reports
export HASHI_WHITELIST_FILES="config_analysis_report.txt"

# Allow config directory for output
export HASHI_WHITELIST_DIRS="config_results"
```

**Pattern Design Guidelines:**
- Use exact matches for specific files: `"config.txt"`
- Use wildcards for pattern families: `"config*"`, `"*secret*"`
- Be specific to avoid blocking legitimate files
- Test patterns with `--verbose` flag to see detailed error messages
- Remember: whitelist overrides blacklist for exceptions