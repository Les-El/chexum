# ADR 002: ZIP File Default Behavior

## Context
Initially, `hashi` was designed to default to internal archive integrity verification (CRC32) when a `.zip` file was provided as an argument. This was intended as a "smart" convenience feature.

## Problem
During code review, it was identified that this behavior violates the Principle of Least Astonishment (POLA). A tool named `hashi` is expected to return the hash of the file itself by default. Defaulting to internal verification makes the tool's behavior inconsistent across different file types and complicates scripting use cases where a uniform output format is expected.

## Decision
1.  **Standardize Default Behavior:** Hashing a `.zip` file will now compute the standard cryptographic hash (e.g., SHA-256) of the archive file itself, consistent with all other file types.
2.  **Explicit Verification:** A new flag `--verify` (no shorthand to avoid conflicts) is introduced to trigger deep integrity verification of supported archive formats.
3.  **Keep Escape Hatch:** The `--raw` flag is retained as a global override to bypass any future "smart" or type-specific processing.

## Rationale
- **Predictability:** Users and scripts get consistent behavior regardless of file extension.
- **Explicit Intent:** The user explicitly requests "deep" verification, avoiding confusion when a file hash is expected for distribution verification.
- **Scalability:** The `--verify` flag can be extended to support other formats (TAR, PDF, EXE) without adding more "magic" defaults.

## Alternatives Considered
- **Keep Smart Default:** Rejected because it prioritizes convenience over predictability and breaks composability.
- **Use `--meta` or `--checksum`:** Rejected in favor of `--verify` which more clearly communicates the action of checking internal integrity.

## Status
Accepted and Implemented.
