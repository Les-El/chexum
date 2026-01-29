# Getting Started with hashi

Welcome to **hashi** - a human-first, intuitive command-line tool for cryptographic hashing. This guide will get you up and running in minutes.

## Installation

### Prerequisites
- Go 1.24.0 or later

### Option 1: Build from Source
```bash
git clone https://github.com/Les-El/hashi.git
cd hashi
go build -o hashi ./cmd/hashi
```

### Option 2: Install via `go install`
```bash
go install github.com/Les-El/hashi/cmd/hashi@latest
```

### Verify Installation
```bash
hashi --version
```

---

## Your First Hash

### Hash a Single File
The simplest use case - compute the hash of one file:

```bash
hashi myfile.txt
```

**Output:**
```text
myfile.txt    9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
```

By default, hashi uses **SHA-256**. The output shows the filename and its hash in a clean, readable format.

---

### Hash Multiple Files
```bash
hashi file1.txt file2.txt file3.txt
```

**Output:**
```text
file1.txt    abc123...
file2.txt    def456...
file3.txt    789ghi...
```

---

### Hash All Files in Current Directory
Running `hashi` without arguments processes all visible files:

```bash
hashi
```

This is perfect for quickly checking a folder's contents.

---

## Verifying File Integrity

### Compare a File Against a Known Hash
Downloaded a file and want to verify it wasn't corrupted? Provide both the file and the expected hash:

```bash
hashi installer.zip a1b2c3d4e5f67890abcdef...
```

**If the hashes match:**
```text
✓ PASS installer.zip
```

**If they don't match:**
```text
✗ FAIL installer.zip
  Expected: a1b2c3d4e5f67890abcdef...
  Computed: f0e1d2c3b4a5...
```

---

## Working with Directories

### Recursive Hashing
Hash all files in the current directory and its subdirectories:

```bash
hashi -r
```

**Example output:**
```text
src/main.go              7d793037a0760186...  
src/utils/helper.go      b1946ac92492d234...  
images/logo.png          55502f40dc8343fc...  
```

---

### Include Hidden Files
By default, hashi ignores hidden files (those starting with `.`). To include them:

```bash
hashi --hidden
```

This is useful for checking configuration files like `.env` or `.gitignore`.

---

## Filtering Files

### By File Pattern
Only hash specific file types:

```bash
hashi -r --include "*.go" --include "*.md"
```

This processes only `.go` and `.md` files throughout the directory tree.

---

### By File Size
Hash only large files (e.g., > 100MB):

```bash
hashi -r --min-size 100MB
```

Or only small files (e.g., < 10KB):

```bash
hashi -r --max-size 10KB
```

---

## Changing the Hash Algorithm

hashi supports SHA-256 (default), SHA-512, SHA-1, MD5, and BLAKE2b:

```bash
hashi --algorithm sha512 secret.pdf
hashi --algorithm md5 legacy.zip
hashi --algorithm blake2b data.bin
```

> **Note**: hashi can auto-detect the algorithm from hash strings, so if you're verifying a hash, you often don't need to specify `--algorithm` manually.

---

## Output Formats

### JSON Output (for scripts)
```bash
hashi --json file1.txt file2.txt
```

**Output:**
```json
{
  "processed": 2,
  "duration_ms": 12,
  "unmatched": [
    {"file": "file1.txt", "hash": "abc123..."},
    {"file": "file2.txt", "hash": "def456..."}
  ]
}
```

Perfect for piping into tools like `jq` or parsing with scripts.

---

### JSONL Output (one JSON object per line)
```bash
hashi --output-format jsonl file1.txt file2.txt
```

Ideal for streaming or processing large batches.

---

## Dry Run Mode

Want to preview what would be hashed without actually computing anything?

```bash
hashi --dry-run -r
```

**Output:**
```text
Dry Run: Previewing files that would be processed

src/main.go             (estimated size: 2.3 KB)
src/utils/helper.go     (estimated size: 1.1 KB)

Summary:
  Files to process: 2
  Aggregate size:   3.4 KB
  Estimated time:   < 1s
```

---

## Quiet and Verbose Modes

### Quiet Mode (`-q`)
Suppress all non-essential output:

```bash
hashi -q myfile.txt
```

Only the hash result is printed - no progress bars, no notices.

---

### Verbose Mode (`--verbose`)
Get detailed error messages and additional information:

```bash
hashi --verbose protected.bin
```

This is your best friend when troubleshooting. hashi uses generic error messages for security, but `--verbose` shows the real details.

---

## Boolean Mode

Need a simple yes/no answer? Use `--bool`:

```bash
hashi --bool file1.txt file2.txt
```

**Output:**
```text
true   # if all files have the same hash
false  # if hashes differ
```

Perfect for shell scripts:

```bash
if hashi --bool file1.txt file2.txt; then
    echo "Files are identical"
fi
```

---

## Advanced: Incremental Hashing with Manifests

For large projects, you can save hashes to a manifest and only re-hash changed files:

### Create a manifest:
```bash
hashi -r --output-manifest checksums.json
```

### Later, only hash changed files:
```bash
hashi -r --manifest checksums.json --only-changed
```

This dramatically speeds up hash verification for codebases with thousands of files.

---

## Performance Tips

### Parallel Processing
hashi automatically uses multiple CPU cores for faster hashing. On an 8-core system, it will use 6 cores by default (leaving 2 free for system responsiveness).

You can manually control parallelism:

```bash
hashi -r --jobs 4   # Use exactly 4 workers
```

---

## Common Use Cases

### 1. Verify Downloaded Files
```bash
hashi installer.dmg 9f86d081884c7d659a2feaa0c55ad015...
```

### 2. Find Duplicate Files
```bash
hashi -r --json | jq '.match_groups'
```

### 3. Check Project Integrity Before Deployment
```bash
hashi -r --include "*.js" --include "*.css" --output-manifest build-hashes.json
```

### 4. Quick Directory Comparison
```bash
# In directory A:
hashi -r > hashes-A.txt

# In directory B:
hashi -r > hashes-B.txt

# Compare:
diff hashes-A.txt hashes-B.txt
```

---

## Getting Help

- **General help**: `hashi --help`
- **Detailed errors**: `hashi --verbose [command]`
- **Flag Cheat Sheet**: See [flags-and-arguments.md](flags-and-arguments.md)
- **Command reference**: See [command-reference.md](command-reference.md)
- **Examples**: See [examples.md](examples.md)
- **Troubleshooting**: See [error-handling.md](error-handling.md)

---

## Next Steps

Now that you're comfortable with the basics:

1. Explore [examples.md](examples.md) for advanced patterns
2. Learn about [filtering.md](filtering.md) for complex file selection
3. Check out [incremental.md](incremental.md) for CI/CD integration
4. Read [scripting.md](scripting.md) for automation techniques

---

Happy hashing! 🎉
