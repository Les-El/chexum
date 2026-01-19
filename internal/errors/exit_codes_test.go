package errors

import (
	"fmt"
	"testing"
	"testing/quick"

	"hashi/internal/config"
	"hashi/internal/hash"
)

func TestDetermineExitCode(t *testing.T) {
	cfg := config.DefaultConfig()

	t.Run("success with no errors", func(t *testing.T) {
		result := &hash.Result{
			Entries: []hash.Entry{{Hash: "abc"}},
			Matches: []hash.MatchGroup{{Hash: "abc", Count: 1}},
		}
		if code := DetermineExitCode(cfg, result); code != config.ExitSuccess {
			t.Errorf("expected 0, got %d", code)
		}
	})

	t.Run("mismatch with match-required", func(t *testing.T) {
		cfgMatch := config.DefaultConfig()
		cfgMatch.MatchRequired = true
		result := &hash.Result{
			Entries:   []hash.Entry{{Hash: "abc"}},
			Unmatched: []hash.Entry{{Hash: "abc"}},
		}
		if code := DetermineExitCode(cfgMatch, result); code != config.ExitNoMatches {
			t.Errorf("expected %d, got %d", config.ExitNoMatches, code)
		}
	})

	t.Run("partial failure with some errors", func(t *testing.T) {
		result := &hash.Result{
			Entries: []hash.Entry{{Hash: "abc"}, {Error: fmt.Errorf("fail")}},
			Errors:  []error{fmt.Errorf("fail")},
		}
		if code := DetermineExitCode(cfg, result); code != config.ExitPartialFailure {
			t.Errorf("expected %d, got %d", config.ExitPartialFailure, code)
		}
	})

	t.Run("specific error: file not found", func(t *testing.T) {
		err := NewFileNotFoundError("missing.txt")
		result := &hash.Result{
			Entries: []hash.Entry{{Error: err}},
			Errors:  []error{err},
		}
		if code := DetermineExitCode(cfg, result); code != config.ExitFileNotFound {
			t.Errorf("expected %d, got %d", config.ExitFileNotFound, code)
		}
	})

	t.Run("specific error: permission denied", func(t *testing.T) {
		err := NewPermissionError("locked.txt")
		result := &hash.Result{
			Entries: []hash.Entry{{Error: err}},
			Errors:  []error{err},
		}
		if code := DetermineExitCode(cfg, result); code != config.ExitPermissionErr {
			t.Errorf("expected %d, got %d", config.ExitPermissionErr, code)
		}
	})
}

// TestProperty_ExitCodes verifies universal exit code properties.
func TestProperty_ExitCodes(t *testing.T) {
	// Property 16: Exit codes reflect processing status
	f := func(matchRequired bool, hasMatches bool, hasErrors bool) bool {
		cfg := config.DefaultConfig()
		cfg.MatchRequired = matchRequired

		result := &hash.Result{}
		if hasMatches {
			result.Matches = []hash.MatchGroup{{Hash: "abc", Count: 1}}
		}
		if hasErrors {
			result.Errors = []error{fmt.Errorf("error")}
			result.Entries = []hash.Entry{{Error: fmt.Errorf("error")}}
		} else {
			result.Entries = []hash.Entry{{Hash: "abc"}}
		}

		code := DetermineExitCode(cfg, result)

		if hasErrors {
			return code != config.ExitSuccess
		}
		if matchRequired && !hasMatches {
			return code == config.ExitNoMatches
		}
		return code == config.ExitSuccess
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
