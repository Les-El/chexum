// Package security provides input validation and path safety checks for hashi.
//
// DESIGN PRINCIPLE: Hashi can't change Hashi
// -----------------------------------------
// One of the most dangerous vulnerabilities in automated tools is
// "Self-Modification" or "Configuration Injection". If an attacker can force
// a tool to overwrite its own security policy, the tool becomes a weapon.
//
// Hashi defends against this with two core mandates:
//  1. READ-ONLY ON SOURCE: Hashi never, under any circumstances, modifies the
//     files it is hashing.
//  2. PROTECTED CONFIGURATION: Hashi cannot write output or logs to its own
//     configuration files or directories.
//
// This package implements these mandates through strict path validation,
// extension whitelisting, and obfuscated error messages that prevent
// attackers from discovering which files are protected.
package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Security options for validation.
type Options struct {
	Verbose        bool
	BlacklistFiles []string
	BlacklistDirs  []string
	WhitelistFiles []string
	WhitelistDirs  []string
}

// Default blacklists
var DefaultBlacklistFiles = []string{
	".env",
	".hashi.toml",
	"*.key",
	"*.pem",
	"id_rsa",
	"id_ed25519",
}

var DefaultBlacklistDirs = []string{
	".git",
	".ssh",
	".aws",
	".config/hashi",
}

// ValidateOutputPath ensures an output path is safe to write to.
func ValidateOutputPath(path string, opts Options) error {
	if path == "" {
		return nil
	}

	// 1. Extension validation
	ext := strings.ToLower(filepath.Ext(path))
	allowedExts := []string{".txt", ".json", ".csv"}
	extAllowed := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			extAllowed = true
			break
		}
	}
	if !extAllowed {
		return fmt.Errorf("output files must have extension: %s (got %s)",
			strings.Join(allowedExts, ", "), ext)
	}

	// 2. Directory traversal check
	// We use Clean to resolve ".." and checks if it tries to escape boundaries
	// strictly speaking, we just want to ensure we aren't being tricked.
	// Simple string check is too aggressive (blocks "file..txt").
	cleaned := filepath.Clean(path)
	// If the path is relative and starts with "..", it's traversing up.
	// But Clean("foo/../../bar") -> "../bar".
	if strings.Contains(filepath.ToSlash(cleaned), "../") || filepath.Base(cleaned) == ".." {
		return fmt.Errorf("directory traversal not allowed in output path")
	}
	// We also keep the original check for safety but strictly for path separators involved
	// actually, let's remove the naive check and trust Clean + analysis.

	// 3. File name validation
	basename := filepath.Base(path)
	if err := ValidateFileName(basename, opts); err != nil {
		return err
	}

	// 4. Directory validation
	if err := ValidateDirPath(path, opts); err != nil {
		return err
	}

	// 5. Symlink check
	// We must ensure that if the file exists, it is not a symlink.
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return formatSecurityError(opts.Verbose, "cannot write to symlink")
		}
	} else if !os.IsNotExist(err) {
		// If we can't check it, fail safe
		return fmt.Errorf("failed to check file status: %w", err)
	}

	return nil
}

// ValidateFileName checks if a filename matches any security patterns.
func ValidateFileName(filename string, opts Options) error {
	if filename == "" {
		return nil
	}

	allBlacklist := append(DefaultBlacklistFiles, opts.BlacklistFiles...)
	filenameLower := strings.ToLower(filename)

	for _, pattern := range allBlacklist {
		patternLower := strings.ToLower(pattern)
		matched, _ := filepath.Match(patternLower, filenameLower)
		if !matched {
			// Also check for prefix match for non-glob patterns
			if !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
				matched = strings.HasPrefix(filenameLower, patternLower)
			}
		}

		if matched {
			// Check whitelist
			for _, white := range opts.WhitelistFiles {
				whiteLower := strings.ToLower(white)
				if wMatched, _ := filepath.Match(whiteLower, filenameLower); wMatched {
					return nil
				}
				if !strings.Contains(white, "*") && !strings.Contains(white, "?") {
					if strings.HasPrefix(filenameLower, whiteLower) {
						return nil
					}
				}
			}
			return formatSecurityError(opts.Verbose, fmt.Sprintf("cannot write to file matching security pattern: %s", pattern))
		}
	}
	return nil
}

// ValidateDirPath checks if any part of the path matches blacklisted directory names.
func ValidateDirPath(path string, opts Options) error {
	if path == "" {
		return nil
	}

	// 1. General directory traversal check for ALL paths
	// Replaced naive strings.Contains checks with strictly cleaned path inspection
	cleaned := filepath.Clean(path)
	// If the path is relative and starts with "..", it's traversing up.
	if strings.Contains(filepath.ToSlash(cleaned), "../") || filepath.Base(cleaned) == ".." {
		return formatSecurityError(opts.Verbose, "directory traversal not allowed in paths")
	}

	// 2. Explicit protection for hashi configuration
	pathLower := strings.ToLower(path)
	if strings.Contains(pathLower, ".hashi") || strings.Contains(pathLower, ".config/hashi") {
		return formatSecurityError(opts.Verbose, "cannot write to configuration directory")
	}

	allBlacklist := append(DefaultBlacklistDirs, opts.BlacklistDirs...)
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}

	parts := strings.Split(filepath.Clean(dir), string(filepath.Separator))
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}
		partLower := strings.ToLower(part)
		for _, pattern := range allBlacklist {
			patternLower := strings.ToLower(pattern)
			matched, _ := filepath.Match(patternLower, partLower)
			if !matched && !strings.Contains(pattern, "*") && !strings.Contains(pattern, "?") {
				matched = strings.HasPrefix(partLower, patternLower)
			}

			if matched {
				// Check whitelist
				for _, white := range opts.WhitelistDirs {
					whiteLower := strings.ToLower(white)
					if wMatched, _ := filepath.Match(whiteLower, partLower); wMatched {
						goto nextPart
					}
				}
				return formatSecurityError(opts.Verbose, fmt.Sprintf("cannot access directory matching security pattern: %s", pattern))
			}
		}
	nextPart:
	}
	return nil
}

// ValidateInputs performs security validation on all provided file paths and hash strings.
func ValidateInputs(files []string, hashes []string, opts Options) error {
	for _, file := range files {
		if file == "-" {
			continue
		}
		if err := ValidateDirPath(file, opts); err != nil {
			return err
		}
	}
	// Hashes are already classified and normalized, but we can do a sanity check
	for _, h := range hashes {
		if !isValidHex(h) {
			return fmt.Errorf("invalid hash string format: %s", h)
		}
	}
	return nil
}

func isValidHex(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ResolveSafePath returns the absolute path while ensuring no traversal attempts.
func ResolveSafePath(path string) (string, error) {
	// Use Clean to checking for traversal
	cleaned := filepath.Clean(path)
	if strings.Contains(filepath.ToSlash(cleaned), "../") || filepath.Base(cleaned) == ".." {
		return "", fmt.Errorf("directory traversal not allowed in paths")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	return abs, nil
}

func formatSecurityError(verbose bool, details string) error {
	if verbose {
		return fmt.Errorf("security policy violation: %s", details)
	}
	return fmt.Errorf("Unknown write/append error") // Obfuscated for security
}
