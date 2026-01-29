# hashi User Documentation

Welcome to the hashi user documentation. This guide will help you get the most out of hashi, from basic usage to advanced scripting and CI/CD integration.

## Documentation Index

### Core Guides
- **[getting-started.md](getting-started.md)** ✅ - Installation and your first hash
- **[examples.md](examples.md)** - Common usage patterns and recipes
- **[command-reference.md](command-reference.md)** ✅ - Complete flag and option reference
- **[flags-and-arguments.md](flags-and-arguments.md)** ✅ - Quick lookup for flags and arguments

### Feature Deep Dives
- **[filtering.md](filtering.md)** - Detailed guide to include/exclude patterns, size, and date filters
- **[incremental.md](incremental.md)** - How to use manifests for high-performance hashing in CI/CD
- **[dry-run.md](dry-run.md)** - Previewing operations and estimating time
- **[output-formats.md](output-formats.md)** - Understanding JSON, JSONL, and plain text output (Coming soon)

### Automation and Configuration
- **[scripting.md](scripting.md)** - Integrating hashi into bash, PowerShell, and more
- **[configuration.md](configuration.md)** - Managing config files and environment variables (Coming soon)

### Troubleshooting
- **[error-handling.md](error-handling.md)** ✅ - Troubleshooting and understanding error messages
- **[test-space-management.md](test-space-management.md)** - Managing disk space during large test runs

## Key Features

### Human-First Design
hashi is designed to be intuitive. It uses colorized output when a TTY is detected, provides progress bars for long-running operations, and offers helpful suggestions when you make a typo.

### Security First
hashi follows the principle that "hashi can't change hashi". It automatically blocks writing output to sensitive system or configuration files and uses obfuscated error messages to prevent information leakage.

### Machine-Friendly
While hashi is great for humans, it's even better for machines. With support for JSON, JSONL, and consistent exit codes, hashi is the perfect companion for your automation scripts.

### Performance at Scale
With advanced filtering and incremental hashing via manifests, hashi can handle codebases with hundreds of thousands of files efficiently.

## Getting Help

If you're ever stuck, remember:
- `hashi --help` for a quick flag reference
- `hashi --verbose` for detailed error information
- Check the [error-handling.md](error-handling.md) guide for common solutions
