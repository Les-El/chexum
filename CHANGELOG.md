# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.12] - 2026-01-20

### Fixed
- Fixed a bug that caused `hashi` to fail silently when installed via `go install` and run with no arguments in a directory.

### Changed
- Improved and clarified installation, uninstallation, and configuration instructions in `README.md`.
- Consolidated duplicate "Environment Variables" and "Security Note" sections in `README.md`.

## [1.0.11] - 2026-01-18

### Fixed
- **CRITICAL**: Fixed configuration precedence violation where explicit default flags were incorrectly overridden by environment variables
- Explicit flags now always override environment variables, even when flag value equals built-in default
- Corrected logic for file/hash classification to prevent misinterpretation of arguments

### Changed
- Refined conflict resolution logic for output format flags (`--json`, `--plain`, `--verbose`, etc.)
- Improved error message for unknown flags with "Did you mean...?" suggestions
- Reorganized `main.go` into distinct modes for clarity and robustness
- Enhanced `README.md` with more detailed sections on configuration precedence and troubleshooting

### Added
- `--match-required` flag to exit with status 0 only if matches are found
- Support for reading file paths from stdin using `-`
- `CHANGELOG.md` to track project history

## [1.0.0] - 2026-01-15

### Added
- Initial release of `hashi`
- Core hashing functionality for multiple algorithms (SHA-256, SHA-512, MD5, etc.)
- File and directory processing (recursive and non-recursive)
- Multiple output formats (default, verbose, JSON, plain)
- Configuration file support (`.hashi.toml`)
- Environment variable support
- Basic error handling and exit codes
- Colorized output with TTY detection
- Progress bar for long operations
- Security features (path validation, sensitive file protection)
- Comprehensive `README.md` with usage and installation instructions
