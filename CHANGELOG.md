# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed
- **CRITICAL**: Fixed configuration precedence violation where explicit default flags were incorrectly overridden by environment variables
  - Example: `export HASHI_ALGORITHM=md5; hashi --algorithm=sha256 file.txt` now correctly uses SHA256 instead of MD5
  - Affects all flags: --algorithm, --format, --verbose, --quiet, --recursive, --hidden, --preserve-order
  - Explicit flags now always override environment variables, even when flag value equals built-in default
  - Maintains correct precedence hierarchy: flags > env vars > config files > defaults

## [1.0.0] - 2026-01-10
- Initial release — first stable public version.

### Notes

- Adds core hashi utilities and CLI entrypoint
- Basic tests and examples included
