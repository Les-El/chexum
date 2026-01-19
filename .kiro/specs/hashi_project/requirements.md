# Requirements Document

## Introduction

This specification defines requirements for reviewing and improving the `hashi` CLI tool against industry-standard CLI design guidelines. The goal is to identify gaps, inconsistencies, and opportunities to enhance usability, discoverability, robustness, and adherence to CLI best practices as outlined in the CLI Guidelines document.

The review will result in concrete, actionable recommendations for improving the tool's user experience, output formatting, error handling, argument parsing, documentation, and overall design philosophy.

`hashi` is a read-only information discovery tool that computes and compares cryptographic hashes. It does not modify files or system state beyond writing output/log files when explicitly requested.

## Glossary

- **Hashi**: The command-line hash comparison tool being reviewed
- **CLI_Guidelines**: The industry-standard CLI design guidelines document used as the reference
- **TTY**: A terminal (teletypewriter) - an interactive terminal session
- **Exit_Code**: A numeric value returned by a program to indicate success (0) or failure (non-zero)
- **Flag**: A named parameter denoted with hyphen(s) (e.g., `-v`, `--verbose`)
- **Argument**: A positional parameter to a command
- **ANSI_Color**: Terminal escape sequences for colored text output
- **XDG_Spec**: X Desktop Group specification for configuration file locations
- **CRC32**: Cyclic Redundancy Check (32-bit) - an error-detecting code used in ZIP files for integrity verification
- **Integrity_Verification**: Confirming that data has not been corrupted during storage or transmission (bits are correct)
- **Authenticity_Verification**: Confirming that data has not been tampered with by a malicious actor (cryptographic proof of origin)
- **Boolean_Output**: Output mode where hashi returns only an exit code (0 or non-zero) with no stdout, optimized for script integration
- **Raw_Mode**: Flag (`--raw`) that treats files as raw bytes, bypassing special file type handling (e.g., ZIP verification)
- **Flag_Conflict**: A situation where two or more flags have incompatible or ambiguous combined behavior
- **Annotated_Edition**: A version of the source code with extensive educational comments (moonshot goal)
- **Fuzzing**: Automated testing technique that generates random inputs to discover unexpected behaviors
- **Dry_Run**: Preview mode that shows what would be processed without actually computing hashes
- **Manifest**: A file containing previous hash computation results used for incremental operations
- **Incremental_Operation**: Processing only files that have changed since a previous run
- **Atomic_Write**: Writing to a temporary file then renaming to prevent data loss on failure
- **Filter_Pattern**: A glob pattern or regular expression used to include or exclude files

## Requirements

### Requirement 1: Help System Compliance

**User Story:** As a user, I want comprehensive and accessible help documentation, so that I can quickly learn how to use the tool without consulting external resources.

#### Acceptance Criteria

1. WHEN a user runs `hashi` with no arguments, THE Hashi SHALL process all non-hidden files in the current directory
2. WHEN a user passes `-h`, `--help`, or `help` as an argument, THE Hashi SHALL display full help documentation
3. WHEN help text is displayed, THE Hashi SHALL include examples of common use cases
4. WHEN help text is displayed, THE Hashi SHALL use formatting (bold, sections) to improve readability
5. WHEN help text is displayed, THE Hashi SHALL provide a link to web-based documentation if available
6. WHEN a user makes a common mistake, THE Hashi SHALL suggest the correct command syntax

### Requirement 2: Output Design and Human-First Principles

**User Story:** As a user, I want clear, well-formatted output that prioritizes human readability, so that I can quickly understand results without parsing dense text.

#### Acceptance Criteria

1. WHEN output is sent to a TTY, THE Hashi SHALL use ANSI colors to highlight important information
2. WHEN output is sent to a non-TTY (pipe or redirect), THE Hashi SHALL disable colors automatically
3. WHEN the `NO_COLOR` environment variable is set, THE Hashi SHALL disable all color output
4. WHEN processing takes longer than 100ms, THE Hashi SHALL display progress indicators
5. WHEN displaying results, THE Hashi SHALL group files by matching hash with blank lines between groups
6. WHEN an operation succeeds, THE Hashi SHALL provide brief confirmation of what was processed
7. WHEN displaying errors, THE Hashi SHALL place the most important information at the end of output
8. WHEN `-q` or `--quiet` flag is passed, THE Hashi SHALL suppress all stdout output and only return exit code

### Requirement 3: Error Handling and User Guidance

**User Story:** As a user, I want clear, actionable error messages, so that I can quickly fix problems without consulting documentation.

#### Acceptance Criteria

1. WHEN an error occurs, THE Hashi SHALL provide a human-readable explanation of what went wrong
2. WHEN an error is fixable, THE Hashi SHALL suggest specific commands or actions to resolve it
3. WHEN a user provides invalid input, THE Hashi SHALL suggest corrections if the intent is clear
4. WHEN an unexpected error occurs, THE Hashi SHALL provide debug information and instructions for reporting bugs
5. WHEN multiple errors occur, THE Hashi SHALL group similar errors to reduce noise
6. WHEN an error message is displayed, THE Hashi SHALL avoid technical jargon unless in verbose mode

### Requirement 4: Argument and Flag Design

**User Story:** As a user, I want consistent, predictable flag names and argument handling, so that I can use the tool efficiently without memorizing arbitrary conventions.

#### Acceptance Criteria

1. WHEN flags are defined, THE Hashi SHALL provide both short (`-v`) and long (`--verbose`) versions
2. WHEN standard flag names exist (e.g., `-h` for help, `-v` for verbose), THE Hashi SHALL use them
3. WHEN accepting file paths, THE Hashi SHALL support `-` to read from stdin
4. WHEN flags and arguments are provided, THE Hashi SHALL accept them in any order
5. WHEN a flag can accept optional values, THE Hashi SHALL support a special word like "none" for empty values

### Requirement 5: Configuration and Environment Variables

**User Story:** As a user, I want flexible configuration options that respect standard conventions, so that I can customize behavior without complex setup.

#### Acceptance Criteria

1. WHEN configuration files are used, THE Hashi SHALL use TOML format for human-readable configuration
2. WHEN environment variables are checked, THE Hashi SHALL respect standard variables (NO_COLOR, DEBUG, TMPDIR, HOME)
3. WHEN configuration is applied, THE Hashi SHALL use precedence: flags > env vars > project config > user config > system config
4. WHEN a `.env` file exists in the working directory, THE Hashi SHALL read environment variables from it
5. WHEN no `--config` flag is provided, THE Hashi SHALL auto-discover config files in standard locations
6. WHEN `--show-config` is used, THE Hashi SHALL display effective configuration safely without exposing sensitive paths

### Requirement 6: Robustness and Reliability

**User Story:** As a user, I want the tool to handle errors gracefully and provide responsive feedback, so that I feel confident using it for important tasks.

#### Acceptance Criteria

1. WHEN user input is received, THE Hashi SHALL validate it before processing
2. WHEN an operation takes longer than 100ms, THE Hashi SHALL display something to indicate progress
3. WHEN processing large batches, THE Hashi SHALL show progress bars or status updates
4. WHEN Ctrl-C is pressed, THE Hashi SHALL exit immediately after displaying cleanup status
5. WHEN Ctrl-C is pressed during cleanup, THE Hashi SHALL skip remaining cleanup and exit
6. WHEN operations are idempotent, THE Hashi SHALL allow re-running after failures

### Requirement 7: Output Formats and Machine Readability

**User Story:** As a user, I want both human-readable and machine-readable output options, so that I can use the tool interactively and in scripts.

#### Acceptance Criteria

1. WHEN `--json` flag is passed, THE Hashi SHALL output results in valid JSON format
2. WHEN `--plain` flag is passed, THE Hashi SHALL output results in plain tabular format for grep/awk
3. WHEN output is piped, THE Hashi SHALL maintain machine-readable format by default
4. WHEN human-readable output would break machine parsing, THE Hashi SHALL provide separate flags for each format
5. WHEN JSON output is requested, THE Hashi SHALL include all relevant data in structured format

### Requirement 8: File Output Handling

**User Story:** As a user, I want safe file output options that prevent accidental data loss, so that I can save results without worrying about overwriting important files.

#### Acceptance Criteria

1. WHEN `--output-file` is specified and the file exists, THE Hashi SHALL prompt for confirmation before overwriting
2. WHEN `--output-file` is specified with `--force` flag, THE Hashi SHALL overwrite without prompting
3. WHEN `--append` flag is used with `--output-file`, THE Hashi SHALL append to existing file instead of overwriting
4. WHEN `--log-file` is specified and the file exists, THE Hashi SHALL append by default
5. WHEN `--log-json` is specified and the file exists, THE Hashi SHALL append valid JSON entries

### Requirement 9: Documentation and Discoverability

**User Story:** As a user, I want easily discoverable features and comprehensive documentation, so that I can learn the tool's capabilities without trial and error.

#### Acceptance Criteria

1. WHEN help is displayed, THE Hashi SHALL show the most common flags and commands first
2. WHEN examples are provided, THE Hashi SHALL lead with them in help text
3. WHEN web documentation exists, THE Hashi SHALL link to it in help text
4. WHEN a man page exists, THE Hashi SHALL make it accessible via `hashi help` or similar
5. WHEN features are complex, THE Hashi SHALL provide tutorial-style documentation

### Requirement 10: Naming and Distribution

**User Story:** As a user, I want a tool with a memorable name that's easy to install and uninstall, so that I can adopt it without friction.

#### Acceptance Criteria

1. THE Hashi SHALL use only lowercase letters in its command name
2. THE Hashi SHALL be distributed as a single binary when possible
3. WHEN installation instructions are provided, THE Hashi SHALL include uninstallation instructions
4. WHEN the tool name is chosen, THE Hashi SHALL avoid conflicts with existing common commands
5. THE Hashi SHALL be easy to type and remember

### Requirement 11: Future-Proofing and Stability

**User Story:** As a user, I want stable interfaces that don't break my scripts, so that I can rely on the tool long-term.

#### Acceptance Criteria

1. WHEN interfaces change, THE Hashi SHALL warn users before making breaking changes
2. WHEN new features are added, THE Hashi SHALL keep changes additive where possible
3. WHEN output format changes, THE Hashi SHALL only change human-readable output, not machine formats
4. WHEN flags are deprecated, THE Hashi SHALL provide migration paths and warnings
5. THE Hashi SHALL not allow arbitrary abbreviations of flags or subcommands

### Requirement 12: Signal Handling and Control

**User Story:** As a user, I want predictable behavior when interrupting the program, so that I can safely stop operations without corruption.

#### Acceptance Criteria

1. WHEN INT signal (Ctrl-C) is received, THE Hashi SHALL exit as soon as possible
2. WHEN cleanup is in progress and Ctrl-C is pressed again, THE Hashi SHALL skip cleanup and exit immediately
3. WHEN exiting, THE Hashi SHALL display what cleanup actions were skipped if any
4. WHEN the program is interrupted, THE Hashi SHALL leave the system in a recoverable state
5. THE Hashi SHALL expect to be started in situations where cleanup has not been run

### Requirement 13: Exit Codes for Scripting

**User Story:** As a script writer, I want meaningful exit codes, so that I can use hashi in conditional logic and automated workflows.

#### Acceptance Criteria

1. WHEN all files are processed successfully, THE Hashi SHALL exit with code 0
2. WHEN any files fail to process, THE Hashi SHALL exit with code 2
3. WHEN invalid arguments are provided, THE Hashi SHALL exit with code 3
4. WHEN files are not found, THE Hashi SHALL exit with code 4
5. WHEN `--match-required` is set and matches are found, THE Hashi SHALL exit with code 0
6. WHEN `--match-required` is set and no matches are found, THE Hashi SHALL exit with code 1
7. WHEN interrupted by Ctrl-C, THE Hashi SHALL exit with code 130

### Requirement 14: Comparison with Current Implementation

**User Story:** As a developer, I want a detailed comparison of the current implementation against CLI guidelines, so that I can prioritize improvements.

#### Acceptance Criteria

1. THE Review SHALL identify all areas where current implementation deviates from guidelines
2. THE Review SHALL categorize deviations by severity (critical, important, nice-to-have)
3. THE Review SHALL provide specific examples of guideline violations in current code
4. THE Review SHALL suggest concrete improvements for each identified issue
5. THE Review SHALL highlight areas where current implementation already follows best practices

### Requirement 15: Archive Integrity Verification (ZIP CRC32)

**User Story:** As a user, I want to verify ZIP file integrity using embedded CRC32 checksums, so that I can quickly confirm files are not corrupted without computing external hashes.

#### Acceptance Criteria

1. WHEN a user runs `hashi file.zip` with no other flags, THE Hashi SHALL verify the internal CRC32 checksums of all entries in the ZIP file
2. WHEN a user runs `hashi file1.zip file2.zip file3.zip`, THE Hashi SHALL verify all ZIP files and return a single boolean result (exit code 0 if all pass, non-zero if any fail)
3. WHEN a user runs `hashi /path/to/directory` containing only ZIP files, THE Hashi SHALL verify all ZIP files in that directory
4. WHEN verifying ZIP integrity, THE Hashi SHALL use only CRC32 regardless of any metadata suggesting alternative algorithms (security hardening)
5. WHEN a user wants to compute a standard hash of a ZIP file instead of verifying internal checksums, THE Hashi SHALL support a `--raw` flag to treat the file as raw bytes
6. WHEN ZIP verification fails, THE Hashi SHALL report which specific entries failed CRC32 verification
7. THE Hashi SHALL clearly document that CRC32 verification confirms integrity (bits not corrupted) but NOT authenticity (file not tampered with)

### Requirement 16: Boolean Output for Script Integration

**User Story:** As a script author, I want hashi to return simple boolean results in specific scenarios, so that I can easily integrate it into automated pipelines and conditional logic.

#### Acceptance Criteria

1. WHEN verifying archive integrity (ZIP CRC32), THE Hashi SHALL default to boolean output (exit code only, no stdout) for seamless pipeline integration
2. WHEN `--quiet` flag is combined with verification operations, THE Hashi SHALL suppress all stdout and return only exit code
3. WHEN boolean output mode is active, THE Hashi SHALL still output errors to stderr so failures can be diagnosed
4. WHEN a user needs verbose output during verification, THE Hashi SHALL support `--verbose` to override boolean defaults
5. THE Hashi SHALL document which operations default to boolean output and how to get detailed output instead

### Requirement 17: Flag Conflict Detection and Resolution

**User Story:** As a user, I want clear behavior when flags conflict, so that I can understand what hashi will do without memorizing complex interaction rules.

#### Acceptance Criteria

1. WHEN mutually exclusive flags are provided (e.g., `--json` and `--plain`), THE Hashi SHALL reject the command with a clear error explaining the conflict
2. WHEN a file type triggers special behavior (e.g., ZIP verification) but the user wants standard hashing, THE Hashi SHALL support `--raw` to override special handling
3. WHEN a config file could be interpreted as both configuration and input (e.g., JSON file), THE Hashi SHALL treat it as input by default and require explicit `--config` flag for configuration
4. THE Hashi SHALL document all known flag conflicts and their resolutions in help text
5. WHEN new flags are added, THE Review process SHALL include conflict analysis against existing flags

### Requirement 18: Flag Precedence and Override System

**User Story:** As a user, I want predictable flag behavior when multiple output flags are specified, so that I can understand which format will be used.

#### Acceptance Criteria

1. WHEN `--bool` flag is provided with other output flags, THE Hashi SHALL use boolean output and override all other format flags
2. WHEN `--bool` and `--quiet` are both provided, THE Hashi SHALL use boolean output (which implies quiet behavior)
3. WHEN `--quiet` and `--verbose` are both provided, THE Hashi SHALL use quiet mode and suppress verbose output
4. WHEN same-level flags are provided (e.g., `--json` and `--plain`), THE Hashi SHALL use the last flag specified
5. THE Hashi SHALL use a matrix-based conflict resolution system for scalable flag interaction management

### Requirement 18: Educational Code Quality (Moonshot)

**User Story:** As a learner, I want the hashi source code to serve as a teaching tool, so that I can learn Go programming and CLI design by studying a real-world example.

#### Acceptance Criteria

1. THE Hashi source code SHALL include comprehensive comments explaining the purpose of each function and significant code block
2. THE Hashi source code SHALL include comments explaining Go idioms and patterns when first used
3. WHERE complex algorithms are implemented, THE Hashi source code SHALL include comments explaining the algorithm step-by-step
4. THE Hashi project SHALL maintain a separate "annotated edition" branch or documentation with even more detailed explanations (moonshot goal)
5. THE Hashi source code SHALL follow consistent formatting and naming conventions that demonstrate Go best practices

### Requirement 19: Conflict Testing Infrastructure (Moonshot)

**User Story:** As a developer, I want automated testing of flag combinations, so that I can catch unexpected interactions before users encounter them.

#### Acceptance Criteria

1. THE Hashi project SHALL document all known flag conflicts in plain English before implementation begins
2. WHEN a new flag is added or changed, THE Review process SHALL include a conflict review against existing flags
3. THE Hashi project SHALL include a test suite that exercises common flag combinations
4. THE Hashi project MAY include a fuzzing tool that generates random flag combinations and records unexpected behaviors (moonshot goal)
5. THE Hashi project SHALL have a "conflict hunt" phase before major releases to identify missed interactions

### Requirement 20: Security-Conscious Design

**User Story:** As a security-minded user, I want hashi to be designed with security in mind from the start, so that I can trust it in automated verification workflows.

#### Acceptance Criteria

1. WHEN returning boolean results for verification, THE Hashi SHALL NOT introduce threat surfaces larger than or different from existing accepted tools
2. THE Hashi SHALL clearly distinguish between integrity verification (data not corrupted) and authenticity verification (data not tampered with)
3. WHEN special file type handling is implemented (e.g., ZIP CRC32), THE Hashi SHALL ignore any metadata that could redirect to different algorithms (preventing algorithm substitution attacks)
4. THE Hashi SHALL document its security model and threat considerations in user-facing documentation
5. THE Hashi project SHALL evaluate security implications of any feature that returns boolean results for automated verification

### Requirement 21: Hash String Detection and Algorithm Identification

**User Story:** As a user, I want hashi to automatically detect hash algorithms from hash strings, so that I can verify files without manually specifying the algorithm.

#### Acceptance Criteria

1. WHEN a hash string is provided, THE Hashi SHALL validate it contains only hexadecimal characters (0-9, a-f, A-F)
2. WHEN a valid hex string is provided, THE Hashi SHALL identify possible algorithms based on string length:
   - 8 characters → CRC32
   - 32 characters → MD5
   - 40 characters → SHA-1
   - 64 characters → SHA-256
   - 128 characters → SHA-512 or BLAKE2b-512 (ambiguous)
3. WHEN a hash length matches the current algorithm, THE Hashi SHALL use it silently
4. WHEN a hash length matches a different algorithm, THE Hashi SHALL return an error with a helpful suggestion (e.g., "This looks like MD5. Try: hashi --algo md5 file.txt [hash]")
5. WHEN a hash length is ambiguous (e.g., 128 chars could be SHA-512 or BLAKE2b), THE Hashi SHALL list all possibilities and suggest specifying with --algo
6. WHEN validating a hash string only (no file), THE Hashi SHALL detect and report all possible algorithms

### Requirement 22: Argument Classification (Files vs Hash Strings)

**User Story:** As a user, I want hashi to automatically distinguish between file paths and hash strings in arguments, so that I can use intuitive command syntax like `hashi file.txt [hash]`.

#### Acceptance Criteria

1. WHEN classifying arguments, THE Hashi SHALL check filesystem existence FIRST (files take precedence over hash strings)
2. WHEN an argument exists as a file, THE Hashi SHALL treat it as a file path regardless of whether it looks like a hash
3. WHEN an argument does not exist as a file AND matches a valid hash format, THE Hashi SHALL treat it as a hash string
4. WHEN an argument does not exist as a file AND does not match a valid hash format, THE Hashi SHALL treat it as a file path (will error later if not found)
5. WHEN a hash string is provided but doesn't match the current algorithm, THE Hashi SHALL return an error with algorithm suggestion before processing begins
6. THE Hashi SHALL normalize hash strings to lowercase before comparison

### Requirement 23: Configuration Auto-Discovery

**User Story:** As a user, I want hashi to automatically load configuration from standard locations, so that I can set default preferences without specifying --config every time.

#### Acceptance Criteria

1. WHEN no --config flag is provided, THE Hashi SHALL search for config files in this order:
   - `./.hashi.toml` (project-specific, highest priority)
   - `$XDG_CONFIG_HOME/hashi/config.toml` (XDG standard)
   - `~/.config/hashi/config.toml` (XDG fallback)
   - `~/.hashi/config.toml` (traditional dotfile)
2. WHEN multiple config files exist, THE Hashi SHALL use the first one found (highest priority)
3. WHEN --config is explicitly provided, THE Hashi SHALL use that file and skip auto-discovery
4. WHEN a config file is auto-loaded, THE Hashi SHALL apply settings with proper precedence (flags > env vars > config file > defaults)
5. WHEN XDG_CONFIG_HOME or HOME environment variables are empty, THE Hashi SHALL skip those locations gracefully
6. THE Hashi SHALL document config file locations and format in help text

### Requirement 24: Hash String Validation Mode

**User Story:** As a user, I want to validate hash string format without a file, so that I can check if a hash string is valid and identify its algorithm.

#### Acceptance Criteria

1. WHEN a user runs `hashi [hash_string]` with no files, THE Hashi SHALL validate the hash format and report results
2. WHEN the hash string is valid, THE Hashi SHALL display which algorithm(s) it could be
3. WHEN the hash string is invalid (wrong characters or length), THE Hashi SHALL display an error explaining why
4. WHEN the hash length is ambiguous (e.g., 128 chars), THE Hashi SHALL list all possible algorithms
5. THE Hashi SHALL exit with code 0 for valid hash format, code 3 for invalid format

### Requirement 25: File and Hash Comparison Mode

**User Story:** As a user, I want to compare a file's hash against a provided hash string, so that I can verify file integrity with a simple command.

#### Acceptance Criteria

1. WHEN a user runs `hashi file.txt [hash_string]`, THE Hashi SHALL compute the file's hash and compare it to the provided hash
2. WHEN the hashes match, THE Hashi SHALL display "PASS" (or equivalent) and exit with code 0
3. WHEN the hashes don't match, THE Hashi SHALL display "FAIL" with both expected and computed hashes, and exit with code 1
4. WHEN the hash string doesn't match the current algorithm, THE Hashi SHALL error with algorithm suggestion before computing
5. WHEN multiple files and multiple hashes are provided, THE Hashi SHALL reject with error "Cannot compare multiple files with hash strings. Use one file at a time."
6. WHEN stdin marker (-) and hash strings are both provided, THE Hashi SHALL reject with error "Cannot use stdin input with hash comparison"

### Requirement 26: Boolean Output Flag

**User Story:** As a script author, I want a --bool flag that outputs only "true" or "false", so that I can easily capture verification results in scripts.

#### Acceptance Criteria

1. WHEN --bool flag is provided with file+hash comparison, THE Hashi SHALL output only "true" or "false" to stdout
2. WHEN --bool flag is provided, THE Hashi SHALL still use appropriate exit codes (0 for match, 1 for no match)
3. WHEN --bool and --quiet are both provided, THE Hashi SHALL reject with error "--bool and --quiet are mutually exclusive"
4. WHEN --bool and --format are both provided, THE Hashi SHALL reject with error "--bool cannot be used with --format"
5. THE Hashi SHALL document --bool flag in help text under OUTPUT FORMATS section

### Requirement 27: Config Command Error Handling

**User Story:** As a user, I want helpful guidance when I try to use a config subcommand, so that I understand how to configure hashi properly.

#### Acceptance Criteria

1. WHEN a user runs `hashi config` or `hashi config [subcommand]`, THE Hashi SHALL reject the command with a helpful error message
2. WHEN the config command error is displayed, THE Hashi SHALL explain that config changes must be made by manually editing config files
3. WHEN the config command error is displayed, THE Hashi SHALL list the config file locations that hashi auto-loads
4. WHEN the config command error is displayed, THE Hashi SHALL provide a link to documentation about configuration
5. THE Hashi SHALL exit with code 3 (invalid arguments) when config command is attempted

### Requirement 28: Advanced Filtering

**User Story:** As a user, I want to filter files by size, date, and patterns before processing, so that I can focus on specific subsets of files without manual pre-filtering.

#### Acceptance Criteria

1. WHEN `--include` flag is provided with a pattern, THE Hashi SHALL process only files matching that pattern
2. WHEN `--exclude` flag is provided with a pattern, THE Hashi SHALL skip files matching that pattern
3. WHEN `--min-size` flag is provided, THE Hashi SHALL process only files larger than or equal to the specified size
4. WHEN `--max-size` flag is provided, THE Hashi SHALL process only files smaller than or equal to the specified size
5. WHEN `--modified-after` flag is provided with a date, THE Hashi SHALL process only files modified after that date
6. WHEN `--modified-before` flag is provided with a date, THE Hashi SHALL process only files modified before that date
7. WHEN multiple filter flags are provided, THE Hashi SHALL apply all filters (AND logic)
8. WHEN `--include` and `--exclude` both match a file, THE Hashi SHALL exclude the file (exclude takes precedence)
9. THE Hashi SHALL support multiple patterns in `--include` and `--exclude` flags (comma-separated or multiple flag instances)

### Requirement 29: Dry Run and Preview Mode

**User Story:** As a user, I want to preview what files would be processed without actually computing hashes, so that I can verify my filters and estimate processing time before running expensive operations.

#### Acceptance Criteria

1. WHEN `--dry-run` flag is provided, THE Hashi SHALL enumerate files that would be processed without computing hashes
2. WHEN dry run mode is active, THE Hashi SHALL display the total number of files that would be processed
3. WHEN dry run mode is active, THE Hashi SHALL display the total size of files that would be processed
4. WHEN dry run mode is active, THE Hashi SHALL provide an estimated processing time based on file sizes
5. WHEN dry run mode is active with filters, THE Hashi SHALL show which filters are applied
6. WHEN dry run mode is active, THE Hashi SHALL exit with code 0 (no actual processing errors can occur)
7. THE Hashi SHALL apply all filters during dry run to show accurate preview

### Requirement 30: Incremental Operations

**User Story:** As a CI/CD engineer, I want to process only files that have changed since the last run, so that I can dramatically reduce processing time for large codebases.

#### Acceptance Criteria

1. WHEN `--manifest` flag is provided with a previous manifest file, THE Hashi SHALL compare current files against the manifest
2. WHEN `--only-changed` flag is provided with a manifest, THE Hashi SHALL process only files that are new, modified, or missing from the manifest
3. WHEN comparing against a manifest, THE Hashi SHALL use file modification time and size for change detection
4. WHEN `--output-manifest` flag is provided, THE Hashi SHALL save current state to a manifest file for future incremental runs
5. WHEN a manifest file is specified but does not exist, THE Hashi SHALL process all files and create a new manifest
6. WHEN manifest format is invalid, THE Hashi SHALL reject with a clear error message
7. THE Hashi SHALL support JSON format for manifest files with file path, hash, size, and modification time

### Requirement 31: Enhanced File Output Safety

**User Story:** As a user, I want comprehensive file output safety features, so that I never accidentally lose important data when saving results.

#### Acceptance Criteria

1. WHEN `--output-file` is specified and the file exists without `--force`, THE Hashi SHALL prompt for confirmation before overwriting
2. WHEN `--output-file` is specified with `--force`, THE Hashi SHALL overwrite without prompting
3. WHEN `--append` flag is used with `--output-file`, THE Hashi SHALL append to the existing file instead of overwriting
4. WHEN writing to an output file, THE Hashi SHALL use atomic writes (write to temp file, then rename)
5. WHEN `--log-file` is specified, THE Hashi SHALL append by default without prompting
6. WHEN `--log-json` is specified with an existing file, THE Hashi SHALL append valid JSON entries maintaining array structure
7. WHEN output file write fails, THE Hashi SHALL preserve the original file and report the error clearly
8. THE Hashi SHALL validate output file path is writable before processing begins
