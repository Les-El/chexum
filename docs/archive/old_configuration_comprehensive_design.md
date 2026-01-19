# Configuration System Design
*Design document for hashi configuration and variable management*

## Purpose
This document defines how hashi handles configuration, environment variables, and user preferences while maintaining security and the read-only principle.

## Core Principles

### 1. Read-Only Stance
**Critical:** Hashi never writes configuration files or modifies environment variables.
- All configuration must be created and edited manually by users
- Hashi only reads configuration, never writes it
- No automatic config generation or modification
- Users maintain full control over their configuration

### 2. Optional Configuration
**Philosophy:** Configuration is convenience, not requirement.
- Hashi must work without any configuration files
- All features accessible via command-line flags
- Configuration provides defaults, not exclusive functionality
- No lock-out scenarios due to missing config

### 3. Clear Precedence
**Hierarchy (highest to lowest priority):**
1. Command-line flags
2. Environment variables (`HASHI_*`)
3. Per-directory config (`.hashi.toml`)
4. User config (`~/.config/hashi/config.toml`)
5. Built-in defaults

---

## Configuration Sources

### 1. Command-Line Flags
**Priority:** Highest
**Scope:** Single invocation
**Examples:**
```bash
hashi --algorithm sha512 --quiet --json file.txt
```

**Characteristics:**
- Always override all other sources
- Explicit and visible
- No persistence between invocations
- Primary interface for all functionality

### 2. Environment Variables
**Priority:** Second highest
**Scope:** Shell session or system-wide
**Naming Convention:** `HASHI_<FLAG_NAME>`

**Examples:**
```bash
export HASHI_ALGORITHM=sha512
export HASHI_OUTPUT_FORMAT=json
export HASHI_RECURSIVE=true
export HASHI_QUIET=false
export HASHI_PARALLEL=8
```

**Supported Variables:**
- `HASHI_ALGORITHM` - Default hash algorithm
- `HASHI_OUTPUT_FORMAT` - Default output format (plain, json, verbose)
- `HASHI_RECURSIVE` - Default recursive behavior (true/false)
- `HASHI_QUIET` - Default quiet mode (true/false)
- `HASHI_PARALLEL` - Default parallelism (number or 0 for auto)
- `HASHI_COLOR` - Color mode (auto, always, never)

**Boolean Values:** Accept `true`/`false`, `1`/`0`, `yes`/`no`

### 3. Per-Directory Configuration
**Priority:** Third highest
**Scope:** Current directory and subdirectories
**File:** `.hashi.toml` in current directory

**Use Cases:**
- Project-specific defaults
- Team consistency
- Repository-specific settings

**Example:**
```toml
# .hashi.toml in project root
[defaults]
algorithm = "sha512"
recursive = true
output_format = "json"

[filters]
exclude = ["*.tmp", "node_modules/*", ".git/*"]
min_size = "1KB"
```

### 4. User Configuration
**Priority:** Fourth highest
**Scope:** User-wide defaults
**File:** `~/.config/hashi/config.toml`

**Use Cases:**
- Personal preferences
- Common bookmarks
- Default algorithms

**Example:**
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

[bookmarks]
downloads = "/home/user/Downloads"
projects = "/home/user/Projects"
logs = "/var/log"

[templates.verify-downloads]
command = "-r /downloads --size +1M --json"
description = "Verify large downloads"

[templates.find-dupes]
command = "-r . --include '*.jpg' --stats"
description = "Find duplicate images"
```

### 5. Built-in Defaults
**Priority:** Lowest
**Scope:** Application defaults
**Source:** Compiled into binary

**Current Defaults:**
- Algorithm: SHA-256
- Output format: Default (grouped)
- Recursive: false
- Quiet: false
- Parallel: auto-detect CPU cores
- Color: auto (TTY detection)

---

## Configuration File Format

### File Format: TOML
**Rationale:**
- Human-readable and editable
- Good Go library support
- Clear syntax for nested structures
- Comments supported
- Widely adopted

### Standard Locations
**User Config:**
- Linux/macOS: `~/.config/hashi/config.toml`
- Windows: `%APPDATA%\hashi\config.toml`

**Per-Directory Config:**
- `.hashi.toml` in current directory
- Searched up directory tree (optional feature)

### Configuration Sections

#### `[defaults]`
Default values for command-line flags:
```toml
[defaults]
algorithm = "sha256"           # --algorithm
output_format = "plain"        # --output
recursive = false              # --recursive
quiet = false                  # --quiet
parallel = 0                   # --parallel (0 = auto)
color = "auto"                 # --color
```

#### `[colors]`
Color scheme customization:
```toml
[colors]
enabled = true
matches = "green"
mismatches = "red"
warnings = "yellow"
info = "blue"
secondary = "gray"
```

#### `[bookmarks]`
Named path shortcuts:
```toml
[bookmarks]
downloads = "/home/user/Downloads"
projects = "/home/user/Projects"
logs = "/var/log"
```

#### `[templates]`
Named command templates:
```toml
[templates.verify-downloads]
command = "-r /downloads --size +1M --json"
description = "Verify large downloads"

[templates.find-dupes]
command = "-r . --include '*.jpg' --stats"
description = "Find duplicate images"
```

#### `[filters]`
Default filtering options:
```toml
[filters]
exclude = ["*.tmp", "node_modules/*"]
include = ["*.txt", "*.md", "*.log"]
min_size = "1KB"
max_size = "100MB"
```

---

## Self-Reporting and Security

### The Security Challenge
Configuration self-reporting can expose sensitive information:
- File system paths
- Directory structure
- Environment variable values
- Project locations
- User patterns

### Threat Model

#### Information Disclosure Risks
**Low Risk:**
- Algorithm preferences (sha256, sha512)
- Output format choices
- Boolean settings (recursive, quiet)

**Medium Risk:**
- Bookmark names (without paths)
- Template names (without commands)
- Filter patterns

**High Risk:**
- Full file paths
- Environment variable values
- Directory structure
- Template command strings

#### Attack Scenarios
1. **System Reconnaissance:** Attacker uses `--show-config` to map user's file system
2. **Environment Leakage:** Sensitive tokens in environment variables exposed
3. **Project Discovery:** Bookmark paths reveal project locations

### Safe Self-Reporting Design

#### Default Mode: Minimal Disclosure
```bash
$ hashi --show-config
Configuration loaded from 4 sources
Algorithm: sha256 (user-config)
Format: plain (default)
Recursive: false (default)
Quiet: true (flag)
Parallel: auto (default)
Color: auto (default)
```

**Safe Information:**
- Effective setting values
- Source type (not path)
- Counts (not contents)

**Hidden Information:**
- Full file paths
- Environment variable values
- Config file contents
- Bookmark paths

#### Verbose Mode: Controlled Disclosure
```bash
$ hashi --show-config --verbose
Warning: Verbose mode may expose file paths
Continue? [y/N] y

Configuration sources:
  Built-in defaults
  User config: ~/.config/hashi/config.toml (found)
  Directory config: .hashi.toml (not found)
  Environment: 3 variables found
  Command-line: --quiet

Effective settings:
  algorithm: sha256 (from user config)
  output_format: plain (from built-in default)
  recursive: false (from built-in default)
  quiet: true (from command-line flag)
  parallel: auto (from built-in default)

Bookmarks: 3 defined
Templates: 2 defined
```

**Additional Information:**
- Config file paths (sanitized with ~)
- Environment variable count
- Bookmark/template counts
- Source attribution

#### Debug Mode: Full Disclosure
```bash
$ hashi --show-config --debug
WARNING: Debug mode exposes sensitive configuration details
This may reveal file paths, environment variables, and system information.
Only use in trusted environments.
Continue? [y/N] y

[Full configuration dump with all details]
```

**Complete Information:**
- All file paths
- All environment variables
- Full config file contents
- All bookmarks and templates

### Implementation Strategy

#### Safe by Default
- Default `--show-config` reveals minimal information
- No sensitive paths or values
- Clear source attribution without exposure

#### Progressive Disclosure
- `--verbose` adds controlled detail with warning
- `--debug` provides complete information with strong warning
- User must explicitly consent to information disclosure

#### Sanitization Rules
1. **Path Sanitization:** Replace home directory with `~`
2. **Environment Filtering:** Show count, not values
3. **Content Filtering:** Show structure, not contents
4. **Relative Paths:** Use relative paths where possible

#### Example Implementation
```go
type ConfigDisplay struct {
    Level DisplayLevel // Safe, Verbose, Debug
}

func (c *ConfigDisplay) ShowConfig() {
    switch c.Level {
    case Safe:
        c.showSafeConfig()
    case Verbose:
        if c.confirmVerbose() {
            c.showVerboseConfig()
        }
    case Debug:
        if c.confirmDebug() {
            c.showDebugConfig()
        }
    }
}

func (c *ConfigDisplay) showSafeConfig() {
    // Only show effective values and source types
    // No paths, no environment values
}
```

---

## Configuration Validation

### Validation Strategy
Focus on "is it working?" rather than "what is it?"

#### Basic Validation
```bash
$ hashi --validate-config
Configuration OK: 4 sources loaded, 8 settings applied
```

#### Detailed Validation
```bash
$ hashi --validate-config --verbose
Sources:
  ✓ Built-in defaults
  ✓ User config: ~/.config/hashi/config.toml
  ✗ Directory config: .hashi.toml (not found)
  ✓ Environment: 3 variables
  ✓ Command-line: 1 flag

Settings:
  ✓ algorithm: sha256 (valid)
  ✓ output_format: plain (valid)
  ✗ parallel: 999 (invalid, using auto)

Warnings:
  Unknown setting 'algoritm' in user config (typo?)
  Bookmark 'old-project' points to non-existent directory
```

#### Error Reporting
```bash
$ hashi --validate-config
Error: Invalid algorithm 'sha999' in environment HASHI_ALGORITHM
Error: Malformed TOML in ~/.config/hashi/config.toml line 15
Warning: Unknown setting 'algoritm' in user config (typo?)
```

### Validation Rules
1. **Algorithm Validation:** Check against supported algorithms
2. **Path Validation:** Verify bookmark paths exist (warn, don't fail)
3. **Format Validation:** Validate TOML syntax
4. **Type Validation:** Check boolean/numeric values
5. **Unknown Settings:** Warn about typos or deprecated options

---

## Security Considerations

## Security Considerations

### Self-Modification Prevention
**Critical:** Hashi implements the "hashi can't change hashi" security principle.

**Simple Protection Strategy:**
- **Extension Whitelist:** Only `.txt`, `.json`, `.csv` files allowed for output
- **Config Name Blacklist:** Block obvious config files (`.hashi.toml`, `config.json`, etc.)
- **Directory Protection:** Prevent writing to `.hashi/` or `.config/hashi/` directories

**Allowed Operations:**
```bash
# Safe - allowed extensions and locations
hashi --output results.txt *.txt
hashi --output /tmp/data.json --json *.txt
hashi --output logs/report.csv *.txt
```

**Blocked Operations:**
```bash
# Blocked - unsafe extensions
hashi --output malicious.sh *.txt
hashi --output script.py *.txt

# Blocked - config file names  
hashi --output .hashi.toml *.txt
hashi --output config.json *.txt

# Blocked - config directories
hashi --output .hashi/output.txt *.txt
hashi --output ~/.config/hashi/out.txt *.txt
```

**Error Messages:**
```
Error: output file: output files must have extension: .txt, .json, .csv (got .sh)
Error: output file: cannot overwrite configuration file: config.json
Error: output file: cannot write to configuration directory
```

### Environment Variable Security
**Risk:** Sensitive values in environment variables
**Mitigation:**
- Never display environment variable values
- Only show counts: "3 environment variables found"
- Warn users about sensitive data in environment

### Configuration File Security
**Risk:** Sensitive data in config files
**Mitigation:**
- Document security best practices
- Recommend file permissions (600)
- Warn against storing secrets in config

### Path Disclosure
**Risk:** File system structure exposure
**Mitigation:**
- Sanitize paths in default mode
- Use relative paths where possible
- Require explicit consent for full paths

### Information Leakage Prevention
**Strategies:**
1. **Minimal Default Display:** Show only essential information
2. **Progressive Disclosure:** Require explicit flags for sensitive info
3. **User Consent:** Warn before exposing sensitive data
4. **Sanitization:** Clean paths and values before display

---

## Implementation Guidelines

### Configuration Loading
```go
type Config struct {
    // Effective configuration after all sources merged
    Algorithm     string
    OutputFormat  string
    Recursive     bool
    Quiet         bool
    Parallel      int
    Color         string
    
    // Source tracking for display
    Sources       []ConfigSource
}

type ConfigSource struct {
    Type     SourceType  // Flag, Env, UserConfig, DirConfig, Default
    Path     string      // File path (if applicable)
    Settings map[string]SourcedValue
}

type SourcedValue struct {
    Value  interface{}
    Source SourceType
}
```

### Loading Order
1. Load built-in defaults
2. Load user config file (if exists)
3. Load directory config file (if exists)
4. Apply environment variables
5. Apply command-line flags
6. Validate final configuration

### Error Handling
- Invalid config files: warn and continue with other sources
- Missing files: silently ignore
- Invalid values: warn and use defaults
- Unknown settings: warn but don't fail

---

## User Documentation Requirements

### Setup Instructions
1. **Manual Config Creation:** Step-by-step guide
2. **Example Files:** Complete, commented examples
3. **Precedence Rules:** Clear explanation of override behavior
4. **Security Notes:** Best practices for sensitive data

### Reference Documentation
1. **All Environment Variables:** Complete list with examples
2. **Config File Schema:** All sections and options
3. **Validation Commands:** How to check configuration
4. **Troubleshooting:** Common issues and solutions

### Security Documentation
1. **Information Disclosure:** What `--show-config` reveals
2. **Safe Practices:** How to avoid exposing sensitive data
3. **File Permissions:** Recommended security settings
4. **Environment Hygiene:** Best practices for environment variables

---

## Future Considerations

### Potential Enhancements
1. **Config File Encryption:** Encrypt sensitive sections
2. **Multiple Profiles:** Switch between named configurations
3. **Config Inheritance:** Directory-based config hierarchies
4. **External Config Sources:** Remote configuration (with security)

### Rejected Features
1. **Automatic Config Generation:** Violates read-only principle
2. **Config Modification Commands:** Users must edit manually
3. **Centralized Config Management:** Adds complexity and security risk
4. **Cloud Config Sync:** Privacy and security concerns

---

## Testing Strategy

### Configuration Testing
1. **Precedence Testing:** Verify override behavior
2. **File Format Testing:** Valid and invalid TOML
3. **Environment Testing:** All supported variables
4. **Validation Testing:** Error detection and reporting

### Security Testing
1. **Information Disclosure:** Verify safe reporting
2. **Path Sanitization:** Test path cleaning
3. **Environment Isolation:** Prevent variable leakage
4. **Permission Testing:** File access controls

### Integration Testing
1. **Cross-Platform:** Different OS config locations
2. **Shell Integration:** Environment variable handling
3. **File System:** Permission and access scenarios
4. **Error Scenarios:** Graceful degradation

---

## Decision Log

### Accepted Decisions
1. **TOML Format:** Human-readable, well-supported
2. **Read-Only Principle:** Never write configuration
3. **Progressive Disclosure:** Safe default, verbose on request
4. **Environment Variables:** Standard `HASHI_*` prefix
5. **File Locations:** Follow XDG Base Directory Specification

### Rejected Alternatives
1. **JSON Config:** Less human-friendly than TOML
2. **YAML Config:** More complex parsing, security issues
3. **INI Format:** Limited nesting capabilities
4. **Automatic Config Writing:** Violates read-only principle
5. **Full Path Disclosure by Default:** Security risk

---

## Implementation Priority

### Phase 1: Basic Configuration
1. Environment variable support
2. User config file loading
3. Basic validation
4. Safe self-reporting

### Phase 2: Advanced Features
1. Per-directory configuration
2. Bookmarks and templates
3. Verbose configuration display
4. Comprehensive validation

### Phase 3: Polish
1. Debug mode with full disclosure
2. Configuration migration tools
3. Advanced security features
4. Performance optimization

---

## Notes

- Configuration is convenience, not requirement
- Security through progressive disclosure
- User maintains full control
- Clear documentation is essential
- Test all security assumptions