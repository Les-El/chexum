# Flag Logic: Precedence, Conflicts, and Modes

This document details the design and implementation of flag handling in the `hashi` CLI. It covers precedence rules, conflict resolution strategies, and special output modes like Boolean and Quiet.

## Design Philosophy

1.  **Forgiving by Default:** Users shouldn't hit walls when combining flags. If a user says `hashi --json --verbose`, the tool should figure out what they mean (JSON implies no verbose commentary) rather than crashing with an error.
2.  **Clear Communication:** When the tool resolves a conflict (e.g., overriding `--verbose` because `--json` was set), it should warn the user, unless the user has explicitly requested silence.
3.  **Predictable Behavior:** Precedence rules must be consistent and documented.
4.  **Scriptability:** Modes like `--bool` and `--quiet` are designed specifically for machine consumption.

---

## Flag Precedence Hierarchy

The `hashi` CLI uses a strict hierarchy to determine the output format. Higher precedence flags override lower ones.

**Precedence:**
1.  **`--bool` / `-b`** (Highest)
    *   *Implies:* `--quiet`, `--match-required` (mostly)
    *   *Output:* `true` or `false`
2.  **`--quiet` / `-q`**
    *   *Behavior:* Suppresses all commentary, warnings, and headers.
3.  **`--json`**
    *   *Output:* Structured JSON.
4.  **`--plain`**
    *   *Output:* minimal plain text.
5.  **`--verbose` / `-v`** (Lowest)
    *   *Output:* Detailed human-readable commentary.

### Resolution Rules

*   **Overrides:** If a higher precedence flag is present, it overrides lower ones.
    *   `hashi --json --verbose` → **JSON** wins. (Warning displayed: "--json overrides --verbose")
    *   `hashi --quiet --verbose` → **Quiet** wins. (Warning suppressed: Quiet mode implies no warnings)
*   **Same-Level Precedence:** If flags have the same precedence (e.g., `--json` and `--plain`), the **last flag specified wins**.
    *   `hashi --json --plain` → **Plain** wins.

---

## Conflict Resolution

The `internal/conflict` package handles flag interactions.

### Conflict Types

1.  **Overrides (Resolvable):**
    *   Conflict between output formats (e.g., `--json` vs `--verbose`).
    *   **Action:** The winner is selected based on precedence. A warning is generated.
2.  **Mutually Exclusive (Fatal):**
    *   Conflict between fundamental operation modes (e.g., `--raw` vs `--verify`).
    *   **Action:** The tool exits with an error.
    *   *Example:* `✗ --raw treats files as bytes, --verify checks internal checksums. Choose one.`

### Smart Warning Suppression

Warnings are designed to educate, not annoy.
*   **Normal Mode:** Warnings are shown. `hashi --json --verbose` prints a warning.
*   **Quiet/Bool Mode:** Warnings are suppressed. `hashi -b --verbose` is silent (except for the result).

---

## Special Modes

### 1. Boolean Mode (`-b` / `--bool`)

**Purpose:** Simple `true`/`false` output for scripting.

**Behavior:**
*   Sets `--quiet` (suppresses commentary).
*   Outputs exactly `true` or `false` to stdout.
*   Exit code reflects the result (0 = true, 1 = false).

**Usage Examples:**

*   **"Do all files match?"** (Default)
    ```bash
    $ hashi -b file1.txt file2.txt
    true
    ```

*   **"Are there any duplicates?"** (With `--match-required`)
    ```bash
    $ hashi -b --match-required *.txt
    ```

*   **"Are all files unique?"** (With negation)
    ```bash
    if ! hashi -b --match-required *.txt; then echo "All unique"; fi
    ```

### 2. Quiet Mode (`--quiet` / `-q`)

**Purpose:** "Just give me the results and exit code, skip the commentary."

**What it Suppresses:**
*   Progress bars/spinners.
*   "Processing X files..." messages.
*   Headers/Footers.
*   **Conflict Warnings.**

**What it Preserves:**
*   **The Output:** The actual hashes or comparison results are still printed.
*   **Errors:** Errors sent to stderr are never suppressed.
*   **Exit Codes:** Remain functional.

**Interaction:**
*   `--quiet` + `--json`: Outputs bare JSON without any surrounding text.
*   `--quiet` + `--verbose`: Quiet wins. No output.

---

## Implementation Summary

*   **Code Location:** `internal/conflict/`, `internal/config/`
*   **Key Logic:** `Resolve()` function checks flags against the hierarchy.
*   **Strategies:** `WarnOnConflict` (default) vs `ErrorOnConflict` (strict/testing).

### Future Considerations
*   `HASHI_STRICT_FLAGS=1` env var to force all conflicts to be fatal errors.
*   `--not-bool` (`-B`) for inverted boolean logic.
