# hashi - User Examples Guide

This document provides detailed examples for every capability of the `hashi` tool, demonstrating how to use its features to solve real-world problems.

---

## 1. Basic Usage

### 1.1 Hash Current Directory
**Scenario:** You want to quickly check the hashes of all files in your current working folder.
**Command:**
```bash
hashi
```
**Output:**
```text
[SHA-256] Computed hashes for 3 files:
e3b0c442...  notes.txt
a1b2c3d4...  image.png
88d4266f...  data.csv
```
**Explanation:** Running `hashi` without arguments processes all visible files in the current directory using the default SHA-256 algorithm.

### 1.2 Hash a Single File
**Scenario:** You need the hash of a specific ISO file to share with a friend.
**Command:**
```bash
hashi ubuntu-24.04.iso
```
**Output:**
```text
[SHA-256] ubuntu-24.04.iso
9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```
**Explanation:** `hashi` calculates the SHA-256 hash for the specified file and displays it clearly.

### 1.3 Verify a File Integrity (Pass)
**Scenario:** You downloaded a file and want to verify it matches the hash provided by the website.
**Command:**
```bash
hashi installer.zip a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2
```
**Output:**
```text
[SHA-256] Verifying installer.zip...
✅ PASS: Hash matches provided string.
```
**Explanation:** You provided both a filename and a hash string. `hashi` automatically detected this, computed the file's hash, compared it to your string, and gave a simple "PASS" result.

### 1.4 Verify a File Integrity (Fail)
**Scenario:** Same as above, but the file was corrupted during download.
**Command:**
```bash
hashi installer.zip a1b2c3d4e5f6... (expected hash)
```
**Output:**
```text
[SHA-256] Verifying installer.zip...
🔴 FAIL: Hash mismatch!
   Expected: a1b2c3d4e5f6...
   Computed: f0e1d2c3b4a5...
```
**Explanation:** `hashi` detected that the computed hash did not match your input, alerting you to potential corruption or tampering.

### 1.5 Validate a Hash String
**Scenario:** You have a hash string and want to check if it's a valid format.
**Command:**
```bash
hashi a1b2c3d4...
```
**Output:**
```text
✅ Valid SHA-256 hash format.
```
**Explanation:** Since no file matched the argument, `hashi` checked if it was a valid hash string format.

---

## 2. Archive Verification

### 2.1 Verify ZIP Integrity (`--verify`)
**Scenario:** You have a ZIP file and want to ensure none of its internal files are corrupted using their embedded CRC32 checksums.
**Command:**
```bash
hashi --verify archive.zip
```
**Output:**
```text
Verifying: archive.zip
  ✓ All 42 entries passed CRC32 verification
```
**Explanation:** The `--verify` flag triggers deep inspection of supported archive formats like ZIP.

### 2.2 Hash ZIP File (Default)
**Scenario:** You want to compute the hash of the ZIP file itself (the standard behavior).
**Command:**
```bash
hashi archive.zip
```
**Output:**
```text
[SHA-256] archive.zip
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```
**Explanation:** Without the `--verify` flag, `hashi` treats ZIP files like any other regular file.

---

## 3. File Selection & Traversal

### 3.1 Recursive Hashing (`-r`)
**Scenario:** You want to hash every file in your project, including those in subfolders `src/` and `images/`.
**Command:**
```bash
hashi -r
```
**Output:**
```text
[SHA-256] Computed hashes for 15 files (recursive):
...
7d793037...  src/main.go
b1946ac9...  src/utils/helper.go
55502f40...  images/logo.png
...
```
**Explanation:** The `-r` (recursive) flag tells `hashi` to traverse the directory tree downwards.

### 3.2 Include Hidden Files (`--hidden`)
**Scenario:** You need to check configuration files like `.bashrc` or `.git/config`.
**Command:**
```bash
hashi --hidden
```
**Output:**
```text
[SHA-256] Computed hashes for 5 files (including hidden):
...
324d26c5...  .gitignore
982d9212...  .env
...
```
**Explanation:** The `--hidden` flag forces `hashi` to include files and directories starting with a dot (`.`), which are usually ignored.

---

## 4. Filtering

### 4.1 Include by Pattern (`--include`)
**Scenario:** You only care about your source code files.
**Command:**
```bash
hashi -r --include "*.go,*.js"
```
**Output:**
```text
[SHA-256] Filtering: include=[*.go, *.js]
Computed hashes for 12 files:
...
```
**Explanation:** `hashi` scanned the tree but only processed files ending in `.go` or `.js`.

### 4.2 Exclude by Pattern (`--exclude`)
**Scenario:** You want to hash everything except temporary log files.
**Command:**
```bash
hashi -r --exclude "*.log"
```
**Output:**
```text
[SHA-256] Filtering: exclude=[*.log]
Computed hashes for 8 files:
...
```
**Explanation:** Files matching `*.log` were skipped.

### 4.3 Filter by Size (`--min-size`, `--max-size`)
**Scenario:** You want to find large ISOs (>1GB) but ignore small text files.
**Command:**
```bash
hashi -r --min-size 1GB
```
**Output:**
```text
[SHA-256] Filtering: size >= 1GB
Computed hashes for 2 files:
e3b0c442...  backups/full_db.dump
a1b2c3d4...  downloads/movie.mkv
```
**Explanation:** Only files meeting the size criteria were hashed.

### 4.4 Filter by Date (`--modified-after`)
**Scenario:** You verified your backup last week. You only want to check files changed since then.
**Command:**
```bash
hashi -r --modified-after 2026-01-10
```
**Output:**
```text
[SHA-256] Filtering: modified >= 2026-01-10
Computed hashes for 4 files:
...
```
**Explanation:** `hashi` checked file metadata and only processed recently modified files.

---

## 5. Advanced Features

### 5.1 Auto-Algorithm Detection
**Scenario:** A website gives you a short 32-character hash (MD5), but you forget to specify `--algorithm md5`.
**Command:**
```bash
hashi myfile.exe 5d41402abc4b2a76b9719d911017c592
```
**Output:**
```text
[INFO] Auto-detected algorithm: MD5 (based on 32-char length)
[MD5] Verifying myfile.exe...
✅ PASS: Hash matches.
```
**Explanation:** `hashi` noticed the string length was 32, inferred it was likely MD5, switched the algorithm for you, and verified the file.

### 5.2 Explicit Algorithm Selection (`--algorithm`)
**Scenario:** You need a SHA-512 hash specifically.
**Command:**
```bash
hashi --algorithm sha512 secure_doc.pdf
```
**Output:**
```text
[SHA-512] secure_doc.pdf
cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce... (truncated)
```
**Explanation:** The `--algorithm` flag overrides the default SHA-256.

---

## 6. Output & Logging

### 6.1 JSON Output (`--format json`)
**Scenario:** You are writing a Python script to process these hashes and need structured data.
**Command:**
```bash
hashi -r --format json
```
**Output:**
```json
{
  "processed": 15,
  "duration_ms": 124,
  "match_groups": [],
  "unmatched": [
    {"file": "notes.txt", "hash": "e3b0c442..."},
    ...
  ],
  "errors": []
}
```
**Explanation:** The output is pure JSON, ready to be piped into `jq` or read by other programs.

### 6.2 Logging to File (`--log-file`)
**Scenario:** You want to record internal events or warnings to a separate file.
**Command:**
```bash
hashi -r --log-file activity.log
```
**Explanation:** Detailed event logs are written to disk, keeping the main output clean.

---

## 7. Configuration Files

### 7.1 Using a Config File (`--config`)
**Scenario:** You have a complex set of defaults you use for a specific project.
**Command:**
```bash
hashi --config .hashi.toml
```
**Content of .hashi.toml:**
```toml
[defaults]
algorithm = "sha512"
recursive = true
output_format = "plain"
```
**Explanation:** `hashi` reads the settings from the TOML file, allowing you to maintain project-specific configurations.