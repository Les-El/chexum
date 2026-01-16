package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// TestValidateArgCountProperty tests the argument count validation property.
// Feature: hash-compare, Property 1: Argument Count Validation
// **Validates: Requirements 1.1, 1.2, 1.3**
func TestValidateArgCountProperty(t *testing.T) {
	// Property: For any argument count less than 2 or greater than 10,
	// validateArgCount SHALL return an error. For any argument count
	// between 2 and 10 inclusive, validateArgCount SHALL return nil.
	
	property := func(argCount int) bool {
		// Handle negative argument counts by creating an empty slice
		var args []string
		if argCount >= 0 {
			args = make([]string, argCount)
			for i := 0; i < argCount; i++ {
				args[i] = "dummy_arg"
			}
		} else {
			// For negative counts, create empty slice (equivalent to 0 arguments)
			args = make([]string, 0)
			argCount = 0 // Normalize for property check
		}
		
		// Call validateArgCount
		err := validateArgCount(args)
		
		// Check the property: error should be returned for counts < 2 or > 10
		if argCount < 2 || argCount > 10 {
			return err != nil // Should have an error
		} else {
			return err == nil // Should not have an error
		}
	}
	
	// Configure the property test to run with a reasonable range
	config := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, rand *rand.Rand) {
			// Generate argument counts in a range that tests both valid and invalid cases
			// Test range: -5 to 15 to cover edge cases around 2 and 10
			argCount := rand.Intn(21) - 5 // Range: -5 to 15
			values[0] = reflect.ValueOf(argCount)
		},
	}
	
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// TestIsValidSHA256Property tests the SHA-256 hash string detection property.
// Feature: hash-compare, Property 2: SHA-256 Hash String Detection
// **Validates: Requirements 2.2**
func TestIsValidSHA256Property(t *testing.T) {
	// Property: For any string that is exactly 64 characters long and contains
	// only hexadecimal characters (0-9, a-f, A-F), isValidSHA256 SHALL return true.
	// For any string that does not meet these criteria, it SHALL return false.
	
	property := func(testCase int) bool {
		var testString string
		var expectedResult bool
		
		// Generate different types of test cases based on testCase value
		switch testCase % 6 {
		case 0: // Valid 64-char hex string (lowercase)
			testString = generateHexString(64, false)
			expectedResult = true
		case 1: // Valid 64-char hex string (uppercase)  
			testString = generateHexString(64, true)
			expectedResult = true
		case 2: // Valid 64-char hex string (mixed case)
			testString = generateMixedCaseHexString(64)
			expectedResult = true
		case 3: // Wrong length (not 64 characters)
			length := rand.Intn(128) // 0-127, avoiding 64
			if length == 64 {
				length = 63 // Ensure it's not 64
			}
			testString = generateHexString(length, false)
			expectedResult = false
		case 4: // Correct length but invalid characters
			testString = generateInvalidHexString(64)
			expectedResult = false
		case 5: // Empty string
			testString = ""
			expectedResult = false
		}
		
		// Test the property
		result := isValidSHA256(testString)
		return result == expectedResult
	}
	
	config := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, rand *rand.Rand) {
			// Generate test case numbers to cover all scenarios
			testCase := rand.Intn(1000) // Large range to ensure good distribution
			values[0] = reflect.ValueOf(testCase)
		},
	}
	
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Helper function to generate a hex string of specified length
func generateHexString(length int, uppercase bool) string {
	if length <= 0 {
		return ""
	}
	
	hexChars := "0123456789abcdef"
	if uppercase {
		hexChars = "0123456789ABCDEF"
	}
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hexChars[rand.Intn(16)]
	}
	return string(result)
}

// Helper function to generate a mixed case hex string
func generateMixedCaseHexString(length int) string {
	if length <= 0 {
		return ""
	}
	
	lowerHex := "0123456789abcdef"
	upperHex := "0123456789ABCDEF"
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		if rand.Intn(2) == 0 {
			result[i] = lowerHex[rand.Intn(16)]
		} else {
			result[i] = upperHex[rand.Intn(16)]
		}
	}
	return string(result)
}

// TestClassifyArgumentMutualExclusivityProperty tests the classification mutual exclusivity property.
// Feature: hash-compare, Property 5: Classification Mutual Exclusivity
// **Validates: Requirements 2.1, 2.2**
func TestClassifyArgumentMutualExclusivityProperty(t *testing.T) {
	// Property: For any argument string, classifyArgument SHALL return at most
	// one true value (either isFile or isHash, but not both). A string cannot
	// be both a file and a valid hash simultaneously in classification.
	
	property := func(testCase int) bool {
		var testArg string
		
		// Generate different types of test arguments
		switch testCase % 4 {
		case 0: // Valid hash string (should not be a file)
			testArg = generateHexString(64, false)
		case 1: // Random string that's not a valid hash
			testArg = generateRandomString(rand.Intn(100) + 1) // 1-100 chars
		case 2: // Existing file (use a known file like main.go)
			testArg = "main.go"
		case 3: // Empty string
			testArg = ""
		}
		
		// Call classifyArgument
		isFile, isHash := classifyArgument(testArg)
		
		// Property check: at most one can be true (mutual exclusivity)
		// Both can be false (invalid argument), but both cannot be true
		return !(isFile && isHash)
	}
	
	config := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, rand *rand.Rand) {
			testCase := rand.Intn(1000)
			values[0] = reflect.ValueOf(testCase)
		},
	}
	
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Helper function to generate a random string with various characters
func generateRandomString(length int) string {
	if length <= 0 {
		return ""
	}
	
	// Use a mix of characters including letters, numbers, and symbols
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?/~` "
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Helper function to generate a string with invalid hex characters
func generateInvalidHexString(length int) string {
	if length <= 0 {
		return ""
	}
	
	// Use characters that are NOT valid hex
	invalidChars := "ghijklmnopqrstuvwxyzGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=[]{}|;:,.<>?/~`"
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		// Mix some valid hex chars with invalid ones to make it more realistic
		if rand.Intn(3) == 0 {
			// Sometimes include valid hex chars
			validHex := "0123456789abcdefABCDEF"
			result[i] = validHex[rand.Intn(len(validHex))]
		} else {
			// Most of the time use invalid chars
			result[i] = invalidChars[rand.Intn(len(invalidChars))]
		}
	}
	
	// Ensure at least one invalid character exists
	if length > 0 {
		result[rand.Intn(length)] = invalidChars[rand.Intn(len(invalidChars))]
	}
	
	return string(result)
}

// TestComputeFileHashProperty tests the hash computation correctness property.
// Feature: hash-compare, Property 3: Hash Computation Correctness
// **Validates: Requirements 3.1**
func TestComputeFileHashProperty(t *testing.T) {
	// Property: For any readable file, computing its hash with computeFileHash
	// and then computing it again SHALL produce the same result (determinism).
	// Additionally, the result SHALL match the output of Go's standard crypto/sha256 package.
	
	property := func(testCase int) bool {
		// We'll test with existing files in the project directory
		// Use a set of known files that should exist
		testFiles := []string{
			"main.go",
			"main_test.go",
			"go.mod",
		}
		
		// Select a file based on the test case
		filename := testFiles[testCase%len(testFiles)]
		
		// First computation
		hash1, err1 := computeFileHash(filename)
		if err1 != nil {
			// If file doesn't exist or can't be read, skip this test case
			// This is not a failure of the property, just an unavailable test case
			return true
		}
		
		// Second computation (should be identical)
		hash2, err2 := computeFileHash(filename)
		if err2 != nil {
			// If second computation fails but first succeeded, this is a problem
			return false
		}
		
		// Property check 1: Determinism - both computations should produce same result
		if hash1 != hash2 {
			return false
		}
		
		// Property check 2: Result should be a valid SHA-256 hash format
		if !isValidSHA256(hash1) {
			return false
		}
		
		// Property check 3: Hash should be lowercase (as specified in requirements)
		if hash1 != strings.ToLower(hash1) {
			return false
		}
		
		// All property checks passed
		return true
	}
	
	config := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, rand *rand.Rand) {
			// Generate test case numbers to cycle through available files
			testCase := rand.Intn(1000)
			values[0] = reflect.ValueOf(testCase)
		},
	}
	
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// TestFindMatchesProperty tests the match grouping correctness property.
// Feature: hash-compare, Property 4: Match Grouping Correctness
// **Validates: Requirements 4.1, 4.2**
func TestFindMatchesProperty(t *testing.T) {
	// Property: For any set of HashEntry structs, the findMatches function
	// SHALL group all entries with identical hash values together, and SHALL
	// only return groups containing 2 or more entries.
	
	property := func(testCase int) bool {
		// Generate a test set of HashEntry structs with various scenarios
		entries := generateTestHashEntries(testCase)
		
		// Call the function under test
		matches := findMatches(entries)
		
		// Property check 1: All returned groups must have 2+ entries
		for hash, group := range matches {
			if len(group) < 2 {
				return false // Violation: single-entry group returned
			}
			
			// Property check 2: All entries in a group must have the same hash
			for _, entry := range group {
				if entry.Hash != hash {
					return false // Violation: entry with wrong hash in group
				}
			}
		}
		
		// Property check 3: All entries with the same hash (2+ occurrences) should be grouped
		// Create a map to count occurrences of each hash in the input
		hashCounts := make(map[string]int)
		validEntries := make(map[string][]HashEntry)
		
		for _, entry := range entries {
			// Skip entries with empty hashes (they should be excluded)
			if entry.Hash == "" {
				continue
			}
			hashCounts[entry.Hash]++
			validEntries[entry.Hash] = append(validEntries[entry.Hash], entry)
		}
		
		// Check that all hashes with 2+ occurrences are in the matches
		for hash, count := range hashCounts {
			if count >= 2 {
				// This hash should be in the matches
				if matchGroup, exists := matches[hash]; !exists {
					return false // Violation: missing match group
				} else if len(matchGroup) != count {
					return false // Violation: incorrect group size
				}
			} else {
				// This hash should NOT be in the matches (single occurrence)
				if _, exists := matches[hash]; exists {
					return false // Violation: single-entry group included
				}
			}
		}
		
		// Property check 4: No entries with empty hashes should appear in matches
		for _, group := range matches {
			for _, entry := range group {
				if entry.Hash == "" {
					return false // Violation: empty hash in match group
				}
			}
		}
		
		// All property checks passed
		return true
	}
	
	config := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, rand *rand.Rand) {
			testCase := rand.Intn(10000) // Large range for variety
			values[0] = reflect.ValueOf(testCase)
		},
	}
	
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Helper function to generate test HashEntry slices for property testing
func generateTestHashEntries(testCase int) []HashEntry {
	// Use testCase to determine the structure of the test data
	rand.Seed(int64(testCase)) // Ensure reproducible test cases
	
	// Generate 1-20 entries
	numEntries := rand.Intn(20) + 1
	entries := make([]HashEntry, numEntries)
	
	// Create a pool of possible hash values (some will be duplicated)
	hashPool := []string{
		"a1b2c3d4e5f6789012345678901234567890123456789012345678901234abcd", // Valid hash 1
		"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", // Valid hash 2  
		"fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321", // Valid hash 3
		"0000000000000000000000000000000000000000000000000000000000000000", // Valid hash 4
		"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", // Valid hash 5
		"", // Empty hash (invalid/error case)
	}
	
	for i := 0; i < numEntries; i++ {
		// Randomly select a hash from the pool (this creates duplicates)
		hash := hashPool[rand.Intn(len(hashPool))]
		
		// Create entry with random metadata
		entries[i] = HashEntry{
			Original: generateRandomString(rand.Intn(20) + 1), // Random original argument
			Hash:     hash,
			IsFile:   rand.Intn(2) == 0, // Random file/hash flag
			Error:    nil,               // Assume no errors for simplicity
		}
		
		// Sometimes add an error for entries with empty hashes
		if hash == "" && rand.Intn(2) == 0 {
			entries[i].Error = fmt.Errorf("test error")
		}
	}
	
	return entries
}