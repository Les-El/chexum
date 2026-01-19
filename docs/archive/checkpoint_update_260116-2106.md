# Checkpoint Update: 260116-2106

## Considerations about hashi Project

  1. What will users use hashi for?
   * De-duplication: Finding files with identical content but different names or locations (facilitated by the grouped output feature).
   * Download Verification: Ensuring a file downloaded from the web matches the checksum provided by the source.
   * Bulk Auditing: Quickly hashing thousands of files to create a manifest or check for changes over time.
   * Archive Integrity: Checking if a ZIP file is corrupted without needing to manually unzip and verify each file.
   * Automation: Integrating hash checks into CI/CD pipelines or local scripts (using --json or --bool).

  2. What questions will users ask?
   * "Are these two folders identical in content?"
   * "Which of my recent downloads are corrupted?"
   * "Did I already back up this specific file somewhere else?"
   * "What is the hash of this file using SHA-512 instead of the default?"
   * "How long is this huge hashing operation going to take?"
   * "Can I compare a flat directory list against a recursive search of another directory?"
   * "Can I provide a file containing a list of arguments and another for flags?"

  3. What information do users need?
   * Clear Match/Mismatch Status: A binary "yes/no" or visual green/red indicator.
   * Progress Feedback: Especially for multi-gigabyte files or directories with 10k+ files (provided by the progress bars and spinners).
   * Error Context: If a file fails, why? (Permission denied, file locked, disk error).
   * Structured Data: For scripts, they need consistent JSON or tab-separated values.

  4. What use cases needs a better/friendlier solution?
   * Visual Clarity: Standard tools like sha256sum just dump a list. hashi solves this by grouping matching hashes, making it immediately obvious which files are duplicates.
   * Automatic Algorithm Detection: Users shouldn't have to specify --algo md5 if they provide a 32-character string. hashi maps input length to the likely algorithm.
   * Human-Readable Filters: Instead of piping find into xargs, hashi provides built-in --min-size and --modified-after flags.
   * Native Archive Handling: Most tools hash the .zip file itself. hashi can look inside to verify the internal CRC32 checksums.

  5. When will users be confused, and how can we prevent it and/or respond?
   * Flag Conflicts: If they use --json and --verbose simultaneously (addressed by the Flag Precedence Hierarchy and clear warnings).
   * Hidden Files: Users might wonder why hashi missed a .config file (it ignores hidden files by default, requiring the -H flag).
   *   **ZIP vs. Raw Hashing:** Users might be confused if hashing a ZIP returns a different result than expected. By defaulting to standard hashing and requiring an explicit `--verify` flag for internal integrity checks, we follow the Principle of Least Astonishment. The `--raw` flag remains as an escape hatch for any future "smart" features.
   * Performance Safety: Running a hash on a 100GB directory by accident could hang a system; the Confirmation Prompt for massive operations prevents this.

## Concepts Proposed

  1. File List vs. Hash List (Bulk Verification)
   * The Problem: sha256sum -c expects a specific format (hash  filename). If a user has a random list of hashes and a random list of files, they have to write a loop.
   * The `hashi` Solution: We can implement a mode where hashi takes a set of target hashes and a set of source files and tells you which sources match which targets.
       * Proposed Command: hashi --targets hashes.txt *.iso (where hashes.txt is just a list of hex strings).

  2. "Fail Fast" (Early Exit)
   * The Problem: In CI/CD, if you’re verifying 1,000 files and the first one fails, you don't want to wait 20 minutes for the other 999.
   * The `hashi` Solution: A "fail-fast" flag.
       * Proposed Command: hashi --fail-fast 1 *.bin (exit immediately on the first mismatch).
       * Proposed Command: hashi --fail-limit 5 *.bin (allow up to 5 mismatches before aborting).

  3. Reporting Filters (Only Matches / Only Mismatches)
   * The Problem: Standard tools show everything. If you have 10,000 matches and 1 mismatch, the mismatch gets lost in the noise.
   * The `hashi` Solution: Negative and positive result filtering.
       * Proposed Command: hashi --return-matches 
       * Proposed Command: hashi --return-mismatches 

  4. Duplicate Thresholds
   * The Problem: Finding any duplicate is easy. Finding "over-duplicated" files (e.g., "show me files that appear more than 3 times") is harder.
   * The `hashi` Solution: A threshold flag for the grouping logic.
       * Proposed Command: hashi --min-matches 3 (only show groups with 3 or more files).

  Should hashi do all this?
  We shouldn't cater to every whim. We can consider those features that align perfectly with our "User-First" and "Script-Friendly" mandates. Common pain points where the "scripting way" (combining find, xargs, grep,
  and awk) is brittle and error-prone are prime candidates.

  I recommend adding the following to our task list:
   1. --fail-fast [N] (Stop after N mismatches).
   2. --return-matches and --return-mismatches flags.
   3. --min-matches [N] (Filter grouping output).
   4. Improved "Bulk Verification" syntax (handling lists of hashes more gracefully).

  5. Directory Set Comparison
   * The Problem: Users often want to know if Directory A and Directory B contain the exact same content (ignoring paths, or respecting them).
   * The `hashi` Solution: A mode that treats directories as sets of hashes and compares the sets.
       * Proposed Command: `hashi --diff dirA dirB` (shows what's in A but not B, and vice versa).

  6. External Argument & Flag Files
   * The Problem: Command lines have length limits (ARG_MAX). Scripts sometimes need to pass 10,000 files.
   * The `hashi` Solution:
       * `@file`: Read arguments from `file` (standard convention in some tools).
       * `--flagfile file`: Load flags from a separate text file. (Note: Task 4 covers config files, but "flagfiles" are often one-shot overrides).


## Manifest-as-Input (The "Infinite Loop" Policy)
To support complex comparisons without "Flag Bleed" (where a filter for Set A accidentally applies to Set B), `hashi` must be able to read its own JSON/Plain output as an input source.
- This allows users to "Pipe" filtered results from one `hashi` operation into another.
- **Example:** `hashi --diff filtered_set_A.json filtered_set_B.json`

## The "Why" Principle (Explainable Logic)
To maintain Developer Continuity and User Trust, no automated decision (like conflict resolution) should be silent.
- **Rule Co-location:** Conflict rules must be stored alongside their "Reasoning" text.
- **Verbose Transparency:** When `--verbose` is active, `hashi` should output the reasoning for every handled conflict (e.g., `* Note: --json was chosen over --verbose to ensure machine-readable output.`).

##  Actionable Error Pattern
`hashi` errors will not just state the problem; they will suggest the solution.
- **Structure:** `[Error Category] + [What Happened] + [Suggested Action]`.
- **Atoms of Messaging:** Messages are not hard-coded into logic Atoms. Instead, Logic Atoms return **Error Codes**, which the `ErrorHandler` Atom maps to human-friendly templates and suggestions.


#### 1.1 Suggested functionality if it can be accomplished atomicaly with base operations
- **Non-Conflict Policy:** The flags --return-matches and --return-mismatches flags are **not** mutually exclusive. 
- **Sequential Output:** If both are provided, `hashi` will output results in batches. 
- **Order Importance:** The order in which the flags appear on the command line determines the order of the output batches.
    - Example: `hashi --return-mismatches --return-matches` shows failures first, then successes.

### 2. Explicit Argument Typing
- **Autodetector Priority:** Files always take precedence over hash strings if an argument matches both.
- **Escape Hatches:**
    - `--hasharg <string>`: Explicitly treats the argument as a hash string, even if a file with that name exists.
    - `--filearg <path>`: Explicitly treats the argument as a file path, even if it looks like a hash string.

### 3. Smart Appending
- **Validation Policy:** `hashi` will only append to an existing file if it is a valid file of the expected type.
- **JSON Append:** Appending to a JSON file must maintain valid JSON structure (e.g., merging into an array or object rather than just appending text to the end).
- **Text Append:** Standard line-based appending for plain text formats.

### 4. Performance & CI/CD Controls
- **Fail-Fast:** `--fail-fast [N]` allows stopping the operation after N mismatches are encountered.
- **Match Thresholds:** `--min-matches [N]` filters the grouped output to only show hashes that appear at least N times (useful for finding widespread duplicates).

## Impact on Implementation Plan
(Argument Classification):** Must incorporate `--hasharg` and `--filearg` overrides.
(Boolean Output):** Update to reflect that `--return-mismatches` is the natural counterpart for failure-seeking scripts.
 (File Output):** Add validation logic for existing files before appending.
-(Conflict Resolver):** 
    - Update rules to allow `--return-matches` + `--return-mismatches`.
    - **Co-location Requirement:** Every rule must include an `Explanation` string for use in `--verbose` or Error messages.
- (Batch Output):** Implement logic to track flag order for result sectioning.

## Operational Philosophy: Atoms & Molecules

To handle both simple and complex user needs without creating "spaghetti code," `hashi` follows a modular composition strategy.

### 1. The Building Blocks
- **Atoms (Unit Operations):** These are single-purpose, low-level functions (e.g., `ComputeHash`, `GetFileSize`, `DetectTTY`, `CompareStrings`).
- **Molecules (Composite Features):** These are user-facing features represented by flags (e.g., `--diff`, `--bool`, `--raw`). They are built by orchestrating multiple **Atoms**.
    - *Example:* `--bool` is a Molecule that combines the `Compare` Atom with the `QuietOutput` and `BooleanFormatter` Atoms.

### 2. The Processing Pipeline (The "Correct Order")
Every request must pass through a strict, linear pipeline to ensure consistency and prevent logic leaks:

1.  **Ingestion:** Collect arguments from CLI, Env, and Config files.
2.  **Classification:** Determine if arguments are Files, Hashes, or Directories.
3.  **Filtration:** Apply size, date, and pattern filters (discarding irrelevant data early).
4.  **Transformation:** Perform the "heavy lifting" (Hashing, ZIP reading, recursive walking).
5.  **Analysis:** Logic gates (Matching, set differencing, integrity checks).
6.  **Formatting:** Map internal results to the requested UI (JSON, Text, Progress bars).
7.  **Emission:** Final delivery to Stdout, Stderr, or a File.

### 3. Composition Rule
Whenever a new "Complex Operation" is proposed, we must first check if the necessary **Atoms** exist in the pipeline. If not, we implement the **Atoms** first. This ensures that a feature like `--diff` benefits from the same Progress Bar and Filtering logic as a simple hash check.

## Feature Evaluation Framework

To maintain project focus and ensure high-quality releases, every new feature proposal must be evaluated against the following criteria:

1.  **Does it provide clear benefit?** Does it solve a real-world problem or significantly improve the user experience?
2.  **Is it worthy of the v1.0 release?** Is it a core necessity or a "nice-to-have" that should be deferred to a later version?
3.  **Can it be implemented by manipulating existing functions?** 
    - If yes, identify the foundational functions needed first. 
    - Prioritize implementing these building blocks to minimize code duplication and maintain architectural integrity.
4.  **Composition vs. Convenience:** If a use case can already be solved by combining existing `hashi` commands (e.g., via piping or multiple passes), is the new syntax worth the complexity?
    - **Weight for:** High-frequency workflows, common user pain points, or operations where manual composition is error-prone.
    - **Weight against:** Niche use cases that can be solved with a simple 2-line shell script or a single pipe.
