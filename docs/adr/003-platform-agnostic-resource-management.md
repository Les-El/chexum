# ADR 003: Platform-Agnostic Resource Management

## Status
Accepted

## Context
The project initially focused on Linux development, leading to the use of Linux-specific terminology and implementations for resource monitoring (e.g., `tmpfs`). This hindered cross-platform reliability, particularly on Windows, where monitoring was disabled via stubs.

## Decision
We have refactored the resource management system to be platform-agnostic:
1.  **Terminology**: All references to `tmpfs` have been replaced with `Storage` or `Resource`.
2.  **Implementation**: Platform-specific logic is now isolated in `storage_unix.go` and `storage_windows.go`.
3.  **Pathing**: The system now uses `os.TempDir()` instead of hardcoding `/tmp`.
4.  **Windows Support**: A real implementation for Windows has been added using the `golang.org/x/sys/windows` package.

## Consequences
- The system now accurately reports storage usage across all supported platforms.
- Documentation is aligned with the actual code structure.
- Future developers can easily understand the resource management system regardless of their OS.
- Reduced reliance on Linux-specific assumptions improves overall project stability.
