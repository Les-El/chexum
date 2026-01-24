// Package conflict implements the "Pipeline of Intent" state machine for
// resolving flag configurations.
//
// It replaces the traditional pairwise conflict checks with a phased approach:
// 1. Intent Collection (Scanning args)
// 2. State Construction (Applying rules like "Last One Wins")
// 3. Validation (Checking for invalid states)
package conflict

import (
	"fmt"
	"strings"
)

// Mode defines the operational mode of the application.
type Mode string

const (
	ModeStandard Mode = "standard"
	ModeBool     Mode = "bool"   // --bool
)

// Format defines the data output format (stdout).
type Format string

const (
	FormatDefault Format = "default"
	FormatJSON    Format = "json"    // --json or --format=json
	FormatPlain   Format = "plain"   // --plain or --format=plain
	FormatVerbose Format = "verbose" // --format=verbose
)

// Verbosity defines the logging level (stderr).
type Verbosity string

const (
	VerbosityNormal  Verbosity = "normal"
	VerbosityQuiet   Verbosity = "quiet"   // --quiet
	VerbosityVerbose Verbosity = "verbose" // --verbose
)

// RunState represents the finalized, resolved behavior of the application.
type RunState struct {
	Mode      Mode
	Format    Format
	Verbosity Verbosity
}

// Intent represents a user's specific request for a behavior.
type intent struct {
	Type     string // "mode", "format", "verbosity"
	Value    string // e.g. "json", "quiet"
	Position int    // Index in os.Args
	Flag     string // The actual flag used (e.g., "--json")
}

// Warning represents a non-fatal conflict resolution.
type Warning struct {
	Message string
}

// ResolveState processes raw arguments and detected flags to produce a consistent RunState.
func ResolveState(args []string, flagSet map[string]bool, explicitFormat string) (*RunState, []Warning, error) {
	warnings := make([]Warning, 0)
	
	// Phase 1: Intent Collection
	lastFormatIntent, lastFormatPos := collectFormatIntent(args, explicitFormat)

	// Phase 2: State Construction
	state := &RunState{
		Mode:      ModeStandard,
		Format:    FormatDefault,
		Verbosity: VerbosityNormal,
	}

	// 2a. Determine Mode
	if flagSet["bool"] {
		state.Mode = ModeBool
		state.Verbosity = VerbosityQuiet
	}

	// 2b. Determine Format
	formatWarn := state.determineFormat(lastFormatIntent, lastFormatPos)
	if formatWarn != "" {
		warnings = append(warnings, Warning{Message: formatWarn})
	}

	// 2c. Determine Verbosity
	verbosityWarns := state.determineVerbosity(flagSet)
	warnings = append(warnings, verbosityWarns...)

	return state, warnings, nil
}

func collectFormatIntent(args []string, explicitFormat string) (string, int) {
	intent := ""
	pos := -1
	if explicitFormat != "" && explicitFormat != "default" {
		intent = explicitFormat
	}

	for i, arg := range args {
		if arg == "--json" || arg == "--plain" {
			intent = strings.TrimPrefix(arg, "--")
			pos = i
		} else if strings.HasPrefix(arg, "--format=") {
			intent = strings.TrimPrefix(arg, "--format=")
			pos = i
		} else if strings.HasPrefix(arg, "-f=") {
			intent = strings.TrimPrefix(arg, "-f=")
			pos = i
		} else if arg == "-f" && i+1 < len(args) {
			intent = args[i+1]
			pos = i
		}
	}
	return intent, pos
}

func (s *RunState) determineFormat(intent string, pos int) string {
	if pos < 0 && intent == "" {
		return ""
	}

	if s.Mode == ModeBool && intent != "" && intent != "default" {
		return fmt.Sprintf("--bool overrides --%s", intent)
	}

	if s.Mode != ModeBool {
		switch intent {
		case "json":
			s.Format = FormatJSON
		case "plain":
			s.Format = FormatPlain
		case "verbose":
			s.Format = FormatVerbose
		case "default":
			s.Format = FormatDefault
		default:
			s.Format = Format(intent)
		}
	}
	return ""
}

func (s *RunState) determineVerbosity(flagSet map[string]bool) []Warning {
	var warns []Warning
	if s.Mode == ModeBool {
		return nil
	}

	if flagSet["quiet"] {
		s.Verbosity = VerbosityQuiet
		if flagSet["verbose"] {
			warns = append(warns, Warning{Message: "--quiet overrides --verbose"})
		}
	} else if flagSet["verbose"] {
		s.Verbosity = VerbosityVerbose
		if s.Format == FormatDefault {
			s.Format = FormatVerbose
		}
	}
	return warns
}

// FormatAllWarnings formats all warnings for display.
func FormatAllWarnings(warnings []Warning) string {
	if len(warnings) == 0 {
		return ""
	}
	
	var sb strings.Builder
	for _, warn := range warnings {
		sb.WriteString(fmt.Sprintf("Warning: %s\n", warn.Message))
	}
	return sb.String()
}