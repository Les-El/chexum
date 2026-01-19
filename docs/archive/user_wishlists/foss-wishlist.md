# FOSS Enthusiast Wishlist for Hashi
*Brainstorming session: 2025-01-17*

## Purpose
This document captures priorities and concerns from a Free and Open Source Software (FOSS) perspective. These represent values and features that matter to users who prioritize transparency, freedom, and community-driven development.

## Core FOSS Values

### Freedom
- Use for any purpose
- Study how it works
- Modify and adapt
- Share modifications

### Transparency
- Open development process
- Public discussions
- Clear decision-making
- Visible roadmap

### Community
- Welcoming to contributors
- Responsive to issues
- Credit for contributions
- Shared ownership

### Sustainability
- Not dependent on single entity
- Clear succession plan
- Forkable if needed
- Long-term maintenance commitment

---

## Critical Requirements (Deal Breakers)

### 1. Open Source License
**Requirement:**
- MIT, Apache 2.0, GPL, or similar OSI-approved license
- No proprietary extensions or "enterprise" features
- All functionality available in source form

**Current Status:** ✓ (Verify LICENSE file exists and is appropriate)

---

### 2. Zero Telemetry
**Requirement:**
- No usage statistics collection
- No crash reporting to external servers
- No "phone home" functionality
- No analytics code in source

**Philosophy:** Not even opt-out telemetry. Just none.

**Implementation:**
- Explicit privacy statement in README
- Code audit to verify no network calls
- No third-party analytics libraries

---

### 3. Offline-First Operation
**Requirement:**
- All features work without internet
- No required network dependencies
- No update checks
- No cloud integration

**Use Cases:**
- Air-gapped systems
- Remote locations
- Security-conscious environments
- Unreliable connectivity

---

### 4. Composability (Unix Philosophy)
**Requirement:**
- Clean stdin/stdout behavior
- Predictable output formats
- Works in pipes and scripts
- No TTY-only features that break automation

**Examples:**
```bash
# Pipe file lists
find . -type f | hashi --stdin-list

# Filter output
hashi *.txt | grep "duplicate"

# Transform with standard tools
hashi --plain *.pdf | awk '$2 == "abc123..."'

# JSON processing
hashi --json *.log | jq '.results[] | select(.count > 1)'
```

---

### 5. Auditable Source Code
**Requirement:**
- Clean, readable code
- Well-documented architecture
- No obfuscation
- Clear separation of concerns

**Supporting Documentation:**
- Architecture Decision Records (ADRs)
- Code comments explaining "why"
- Design documentation
- Security model documentation

---

## High Priority Features

### 6. Man Pages
**Requirement:**
- Proper Unix documentation format
- Installed to standard locations
- Generated from source (not manually maintained)
- Comprehensive coverage of all flags

**Standard Sections:**
- NAME
- SYNOPSIS
- DESCRIPTION
- OPTIONS
- EXAMPLES
- EXIT STATUS
- ENVIRONMENT
- FILES
- SEE ALSO
- BUGS

**Implementation:**
- Generate from markdown or code
- Include in installation process
- Keep in sync with --help output

---

### 7. Shell Completions
**Requirement:**
- Bash completion
- Zsh completion
- Fish completion
- Installed to standard locations

**Functionality:**
- Flag completion
- File path completion
- Algorithm name completion
- Context-aware suggestions

---

### 8. Reproducible Builds
**Requirement:**
- Same source produces identical binary
- Build instructions in documentation
- Version information embedded in binary
- Checksums published for releases

**Benefits:**
- Verify binary matches source
- Security auditing
- Trust in distribution
- Supply chain security

**Implementation:**
```bash
# Build with version info
go build -ldflags "-X main.version=$(git describe --tags)"

# Verify build
hashi --version  # Shows commit hash
hashi --verify-build  # Compares to known checksums
```

---

### 9. Stable CLI Interface
**Requirement:**
- Semantic versioning
- Deprecation warnings before removal
- Backward compatibility guarantees
- Clear migration guides

**Stability Promise:**
- Output formats don't change within major version
- New flags don't break existing scripts
- Exit codes remain consistent
- Manifest formats are versioned

**Example Deprecation:**
```bash
# Version 1.5: Deprecation warning
$ hashi --old-flag
Warning: --old-flag is deprecated, use --new-flag instead
Will be removed in version 2.0

# Version 2.0: Removed with clear error
$ hashi --old-flag
Error: --old-flag was removed in 2.0, use --new-flag
See: https://github.com/example/hashi/blob/main/MIGRATION.md
```

---

### 10. Minimal Dependencies
**Requirement:**
- Small dependency tree
- Well-maintained dependencies
- Security-audited dependencies
- Optional: vendored dependencies for reproducibility

**Current Dependencies (from go.mod):**
- github.com/fatih/color
- github.com/schollz/progressbar
- github.com/spf13/pflag
- github.com/joho/godotenv
- golang.org/x/term

**Evaluation Criteria:**
- Is dependency actively maintained?
- Is it widely used and audited?
- Could we implement it ourselves if needed?
- Does it have its own large dependency tree?

---

## Medium Priority Features

### 11. Configuration File Support
**Requirement:**
- Optional, not required
- Standard locations (~/.config/hashi/config)
- Simple format (TOML, YAML, or INI)
- Environment variables override config
- Flags override everything

**Philosophy:** Convenience, not requirement. Tool must work without config file.

**Example Config:**
```toml
# ~/.config/hashi/config.toml
[defaults]
algorithm = "sha256"
output_format = "plain"
recursive = false

[colors]
enabled = true
theme = "default"
```

---

### 12. Multiple Hash Algorithms
**Requirement:**
- User choice of algorithm
- Clear default (SHA-256)
- Support for common algorithms
- Extensible architecture

**Algorithms to Support:**
- MD5 (with security warning)
- SHA-1 (with security warning)
- SHA-256 (default)
- SHA-512
- BLAKE2b
- BLAKE3 (if feasible)

**Usage:**
```bash
hashi --algorithm sha512 file.txt
hashi --algo blake2b file.txt
hashi -a md5 file.txt  # Shows security warning
```

---

### 13. Packaging Helpers
**Requirement:**
- Scripts for common package formats
- Installation respects PREFIX
- Standard directory structure
- Uninstall support

**Package Formats:**
- Debian (.deb)
- RPM (.rpm)
- Arch (PKGBUILD for AUR)
- Homebrew (formula)
- Nix (derivation)

**Standard Installation:**
```bash
make PREFIX=/usr/local install
# Installs:
# /usr/local/bin/hashi
# /usr/local/share/man/man1/hashi.1
# /usr/local/share/bash-completion/completions/hashi
```

---

### 14. Architecture Documentation
**Requirement:**
- High-level design overview
- Component interaction diagrams
- Data flow documentation
- Extension points

**Documentation Structure:**
```
docs/
├── architecture/
│   ├── overview.md
│   ├── components.md
│   ├── data-flow.md
│   └── extension-points.md
├── adr/
│   └── 001-example-decision.md
└── contributing/
    ├── code-style.md
    └── testing.md
```

---

### 15. Contribution Guidelines
**Requirement:**
- CONTRIBUTING.md in repo
- Clear process for PRs
- Issue templates
- Code of conduct
- Development setup guide

**Contents:**
- How to set up dev environment
- How to run tests
- Code style guidelines
- PR submission process
- Review expectations
- Credit and attribution policy

---

## Low Priority Features

### 16. Plugin System
**Concept:**
- Extend functionality without forking
- Custom hash algorithms
- Custom output formatters
- Custom input parsers

**Considerations:**
- Security implications
- Maintenance burden
- API stability
- Documentation requirements

**Possible Implementation:**
```bash
# Custom output format
hashi --format-plugin ~/my-formatter.so *.txt

# Custom hash algorithm
hashi --hash-plugin ~/blake3.so file.txt
```

**Status:** Interesting but complex. Evaluate after v1.0.

---

### 17. Custom Output Formatters
**Concept:**
- User-provided scripts for output formatting
- Receive JSON on stdin
- Output custom format

**Example:**
```bash
# Custom formatter script
hashi --json *.txt | ./my-formatter.sh

# Or built-in support
hashi --format-script ./my-formatter.sh *.txt
```

**Status:** May be unnecessary if JSON output is good enough.

---

### 18. Library Interface
**Concept:**
- Use hashi as a Go library
- Stable API for programmatic use
- Separate CLI and library concerns

**Considerations:**
- API stability guarantees
- Documentation burden
- Versioning complexity

**Status:** Post-v1.0 consideration.

---

### 19. Benchmark Suite
**Concept:**
- Performance regression testing
- Compare against other tools
- Track performance over time

**Implementation:**
```bash
make bench
# Runs benchmarks, compares to baseline
# Reports any regressions
```

**Status:** Nice to have for performance-critical features.

---

### 20. Multi-Architecture Support
**Requirement:**
- Linux: x86_64, ARM64, ARM32
- BSD: FreeBSD, OpenBSD
- macOS: x86_64, ARM64 (Apple Silicon)
- Windows: x86_64 (lower priority)

**Implementation:**
- Cross-compilation support
- CI builds for all platforms
- Platform-specific testing
- Release artifacts for each platform

---

## Anti-Requirements (What NOT to Include)

### Red Flags
- Telemetry (even "anonymous" or "opt-out")
- Cloud integration or sync features
- Proprietary extensions
- Bundled analytics
- Auto-update mechanisms
- Required registration or accounts
- Closed issue tracker
- Contributor License Agreements (CLAs)

### Annoyances
- Electron or web-based UI
- Massive dependency trees
- Requires Docker to build
- Config in proprietary formats
- Documentation behind login walls
- "Enterprise" vs "Community" editions
- Artificial feature limitations

---

## Privacy and Security Principles

### Privacy Guarantees
1. **No data leaves the machine** - All operations are local
2. **No metadata collection** - No usage statistics
3. **No crash reporting** - No external error reporting
4. **No update checks** - No network calls

### Security Model
1. **Read-only operations** - Never modify source files
2. **No code execution** - Don't execute file contents
3. **Path traversal protection** - Validate all file paths
4. **Resource limits** - Prevent DoS via resource exhaustion

### Documentation Requirements
- Explicit privacy policy in README
- Security model documentation
- Threat model (post-v1.0)
- Responsible disclosure policy

---

## Distribution Considerations

### Build Process
**Requirements:**
- Standard Go build process
- No custom build tools required
- Reproducible builds
- Clear build instructions

**Example:**
```bash
# Simple build
go build -o hashi ./cmd/hashi

# Production build with version info
make build VERSION=$(git describe --tags)

# Install
make install PREFIX=/usr/local
```

### Package Maintainer Needs
1. **Minimal dependencies** - Easier to package
2. **Standard paths** - Follows FHS (Filesystem Hierarchy Standard)
3. **Man pages** - Proper documentation
4. **License file** - Clear licensing
5. **Changelog** - Track changes between versions

### Release Process
1. Tag release in git
2. Generate changelog
3. Build binaries for all platforms
4. Generate checksums
5. Sign release (GPG)
6. Publish to GitHub releases
7. Update package repositories

---

## Community Governance

### Decision Making
- Public discussion of major features
- ADRs for architectural decisions
- Transparent roadmap
- Community input on direction

### Contribution Recognition
- Contributors listed in CONTRIBUTORS file
- Credit in release notes
- Co-authorship on commits when appropriate
- Recognition in documentation

### Maintenance
- Clear maintainer responsibilities
- Succession planning
- Bus factor > 1
- Responsive to security issues

---

## Success Metrics (FOSS Perspective)

### Adoption Indicators
- GitHub stars and forks
- Package availability (distros, Homebrew, etc.)
- Community contributions
- Issue/PR activity
- Mentions in blogs/tutorials

### Health Indicators
- Active maintenance
- Responsive to issues
- Regular releases
- Growing contributor base
- Low bug backlog

### Trust Indicators
- Security audits
- Reproducible builds
- Transparent development
- Clear communication
- Stable API

---

## Integration with Project Goals

All features in this document should align with hashi's core principles:

1. **Developer Continuity** - Documentation enables future contributors
2. **User-First Design** - Features serve real user needs
3. **No Lock-Out** - Users maintain control and freedom

### Feature Assessment
Each suggestion must pass through standard criteria:
- Clear user benefit
- Fills real workflow gap
- Addresses common pain points
- Doesn't add unnecessary complexity

---

## Next Steps

1. Review current project against FOSS requirements
2. Identify gaps in documentation or governance
3. Prioritize features based on community feedback
4. Create issues for accepted features
5. Document rejected features with rationale
6. Establish contribution process
7. Build community engagement plan

---

## Notes

- FOSS values emphasize process and governance, not just features
- Many requirements are about "how" not "what"
- Community trust is earned through transparency and consistency
- Long-term sustainability requires more than just code
- Documentation and governance are first-class concerns
