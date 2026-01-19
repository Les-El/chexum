# Gemini Context: hashi Project

## Project Overview

`hashi` is a modern, unified command-line tool for file hashing and verification, written in Go. It aims to replace traditional utilities like `md5sum` and `sha256sum` by offering a human-centric interface, robust configuration, and advanced features like automatic algorithm detection and archive verification.

### Core Principles
1.  **Developer Continuity:** Every function and decision is documented to ensure any developer can pick it up the project. Intent is prioritized over implementation details.
2.  **User-First Design:** Features and defaults are designed around user needs and expectations (e.g., colorized output, progress bars, helpful error suggestions).
3.  **No Lock-Out:** "Smart" features always have escape hatches.

## Key Technologies & Libraries
- **Language:** Go 1.24+
- **Flag Parsing:** `github.com/spf13/pflag` (POSIX-compliant).
- **Configuration:** `github.com/BurntSushi/toml` (TOML format support).
- **UI/UX:** `github.com/fatih/color` (Colorized output), `github.com/schollz/progressbar/v3` (Progress indicators).
- **Crypto:** `golang.org/x/crypto` (BLAKE2b support).

## Project Structure

- **`cmd/hashi/`**: CLI entry point and high-level execution flow.
- **`internal/`**: Core application logic.
    - `archive/`: ZIP verification using CRC32.
    - `color/`: TTY-aware color handling and `NO_COLOR` support.
    - `config/`: Argument parsing, precedence logic, and auto-discovery.
    - `conflict/`: Matrix-based flag conflict resolution.
    - `console/`: Split stream management (Stdout for data, Stderr for context).
    - `errors/`: User-friendly error formatting and exit code mapping.
    - `hash/`: Core hashing algorithms and streaming file processing.
    - `output/`: Formatters (Default, Verbose, JSON, Plain, Boolean).
    - `progress/`: Progress bar implementation.
    - `security/`: Path validation and input sanitization.
    - `signals/`: Graceful handling of SIGINT (Ctrl-C).

## Building and Running

### Commands
- **Build:** `go build -o hashi ./cmd/hashi`
- **Run:** `go run cmd/hashi/main.go [flags] [files]`
- **Test:** `go test ./...`

### Configuration Precedence
Settings are applied in the following order (highest to lowest):
1.  **Command-line flags**
2.  **Environment variables** (`HASHI_*`)
3.  **Configuration file** (TOML)
4.  **Built-in defaults** (SHA-256)

Auto-discovery searches for: `./.hashi.toml`, `$XDG_CONFIG_HOME/hashi/config.toml`, `~/.config/hashi/config.toml`, and `~/.hashi/config.toml`.

## Development Conventions

- **Split Streams:** Data (hashes, results) goes to **Stdout**. Context (progress bars, warnings, errors) goes to **Stderr**.
- **Error Handling:** Errors must be actionable. Security-sensitive errors are sanitized unless `--verbose` is used.
- **Formatting:** Adhere to `gofmt`.
- **Modularity:** Keep hashing logic separate from UI and configuration.
- **Documentation:** Maintain `hashi_features.md`, `hashi_help_screen.md`, and ADRs in `docs/`.

## Current Status (from `tasks.md`)
- Core hashing, output formatting, and configuration are complete.
- ZIP archive verification (`internal/archive`) is implemented and wired into `main.go` via the explicit `--verify` flag.
- Documentation polish and man page generation are in progress.
