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
notes.txt    e3b0c442...
image.png    a1b2c3d4...
data.csv    88d4266f...
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
ubuntu-24.04.iso    9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```
**Explanation:** `hashi` calculates the SHA-256 hash for the specified file and displays it clearly.

### 1.3 Verify a File Integrity (Pass)
**Scenario:** You downloaded a file and want to verify it matches the hash provided by the website.
**Command:**
```bash
hashi installer.zip a1b2c3d4e5f6...
```
**Output:**
```text
PASS installer.zip
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
FAIL installer.zip
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
✓ a1b2c3d4... - Valid hash
  Algorithm: sha256
```
**Explanation:** Since no file matched the argument, `hashi` checked if it was a valid hash string format.

---

## 2. File Selection & Traversal

### 2.1 Recursive Hashing (`-r`)
**Scenario:** You want to hash every file in your project, including those in subfolders `src/` and `images/`.
**Command:**
```bash
hashi -r
```
**Output:**
```text
src/main.go    7d793037...
src/utils/helper.go    b1946ac9...
images/logo.png    55502f40...
```
**Explanation:** The `-r` flag tells `hashi` to traverse the directory tree downwards.

### 2.2 Include Hidden Files (`--hidden`)
**Scenario:** You need to check configuration files like `.bashrc` or `.env`.
**Command:**
```bash
hashi --hidden
```
**Output:**
```text
.gitignore    324d26c5...
.env    982d9212...
```
**Explanation:** The `--hidden` flag forces `hashi` to include files and directories starting with a dot (`.`), which are usually ignored.

---

## 3. Filtering

### 3.1 Include by Pattern (`--include`)
**Scenario:** You only care about your source code files.
**Command:**
```bash
hashi -r --include "*.go" --include "*.js"
```
**Explanation:** `hashi` scanned the tree but only processed files ending in `.go` or `.js`.

### 3.2 Filter by Size (`--min-size`)
**Scenario:** You want to find large files (>1GB) but ignore small text files.
**Command:**
```bash
hashi -r --min-size 1GB
```
**Explanation:** Only files meeting the size criteria were hashed.

---

## 4. Advanced Features

### 4.1 Auto-Algorithm Detection
**Scenario:** A website gives you a short 32-character hash (MD5), but you forget to specify `--algorithm md5`.
**Command:**
```bash
hashi myfile.exe 5d41402abc4b2a76b9719d911017c592
```
**Output:**
```text
PASS myfile.exe
```
**Explanation:** `hashi` noticed the string length was 32, inferred it was likely MD5, switched the algorithm for you, and verified the file. (Note: Output matches the current standard comparison format).

### 4.2 Explicit Algorithm Selection (`--algorithm`)
**Scenario:** You need a SHA-512 hash specifically.
**Command:**
```bash
hashi --algorithm sha512 secure_doc.pdf
```
**Output:**
```text
secure_doc.pdf    cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce...
```
**Explanation:** The `--algorithm` flag overrides the default SHA-256.

---

## 5. Output & Logging

### 5.1 JSON Output (`--json`)
**Scenario:** You are writing a script to process these hashes and need structured data.
**Command:**
```bash
hashi -r --json
```
**Output:**
```json
{
  "processed": 2,
  "duration_ms": 5,
  "match_groups": [],
  "unmatched": [
    {
      "file": "notes.txt",
      "hash": "e3b0c442..."
    }
  ],
  "errors": []
}
```
**Explanation:** The output is pure JSON, ready to be piped into `jq` or read by other programs.

### 5.2 Logging to File (`--log-file`)
**Scenario:** You are running a long operation and want a record of any errors.
**Command:**
```bash
hashi -r --log-file verify.log
```
**Explanation:** Progress and errors are written to the log file while stdout remains for the results.
