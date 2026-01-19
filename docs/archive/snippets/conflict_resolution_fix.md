# Conflict Resolution Fix

**Date**: 2026-01-16  
**Issue**: Duplicate validation logic between `internal/conflict` and `internal/config`

## Problem

A code review identified that flag conflict validation was duplicated:
- `internal/conflict/conflict.go` - Robust rule-based conflict detection system
- `internal/config/config.go` - Hardcoded checks like `if cfg.Quiet && cfg.Verbose`

This violated DRY principles and could lead to inconsistencies.

## Solution

### 1. Enhanced the Conflict Package

Added optional logging support for debugging during development:

```go
resolver := conflict.NewResolver()
resolver.SetLogger(os.Stderr)  // Enable debug logging
err := resolver.Check(flags)
```

Logging output shows:
- Which flags are being checked
- Which rules are being evaluated
- Whether conflicts are detected
- Detailed flag state (set/not set, short form detection)

### 2. Integrated Conflict Package into Config

Removed hardcoded validation from `ValidateConfig()` and replaced with conflict resolver:

```go
// OLD (removed):
if cfg.Quiet && cfg.Verbose {
    return fmt.Errorf("--quiet and --verbose are mutually exclusive")
}

// NEW (using conflict package):
resolver := conflict.NewResolver()
flags := conflict.FlagSet{
    "--quiet":   cfg.Quiet,
    "--verbose": cfg.Verbose,
}
err := resolver.Check(flags)
```

### 3. Special Handling for Format Shorthands

Since `--json` and `--plain` are converted to `OutputFormat` during parsing, their conflicts are checked in `ParseArgs()` before conversion:

```go
// Check conflicts before applying shorthands
if jsonOutput && plainOutput {
    return nil, fmt.Errorf("Cannot use --json and --plain together...")
}
if (jsonOutput || plainOutput) && cfg.Verbose {
    return nil, fmt.Errorf("Cannot use --json and --verbose together...")
}
```

### 4. Updated Documentation

- Added clear note in `internal/config/config.go` package doc
- Updated `.kiro/specs/hashi_project/design.md` to specify conflict package usage
- Added warning: "Do NOT implement hardcoded conflict checks"

### 5. Updated Tests

- Fixed test cases that used invalid flag combinations (`--json` + `--verbose`)
- Updated error message expectations to match conflict package messages
- Added comprehensive tests for conflict package including logging

## Files Changed

- `internal/conflict/conflict.go` - Added logging support
- `internal/conflict/conflict_test.go` - New comprehensive tests
- `internal/conflict/example_test.go` - New usage examples
- `internal/config/config.go` - Integrated conflict resolver, removed hardcoded checks
- `internal/config/config_test.go` - Updated test expectations
- `.kiro/specs/hashi_project/design.md` - Clarified conflict detection approach

## Testing

All tests pass:
```bash
go test ./...
# All packages: ok
```

## Benefits

1. **Single Source of Truth**: All conflict rules in one place
2. **Extensibility**: Easy to add new conflict rules
3. **Debugging**: Optional logging helps during development
4. **Consistency**: Same error messages across all conflicts
5. **Maintainability**: No risk of forgetting to update multiple locations

## Usage for Developers

When adding new mutually exclusive flags:

1. Add rule to `internal/conflict/conflict.go` in `addDefaultRules()`
2. Update `ValidateConfig()` to include the new flags in the FlagSet
3. Tests will automatically validate the new rule

DO NOT add hardcoded checks like `if cfg.FlagA && cfg.FlagB`.
