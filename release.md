# hashi v1.0.1 — Release Notes

## 🚀 Overview
This release introduces significant new functionality for verifying ZIP file integrity and a robust new internal engine for handling complex command-line flags. We have also resolved a critical configuration precedence bug.

## ✨ New Features

### 📦 ZIP Integrity Verification (`--verify`)
Hashi can now look *inside* ZIP archives to verify their internal data integrity.
- **Action:** Uses embedded CRC32 checksums to ensure no bits were corrupted.
- **Security Hardening:** Always uses the CRC32 algorithm regardless of archive metadata to prevent algorithm-substitution attacks.
- **Smart Output:** Defaults to "Boolean Mode" (exit code only) for clean use in shell scripts.
- **Escape Hatch:** Use the `--raw` flag if you want to hash the ZIP file itself as a single unit.

### 🛣️ The "Pipeline of Intent" (Conflict Resolution)
We have completely rewritten the flag handling logic to be more predictable.
- **Last One Wins:** If you specify multiple output formats, the last one you typed takes precedence.
- **Mode Overrides:** The new `--bool` mode overrides other formatting flags, ensuring your scripts always get a simple `true`/`false`.
- **Clear Errors:** Incompatible modes (like `--raw` and `--verify`) now trigger helpful error messages.

### 🧪 Improved Scripting Support
- **Boolean Mode (`--bool` / `-b`):** Designed specifically for `if/then` logic in scripts.
- **Exit Code 6:** New dedicated exit code for Archive Integrity Failures.

## 🛠️ Critical Bug Fixes
- **Configuration Precedence:** Fixed an issue where environment variables would incorrectly override explicit command-line flags when the flag was set to its default value. **Flags now strictly override everything else.**

## 📦 Binary Downloads
This release includes pre-compiled, static binaries for:
- **Linux:** x86-64 & ARM64
- **macOS:** Apple Silicon (M-series) & Intel
- **Windows:** x86-64 & ARM64

## 📖 Quick Examples
```bash
# Verify a ZIP file and use the result in a script
if hashi --verify backup.zip; then
  echo "Backup is healthy!"
fi

# Compare a file to a hash and get a simple true/false
hashi -b file.txt e3b0c442... 

# Cross-compile for a friend on a Mac M1 from your Linux machine
GOOS=darwin GOARCH=arm64 go build -o hashi-mac ./cmd/hashi
```
