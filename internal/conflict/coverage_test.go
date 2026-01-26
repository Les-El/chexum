package conflict

import (
	"testing"
)

func TestCollectFormatIntent_Detailed(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		explicit string
		want     string
		wantPos  int
	}{
		{"ShortFlagWithEquals", []string{"-f=json"}, "", "json", 0},
		{"ShortFlagSeparate", []string{"-f", "json"}, "", "json", 0},
		{"LongFlagWithEquals", []string{"--format=plain"}, "", "plain", 0},
		{"ExplicitFormatOnly", []string{}, "json", "json", -1},
		{"EmptyArgsAndExplicit", []string{}, "", "", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotPos := collectFormatIntent(tt.args, tt.explicit)
			if got != tt.want || gotPos != tt.wantPos {
				t.Errorf("collectFormatIntent(%v, %q) = (%q, %d); want (%q, %d)", tt.args, tt.explicit, got, gotPos, tt.want, tt.wantPos)
			}
		})
	}
}

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
