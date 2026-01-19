# CLI Power User Wishlist for Hashi
*Brainstorming session: 2025-01-17*

## Purpose
This document captures priorities and pain points from a CLI power user perspective. These users live in the terminal, value efficiency, and have strong muscle memory for command patterns.

## Core Power User Values

### Efficiency
- Every keystroke counts
- Muscle memory is sacred
- Speed is a feature
- Defaults should be smart

### Predictability
- Same input = same output
- No surprises
- Clear behavior
- Documented edge cases

### Composability
- Works with other tools
- Clean input/output
- Respects conventions
- No special cases

### Control
- Override any default
- Fine-grained options
- Escape hatches everywhere
- No forced behavior

---

## Critical Requirements

### 1. Short Flags for Everything
**Current Pain:**
```bash
# Too much typing
hashi --recursive --algorithm sha256 --output json --quiet
```

**Desired:**
```bash
# Minimal keystrokes
hashi -ra sha256 -oj -q
```

**Requirements:**
- Every long flag has a short equivalent
- Logical, memorable abbreviations
- Consistent patterns across flags
- No conflicts between short flags
- Document both forms in help

**Common Patterns:**
- `-r` for recursive
- `-q` for quiet
- `-v` for verbose
- `-o` for output format
- `-a` for algorithm
- `-j` for JSON (or `-oj` for output json)

---

### 2. Tab Completion That Actually Works
**Requirement:**
- Bash completion
- Zsh completion with descriptions
- Fish completion with inline help
- Context-aware suggestions

**Functionality:**
```bash
hashi --alg<TAB>           # completes to --algorithm
hashi -a sh<TAB>           # completes to sha256, sha512
hashi --output <TAB>       # shows: json, plain, verbose
hashi file<TAB>            # completes filenames intelligently
```

**Implementation:**
- Generate completion scripts from flag definitions
- Install to standard locations
- Keep in sync with actual flags
- Test on all supported shells

---

### 3. Environment Variable Defaults
**Requirement:**
- Common options configurable via environment
- Clear naming convention
- Easy to discover
- Flags override environment variables

**Examples:**
```bash
# In .bashrc or .zshrc
export HASHI_ALGORITHM=sha512
export HASHI_OUTPUT_FORMAT=plain
export HASHI_RECURSIVE=true
export HASHI_QUIET=false

# Now this uses sha512 by default
hashi file.txt

# But can still override
hashi -a sha256 file.txt
```

**Naming Convention:**
- Prefix: `HASHI_`
- Uppercase flag name
- Underscores for multi-word flags
- Boolean values: true/false, 1/0, yes/no

**Documentation:**
- List all supported environment variables
- Show examples in README
- Include in man page
- `hashi --show-env` to display current values

---

### 4. Fast Startup Time
**Requirement:**
- Startup time < 50ms
- No unnecessary initialization
- Lazy loading where possible
- Compiled binary (not interpreted)

**Measurement:**
```bash
time hashi --version
# Target: real 0m0.010s
# Acceptable: real 0m0.050s
# Unacceptable: real 0m0.100s+
```

**Implementation:**
- Avoid heavy imports in main package
- Defer initialization until needed
- No network calls on startup
- No config file parsing unless used

---

### 5. Clean Pipe Behavior
**Requirement:**
- Auto-detect TTY vs pipe
- No progress bars in pipes
- No color in pipes (unless forced)
- Consistent output format

**Examples:**
```bash
# Interactive: shows progress, color
hashi *.txt

# Piped: clean output, no progress
hashi *.txt | grep "abc123"

# Extract just hashes
hashi --plain *.txt | cut -f2

# Process JSON
hashi --json *.log | jq -r '.results[].files[]'
```

**Auto-Detection:**
- Check if stdout is TTY
- Disable progress bars in pipes
- Disable color in pipes (respect NO_COLOR)
- Allow override with --color=always

---

## High Priority Features

### 6. Rich Filtering Options
**Requirement:**
- Filter files before hashing
- Size-based filtering
- Date-based filtering
- Pattern-based filtering
- Combine multiple filters

**Examples:**
```bash
# Size filters
hashi --min-size 1M --max-size 100M *.bin
hashi --size +1M --size -100M *.bin  # Alternative syntax

# Date filters
hashi --newer-than 2025-01-01 /logs
hashi --older-than 7d /archive
hashi --modified-within 24h /tmp

# Pattern filters
hashi --include '*.{txt,md,log}' -r /docs
hashi --exclude '*.tmp' --exclude 'node_modules/*' -r .

# Combine filters
hashi --size +1M --newer-than 7d --include '*.log' -r /var/log
```

**Implementation:**
- Apply filters before hashing (efficiency)
- Support glob patterns
- Support regex (with flag)
- Clear error messages for invalid patterns

---

### 7. Parallel Processing by Default
**Requirement:**
- Use multiple cores automatically
- Configurable parallelism
- Efficient for large file sets
- No output corruption

**Examples:**
```bash
# Auto-detect cores
hashi -r /large/directory

# Specify parallelism
hashi --parallel 8 *.iso
hashi -p 8 *.iso

# Disable parallelism
hashi --parallel 1 *.txt
hashi --no-parallel *.txt
```

**Implementation:**
- Default: number of CPU cores
- Worker pool pattern
- Ordered output (maintain input order)
- Progress bar shows total progress

---

### 8. Batch Operations Without Loops
**Requirement:**
- Verify multiple files against hashes
- No manual for loops needed
- Clear success/failure reporting
- Useful exit codes

**Examples:**
```bash
# Verify all files against hash list
hashi --verify-all *.txt --against hashes.txt

# Report only failures
hashi --verify-all *.txt --against hashes.txt --report-failures

# Exit on first failure
hashi --verify-all *.txt --against hashes.txt --fail-fast
```

**Output:**
```
✓ file1.txt
✓ file2.txt
✗ file3.txt (hash mismatch)
✓ file4.txt

Summary: 3/4 passed, 1 failed
```

---

### 9. Dry Run Mode
**Requirement:**
- Show what would be processed
- No actual hashing
- Fast preview
- Useful for large operations

**Examples:**
```bash
# Preview what would be hashed
hashi --dry-run -r /huge/directory

# Output:
# Would process 10,234 files (45.2 GB)
# Estimated time: 2m 34s
```

**Implementation:**
- Walk directory structure
- Apply filters
- Count files and sizes
- Estimate time based on size
- No actual hashing

---

### 10. Multiple Verbosity Levels
**Requirement:**
- Control output detail
- Multiple levels
- Cumulative flags
- Debug mode for troubleshooting

**Examples:**
```bash
# Normal output
hashi file.txt

# Verbose: show details
hashi --verbose file.txt
hashi -v file.txt

# Very verbose: show more details
hashi -vv file.txt

# Debug: show everything
hashi --debug file.txt
hashi -vvv file.txt

# Quiet: only errors
hashi --quiet file.txt
hashi -q file.txt
```

**Verbosity Levels:**
- Default: results only
- `-v`: show files being processed
- `-vv`: show timing and statistics
- `-vvv` / `--debug`: show internal operations
- `-q`: suppress all non-error output

---

## Medium Priority Features

### 11. Smart Defaults with Easy Overrides
**Requirement:**
- Intelligent behavior based on arguments
- Explicit flags when needed
- No ambiguity
- Predictable behavior

**Examples:**
```bash
# Auto-detect mode
hashi file.txt abc123...        # Compares automatically
hashi *.txt                     # Finds duplicates
hashi --verify archive.zip      # Verifies integrity (explicit)

# Standard behavior
hashi archive.zip               # Hash ZIP file itself
```

**Auto-Detection Logic:**
- 1 file + 1 hash string = compare
- Multiple files = find duplicates
- `--verify` + .zip file = verify integrity
- Directory = recursive hash

---

### 12. Null-Terminated Input Support
**Requirement:**
- Handle filenames with spaces/newlines
- Compatible with find -print0
- Compatible with xargs -0

**Examples:**
```bash
# With find
find . -type f -name '*.log' -print0 | hashi --stdin0

# With xargs
cat files.txt | xargs -0 -I {} hashi {}

# Output null-terminated
hashi --print0 *.txt | xargs -0 -I {} echo "Processing: {}"
```

**Implementation:**
- `--stdin0` or `-0`: read null-terminated input
- `--print0` or `-z`: output null-terminated
- Compatible with GNU tools

---

### 13. Color Control
**Requirement:**
- Auto-detect TTY
- Respect NO_COLOR
- Force color when needed
- Sensible color scheme

**Examples:**
```bash
# Auto-detect (default)
hashi *.txt              # Colorized if TTY
hashi *.txt | less       # No color (pipe detected)

# Force color
hashi --color=always *.txt | less -R
hashi --color *.txt | less -R

# Disable color
hashi --color=never *.txt
hashi --no-color *.txt

# Respect environment
NO_COLOR=1 hashi *.txt   # No color
```

**Color Usage:**
- Green: matches, success
- Red: mismatches, errors
- Yellow: warnings
- Blue: informational
- Gray: secondary info

---

### 14. Statistics and Summaries
**Requirement:**
- Optional statistics
- Performance metrics
- Duplicate analysis
- Time estimates

**Examples:**
```bash
hashi -r /huge/directory --stats

# Output:
# Processed: 10,234 files
# Total size: 45.2 GB
# Duplicates: 234 files (1.2 GB wasted)
# Unique hashes: 10,000
# Time: 2m 34s
# Speed: 297 MB/s
```

**Statistics Include:**
- File count
- Total size
- Duplicate count and size
- Unique hash count
- Processing time
- Throughput (MB/s)

---

### 15. Diff and Compare Modes
**Requirement:**
- Compare manifests
- Show what changed
- Visual diff output
- Machine-readable format

**Examples:**
```bash
# Compare two manifests
hashi --diff old-manifest.txt new-manifest.txt

# Output:
# Added (3 files):
#   + new-file1.txt
#   + new-file2.txt
#   + new-file3.txt
#
# Removed (1 file):
#   - old-file.txt
#
# Modified (2 files):
#   M changed-file1.txt (hash changed)
#   M changed-file2.txt (hash changed)

# Machine-readable
hashi --diff old.txt new.txt --format json
```

**Diff Types:**
- Added files
- Removed files
- Modified files (hash changed)
- Renamed files (same hash, different path)

---

### 16. Watch Mode
**Requirement:**
- Monitor for changes
- Configurable intervals
- Alert mechanisms
- Efficient (only check changed files)

**Examples:**
```bash
# Watch directory for changes
hashi --watch /important/files

# With interval
hashi --watch /important/files --interval 60s

# Alert on change
hashi --watch /important/files --on-change 'notify-send "Files changed"'

# Continuous verification
hashi --watch-manifest manifest.txt --interval 5m
```

**Implementation:**
- Use file system timestamps
- Only rehash changed files
- Configurable check interval
- Exit code or command on change

---

### 17. Template-Based Output
**Requirement:**
- Custom output format
- Simple template syntax
- No need for external tools
- Common use cases covered

**Examples:**
```bash
# Custom format
hashi --template '{{.File}}: {{.Hash}}' *.txt
# Output: file1.txt: abc123...

# With conditionals
hashi --template '{{if .Match}}MATCH: {{end}}{{.File}}' *.txt

# JSON-like but custom
hashi --template '{"file":"{{.File}}","hash":"{{.Hash}}"}' *.txt
```

**Template Variables:**
- `{{.File}}` - filename
- `{{.Hash}}` - hash value
- `{{.Size}}` - file size
- `{{.Modified}}` - modification time
- `{{.Match}}` - boolean, has match
- `{{.Algorithm}}` - hash algorithm used

---

### 18. Expression-Based Filters
**Requirement:**
- Complex filtering logic
- Simple expression language
- No need for external tools
- Common operations supported

**Examples:**
```bash
# Size and extension
hashi --filter 'size > 1MB && ext == ".log"' -r /var/log

# Date-based
hashi --filter 'modified > "2025-01-01"' /archive

# Complex logic
hashi --filter '(size > 1MB || ext == ".iso") && modified < 30d' -r .
```

**Expression Language:**
- Operators: `==`, `!=`, `>`, `<`, `>=`, `<=`, `&&`, `||`
- Variables: `size`, `ext`, `modified`, `name`, `path`
- Functions: `contains()`, `matches()`, `startswith()`, `endswith()`
- Units: `KB`, `MB`, `GB`, `d` (days), `h` (hours)

---

## Low Priority Features

### 19. Command History
**Requirement:**
- Optional feature (privacy-respecting)
- Store recent commands
- Easy recall
- Clear history

**Examples:**
```bash
# Show history
hashi --history
# 1: hashi -r /var/log --size +1M
# 2: hashi *.txt --json
# 3: hashi file.txt abc123...

# Repeat command
hashi --replay 1

# Show last command
hashi --last-command
```

**Implementation:**
- Store in `~/.hashi_history`
- Configurable via `HASHI_HISTORY_FILE`
- Disable with `HASHI_HISTORY=0`
- Clear with `hashi --clear-history`

**Privacy:**
- Opt-in, not opt-out
- Clear documentation
- Easy to disable
- Easy to clear

---

### 20. Named Command Templates
**Requirement:**
- Save complex commands
- Easy recall
- Per-user storage
- Simple management

**Examples:**
```bash
# Save command template
hashi -r /downloads --size +1M --json > /dev/null
# (Manually save to config)

# Later, use template
hashi --run verify-downloads

# List templates
hashi --list-templates
```

**Implementation:**
- Store in config file
- Simple TOML format
- User-defined names
- Variable substitution

**Config Example:**
```toml
[templates.verify-downloads]
command = "-r /downloads --size +1M --json"
description = "Verify large downloads"

[templates.find-dupes]
command = "-r . --include '*.jpg' --stats"
description = "Find duplicate images"
```

---

### 21. Per-Directory Configuration
**Requirement:**
- Project-specific defaults
- Manual setup (hashi doesn't write config)
- Override global defaults
- Clear precedence

**Examples:**
```bash
# User creates .hashi.toml in project directory
cd ~/project
cat .hashi.toml
# [defaults]
# algorithm = "sha512"
# recursive = true
# output_format = "json"

# Now hashi uses these defaults in this directory
hashi *.txt  # Uses sha512, recursive, json output
```

**Precedence (highest to lowest):**
1. Command-line flags
2. Environment variables
3. Per-directory config (`.hashi.toml`)
4. User config (`~/.config/hashi/config.toml`)
5. Built-in defaults

**Documentation:**
- Clear instructions for manual setup
- Example config files
- Precedence rules explained
- `hashi --show-config` to display effective config

**Important:** Hashi never writes config files. Users must create and edit them manually.

---

### 22. Bookmarks for Common Paths
**Requirement:**
- Quick access to common directories
- Manual setup in config file
- Simple syntax
- Easy to use

**Examples:**
```bash
# User manually adds to config file:
# [bookmarks]
# downloads = "/home/user/Downloads"
# projects = "/home/user/Projects"
# logs = "/var/log"

# Use bookmark
hashi --bookmark downloads --verify manifest.txt
hashi -b downloads -r

# List bookmarks
hashi --list-bookmarks
```

**Implementation:**
- Stored in user config file
- User edits manually
- Simple path substitution
- Clear error if bookmark not found

---

### 23. Hooks for Custom Actions
**Requirement:**
- Execute commands on events
- Simple hook system
- Common use cases
- No complex scripting

**Examples:**
```bash
# On match found
hashi --on-match 'echo "Found duplicate: {file}"' *.txt

# On mismatch
hashi --on-mismatch 'logger "Hash mismatch: {file}"' *.txt

# On completion
hashi --on-complete 'notify-send "Hashing complete"' -r /large
```

**Hook Variables:**
- `{file}` - current filename
- `{hash}` - computed hash
- `{expected}` - expected hash (if comparing)
- `{count}` - number of matches

**Events:**
- `--on-match` - file matches another
- `--on-mismatch` - hash doesn't match expected
- `--on-error` - error processing file
- `--on-complete` - all processing done

---

## Configuration System

### Read-Only Principle
**Critical:** Hashi never writes configuration files. All config must be created and edited manually by users.

### Configuration Sources (Precedence Order)
1. **Command-line flags** (highest priority)
2. **Environment variables** (`HASHI_*`)
3. **Per-directory config** (`.hashi.toml` in current directory)
4. **User config** (`~/.config/hashi/config.toml`)
5. **Built-in defaults** (lowest priority)

### Configuration File Format
**Location:** `~/.config/hashi/config.toml` (user) or `.hashi.toml` (per-directory)

**Example:**
```toml
[defaults]
algorithm = "sha256"
output_format = "plain"
recursive = false
quiet = false
parallel = 0  # 0 = auto-detect cores

[colors]
enabled = true
matches = "green"
mismatches = "red"
warnings = "yellow"

[bookmarks]
downloads = "/home/user/Downloads"
projects = "/home/user/Projects"

[templates.verify-downloads]
command = "-r /downloads --size +1M --json"
description = "Verify large downloads"
```

### Environment Variables
All flags can be set via environment variables:
- Format: `HASHI_<FLAG_NAME>`
- Example: `HASHI_ALGORITHM=sha512`
- Boolean: `true`/`false`, `1`/`0`, `yes`/`no`
- Override with command-line flags

### Display Current Configuration
```bash
# Show effective configuration
hashi --show-config

# Output:
# Configuration sources:
#   Built-in defaults
#   User config: ~/.config/hashi/config.toml
#   Directory config: .hashi.toml
#   Environment: HASHI_ALGORITHM=sha512
#   Flags: --quiet
#
# Effective settings:
#   algorithm: sha512 (from environment)
#   output_format: plain (from user config)
#   recursive: false (from built-in default)
#   quiet: true (from flag)
```

### Setup Instructions
Documentation must include:
1. How to create config file manually
2. Example config files
3. All available options
4. Precedence rules
5. How to verify effective config

---

## Anti-Requirements

### What NOT to Include
- Slow startup (>100ms)
- Required configuration file
- Interactive prompts
- Automatic config file creation/modification
- Breaking changes without warning
- Inconsistent flag naming
- No way to override defaults

### Red Flags
- Tries to be too smart (wrong guesses)
- Unpredictable behavior
- Poor error messages
- No examples in help
- Verbose output by default

---

## Success Metrics (Power User Perspective)

### Adoption Indicators
- Becomes part of daily workflow
- Replaces existing tools
- Recommended to others
- Used in scripts and aliases

### Quality Indicators
- Fast enough for interactive use
- Predictable behavior
- Good error messages
- Complete documentation
- Active maintenance

### Efficiency Indicators
- Reduces keystrokes
- Eliminates manual loops
- Speeds up common tasks
- Integrates with existing workflow

---

## Integration with Project Goals

All features must align with hashi's core principles:
1. **Developer Continuity** - Clear documentation
2. **User-First Design** - Serves real user needs
3. **No Lock-Out** - Users maintain control

### Feature Assessment
Each suggestion must pass through standard criteria:
- Clear user benefit
- Fills real workflow gap
- Addresses common pain points
- Doesn't add unnecessary complexity

---

## Next Steps

1. Review current CLI against power user requirements
2. Identify quick wins (short flags, tab completion)
3. Prioritize features based on impact
4. Create issues for accepted features
5. Document rejected features with rationale
6. Build examples for common workflows
7. Create comprehensive man pages

---

## Notes

- Power users value efficiency and predictability above all
- Muscle memory is critical - don't break it
- Speed matters - both startup and execution
- Configuration is convenience, not requirement
- Hashi never modifies its own configuration
- Users must manually create and edit config files
- Clear documentation is essential for manual setup
