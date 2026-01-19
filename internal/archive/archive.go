// Package archive provides ZIP file verification using CRC32 checksums.
//
// It verifies the internal integrity of ZIP files by checking CRC32
// checksums of all entries. This confirms data integrity (bits not
// corrupted) but NOT authenticity (not tampered with).
//
// Security Note: CRC32 is always used regardless of any metadata in
// the ZIP file suggesting alternative algorithms. This prevents
// algorithm substitution attacks.
package archive

import (
	"archive/zip"
	"fmt"
	"hash/crc32"
	"io"
	"path/filepath"
	"strings"
)

// VerificationResult holds the result of verifying an archive.
type VerificationResult struct {
	FilePath      string   // Path to the archive file
	Passed        bool     // True if all entries passed verification
	FailedEntries []string // Names of entries that failed CRC32
	TotalEntries  int      // Total number of entries checked
	Error         error    // Error if verification could not complete
}

// Verifier handles archive integrity verification.
type Verifier struct {
	verbose bool
}

// NewVerifier creates a new archive verifier.
func NewVerifier() *Verifier {
	return &Verifier{}
}

// SetVerbose enables or disables verbose output.
func (v *Verifier) SetVerbose(verbose bool) {
	v.verbose = verbose
}

// IsArchiveFile checks if a file is a supported archive type.
func (v *Verifier) IsArchiveFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".zip"
}

// VerifyZIP verifies the CRC32 checksums of all entries in a ZIP file.
func (v *Verifier) VerifyZIP(path string) (*VerificationResult, error) {
	result := &VerificationResult{
		FilePath:      path,
		Passed:        true,
		FailedEntries: make([]string, 0),
	}

	// Open the ZIP file
	reader, err := zip.OpenReader(path)
	if err != nil {
		result.Error = fmt.Errorf("failed to open ZIP file: %w", err)
		result.Passed = false
		return result, result.Error
	}
	defer reader.Close()

	result.TotalEntries = len(reader.File)

	// Verify each entry
	for _, file := range reader.File {
		// Skip directories
		if file.FileInfo().IsDir() {
			continue
		}

		// Verify the entry's CRC32
		if err := v.verifyEntry(file); err != nil {
			result.Passed = false
			result.FailedEntries = append(result.FailedEntries, file.Name)
		}
	}

	return result, nil
}

// verifyEntry verifies the CRC32 of a single ZIP entry.
func (v *Verifier) verifyEntry(file *zip.File) error {
	// Open the entry for reading
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open entry: %w", err)
	}
	defer rc.Close()

	// Compute CRC32 of the entry's contents
	hasher := crc32.NewIEEE()
	if _, err := io.Copy(hasher, rc); err != nil {
		return fmt.Errorf("failed to read entry: %w", err)
	}

	// Compare with stored CRC32
	// Note: We always use CRC32 regardless of any metadata suggesting alternatives
	// This is a security hardening measure to prevent algorithm substitution attacks
	computed := hasher.Sum32()
	expected := file.CRC32

	if computed != expected {
		return fmt.Errorf("CRC32 mismatch: expected %08x, got %08x", expected, computed)
	}

	return nil
}

// VerifyMultiple verifies multiple archive files and returns a combined result.
func (v *Verifier) VerifyMultiple(paths []string) ([]*VerificationResult, bool) {
	results := make([]*VerificationResult, 0, len(paths))
	allPassed := true

	for _, path := range paths {
		if !v.IsArchiveFile(path) {
			continue
		}

		result, _ := v.VerifyZIP(path)
		results = append(results, result)

		if !result.Passed {
			allPassed = false
		}
	}

	return results, allPassed
}

// FormatResult formats a verification result for display.
func (v *Verifier) FormatResult(result *VerificationResult, verbose bool) string {
	if !verbose {
		// Boolean mode: no output, just exit code
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Verifying: %s\n", result.FilePath))

	if result.Error != nil {
		sb.WriteString(fmt.Sprintf("  Error: %s\n", result.Error))
		return sb.String()
	}

	if result.Passed {
		sb.WriteString(fmt.Sprintf("  ✓ All %d entries passed CRC32 verification\n", result.TotalEntries))
	} else {
		sb.WriteString(fmt.Sprintf("  ✗ %d of %d entries failed CRC32 verification:\n",
			len(result.FailedEntries), result.TotalEntries))
		for _, entry := range result.FailedEntries {
			sb.WriteString(fmt.Sprintf("    - %s\n", entry))
		}
	}

	return sb.String()
}
