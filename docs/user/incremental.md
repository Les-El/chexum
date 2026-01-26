# Incremental Hashing with Manifests

For large codebases or CI/CD pipelines, hashing every file every time can be slow. hashi's manifest system allows you to skip files that haven't changed, dramatically improving performance.

## What is a Manifest?

A manifest is a JSON file that records the state of your files at a specific point in time. It stores:
- File paths
- File sizes
- Last modification times
- Computed hashes

## Saving a Manifest (`--output-manifest`)

To create a manifest, use the `--output-manifest` flag when running hashi.

```bash
# Hash all files and save the results to baseline.json
hashi --output-manifest baseline.json
```

## Using a Manifest (`--manifest`, `--only-changed`)

To use a previous manifest for incremental hashing, provide it via `--manifest` and use the `--only-changed` flag.

```bash
# Only process files that have changed since baseline.json was created
hashi --manifest baseline.json --only-changed
```

### How Change Detection Works
hashi considers a file "changed" if:
1. It exists in the current run but is missing from the manifest (Added)
2. Its size has changed (Modified)
3. Its modification time has changed (Modified)

If a file's size and modification time match the manifest, hashi assumes the content hasn't changed and skips it.

## CI/CD Workflow Example

A common pattern in CI/CD is to compare the current branch against a baseline (like the `main` branch).

```bash
# 1. On main branch: Create baseline
hashi -r --output-manifest main-baseline.json

# 2. On feature branch: Only hash what changed
hashi -r --manifest main-baseline.json --only-changed --json > changes.json

# 3. Analyze changes
jq '.processed' changes.json
```

## Best Practices

- **Atomic Updates**: hashi uses atomic writes when saving manifests, so your baseline won't be corrupted if the process is interrupted.
- **Algorithm Consistency**: Ensure you use the same hash algorithm (`--algo`) when creating and using manifests.
- **Relative Paths**: hashi stores paths as provided on the command line. For best results, run hashi from the same root directory each time.
