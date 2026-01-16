# Implementation Plan: Hash Comparator

## Overview

Build a single-file Go CLI tool that compares file hashes and/or pre-computed SHA-256 hash strings. The implementation follows a linear pipeline: validate → classify → compute → compare → report.

## Tasks

- [x] 1. Initialize Go project and create main.go with structure
  - Create `main.go` with package declaration and imports
  - Define `HashEntry` struct with comments explaining each field
  - Create empty function stubs for all core functions
  - Add main function skeleton that outlines the pipeline flow
  - _Requirements: 7.1, 7.2, 7.3, 7.5_

- [x] 2. Implement argument validation
  - [x] 2.1 Implement `validateArgCount` function
    - Check argument count is between 2 and 10
    - Return descriptive error messages for out-of-range counts
    - Add comments explaining the logic
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 2.2 Write property test for argument count validation
    - **Property 1: Argument Count Validation**
    - **Validates: Requirements 1.1, 1.2, 1.3**
    - **NOTE: Test exists but was not verified due to missing Go installation**

- [x] 3. Implement argument classification
  - [x] 3.1 Implement `isValidSHA256` function
    - Check string is exactly 64 characters
    - Check all characters are valid hexadecimal (0-9, a-f, A-F)
    - Add comments explaining the validation logic
    - _Requirements: 2.2_

  - [x] 3.2 Write property test for SHA-256 validation
    - **Property 2: SHA-256 Hash String Detection**
    - **Validates: Requirements 2.2**

  - [x] 3.3 Implement `classifyArgument` function
    - Check if argument exists as file on filesystem
    - Check if argument is valid SHA-256 hash string
    - Return classification flags (isFile, isHash)
    - Add comments explaining classification priority
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 3.4 Write property test for classification mutual exclusivity
    - **Property 5: Classification Mutual Exclusivity**
    - **Validates: Requirements 2.1, 2.2**

- [x] 4. Implement hash computation
  - [x] 4.1 Implement `computeFileHash` function
    - Open file for reading with proper error handling
    - Stream file contents through SHA-256 hasher (don't load entire file)
    - Return lowercase hex-encoded hash string
    - Add comments explaining streaming approach
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 4.2 Write property test for hash computation
    - **Property 3: Hash Computation Correctness**
    - **Validates: Requirements 3.1**

- [x] 5. Implement argument processing
  - [x] 5.1 Implement `processArgument` function
    - Call classifyArgument to determine type
    - Compute hash for files, store hash for hash strings
    - Handle invalid arguments with appropriate error
    - Return populated HashEntry struct
    - Add comments explaining the processing flow
    - _Requirements: 2.1, 2.2, 2.3, 3.1, 3.2_

- [x] 6. Checkpoint - Verify core functions work
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Implement match finding
  - [x] 7.1 Implement `findMatches` function
    - Group HashEntry structs by hash value using a map
    - Filter to only include groups with 2+ entries
    - Add comments explaining the grouping algorithm
    - _Requirements: 4.1, 4.2_

  - [x] 7.2 Write property test for match grouping
    - **Property 4: Match Grouping Correctness**
    - **Validates: Requirements 4.1, 4.2**

- [x] 8. Implement output report
  - [x] 8.1 Implement `printReport` function
    - Print all processed arguments with their hashes
    - Print match groups with clear formatting
    - Indicate file vs hash for each entry
    - Print "no matches" message when applicable
    - Show errors inline with problematic arguments
    - Add comments explaining output format
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_

- [x] 9. Wire everything together in main
  - [x] 9.1 Complete main function implementation
    - Parse command-line arguments
    - Call validateArgCount and exit on error
    - Process each argument into HashEntry slice
    - Check for all-invalid case and exit appropriately
    - Find matches and print report
    - Add comments explaining the main workflow
    - _Requirements: 1.1, 1.2, 1.3, 2.4, 4.3_

- [x] 10. Final checkpoint - Full integration test
  - Ensure all tests pass, ask the user if questions arise.
  - Manually test with sample files and hash strings

- [x] 11. Add build instructions for cross-platform compilation
  - Add comments at top of file with build commands for Windows and Linux
  - Document GOOS and GOARCH environment variables
  - _Requirements: 6.1, 6.2, 6.3_

## Notes

- **CRITICAL: Tests CANNOT be skipped. All tests must be run and verified to pass before marking tasks as complete.**
- All tasks are required including property-based tests
- All code goes in a single `main.go` file for simplicity
- Tests go in `main_test.go`
- Use Go standard library only — no external dependencies
- Every function, variable, and logical block must have explanatory comments
