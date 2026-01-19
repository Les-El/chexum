# CI/CD Wishlist for Hashi
*Brainstorming session: 2025-01-17*

## Purpose
This document captures pain points and feature requests from a CI/CD perspective. All suggestions are subject to the feature assessment criteria defined in project documentation before implementation.

## Core Pain Points

### 1. Artifact Verification Workflow
**Current Pain:**
```bash
curl -O https://cdn.example.com/release-v1.2.3.tar.gz
curl -O https://cdn.example.com/release-v1.2.3.tar.gz.sha256
sha256sum -c release-v1.2.3.tar.gz.sha256
```

**Desired Workflow:**
- Single command to verify artifact against checksum file
- Auto-detect checksum file format
- Clean exit codes for automation
- Structured output option

**Potential Solution:**
`hashi --verify release-v1.2.3.tar.gz --against release-v1.2.3.tar.gz.sha256`

---

### 2. Multi-Stage Build Cache Detection
**Current Pain:**
- Need to detect if ANY file changed to decide on rebuild
- Manual manifest creation and comparison
- No structured diff output

**Desired Workflow:**
- Generate manifest of directory state
- Compare manifests to detect changes
- JSON output showing what changed
- Exit codes: 0 = identical, 1 = changes detected

**Potential Solutions:**
- `hashi --manifest ./src > current.manifest`
- `hashi --compare-manifests previous.manifest current.manifest --json`

---

### 3. Parallel Processing
**Current Pain:**
- Hashing large artifacts is slow
- Manual parallelization is error-prone
- Output collection from multiple processes is messy

**Desired Workflow:**
- Built-in parallel processing
- Automatic core detection or manual specification
- Clean structured output despite parallelization

**Potential Solution:**
`hashi --parallel 8 *.iso --json`

---

### 4. Conditional Verification
**Current Pain:**
- Different file classes need different failure behaviors
- Critical files should fail-fast
- Optional files should warn but not fail build

**Desired Workflow:**
- Separate verification rules for different file sets
- Configurable failure behavior
- Clear distinction between errors and warnings

**Potential Solutions:**
- `hashi --filelist critical.txt --fail-fast 1`
- `hashi --filelist optional.txt --warn-only`

---

### 5. Incremental Verification
**Current Pain:**
- Large artifact sets require full rehashing
- Most files unchanged between builds
- Wasted CI time on redundant operations

**Desired Workflow:**
- Only hash files that changed (by timestamp/size)
- Reference previous manifest
- Massive time savings for large artifact sets

**Potential Solution:**
`hashi --manifest previous.manifest --only-changed ./artifacts`

---

### 6. Format Interoperability
**Current Pain:**
Different tools produce different formats:
- `sha256sum`: `hash  filename`
- Others: `filename: hash`
- JSON: `{"file": "...", "hash": "..."}`
- Comments, blank lines, varying whitespace

**Desired Workflow:**
- Universal import of common hash formats
- Standardized export options
- Format translation capability

**Potential Solution:**
`hashi --import <any-format> --export json`

---

### 7. Timeout and Resource Control
**Current Pain:**
- CI has time limits
- Large files can cause timeouts
- No graceful handling of resource constraints

**Desired Workflow:**
- Time-based limits on operations
- Size-based filtering
- Time budget with partial results

**Potential Solutions:**
- `hashi --timeout 5m huge-file.iso`
- `hashi --max-size 10G ./artifacts`
- `hashi --budget 30s ./artifacts` (hash what's possible, report skipped)

---

### 8. Dependency Verification
**Current Pain:**
- Each package ecosystem requires custom verification scripts
- Lockfiles exist but verification is manual
- No standard approach across ecosystems

**Desired Workflow:**
- Native understanding of common lockfile formats
- Verify dependencies match lockfile
- Ecosystem-agnostic interface

**Potential Solutions:**
- `hashi --lockfile package-lock.json --verify node_modules`
- `hashi --lockfile go.sum --verify vendor/`

---

### 9. Audit Trail
**Current Pain:**
- Compliance requires proof of verification
- Log parsing is fragile
- No standard audit format

**Desired Workflow:**
- Structured, timestamped verification logs
- Append-only audit trail
- Includes context: what, when, result, version, environment

**Potential Solution:**
`hashi --audit-log verification.log --json-append`

---

### 10. Failure Diagnostics
**Current Pain:**
When verification fails:
- Limited context for debugging
- Remote CI makes investigation difficult
- Common causes not surfaced

**Desired Workflow:**
- Detailed mismatch explanation
- Show actual vs expected
- Suggest common causes
- Actionable debugging information

**Potential Solution:**
`hashi --verbose --explain-mismatch file.txt expected-hash`

Output includes:
- File size
- Actual hash
- Expected hash
- Algorithm used
- Common causes (corruption, wrong algorithm, encoding)

---

## Feature Priority Assessment

### High Impact (Solves frequent, high-friction problems)
1. **Manifest comparison** - Core CI/CD workflow
2. **Parallel processing** - Performance multiplier
3. **Format agnostic import** - Eliminates custom parsers
4. **Fail-fast controls** - Time savings on failures
5. **Structured audit logs** - Compliance requirement

### Medium Impact (Valuable but workarounds exist)
6. **Timeout/resource controls** - Safety and predictability
7. **Incremental verification** - Optimization for large sets
8. **Lockfile integration** - Ecosystem-specific value
9. **Conditional verification** - Workflow flexibility

### Nice to Have (Convenience features)
10. **Failure diagnostics** - Better debugging experience
11. **Progress estimation** - User experience improvement
12. **Retry logic** - Resilience enhancement

---

## Anti-Requirements (What CI/CD Does NOT Need)

- GUI or interactive modes (CI is headless)
- Color output (though NO_COLOR respect is good)
- Fancy formatting (JSON is king)
- Configuration files (flags and env vars sufficient)
- Real-time progress bars (can interfere with log capture)

---

## Core CI/CD Values

### Trust
- Verification must be reliable
- Exit codes must be meaningful
- Output must be parseable
- Audit trail must be complete

### Speed
- Parallel processing where possible
- Incremental operations when feasible
- Fail-fast to save time
- Resource controls to prevent hangs

### Automation
- Structured output (JSON)
- Predictable behavior
- No interactive prompts
- Clear exit codes

### Observability
- What was verified
- When it was verified
- What the result was
- Why it failed (if applicable)

---

## Integration Notes

All features in this document must pass through the project's feature assessment criteria:

**Accept features that:**
1. Provide clear user benefit
2. Fill real workflow gaps that multiply productivity
3. Address high-frequency workflows or common pain points
4. Solve error-prone manual processes

**Reject features that:**
- Are niche use cases solvable with simple shell scripts
- Add complexity without sufficient benefit
- Can be achieved by combining existing functionality

---

## Next Steps

1. Review each pain point against feature assessment criteria
2. Identify which can be solved with existing hashi features
3. Identify which require new flags/features
4. Prioritize based on impact and implementation cost
5. Integrate accepted features into project roadmap
6. Document rejected features with rationale

---

## Notes

- This is a brainstorming document, not a commitment
- Many suggestions introduce new flags that need evaluation
- Some features may be achievable through better documentation of existing capabilities
- Focus on solving real problems, not adding features for completeness
