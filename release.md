# hashi v1.0.2 — Installation Fix

## 🚀 Overview
This is a quick patch release to fix an issue preventing the tool from being installed directly via `go install`.

## 🛠️ Fixes
- **Module Path Alignment**: The module is now correctly declared as `github.com/Les-El/hashi`. This resolves the "module declares its path as: hashi" error when installing from GitHub.

## 📦 Installation
You can now install the latest version directly using:
```bash
go install github.com/Les-El/hashi/cmd/hashi@latest
```

---
*For more details on the features introduced in v1.0.1 (Archive Verification, Boolean Mode, etc.), please see the previous release notes.*