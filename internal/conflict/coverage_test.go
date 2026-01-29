package conflict

import (
	"testing"
)

func TestFormatAllWarnings_Empty(t *testing.T) {
	if FormatAllWarnings(nil) != "" {
		t.Error("Expected empty string for nil warnings")
	}
	if FormatAllWarnings([]Warning{}) != "" {
		t.Error("Expected empty string for empty warnings")
	}
}

func TestDetermineVerbosity_Coverage(t *testing.T) {
	s := &RunState{Mode: ModeStandard, Format: FormatDefault}

	// Test verbose with default format changes format to verbose
	flagSet := map[string]bool{"verbose": true}
	s.determineVerbosity(flagSet)
	if s.Format != FormatVerbose {
		t.Errorf("Expected FormatVerbose, got %v", s.Format)
	}

	// Test quiet in bool mode returns nil warns
	s.Mode = ModeBool
	warns := s.determineVerbosity(map[string]bool{"quiet": true})
	if warns != nil {
		t.Error("Expected nil warns for ModeBool")
	}
}
