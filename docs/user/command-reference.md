# Command Reference

This document provides a complete reference for all command-line flags available in `hashi`.

## Core Flags

### `--recursive`, `-r`
Recursively traverse directories.
- **Default**: false

### `--hidden`
Include hidden files and directories in the analysis.
- **Default**: false

### `--algorithm`, `-a`
Specify the hashing algorithm to use (`sha256`, `sha1`, `md5`, `sha512`, `blake2b`).
- **Default**: `sha256`

### `--config`, `-c`
Path to a configuration file.

### `--help`, `-h`
Show help text and exit.

### `--version`, `-V`
Display the version of `hashi` and exit.

## Filtering Flags

### `--include`, `-i`
Glob patterns to include. Can be specified multiple times.

### `--exclude`, `-e`
Glob patterns to exclude. Can be specified multiple times.

### `--min-size`
Minimum file size (e.g., 100KB, 1MB).

### `--max-size`
Maximum file size (e.g., 1GB).

### `--modified-after`
Only process files modified after this date (YYYY-MM-DD).

### `--modified-before`
Only process files modified before this date (YYYY-MM-DD).

## Output Control

### `--quiet`, `-q`
Suppresses all non-essential output. Only critical errors will be displayed.
- **Default**: false

### `--verbose`, `-v`
Enable verbose logging.

### `--bool`, `-b`
Output only a boolean result (`true` or `false`) indicating success or failure.
- **Default**: false

### `--format`, `-f`
Specify the output format (`default`, `verbose`, `json`, `jsonl`, `plain`).
- **Default**: `default`

### `--json`
Shortcut for `--format json`.

### `--jsonl`
Shortcut for `--format jsonl`.

### `--plain`
Shortcut for `--format plain`.

### `--output`, `-o`
Write output results to the specified file.

### `--append`
Append results to the specified output file instead of overwriting it.

### `--force`
Overwrite files without prompting.

### `--log-file`
File for logging context and errors.

### `--log-json`
File for JSON formatted logging.

## Advanced & Incremental Behavior

### `--preserve-order`
Ensure that file discovery and processing maintain alphabetical order.
- **Default**: false

### `--match-required`
Exit 0 only if matches were found.

### `--manifest`
Baseline manifest for incremental operations.

### `--only-changed`
Only process files that have changed relative to the manifest.

### `--output-manifest`
Save the results as a new manifest file.
