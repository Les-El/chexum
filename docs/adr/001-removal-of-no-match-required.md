# Removal of --no-match-required Flag

## Decision

Removed the `--no-match-required` flag from hashi. This flag was redundant and could be replaced with simple negation.

## Rationale

### 1. Redundant with Negation
Everything `--no-match-required` did can be accomplished with `!`:

```bash
# Old way (with --no-match-required)
hashi --no-match-required *.txt && echo "All unique"

# New way (with negation)
! hashi --match-required *.txt && echo "All unique"
# or
hashi --match-required *.txt || echo "All unique"
```

### 2. Confusing Name
The name "no-match-required" was ambiguous:
- Could mean "matching is optional" (not what it did)
- Actually meant "require that files DON'T match"
- Didn't clearly communicate intent

### 3. Rare Use Case
- Most users want to check if files match, not if they don't
- Checking for uniqueness is less common than checking for matches
- The flag would be used by <10% of users

### 4. Simpler API
- Fewer flags = easier to learn
- Less cognitive load for users
- Cleaner conflict resolution logic

## What Was Removed

### Code Changes
1. **Config struct**: Removed `NoMatchRequired` field
2. **ConfigFile struct**: Removed `NoMatchRequired` field
3. **Flag definition**: Removed `--no-match-required` flag
4. **Conflict rules**: Removed mutual exclusivity rule
5. **Validation logic**: Removed checks for `NoMatchRequired`
6. **Tests**: Removed tests for `--no-match-required`

### Documentation Changes
1. **Help text**: Removed from EXIT CODE CONTROL section
2. **BOOL_MODE.md**: Updated to show negation approach
3. **FLAG_PRECEDENCE.md**: Removed from conflicts section
4. **IMPLEMENTATION_SUMMARY.md**: Removed references

## Migration Guide

For users who might have been using `--no-match-required`:

### Before (with --no-match-required)
```bash
# Check if all files are unique
hashi --no-match-required *.txt
exit_code=$?

# With boolean mode
hashi -b --no-match-required *.txt
```

### After (with negation)
```bash
# Check if all files are unique
! hashi --match-required *.txt
exit_code=$?

# With boolean mode
! hashi -b --match-required *.txt
```

### Scripting Examples

#### Verify Uniqueness
```bash
# Before
if hashi --no-match-required *.txt; then
    echo "All unique"
fi

# After
if ! hashi --match-required *.txt; then
    echo "All unique"
fi
```

#### Exit on Duplicates
```bash
# Before
hashi --no-match-required *.txt || {
    echo "Duplicates found!"
    exit 1
}

# After
hashi --match-required *.txt && {
    echo "Duplicates found!"
    exit 1
}
```

## Benefits of Removal

### 1. Clearer Intent
Negation makes the logic explicit:
```bash
# Clear: "if NOT (matches found)"
! hashi --match-required *.txt

# Was unclear: "if no-match-required"
hashi --no-match-required *.txt
```

### 2. Standard Unix Pattern
Using `!` for negation is a standard Unix pattern that all shell users understand.

### 3. Simpler Mental Model
Users only need to understand:
- Default: check if all match
- `--match-required`: check if any match
- Negation: invert the result

### 4. Fewer Edge Cases
No need to handle:
- `--match-required` + `--no-match-required` conflicts
- `-b --no-match-required` combinations
- Documentation of two opposite flags

## Testing

All tests pass after removal:
- ✓ Config parsing tests updated
- ✓ Conflict resolution tests updated
- ✓ Boolean mode tests updated
- ✓ Help text updated
- ✓ Documentation updated

## Status

✅ **COMPLETE** - `--no-match-required` has been completely removed from the codebase.

The project is now simpler, clearer, and easier to maintain.
