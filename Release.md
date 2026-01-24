# Release v0.0.22

This release completes the major project stabilization phase, bringing comprehensive testing, quality improvements, and infrastructure refinements to the hashi project.

## ✨ Features
- **Comprehensive Testing**: Added 100% coverage for the configuration system, new CLI integration tests, and property-based tests for core hashing and conflict resolution logic.
- **Quality Infrastructure**: Enhanced the quality engine to support "Reviewed: LONG-FUNCTION" markers, allowing for intentional deviations from standards where appropriate.
- **Improved Test Discovery**: Optimized the testing battery to be package-aware, improving analysis speed and accuracy.
- **Benchmarking**: Added performance benchmarks for hashing operations to prevent future regressions.

## 🐛 Bug Fixes
- Resolved various quality issues identified during the stabilization analysis.
- Improved error handling and validation in the configuration parsing logic.

## 📦 Available Binaries (dist/)

| Platform | Architecture | Filename |
| :--- | :--- | :--- |
| **Linux** | amd64 | `hashi-linux-amd64` |
| **Linux** | arm64 | `hashi-linux-arm64` |
| **Windows** | amd64 | `hashi-windows-amd64.exe` |
| **Windows** | arm64 | `hashi-windows-arm64.exe` |
| **macOS** | amd64 | `hashi-darwin-amd64` |
| **macOS** | arm64 | `hashi-darwin-arm64` |

## 🚀 Installation

### Via Go (Recommended)
```bash
GOPROXY=direct go install github.com/Les-El/hashi/cmd/hashi@latest
```

### Via Binary Download
1. Download the appropriate binary for your system from the dist/ directory.
2. (Linux/macOS) Make the binary executable: `chmod +x hashi-<platform>-<arch>`
3. Move the binary to a directory in your system's PATH (e.g., `/usr/local/bin/hashi`).

---
*Verified and Signed Off for Linux production release on 2026-01-24.*
