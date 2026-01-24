package security

import (
	"strings"
	"testing"
	"testing/quick"
)

func TestValidateOutputPath(t *testing.T) {
	opts := Options{Verbose: true}

	t.Run("valid path", func(t *testing.T) {
		if err := ValidateOutputPath("results.txt", opts); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid extension", func(t *testing.T) {
		if err := ValidateOutputPath("results.exe", opts); err == nil {
			t.Error("expected error for .exe")
		}
	})

	t.Run("directory traversal", func(t *testing.T) {
		if err := ValidateOutputPath("../../etc/passwd.txt", opts); err == nil {
			t.Error("expected error for traversal")
		}
	})

	t.Run("blacklisted file", func(t *testing.T) {
		if err := ValidateOutputPath("secret_data.txt", opts); err == nil {
			t.Error("expected error for 'secret' file")
		}
	})

	t.Run("blacklisted dir", func(t *testing.T) {
		if err := ValidateOutputPath("config/results.txt", opts); err == nil {
			t.Error("expected error for 'config' dir")
		}
	})

	t.Run("whitelist override", func(t *testing.T) {
		wOpts := opts
		wOpts.WhitelistFiles = []string{"secret_report.txt"}
		if err := ValidateOutputPath("secret_report.txt", wOpts); err != nil {
			t.Errorf("expected whitelist to allow file, got %v", err)
		}
	})
}

func TestProperty_SecurityValidation(t *testing.T) {
	// Property 8: Input validation occurs before processing
	// We verify that blacklisted patterns are always rejected unless whitelisted
	f := func(name string) bool {
		if name == "" {
			return true
		}
		opts := Options{Verbose: true}
		
		// If name contains a blacklist word, it should be rejected
		isBlacklisted := false
		for _, b := range DefaultBlacklistFiles {
			if strings.Contains(strings.ToLower(name), strings.ToLower(b)) {
				isBlacklisted = true
				break
			}
		}
		
		err := ValidateFileName(name, opts)
		if isBlacklisted && !strings.Contains(name, "*") && !strings.Contains(name, "?") {
			return err != nil
		}
		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestResolveSafePath(t *testing.T) {
	t.Run("safe path", func(t *testing.T) {
		_, err := ResolveSafePath("file.txt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("traversal", func(t *testing.T) {
		_, err := ResolveSafePath("../outside")
		if err == nil {
			t.Error("expected error for ..")
		}
	})
}

func TestValidateFileName(t *testing.T) {
	opts := Options{Verbose: true}
	if err := ValidateFileName("secret.txt", opts); err == nil {
		t.Error("Expected error for secret.txt")
	}
	if err := ValidateFileName("safe.txt", opts); err != nil {
		t.Errorf("Unexpected error for safe.txt: %v", err)
	}
}

func TestValidateDirPath(t *testing.T) {
	opts := Options{Verbose: true}
	if err := ValidateDirPath("config/file.txt", opts); err == nil {
		t.Error("Expected error for config/file.txt")
	}
	if err := ValidateDirPath("safe/file.txt", opts); err != nil {
		t.Errorf("Unexpected error for safe/file.txt: %v", err)
	}
}

func TestValidateInputs(t *testing.T) {
	opts := Options{Verbose: true}
	files := []string{"safe.txt"}
	hashes := []string{"abc123"}
	if err := ValidateInputs(files, hashes, opts); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	badHashes := []string{"not-hex"}
	if err := ValidateInputs(files, badHashes, opts); err == nil {
		t.Error("Expected error for invalid hex hash")
	}
}
