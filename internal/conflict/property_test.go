package conflict

import (
	"testing"
	"testing/quick"
)

func TestProperty_VerbosityConsistency(t *testing.T) {
	f := func(verbose, quiet bool) bool {
		flagSet := map[string]bool{
			"verbose": verbose,
			"quiet":   quiet,
		}

		state, _, _ := ResolveState([]string{}, flagSet, "default")

		if quiet {
			return state.Verbosity == VerbosityQuiet
		}
		if verbose {
			return state.Verbosity == VerbosityVerbose
		}
		return state.Verbosity == VerbosityNormal
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestProperty_ModeConsistency(t *testing.T) {
	f := func(isBool bool) bool {
		flagSet := map[string]bool{
			"bool": isBool,
		}

		state, _, _ := ResolveState([]string{}, flagSet, "default")

		if isBool {
			return state.Mode == ModeBool && state.Verbosity == VerbosityQuiet
		}
		return state.Mode == ModeStandard
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
