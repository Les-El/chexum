# hashi Backlog: Ideas and Refinements

This document tracks future improvements, "leftover" ideas from initial development, and refinements for subsequent versions of `hashi`.

## High Priority Refinements
- **Parallel Processing**: Implement concurrent hashing for multi-core speedup on large batches.
- **Progress Bar Polish**: Improve progress bar behavior when multiple files fail sequentially to avoid flickering.
- **Detailed Flag Reference**: Complete `docs/user/command-reference.md`.

## Medium Priority Ideas
- **Custom JSON Envelopes**: Allow users to provide a template for JSON output metadata.
- **URL Support**: Allow hashing files directly from HTTPS URLs.
- **Archive Deep Scan**: Option to hash individual files *inside* ZIP archives without extracting to disk.

## Moonshot Goals
- **Watch Mode**: A background daemon that monitors a directory and updates a manifest in real-time.
- **Web UI**: A lightweight local web interface for visualizing duplicate groups and manifest changes.
- **Cryptographic Signatures**: Support for signing manifest files to verify authenticity.

## Technical Debt / Leftovers
- **Deprecation Framework**: Standardize how we warn users about upcoming flag changes.
- **Test Optimization**: Reduce execution time of property tests in `internal/signals`.
- **Man Page Integration**: Automate the generation of the man page from help text metadata.
