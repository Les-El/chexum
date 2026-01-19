# hashi - Technical Specification

This document details the features, interface, and design specifications for `hashi`.

## Overview
`hashi` is a modern CLI tool for file hashing and verification. It aims to replace multiple standard utilities (`md5sum`, `sha256sum`, etc.) with a unified, human-first interface.

---

## Core Functionality

### 1. Default Behavior
*   **Target:** All non-hidden files in the current directory.
*   **Algorithm:** SHA-256.
*   **Recursion:** Disabled by default.
*   **Output:** Concise summary.

### 2. Modes of Operation
*   **Hashing:** Compute hashes for provided files.
*   **Verification:** Compare a file against a provided hash string.
*   **Comparison:** Compare two files or directory trees.
*   **Validation:** Check if a string is a valid hash for a given algorithm.

### 3. Supported Algorithms
*   SHA-256 (Default)
*   MD5
*   SHA-1
*   SHA-512
*   Blake2b / Blake2s

---

## Command Line Interface (CLI)

### Usage
`hashi [OPTIONS] [FILE_OR_HASH...]`

### General Options
*   `-h`, `--help`: Display usage information.
*   `-V`, `--version`: Display version information.
*   `-v`, `--verbose`: Enable detailed output.
*   `-q`, `--quiet`: Suppress all commentary and non-result output.
*   `-b`, `--bool`: Boolean output mode (returns `true`/`false`).

### File Selection & Traversal
*   `-r`, `--recursive`: Recursively process subdirectories.
*   `-H`, `--hidden`: Include hidden files and directories.
*   `--include <patterns>`: Glob patterns to include (e.g., `"*.txt"`).
*   `--exclude <patterns>`: Glob patterns to exclude.

### Filtering
*   `--min-size`, `--max-size`: Filter by file size (e.g., `10MB`).
*   `--modified-after`, `--modified-before`: Filter by modification date.

### Configuration
*   `--algo <algorithm>`: Explicitly set the hashing algorithm.
*   `--config <file_path>`: Load flags and arguments from a JSON or text file.

### Output & Logging
*   `--output-format <summary|verbose|json>`: Choose the output format.
*   `--output-file <file_path>`: Write result to a file.
*   `--log-file <file_path>`, `--log-json <file_path>`: Write operational logs.

---

## Implementation Details

### Auto-Algorithm Detection
`hashi` detects the intended algorithm based on the length of provided hash strings:
*   32 chars: MD5
*   40 chars: SHA-1
*   64 chars: SHA-256
*   128 chars: SHA-512

### Visuals
*   **Colorized Output:** Green for success, red for failure, yellow for warnings (TTY only).
*   **Progress Indicators:** Spinners for directory traversal, progress bars for large files (>100MB).

### Safety
*   **Confirmation Prompt:** Required for operations involving a massive number of files (>10,000) or very large files (>10GB cumulative) to prevent accidental system hangs.

---

## Exit Codes
*   `0`: Success / Match found.
*   `1`: Mismatch / Error encountered.
*   `2`: Invalid arguments or configuration.
*   `130`: Interrupted by user (SIGINT).
