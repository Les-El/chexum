# Requirements Document

## Introduction

A command-line utility that compares file hashes and/or pre-computed hash strings. The tool accepts between 2 and 10 arguments, automatically detects whether each argument is a file path or a hash string, computes hashes for files, and reports all matches found. Designed for simplicity, cross-platform compatibility (Windows and Linux), and elegant implementation in Go.

## Glossary

- **hashi**: The command-line tool being developed
- **Argument**: A command-line input that is either a file path or a hash string
- **Hash_String**: A hexadecimal string representing a computed hash (e.g., SHA-256)
- **File_Path**: A path to a file on the filesystem
- **Hash_Entry**: An internal representation containing the original argument, its hash value, and whether it was a file or provided hash
- **Match**: Two or more Hash_Entries that share the same hash value

## Requirements

### Requirement 1: Argument Acceptance

**User Story:** As a user, I want to provide between 2 and 10 arguments so that I can compare multiple files and/or hashes in a single invocation.

#### Acceptance Criteria

1. WHEN the user provides fewer than 2 arguments THEN THE hashi SHALL display an error message and exit with a non-zero status
2. WHEN the user provides more than 10 arguments THEN THE hashi SHALL display an error message and exit with a non-zero status
3. WHEN the user provides between 2 and 10 arguments THEN THE hashi SHALL accept and process all arguments

### Requirement 2: Argument Type Detection

**User Story:** As a user, I want to provide file paths and hash strings in any order so that I don't have to remember special flags or ordering rules.

#### Acceptance Criteria

1. WHEN an argument matches a valid file path that exists on the filesystem THEN THE hashi SHALL treat it as a file and compute its hash
2. WHEN an argument is a 64-character hexadecimal string (SHA-256 format) THEN THE hashi SHALL treat it as a pre-computed hash
3. WHEN an argument does not exist as a file AND is not a valid hash format THEN THE hashi SHALL display an error for that argument and continue processing remaining arguments
4. WHEN the user provides ONLY INVALID arguments THEN THE hashi SHALL display an error message and exit with a non-zero status

### Requirement 3: Hash Computation

**User Story:** As a user, I want the tool to compute SHA-256 hashes for files so that I can compare file contents reliably.

#### Acceptance Criteria

1. WHEN a file argument is provided THEN THE hashi SHALL compute its SHA-256 hash
2. WHEN a file cannot be read THEN THE hashi SHALL display an error for that file and continue processing remaining arguments
3. THE hashi SHALL compute hashes efficiently by streaming file contents rather than loading entire files into memory

### Requirement 4: Hash Comparison

**User Story:** As a user, I want all hashes compared against each other so that I can find all matching files and hashes in one run.

#### Acceptance Criteria

1. WHEN processing completes THEN THE hashi SHALL compare every hash against every other hash
2. WHEN two or more hashes match THEN THE hashi SHALL group them together in the output
3. WHEN no matches are found THEN THE hashi SHALL report that no matches were found

### Requirement 5: Output Report

**User Story:** As a user, I want a clear report showing all processed arguments and which items match so that I can quickly understand the results.

#### Acceptance Criteria

1. THE hashi SHALL display ALL processed arguments with their computed or provided hash values
2. WHEN matches are found THEN THE hashi SHALL display each match group with all matching arguments listed
3. WHEN displaying a match THEN THE hashi SHALL indicate whether each item was a file or a provided hash
4. WHEN displaying a file match THEN THE hashi SHALL show the file path
5. WHEN displaying a hash match THEN THE hashi SHALL show the hash value (or a truncated version for readability)
6. THE hashi SHALL display the shared hash value for each match group
7. WHEN no matches are found THEN THE hashi SHALL report that no matches were found

### Requirement 6: Cross-Platform Compatibility

**User Story:** As a user, I want the tool to work on both Windows and Linux so that I can use it across different systems.

#### Acceptance Criteria

1. THE hashi SHALL compile to a native binary for Windows (x86-64)
2. THE hashi SHALL compile to a native binary for Linux (x86-64)
3. THE hashi SHALL handle file paths correctly on both Windows and Linux

### Requirement 7: Code Quality

**User Story:** As a developer, I want well-commented and efficent code so that the program is lightweight, maintainable and understandable.

#### Acceptance Criteria

1. THE hashi SHALL have comments explaining every function
2. THE hashi SHALL have comments explaining every variable declaration
3. THE hashi SHALL have comments explaining every logical workflow
4. THE hashi SHALL follow Go idioms and best practices
5. THE hashi SHALL be implemented in a single source file for simplicity
6. THE hashi SHALL be designed thoughfully and elegantly
7. THE hashi SHALL be efficent WITHOUT being obscure
8. THE hashi SHALL NOT brute force the objectives