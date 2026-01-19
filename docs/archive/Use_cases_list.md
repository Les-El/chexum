# Hashi Use Cases List

*A comprehensive catalog of everything we might want to do with hashi, organized by feasibility, justification, and atomic structure.*

---

## Document Purpose

This document serves as a central repository for all potential hashi use cases, derived from the project specifications, requirements, and design documents. Each use case is evaluated for:

- **Feasibility**: Technical complexity and implementation effort
- **Justification**: User value and alignment with CLI guidelines
- **Atomic Structure**: How the use case breaks down into discrete components

---

## Core Use Cases (High Priority - Implemented/In Progress)

### UC-001: Basic File Hashing
**Description**: Compute cryptographic hashes for individual files or directories
**Feasibility**: ✅ High (Core functionality)
**Justification**: ✅ Essential - Primary purpose of the tool
**Components**:
- File reading and streaming
- Hash algorithm selection (SHA-256, SHA-512, MD5, SHA-1, CRC32)
- Progress indicators for large files
- Error handling for file access issues

**Examples**:
```bash
hashi file.txt                    # Hash single file
hashi                            # Hash current directory
hashi -a                         # Recursive hashing
hashi --algo sha512 file.pdf     # Specific algorithm
```

### UC-002: Hash Verification
**Description**: Compare computed hashes against provided hash strings
**Feasibility**: ✅ High (Straightforward comparison)
**Justification**: ✅ Essential - Critical for integrity verification
**Components**:
- Hash string validation and normalization
- Algorithm auto-detection from hash length
- File vs hash argument classification
- Pass/fail reporting with clear output

**Examples**:
```bash
hashi file.txt a1b2c3d4...       # Verify file against hash
hashi installer.zip <expected>   # Download verification
```

### UC-003: Archive Integrity Verification
**Description**: Verify internal checksums of archive files (ZIP CRC32) using the explicit `--verify` flag.
**Feasibility**: ✅ Medium (ZIP parsing required)
**Justification**: ✅ High - Unique value for integrity checking
**Components**:
- ZIP file format parsing
- CRC32 verification per entry
- Security hardening (algorithm substitution prevention)
- Boolean output for scripting

**Examples**:
```bash
hashi --verify file.zip           # Verify ZIP integrity
hashi file.zip                    # Hash ZIP as raw bytes (standard behavior)
hashi --verify *.zip              # Batch ZIP verification
```

### UC-004: Duplicate Detection
**Description**: Identify files with matching hashes (duplicates)
**Feasibility**: ✅ High (Hash grouping)
**Justification**: ✅ High - Common user need
**Components**:
- Hash computation and grouping
- Match group identification
- Output formatting (grouped display)
- Statistics and summaries

**Examples**:
```bash
hashi -a                         # Find duplicates recursively
hashi --include "*.jpg" -a       # Find duplicate images
hashi --output-format json -a    # Machine-readable duplicates
```

### UC-005: Configuration Management
**Description**: Flexible configuration through files, environment variables, and flags
**Feasibility**: ✅ Medium (Multiple config sources)
**Justification**: ✅ High - CLI guidelines compliance
**Components**:
- Config file auto-discovery (.hashi.json)
- Environment variable support (HASHI_*)
- Precedence handling (flags > env > config > defaults)
- Configuration validation and reporting

**Examples**:
```bash
hashi --config project.json      # Explicit config
export HASHI_ALGORITHM=sha512    # Environment config
# Auto-loads from ./.hashi.json, ~/.config/hashi/config.json
```

---

## Advanced Use Cases (Medium Priority)

### UC-006: Output Format Flexibility
**Description**: Multiple output formats for different use cases
**Feasibility**: ✅ High (Format abstraction)
**Justification**: ✅ High - CLI guidelines requirement
**Components**:
- Default formatter (grouped by matches)
- JSON formatter (machine-readable)
- Plain formatter (tab-separated for scripting)
- Verbose formatter (detailed summaries)
- Boolean output (-b flag for true/false scripts)
- Quiet mode (minimal output for piping)

**Examples**:
```bash
hashi --json files/              # JSON output
hashi --plain files/             # Tab-separated
hashi -b file1.txt file2.txt     # Boolean result (true/false)
hashi --quiet file.txt           # Just the hash
hashi --verbose -a               # Detailed summary
```

### UC-007: Advanced Filtering
**Description**: Filter files by various criteria before processing
**Feasibility**: ✅ Medium (Multiple filter types)
**Justification**: ✅ Medium - Useful for large datasets
**Components**:
- Pattern matching (include/exclude)
- Size filtering (min/max)
- Date filtering (modified before/after)
- Hidden file handling
- Filter combination logic

**Examples**:
```bash
hashi --include "*.pdf" --min-size 1MB    # Large PDFs only
hashi --exclude "*.tmp" -a                # Skip temp files
hashi --modified-after 2026-01-01 -a      # Recent files only
```

### UC-008: Scripting Integration
**Description**: Exit codes and quiet modes for automation
**Feasibility**: ✅ High (Exit code logic)
**Justification**: ✅ High - Essential for automation
**Components**:
- Meaningful exit codes (0-130)
- Boolean mode (-b flag) for true/false output
- Quiet mode (suppress stdout, keep stderr)
- Match requirement flags
- Error code differentiation
- Parallel processing support

**Examples**:
```bash
# Boolean mode (simplest)
hashi -b file1.txt file2.txt && echo "match" || echo "differ"
MATCH=$(hashi -b file1.txt file2.txt)

# Exit code based
if hashi --quiet --match-required *.jpg; then
    echo "Duplicates found"
fi

# Hash extraction
HASH=$(hashi --quiet file.txt)

# Parallel processing
find . -type f | parallel hashi --quiet {}
```

### UC-009: Progress and Feedback
**Description**: Progress indicators and responsive feedback
**Feasibility**: ✅ Medium (TTY detection and timing)
**Justification**: ✅ High - CLI guidelines requirement
**Components**:
- TTY detection for progress display
- Progress bars for long operations (>100ms)
- ETA calculation
- Interrupt handling (Ctrl-C)
- Status reporting

**Examples**:
```bash
hashi -a /large/directory/       # Shows progress bar
# Processing files... [████████████░░░░░░░░] 60% (6/10) ETA: 2s
```

### UC-010: Error Handling and Recovery
**Description**: User-friendly error messages with actionable suggestions
**Feasibility**: ✅ Medium (Error transformation)
**Justification**: ✅ High - CLI guidelines requirement
**Components**:
- Error message transformation
- Actionable suggestions
- Error grouping and deduplication
- Path sanitization
- Recovery guidance

**Examples**:
```bash
# Instead of: "no such file or directory"
# Shows: "Cannot find file: document.pdf
#         Try: Check spelling, use 'ls' to see files"
```

---

## Specialized Use Cases (Lower Priority)

### UC-011: Hash String Validation
**Description**: Validate hash string format without files
**Feasibility**: ✅ High (String validation)
**Justification**: ✅ Medium - Useful utility function
**Components**:
- Hex character validation
- Length-based algorithm detection
- Ambiguity handling (128-char hashes)
- Format error reporting

**Examples**:
```bash
hashi a1b2c3d4...               # Validate hash format
# Output: ✅ Valid SHA-256 hash format (64 hex characters)
```

### UC-012: Directory Comparison
**Description**: Compare two directories for differences
**Feasibility**: 🟡 Medium-High (Complex logic)
**Justification**: 🟡 Medium - Useful but not core
**Components**:
- Recursive directory traversal
- Hash-based comparison
- Difference categorization (missing, extra, modified)
- Detailed reporting
- Performance optimization for large trees

**Examples**:
```bash
hashi --compare ./source ./backup    # Compare directories
# Shows: matches, mismatches, unique files
```

### UC-013: File Output Management
**Description**: Save results to files with safety features
**Feasibility**: ✅ Medium (File I/O with safety)
**Justification**: ✅ Medium - Useful for reporting
**Components**:
- Overwrite protection
- Append mode support
- Atomic writes (temp + rename)
- JSON log append (maintain validity)
- Force flag override

**Examples**:
```bash
hashi -a --output-file report.txt       # Save to file
hashi -a --log-file audit.log           # Log events
hashi --json --output-file data.json    # JSON output
```

### UC-014: Signal Handling
**Description**: Graceful interruption and cleanup
**Feasibility**: ✅ Medium (Signal handling)
**Justification**: ✅ High - CLI guidelines requirement
**Components**:
- SIGINT (Ctrl-C) handling
- Cleanup timeout management
- Double Ctrl-C force quit
- Status reporting on interruption
- Recoverable state guarantee

**Examples**:
```bash
hashi -a /huge/directory/
# ^C - Interrupted. Cleaning up... (Press Ctrl-C again to force quit)
```

### UC-015: Backup Verification Workflows
**Description**: Automated backup integrity checking and change detection
**Feasibility**: ✅ High (Combines existing features)
**Justification**: ✅ High - Common real-world use case
**Components**:
- Boolean comparison for change detection
- Archive integrity verification
- Exit code handling for automation
- Quiet mode for clean scripting
- Error handling and reporting

**Examples**:
```bash
# Backup verification script
if hashi -b original.tar.gz backup.tar.gz; then
    echo "Backup verified successfully"
else
    echo "ERROR: Backup verification failed!" >&2
    exit 1
fi

# Change detection
if ! hashi -b config.yaml config.yaml.bak; then
    echo "Config changed, reloading..."
    cp config.yaml config.yaml.bak
    systemctl reload myapp
fi
```

### UC-016: Duplicate File Management
**Description**: Find and manage duplicate files across directory trees
**Feasibility**: ✅ High (Hash grouping + scripting)
**Justification**: ✅ High - Very common user need
**Components**:
- Recursive directory traversal
- Hash computation and grouping
- Duplicate identification
- Scriptable output for processing
- Size and pattern filtering

**Examples**:
```bash
# Find all duplicates
find . -type f -print0 | \
    xargs -0 hashi --quiet | \
    sort | uniq -d -w 64 | cut -d' ' -f2-

# Count unique files
hashi --quiet *.txt | sort -u | wc -l

# Find duplicate images only
hashi --include "*.jpg,*.png" -a --json | jq '.match_groups[]'
```

### UC-017: File Integrity Monitoring
**Description**: Monitor file changes and integrity over time
**Feasibility**: ✅ Medium (Requires state tracking)
**Justification**: ✅ High - Security and maintenance use case
**Components**:
- Baseline hash computation
- Change detection logic
- Timestamp and metadata tracking
- Alert mechanisms
- Batch processing capabilities

**Examples**:
```bash
# Monitor critical files
for file in /etc/passwd /etc/shadow /etc/sudoers; do
    if ! hashi -b "$file" "${file}.baseline"; then
        echo "ALERT: $file has been modified!" >&2
    fi
done

# System integrity check
hashi --quiet /bin/* | diff - /var/lib/hashi/bin.hashes
```

### UC-018: Multi-Algorithm Support
**Description**: Support multiple hash algorithms with auto-detection
**Feasibility**: ✅ High (Algorithm abstraction)
**Justification**: ✅ High - Different use cases need different algorithms
**Components**:
- Algorithm selection (SHA-256, SHA-512, MD5, SHA-1, CRC32)
- Auto-detection from hash string length
- Algorithm-specific optimizations
- Backward compatibility
- Performance considerations

**Examples**:
```bash
hashi --algo sha512 secure_doc.pdf    # Explicit algorithm
hashi file.exe 5d41402abc4b2a76...     # Auto-detects MD5
hashi --algo md5 legacy_files/         # Legacy compatibility
```

### UC-019: Stdin and Pipe Integration
**Description**: Process data from stdin and integrate with Unix pipes
**Feasibility**: ✅ High (Stream processing)
**Justification**: ✅ High - Unix philosophy compliance
**Components**:
- Stdin reading with `-` marker
- Stream processing capabilities
- Pipe-friendly output formats
- Progress indication for streams
- Memory-efficient processing

**Examples**:
```bash
# Hash from stdin
echo "hello world" | hashi -
cat large_file.dat | hashi -

# Pipe integration
find . -name "*.txt" | xargs hashi --quiet
hashi --quiet *.log | sort | uniq -d
```

### UC-020: Dry Run and Preview Mode
**Description**: Preview operations without executing them
**Feasibility**: ✅ High (File system traversal without hashing)
**Justification**: ✅ Medium - Useful for large operations
**Components**:
- Directory traversal and file enumeration
- Size calculation and time estimation
- Filter application preview
- No actual hash computation

**Examples**:
```bash
hashi --dry-run -r /huge/directory     # Preview what would be processed
# Output: Would process 10,234 files (45.2 GB), estimated time: 2m 34s

hashi --dry-run --include "*.pdf" --min-size 1MB -r .
# Output: Would process 45 PDF files (2.1 GB), estimated time: 15s
```

### UC-021: Incremental Operations
**Description**: Only process files that have changed since last run
**Feasibility**: 🟡 Medium (Requires state tracking)
**Justification**: ✅ High - Major performance optimization for CI/CD
**Components**:
- Timestamp and size-based change detection
- Previous manifest comparison
- Incremental hash computation
- State file management

**Examples**:
```bash
hashi --manifest previous.manifest --only-changed ./artifacts
hashi --incremental --state-file .hashi-state ./src
hashi --since-manifest baseline.json --output-changes changes.json
```

---

## Future/Moonshot Use Cases (Research Phase)

### UC-015: Educational Code Quality
**Description**: Source code as a teaching tool for Go and CLI design
**Feasibility**: 🟡 Medium (Documentation effort)
**Justification**: 🟡 Low-Medium - Nice to have
**Components**:
- Comprehensive code comments
- Algorithm explanations
- Go idiom documentation
- Annotated edition branch
- Best practices demonstration

### UC-016: Fuzzing and Conflict Testing
**Description**: Automated testing of flag combinations and edge cases
**Feasibility**: 🟡 Medium-High (Testing infrastructure)
**Justification**: 🟡 Medium - Quality assurance
**Components**:
- Flag combination generator
- Random input generation
- Behavior recording and analysis
- Conflict detection automation
- Regression testing

### UC-017: Performance Optimization
**Description**: High-performance hashing for large datasets
**Feasibility**: 🟡 Medium-High (Optimization complexity)
**Justification**: 🟡 Medium - Scalability
**Components**:
- Parallel processing
- Memory-mapped files
- SIMD optimizations
- Streaming algorithms
- Benchmark suite

### UC-018: Extended Archive Support
**Description**: Support for more archive formats (TAR, 7Z, RAR)
**Feasibility**: 🔴 High (Format complexity)
**Justification**: 🟡 Medium - Broader utility
**Components**:
- Multiple format parsers
- Format-specific integrity checks
- Unified verification interface
- Security considerations per format

### UC-019: Network Integration
**Description**: Hash verification over network protocols
**Feasibility**: 🔴 High (Network complexity)
**Justification**: 🔴 Low - Scope creep risk
**Components**:
- HTTP/HTTPS support
- Remote hash fetching
- Certificate validation
- Timeout handling
- Security considerations

### UC-021: Backup Verification Workflows
**Description**: Automated backup integrity checking and change detection
**Feasibility**: ✅ High (Combines existing features)
**Justification**: ✅ High - Common real-world use case
**Components**:
- Boolean comparison for change detection
- Archive integrity verification
- Exit code handling for automation
- Quiet mode for clean scripting
- Error handling and reporting

**Examples**:
```bash
# Backup verification script
if hashi -b original.tar.gz backup.tar.gz; then
    echo "Backup verified successfully"
else
    echo "ERROR: Backup verification failed!" >&2
    exit 1
fi

# Change detection
if ! hashi -b config.yaml config.yaml.bak; then
    echo "Config changed, reloading..."
    cp config.yaml config.yaml.bak
    systemctl reload myapp
fi
```

### UC-022: Duplicate File Management
**Description**: Find and manage duplicate files across directory trees
**Feasibility**: ✅ High (Hash grouping + scripting)
**Justification**: ✅ High - Very common user need
**Components**:
- Recursive directory traversal
- Hash computation and grouping
- Duplicate identification
- Scriptable output for processing
- Size and pattern filtering

**Examples**:
```bash
# Find all duplicates
find . -type f -print0 | \
    xargs -0 hashi --quiet | \
    sort | uniq -d -w 64 | cut -d' ' -f2-

# Count unique files
hashi --quiet *.txt | sort -u | wc -l

# Find duplicate images only
hashi --include "*.jpg,*.png" -a --json | jq '.match_groups[]'
```

### UC-023: File Integrity Monitoring
**Description**: Monitor file changes and integrity over time
**Feasibility**: ✅ Medium (Requires state tracking)
**Justification**: ✅ High - Security and maintenance use case
**Components**:
- Baseline hash computation
- Change detection logic
- Timestamp and metadata tracking
- Alert mechanisms
- Batch processing capabilities

**Examples**:
```bash
# Monitor critical files
for file in /etc/passwd /etc/shadow /etc/sudoers; do
    if ! hashi -b "$file" "${file}.baseline"; then
        echo "ALERT: $file has been modified!" >&2
    fi
done

# System integrity check
hashi --quiet /bin/* | diff - /var/lib/hashi/bin.hashes
```

---

## Cross-Cutting Concerns

### Security Considerations
**Applicable to**: All use cases
**Components**:
- Input validation and sanitization
- Path traversal prevention
- Algorithm substitution protection
- Information disclosure prevention
- Secure defaults

### Performance Considerations
**Applicable to**: UC-001, UC-004, UC-012, UC-017
**Components**:
- Streaming file processing
- Memory efficiency
- CPU utilization
- I/O optimization
- Progress reporting overhead

### Usability Considerations
**Applicable to**: All use cases
**Components**:
- Intuitive flag names
- Clear error messages
- Consistent output formatting
- Help system quality
- Discovery mechanisms

### Compatibility Considerations
**Applicable to**: All use cases
**Components**:
- Cross-platform support
- Shell integration
- Terminal compatibility
- File system differences
- Character encoding

---

## Implementation Priority Matrix

| Use Case | Feasibility | Justification | Priority | Status |
|----------|-------------|---------------|----------|---------|
| UC-001 | High | Essential | P0 | ✅ Complete |
| UC-002 | High | Essential | P0 | ✅ Complete |
| UC-003 | Medium | High | P1 | 🚧 In Progress |
| UC-004 | High | High | P1 | ✅ Complete |
| UC-005 | Medium | High | P1 | ✅ Complete |
| UC-006 | High | High | P1 | ✅ Complete |
| UC-007 | Medium | Medium | P2 | 📋 Planned |
| UC-008 | High | High | P1 | ✅ Complete |
| UC-009 | Medium | High | P1 | ✅ Complete |
| UC-010 | Medium | High | P1 | ✅ Complete |
| UC-011 | High | Medium | P2 | 📋 Planned |
| UC-012 | Medium-High | Medium | P3 | 💭 Future |
| UC-013 | Medium | Medium | P2 | 📋 Planned |
| UC-014 | Medium | High | P1 | ✅ Complete |
| UC-015 | High | High | P1 | ✅ Complete |
| UC-016 | High | High | P1 | ✅ Complete |
| UC-017 | Medium | High | P1 | ✅ Complete |
| UC-018 | High | High | P1 | ✅ Complete |
| UC-019 | High | High | P1 | ✅ Complete |
| UC-020 | Medium | Low-Medium | P4 | 💭 Moonshot |
| UC-021 | Medium-High | Medium | P4 | 💭 Moonshot |
| UC-022 | Medium-High | Medium | P3 | 💭 Future |
| UC-023 | High | Medium | P3 | 💭 Future |
| UC-024 | High | Low | P5 | ❌ Rejected |
| UC-025 | High | Low | P5 | ❌ Rejected |
| UC-026 | High | Medium | P2 | 📋 Planned |
| UC-027 | Medium | High | P2 | 📋 Planned |
| UC-028 | Medium-High | Medium | P3 | 💭 Future |
| TBS-001 | High | Medium | TBS | 🔬 Study |
| TBS-002 | High | Low | TBS | 🔬 Study |
| TBS-003 | Medium-High | Low | TBS | 🔬 Study |
| TBS-004 | Medium | Low | TBS | 🔬 Study |
| TBS-005 | Medium | Medium | TBS | 🔬 Study |

**Legend**:
- ✅ Complete: Implemented and tested
- 🚧 In Progress: Currently being developed
- 📋 Planned: Scheduled for implementation
- 💭 Future: Under consideration
- ❌ Rejected: Not aligned with project goals

---

## Atomic Component Breakdown

### Core Components (Required for multiple use cases)
1. **File System Interface** (UC-001, UC-004, UC-007, UC-012)
2. **Hash Engine** (UC-001, UC-002, UC-003, UC-004)
3. **Configuration System** (UC-005, All)
4. **Output Formatters** (UC-006, All)
5. **Error Handler** (UC-010, All)
6. **Progress System** (UC-009, UC-001, UC-004)
7. **Signal Handler** (UC-014, All)

### Specialized Components (Single use case focus)
1. **Archive Verifier** (UC-003)
2. **Hash Validator** (UC-011)
3. **Directory Comparator** (UC-012)
4. **File Output Manager** (UC-013)
5. **Filter Engine** (UC-007)

### Support Components (Cross-cutting)
1. **TTY Detector** (UC-006, UC-009)
2. **Color Handler** (UC-006, UC-010)
3. **Flag Conflict Resolver** (UC-005, All)
4. **Path Sanitizer** (UC-010, Security)
5. **Exit Code Manager** (UC-008, All)

---

## Decision Framework

### Inclusion Criteria
1. **Alignment**: Does it support the core mission of hash computation and verification?
2. **User Value**: Does it solve a real user problem?
3. **CLI Guidelines**: Does it follow industry-standard CLI design principles?
4. **Complexity**: Is the implementation complexity justified by the value?
5. **Maintenance**: Can we maintain it long-term without significant burden?

### Exclusion Criteria
1. **Scope Creep**: Features that turn hashi into a different type of tool
2. **Security Risk**: Features that introduce significant attack surface
3. **Complexity Explosion**: Features requiring disproportionate implementation effort
4. **Maintenance Burden**: Features requiring ongoing complex maintenance
5. **User Confusion**: Features that make the tool harder to understand or use

---

## Next Steps

1. **Complete P1 Use Cases**: Finish implementation of high-priority features
2. **Validate P2 Use Cases**: Gather user feedback on planned features
3. **Research P3 Use Cases**: Investigate feasibility and user demand
4. **Document Decisions**: Record rationale for included/excluded features
5. **Update Regularly**: Keep this document current as requirements evolve

---

*This document serves as the authoritative source for hashi feature planning and should be updated as new use cases are identified or priorities change.*

---

## Wishlist Analysis Results

After reviewing the three wishlist documents (`cli-wishlist.md`, `ci-cd-wishlist.md`, `foss-wishlist.md`), I've categorized the suggestions:

### ✅ Already Covered by Existing Use Cases (No Action Needed)
- **Short flags** → UC-006 (Output Format Flexibility) - already planned
- **Tab completion** → UC-006 (Output Format Flexibility) - shell integration
- **Environment variables** → UC-005 (Configuration Management) - already implemented
- **Fast startup** → UC-022 (Performance Optimization) - already planned
- **Clean pipe behavior** → UC-019 (Stdin and Pipe Integration) - already covered
- **Parallel processing** → UC-022 (Performance Optimization) - already planned
- **Multiple algorithms** → UC-018 (Multi-Algorithm Support) - already covered
- **Null-terminated input** → UC-019 (Stdin and Pipe Integration) - already covered
- **Color control** → UC-006 (Output Format Flexibility) - already implemented
- **Statistics** → UC-004 (Duplicate Detection) - already covered
- **Manifest comparison** → UC-012 (Directory Comparison) - already planned
- **Format interoperability** → UC-006 (Output Format Flexibility) - already covered
- **Audit trail** → UC-013 (File Output Management) - already covered
- **Man pages** → Documentation (not a use case)
- **Shell completions** → Documentation (not a use case)
- **Reproducible builds** → Build process (not a use case)
- **Minimal dependencies** → Architecture decision (not a use case)

### 🔄 Enhancements to Existing Use Cases
Several suggestions enhance existing use cases without creating new ones:

**UC-007 (Advanced Filtering) Enhancements:**
- Expression-based filters: `--filter 'size > 1MB && ext == ".log"'`
- Date-based filtering: `--newer-than`, `--older-than`
- Complex pattern matching

**UC-008 (Scripting Integration) Enhancements:**
- Batch verification: `--verify-all *.txt --against hashes.txt`
- Fail-fast mode: `--fail-fast`
- Hook system: `--on-match`, `--on-mismatch`, `--on-complete`

**UC-013 (File Output Management) Enhancements:**
- Template-based output: `--template '{{.File}}: {{.Hash}}'`
- Structured audit logs with timestamps

### 📋 New Use Cases Identified

### UC-026: Dry Run and Preview Mode
**Description**: Preview operations without executing them
**Feasibility**: ✅ High (File system traversal without hashing)
**Justification**: ✅ Medium - Useful for large operations
**Components**:
- Directory traversal and file enumeration
- Size calculation and time estimation
- Filter application preview
- No actual hash computation

**Examples**:
```bash
hashi --dry-run -r /huge/directory     # Preview what would be processed
# Output: Would process 10,234 files (45.2 GB), estimated time: 2m 34s
```

### UC-027: Incremental Operations
**Description**: Only process files that have changed since last run
**Feasibility**: 🟡 Medium (Requires state tracking)
**Justification**: ✅ High - Major performance optimization for CI/CD
**Components**:
- Timestamp and size-based change detection
- Previous manifest comparison
- Incremental hash computation
- State file management

**Examples**:
```bash
hashi --manifest previous.manifest --only-changed ./artifacts
hashi --incremental --state-file .hashi-state ./src
```

### UC-028: Watch Mode and Monitoring
**Description**: Continuously monitor directories for changes
**Feasibility**: 🟡 Medium-High (File system watching)
**Justification**: 🟡 Medium - Useful for development/monitoring
**Components**:
- File system event monitoring
- Configurable check intervals
- Change detection and alerting
- Background operation mode

**Examples**:
```bash
hashi --watch /important/files --interval 60s
hashi --watch-manifest manifest.txt --on-change 'notify-send "Changed"'
```

---

## To-Be-Studied Category (Complex/Different Use Cases)

### TBS-001: Dependency Verification System
**Description**: Verify package dependencies against lockfiles
**Complexity**: 🔴 High - Requires ecosystem-specific parsers
**Rationale**: Each package ecosystem (npm, go, pip, etc.) has different formats and verification requirements. This would require significant ecosystem-specific code.

**Examples from wishlist**:
```bash
hashi --lockfile package-lock.json --verify node_modules
hashi --lockfile go.sum --verify vendor/
```

**Study Questions**:
- Which ecosystems to support?
- How to handle version mismatches vs hash mismatches?
- Integration with existing package managers?
- Maintenance burden for multiple formats?

### TBS-002: Plugin System Architecture
**Description**: Extensible plugin system for custom algorithms and formatters
**Complexity**: 🔴 High - Security, API stability, maintenance burden
**Rationale**: Plugin systems introduce significant complexity around security, API versioning, and maintenance.

**Examples from wishlist**:
```bash
hashi --format-plugin ~/my-formatter.so *.txt
hashi --hash-plugin ~/blake3.so file.txt
```

**Study Questions**:
- Security model for untrusted plugins?
- API stability guarantees?
- Plugin discovery and management?
- Cross-platform compatibility?

### TBS-003: Advanced Template System
**Description**: Complex templating with conditionals and functions
**Complexity**: 🟡 Medium-High - Template language design and implementation
**Rationale**: Goes beyond simple string substitution into a mini-language.

**Examples from wishlist**:
```bash
hashi --template '{{if .Match}}MATCH: {{end}}{{.File}}' *.txt
hashi --template '{{.File | uppercase}}: {{.Hash | truncate 8}}' *.txt
```

**Study Questions**:
- How complex should the template language be?
- Security implications of template execution?
- Maintenance burden vs JSON + external tools?

### TBS-004: Command History and Templates
**Description**: Store and replay complex command patterns
**Complexity**: 🟡 Medium - Privacy, storage, management
**Rationale**: Introduces state management and privacy concerns.

**Examples from wishlist**:
```bash
hashi --history                    # Show recent commands
hashi --replay 1                   # Repeat command #1
hashi --save-template verify-downloads "hashi -r /downloads --size +1M"
```

**Study Questions**:
- Privacy implications of command storage?
- How to handle sensitive file paths?
- Template variable substitution?
- Cross-session persistence?

### TBS-005: Resource Control and Budgeting
**Description**: Time and resource-based operation limits
**Complexity**: 🟡 Medium - Requires sophisticated scheduling and estimation
**Rationale**: Complex heuristics for time estimation and resource management.

**Examples from wishlist**:
```bash
hashi --timeout 5m huge-file.iso
hashi --budget 30s ./artifacts     # Hash what's possible in 30s
hashi --max-memory 1GB ./large-files
```

**Study Questions**:
- How to accurately estimate processing time?
- Graceful degradation strategies?
- Resource monitoring overhead?
- Cross-platform resource limits?

---

## Rejected Categories

### ❌ Scope Creep (Outside Core Mission)
- Network integration (downloading files, remote verification)
- Database integration (storing hashes in databases)
- GUI interfaces
- Cloud synchronization
- Package management features
- Build system integration beyond basic verification

### ❌ Better Served by External Tools
- Complex log parsing (use `jq`, `awk`, `grep`)
- File system operations (use `find`, `xargs`)
- Notification systems (use `notify-send`, `mail`)
- Process management (use `systemd`, `cron`)
- Configuration management (use dedicated config tools)

### ❌ Anti-Requirements (Against Project Principles)
- Telemetry or analytics
- Automatic configuration file modification
- Interactive prompts in scripts
- Breaking CLI compatibility
- Required network connectivity
- Proprietary extensions

---

## Key Insights from Documentation Analysis

### From Scripting Documentation (`docs/user/scripting.md`):
- **Boolean Mode (-b flag)** is the primary scripting interface - outputs just "true" or "false"
- **Quiet Mode (--quiet)** is essential for hash extraction and piping
- **Exit Codes** are comprehensive (0-130) and meaningful for automation
- **Parallel Processing** support with tools like GNU parallel
- **Real-world Patterns**: Backup verification, change detection, duplicate finding

### From Examples Documentation (`docs/user/examples.md`):
- **Auto-Algorithm Detection** from hash string length is a key feature
- **Directory Comparison** is a documented advanced feature
- **Archive Verification** defaults to boolean output for scripting
- **Filtering** by size, date, and patterns is comprehensive
- **Configuration Files** support complex daily workflows

### From README (`README.md`):
- **Core Principles** emphasize user-first design and no lock-out
- **Security Model** clearly distinguishes integrity vs authenticity
- **Project Structure** shows modular architecture with clear separation
- **Dependencies** are minimal and well-chosen
- **Output Formats** are designed for both human and machine consumption

### Missing Use Cases Identified:
The documentation review revealed several important use cases that weren't initially captured:
1. **Backup Verification Workflows** (UC-015) - Critical real-world use case from scripting examples
2. **Duplicate File Management** (UC-016) - Very common user need with specific scripting patterns
3. **File Integrity Monitoring** (UC-017) - Security/maintenance use case for system administrators
4. **Multi-Algorithm Support** (UC-018) - Already partially implemented, needs full documentation
5. **Stdin/Pipe Integration** (UC-019) - Unix philosophy compliance, essential for composability

These additions bring the total use case count to 25, providing comprehensive coverage of hashi's potential functionality.

---

## To-Be-Studied Use Cases (Complex/Different from Core Mission)

### TBS-001: Dependency Verification System
**Description**: Verify package dependencies against lockfiles (package-lock.json, go.sum, etc.)
**Complexity**: 🔴 High - Requires ecosystem-specific parsers
**Rationale**: Each package ecosystem has different formats and verification requirements
**Study Questions**:
- Which ecosystems to support first?
- How to handle version vs hash mismatches?
- Integration with existing package managers?
- Maintenance burden for multiple formats?

### TBS-002: Plugin System Architecture
**Description**: Extensible plugin system for custom algorithms and formatters
**Complexity**: 🔴 High - Security, API stability, maintenance burden
**Rationale**: Plugin systems introduce significant complexity around security and API versioning
**Study Questions**:
- Security model for untrusted plugins?
- API stability guarantees?
- Plugin discovery and management?
- Cross-platform compatibility?

### TBS-003: Advanced Template System
**Description**: Complex templating with conditionals and functions beyond simple substitution
**Complexity**: 🟡 Medium-High - Template language design and implementation
**Rationale**: Goes beyond simple string substitution into a mini-language
**Study Questions**:
- How complex should the template language be?
- Security implications of template execution?
- Maintenance burden vs JSON + external tools?

### TBS-004: Command History and Templates
**Description**: Store and replay complex command patterns with variable substitution
**Complexity**: 🟡 Medium - Privacy, storage, management
**Rationale**: Introduces state management and privacy concerns
**Study Questions**:
- Privacy implications of command storage?
- How to handle sensitive file paths?
- Template variable substitution complexity?
- Cross-session persistence requirements?

### TBS-005: Resource Control and Budgeting
**Description**: Time and resource-based operation limits with intelligent scheduling
**Complexity**: 🟡 Medium - Requires sophisticated scheduling and estimation
**Rationale**: Complex heuristics for time estimation and resource management
**Study Questions**:
- How to accurately estimate processing time?
- Graceful degradation strategies?
- Resource monitoring overhead?
- Cross-platform resource limits?