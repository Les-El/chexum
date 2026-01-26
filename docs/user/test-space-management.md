# Test Space Management

## Problem

During test execution, particularly with Go's testing infrastructure, accumulated temporary files in `/tmp` can cause disk space exhaustion. This manifests as errors like:
- `open /tmp/go-build*/file: no such file or directory`
- Build failures due to insufficient /tmp space

## Solution

The hashi project includes automated cleanup mechanisms to address this:

### 1. Automatic Cleanup After Tests

Each test package (cmd/hashi, cmd/checkpoint, cmd/cleanup, internal/checkpoint) includes a `TestMain()` function that automatically removes test-created temporary files after tests complete:

```bash
go test ./cmd/...
# Automatically removes /tmp/hashi-*, /tmp/checkpoint-*, /tmp/test-* after completion
```

### 2. Pre-Test Cleanup Script

For CI/CD pipelines or before running extensive test suites, use the cleanup script:

```bash
bash scripts/cleanup-before-tests.sh
go test ./...
```

### 3. CleanupManager Integration

The [CleanupManager](../internal/checkpoint/cleanup.go) is available for manual use:

```go
import "github.com/Les-El/hashi/internal/checkpoint"

cm := checkpoint.NewCleanupManager(false)
result, err := cm.CleanupTemporaryFiles()
```

Or from the command line:

```bash
./checkpoint cleanup --dry-run          # Preview cleanup
./checkpoint cleanup --force             # Force cleanup immediately
```

### 4. Configuration

Cleanup behavior can be customized via:
- Environment variables
- Configuration files (see [Cleanup Config](../internal/checkpoint/cleanup.go#L15-L30))
- Command-line flags (see [cmd/cleanup/main.go](../cmd/cleanup/main.go))

## Best Practices

1. **Before Large Test Runs**: Always run the pre-test cleanup script
   ```bash
   bash scripts/cleanup-before-tests.sh && go test ./...
   ```

2. **In CI/CD**: Add cleanup to your pipeline before test execution:
   ```yaml
   - name: Clean tmp before tests
     run: bash scripts/cleanup-before-tests.sh
   ```

3. **Monitoring**: Check /tmp usage with:
   ```bash
   df -h /tmp
   du -sh /tmp  # Requires elevated privileges for some dirs
   ```

4. **Safe Patterns**: The cleanup system only removes patterns known to be safe:
   - `/tmp/go-build*` - Go compiler artifacts
   - `/tmp/hashi-*` - Project-specific temp files
   - `/tmp/checkpoint-*` - Checkpoint artifacts
   - `/tmp/test-*` - Test temporary files

Never remove the root `/tmp` or system temporary directories!

## Troubleshooting

If tests still fail with disk space errors after running cleanup:

1. Check available disk space:
   ```bash
   df -h /tmp
   ```

2. Manually clean aggressive patterns:
   ```bash
   rm -rf /tmp/go-build* /tmp/hashi-* /tmp/checkpoint-* /tmp/test-*
   ```

3. Check for other processes using /tmp:
   ```bash
   lsof /tmp | head -20
   ```

4. Consider using a larger tmpfs mount or switching to a different temporary directory:
   ```bash
   export TMPDIR=/var/tmp
   go test ./...
   ```

## See Also

- [CleanupManager API Documentation](../internal/checkpoint/cleanup.go)
- [Cleanup Command](../cmd/cleanup/main.go)
- [Requirements Document - Requirement 3](./requirements.md#requirement-3-cleanup-manager-enhancement)
