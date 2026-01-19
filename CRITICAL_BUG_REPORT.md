# CRITICAL BUG REPORT: Configuration Precedence Violation

**Date**: January 17, 2026  
**Severity**: CRITICAL  
**Status**: IDENTIFIED - Requires immediate fix  
**Task**: #47 in tasks.md  

## Summary

The configuration precedence system has a critical flaw that violates the documented hierarchy "flags > environment variables > config files". When users explicitly set command-line flags to their default values, those flags are incorrectly overridden by environment variables.

## Impact Assessment

### User Impact: HIGH
- Users cannot override environment variables with explicit flags in common scenarios
- Breaks expected CLI behavior and documented precedence
- Affects both string flags (algorithm, format) and boolean flags (verbose, quiet, etc.)
- Creates confusion and unpredictable behavior

### Development Impact: HIGH  
- Current tests avoid this bug by using non-default values
- Future development may introduce more precedence bugs
- Violates fundamental CLI design principles

## Technical Details

### Root Cause
The `ApplyEnvConfig` function uses hardcoded default value comparisons instead of detecting whether flags were explicitly set:

```go
// BUGGY CODE in internal/config/config.go:500-521
if cfg.Algorithm == "sha256" && env.HashiAlgorithm != "" {
    cfg.Algorithm = env.HashiAlgorithm  // WRONG: Overrides explicit --algorithm=sha256
}
if !cfg.Verbose && env.HashiVerbose {
    cfg.Verbose = env.HashiVerbose      // WRONG: Overrides explicit --verbose=false
}
```

### Affected Configuration Options
1. **String Flags**:
   - `--algorithm=sha256` (default: sha256)
   - `--format=default` (default: default)

2. **Boolean Flags**:
   - `--recursive=false` (default: false)
   - `--verbose=false` (default: false) 
   - `--quiet=false` (default: false)
   - `--hidden=false` (default: false)
   - `--preserve-order=false` (default: false)

## Reproduction Steps

1. Set environment variable: `export HASHI_ALGORITHM=md5`
2. Run with explicit default flag: `hashi --algorithm=sha256 file.txt`
3. **Expected**: Uses SHA256 (flag overrides env var)
4. **Actual**: Uses MD5 (env var incorrectly overrides flag)

## Test Evidence

The bug is confirmed by this failing test:

```go
func TestCriticalBug(t *testing.T) {
    os.Setenv("HASHI_ALGORITHM", "md5")
    defer os.Unsetenv("HASHI_ALGORITHM")
    
    cfg, _, err := ParseArgs([]string{"--algorithm=sha256"})
    // ... apply env config ...
    
    // FAILS: cfg.Algorithm == "md5" instead of "sha256"
    assert.Equal(t, "sha256", cfg.Algorithm)
}
```

## Solution Approach

Use `pflag.Changed()` to detect explicitly set flags:

```go
// CORRECT APPROACH
func (env *EnvConfig) ApplyEnvConfig(cfg *Config, flagSet *pflag.FlagSet) {
    if !flagSet.Changed("algorithm") && env.HashiAlgorithm != "" {
        cfg.Algorithm = env.HashiAlgorithm
    }
    if !flagSet.Changed("verbose") && env.HashiVerbose {
        cfg.Verbose = env.HashiVerbose
    }
    // ... etc for all flags
}
```

## Priority Justification

This is marked CRITICAL because:
1. **Breaks core functionality**: CLI precedence is fundamental
2. **User-facing**: Directly affects user experience and expectations
3. **Silent failure**: Users may not notice incorrect behavior immediately
4. **Widespread**: Affects multiple configuration options
5. **Design violation**: Contradicts documented behavior

## Next Steps

1. Implement Task #47 to fix the precedence logic
2. Add comprehensive tests for all affected flags
3. Update documentation to clarify precedence behavior
4. Consider adding integration tests with real environment variables

## Related Files

- `internal/config/config.go` - Contains buggy ApplyEnvConfig function
- `internal/config/config_test.go` - Contains test demonstrating the bug
- `.kiro/specs/hashi_project/tasks.md` - Task #47 for the fix