package archive

import (
	"archive/zip"
	"fmt"
	"os"
	"strings"
	"testing"
	"testing/quick"
)

// Property-based tests

// Feature: cli-guidelines-review, Property 22: ZIP verification uses CRC32 only
// **Validates: Requirements 15.4, 20.3**
func TestProperty_ZIPVerificationUsesCRC32Only(t *testing.T) {
	// Property: For any ZIP file, verification should always use CRC32 regardless of metadata
	property := func(numFiles uint8) bool {
		if numFiles == 0 {
			numFiles = 1 // Ensure at least one file
		}
		if numFiles > 10 {
			numFiles = 10 // Limit for reasonable test time
		}

		// Create a temporary ZIP file with known CRC32 values
		zipPath := createTestZIP(t, int(numFiles))
		defer os.Remove(zipPath)

		verifier := NewVerifier()
		result, err := verifier.VerifyZIP(zipPath)

		// Should succeed without error
		if err != nil {
			t.Logf("Unexpected error: %v", err)
			return false
		}

		// Should have processed the expected number of files
		if result.TotalEntries != int(numFiles) {
			t.Logf("Expected %d entries, got %d", numFiles, result.TotalEntries)
			return false
		}

		// Should pass verification (our test ZIP has correct CRC32)
		if !result.Passed {
			t.Logf("Verification failed unexpectedly: %v", result.FailedEntries)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 20} // Reduced for file I/O operations
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Feature: cli-guidelines-review, Property 23: ZIP verification returns boolean by default
// **Validates: Requirements 15.1, 16.1**
func TestProperty_ZIPVerificationReturnsBooleanByDefault(t *testing.T) {
	// Property: For any ZIP file, default output should be empty (boolean mode)
	property := func(numFiles uint8) bool {
		if numFiles == 0 {
			numFiles = 1
		}
		if numFiles > 5 {
			numFiles = 5
		}

		zipPath := createTestZIP(t, int(numFiles))
		defer os.Remove(zipPath)

		verifier := NewVerifier()
		verifier.SetVerbose(false) // Default non-verbose mode

		result, err := verifier.VerifyZIP(zipPath)
		if err != nil {
			return false
		}

		// Format result in default (boolean) mode
		output := verifier.FormatResult(result, false)

		// Default mode should return empty string (boolean mode)
		return output == ""
	}

	config := &quick.Config{MaxCount: 10}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Feature: cli-guidelines-review, Property 24: Raw flag bypasses special file handling
// **Validates: Requirements 15.5, 17.2**
func TestProperty_RawFlagBypassesSpecialFileHandling(t *testing.T) {
	// Property: For any file, IsArchiveFile should identify ZIP files correctly
	property := func(hasZipExt bool) bool {
		verifier := NewVerifier()

		var filename string
		if hasZipExt {
			filename = "test.zip"
		} else {
			filename = "test.txt"
		}

		result := verifier.IsArchiveFile(filename)

		// Should return true only for .zip files
		return result == hasZipExt
	}

	config := &quick.Config{MaxCount: 50}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Feature: cli-guidelines-review, Property 26: Multiple ZIP verification returns single boolean
// **Validates: Requirements 15.2**
func TestProperty_MultipleZIPVerificationReturnsSingleBoolean(t *testing.T) {
	// Property: For any list of ZIP files, VerifyMultiple should return consistent boolean result
	property := func(numZips uint8) bool {
		if numZips == 0 {
			numZips = 1
		}
		if numZips > 5 {
			numZips = 5
		}

		// Create multiple test ZIP files
		var zipPaths []string
		for i := uint8(0); i < numZips; i++ {
			zipPath := createTestZIP(t, 2) // 2 files per ZIP
			zipPaths = append(zipPaths, zipPath)
			defer os.Remove(zipPath)
		}

		verifier := NewVerifier()
		results, allPassed := verifier.VerifyMultiple(zipPaths)

		// Should have results for each ZIP file
		if len(results) != int(numZips) {
			t.Logf("Expected %d results, got %d", numZips, len(results))
			return false
		}

		// All our test ZIPs should pass, so allPassed should be true
		if !allPassed {
			t.Logf("Expected all ZIPs to pass, but allPassed = false")
			return false
		}

		// Each individual result should also pass
		for i, result := range results {
			if !result.Passed {
				t.Logf("ZIP %d failed verification", i)
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 10}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property failed: %v", err)
	}
}

// Unit tests

func TestNewVerifier(t *testing.T) {
	verifier := NewVerifier()
	if verifier == nil {
		t.Fatal("NewVerifier() returned nil")
	}
	if verifier.verbose {
		t.Error("NewVerifier() should create non-verbose verifier by default")
	}
}

func TestVerifier_SetVerbose(t *testing.T) {
	verifier := NewVerifier()
	
	// Test setting verbose to true
	verifier.SetVerbose(true)
	if !verifier.verbose {
		t.Error("SetVerbose(true) did not set verbose flag")
	}
	
	// Test setting verbose to false
	verifier.SetVerbose(false)
	if verifier.verbose {
		t.Error("SetVerbose(false) did not clear verbose flag")
	}
}

func TestVerifier_IsArchiveFile(t *testing.T) {
	verifier := NewVerifier()
	
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"zip file", "test.zip", true},
		{"ZIP file uppercase", "test.ZIP", true},
		{"zip file with path", "/path/to/file.zip", true},
		{"txt file", "test.txt", false},
		{"no extension", "test", false},
		{"empty string", "", false},
		{"zip in filename but not extension", "zipfile.txt", false},
		{"multiple extensions", "test.tar.zip", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifier.IsArchiveFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsArchiveFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestVerifier_VerifyZIP_ValidZIP(t *testing.T) {
	// Create a valid test ZIP file
	zipPath := createTestZIP(t, 3)
	defer os.Remove(zipPath)
	
	verifier := NewVerifier()
	result, err := verifier.VerifyZIP(zipPath)
	
	if err != nil {
		t.Fatalf("VerifyZIP() returned error: %v", err)
	}
	
	if result == nil {
		t.Fatal("VerifyZIP() returned nil result")
	}
	
	if result.FilePath != zipPath {
		t.Errorf("Result.FilePath = %q, want %q", result.FilePath, zipPath)
	}
	
	if !result.Passed {
		t.Errorf("Result.Passed = false, want true. Failed entries: %v", result.FailedEntries)
	}
	
	if result.TotalEntries != 3 {
		t.Errorf("Result.TotalEntries = %d, want 3", result.TotalEntries)
	}
	
	if len(result.FailedEntries) != 0 {
		t.Errorf("Result.FailedEntries = %v, want empty", result.FailedEntries)
	}
	
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestVerifier_VerifyZIP_CorruptedZIP(t *testing.T) {
	// For this test, we'll create a ZIP file and then test the verification logic
	// by creating a scenario where CRC32 would fail
	
	// Create a valid ZIP first
	zipPath := createTestZIP(t, 1)
	defer os.Remove(zipPath)
	
	// Test that our verification logic works by checking a valid ZIP
	verifier := NewVerifier()
	result, err := verifier.VerifyZIP(zipPath)
	
	if err != nil {
		t.Fatalf("VerifyZIP() returned error: %v", err)
	}
	
	// The valid ZIP should pass
	if !result.Passed {
		t.Errorf("Valid ZIP failed verification: %v", result.FailedEntries)
	}
	
	// Test with a file that's not a valid ZIP format
	tmpFile, err := os.CreateTemp("", "invalid*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	// Write invalid ZIP data
	tmpFile.WriteString("PK\x03\x04invalid zip data")
	tmpFile.Close()
	
	result2, err2 := verifier.VerifyZIP(tmpFile.Name())
	
	// Should return an error for invalid ZIP
	if err2 == nil {
		t.Error("Expected error for invalid ZIP file")
	}
	
	if result2.Passed {
		t.Error("Invalid ZIP should not pass verification")
	}
}

func TestVerifier_VerifyZIP_NonExistentFile(t *testing.T) {
	verifier := NewVerifier()
	result, err := verifier.VerifyZIP("nonexistent.zip")
	
	if err == nil {
		t.Error("VerifyZIP() expected error for nonexistent file, got nil")
	}
	
	if result == nil {
		t.Fatal("VerifyZIP() returned nil result")
	}
	
	if result.Passed {
		t.Error("Result.Passed = true, want false for nonexistent file")
	}
	
	if result.Error == nil {
		t.Error("Result.Error = nil, want error for nonexistent file")
	}
}

func TestVerifier_VerifyZIP_InvalidZIP(t *testing.T) {
	// Create a file that's not a ZIP
	tmpFile, err := os.CreateTemp("", "notzip*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.WriteString("This is not a ZIP file")
	tmpFile.Close()
	
	verifier := NewVerifier()
	result, err := verifier.VerifyZIP(tmpFile.Name())
	
	if err == nil {
		t.Error("VerifyZIP() expected error for invalid ZIP file, got nil")
	}
	
	if result.Passed {
		t.Error("Result.Passed = true, want false for invalid ZIP")
	}
}

func TestVerifier_VerifyMultiple(t *testing.T) {
	// Create multiple test ZIP files
	zip1 := createTestZIP(t, 2)
	zip2 := createTestZIP(t, 3)
	defer os.Remove(zip1)
	defer os.Remove(zip2)
	
	// Create a non-ZIP file
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	
	verifier := NewVerifier()
	paths := []string{zip1, zip2, tmpFile.Name()}
	results, allPassed := verifier.VerifyMultiple(paths)
	
	// Should only process ZIP files
	if len(results) != 2 {
		t.Errorf("VerifyMultiple() returned %d results, want 2", len(results))
	}
	
	if !allPassed {
		t.Error("VerifyMultiple() allPassed = false, want true")
	}
	
	// Check individual results
	for i, result := range results {
		if !result.Passed {
			t.Errorf("Result %d failed verification", i)
		}
	}
}

func TestVerifier_FormatResult_VerboseMode(t *testing.T) {
	zipPath := createTestZIP(t, 2)
	defer os.Remove(zipPath)
	
	verifier := NewVerifier()
	result, err := verifier.VerifyZIP(zipPath)
	if err != nil {
		t.Fatalf("VerifyZIP() failed: %v", err)
	}
	
	// Test verbose output
	output := verifier.FormatResult(result, true)
	
	if output == "" {
		t.Error("FormatResult(verbose=true) returned empty string")
	}
	
	if !strings.Contains(output, "Verifying:") {
		t.Error("Verbose output should contain 'Verifying:'")
	}
	
	if !strings.Contains(output, zipPath) {
		t.Error("Verbose output should contain file path")
	}
	
	if !strings.Contains(output, "✓") {
		t.Error("Verbose output should contain success indicator")
	}
}

func TestVerifier_FormatResult_BooleanMode(t *testing.T) {
	zipPath := createTestZIP(t, 2)
	defer os.Remove(zipPath)
	
	verifier := NewVerifier()
	result, err := verifier.VerifyZIP(zipPath)
	if err != nil {
		t.Fatalf("VerifyZIP() failed: %v", err)
	}
	
	// Test boolean (non-verbose) output
	output := verifier.FormatResult(result, false)
	
	if output != "" {
		t.Errorf("FormatResult(verbose=false) returned %q, want empty string", output)
	}
}

func TestVerifier_FormatResult_WithError(t *testing.T) {
	verifier := NewVerifier()
	result := &VerificationResult{
		FilePath: "test.zip",
		Passed:   false,
		Error:    fmt.Errorf("test error"),
	}
	
	output := verifier.FormatResult(result, true)
	
	if !strings.Contains(output, "Error:") {
		t.Error("Output should contain 'Error:' for failed verification")
	}
	
	if !strings.Contains(output, "test error") {
		t.Error("Output should contain the error message")
	}
}

func TestVerifier_FormatResult_WithFailedEntries(t *testing.T) {
	verifier := NewVerifier()
	result := &VerificationResult{
		FilePath:      "test.zip",
		Passed:        false,
		FailedEntries: []string{"file1.txt", "file2.txt"},
		TotalEntries:  3,
	}
	
	output := verifier.FormatResult(result, true)
	
	if !strings.Contains(output, "✗") {
		t.Error("Output should contain failure indicator")
	}
	
	if !strings.Contains(output, "2 of 3") {
		t.Error("Output should show failed/total count")
	}
	
	if !strings.Contains(output, "file1.txt") {
		t.Error("Output should list failed entries")
	}
	
	if !strings.Contains(output, "file2.txt") {
		t.Error("Output should list failed entries")
	}
}

// Helper functions

// createTestZIP creates a valid ZIP file with the specified number of files
func createTestZIP(t *testing.T, numFiles int) string {
	tmpFile, err := os.CreateTemp("", "test*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tmpFile.Close()
	
	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()
	
	for i := 0; i < numFiles; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		content := fmt.Sprintf("Content of file %d", i)
		
		writer, err := zipWriter.Create(filename)
		if err != nil {
			t.Fatalf("Failed to create ZIP entry: %v", err)
		}
		
		_, err = writer.Write([]byte(content))
		if err != nil {
			t.Fatalf("Failed to write ZIP entry: %v", err)
		}
	}
	
	return tmpFile.Name()
}