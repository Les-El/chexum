# Refactoring Temporary File Management: Afero + Workspace Pattern

## Status
**Proposed**

## Context
The current `CleanupManager` implementation uses a "shattershot" approach to manage temporary files. It scans the system-wide `/tmp` directory for files matching specific prefixes (`hashi-*`, `test-*`, etc.) and deletes them based on pattern matching. 

### Problems with the Current Approach:
1.  **Safety**: Scanning a shared system directory like `/tmp` and deleting based on prefixes is risky. There is a small but non-zero chance of name collisions with other applications.
2.  **Universal Support**: The current implementation relies on platform-specific logic (e.g., `syscall.Statfs`) to check disk usage, making cross-compilation and universal support more complex.
3.  **Efficiency**: It requires constant disk I/O to check for leaks and manage state, even for small metadata artifacts that could live in RAM.
4.  **Ownership**: The app doesn't truly "own" its temporary files; it just hopes it can find them later based on their names.

## Decision
We will refactor the temporary resource management by merging two established patterns:

### 1. Filesystem Abstraction (Afero)
We will integrate `github.com/spf13/afero` as our filesystem layer. 
- **Benefits**: It allows us to swap between `OsFs` (real disk) and `MemMapFs` (RAM) seamlessly. 
- **Usage**: Analysis engines will write their temporary findings to a virtual filesystem. Small artifacts will stay in RAM, while larger ones can be flushed to disk if needed.

### 2. The Workspace Pattern
Instead of scattered files in `/tmp`, we will adopt a "Managed Workspace" approach:
- **Root Isolation**: On startup, the application creates a single, uniquely named root directory (e.g., `os.TempDir()/hashi-workspace-<unique-id>`).
- **Strict Ownership**: All temporary artifacts created by the app **must** reside within this directory.
- **Deterministic Cleanup**: Cleanup becomes a simple `os.RemoveAll(workspaceRoot)`. The app no longer needs to search for files; it simply deletes its own container.

## Consequences
- **Improved Safety**: We will never touch a file outside of our own uniquely named workspace.
- **True Portability**: By using Go's standard `os.TempDir()` and Afero's abstraction, the tool will behave identically on Linux, Windows, and macOS.
- **Clean Architecture**: Engines will no longer need to know about the system's `/tmp` structure; they will simply be handed an `afero.Fs` to work within.
- **Simplified Testing**: We can run the entire analysis suite in memory using `afero.NewMemMapFs()`, making tests faster and ensuring zero disk pollution.

## Implementation Steps
1.  Initialize `github.com/spf13/afero` in the project.
2.  Implement a `Workspace` struct that manages the lifecycle of the temporary root.
3.  Refactor `CleanupManager` to focus on Workspace disposal rather than global pattern matching.
4.  Update Analysis Engines to utilize the Workspace abstraction for all temporary storage.
