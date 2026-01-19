package config

import (
	"strings"
	"testing"
)

func TestSuggestFlag(t *testing.T) {
	tests := []struct {
		unknown string
		want    string
	}{
		{"verboze", "--verbose"},
		{"quied", "--quiet"},
		{"recursiv", "--recursive"},
		{"algo", "--algorithm"},
		{"h", "-h"}, // Should find short flag if close (wait, KnownFlags doesn't have short flags yet)
		{"ver", "--version"},
		{"totally_wrong", ""},
	}

	for _, tt := range tests {
		got := SuggestFlag(tt.unknown)
		if got != tt.want {
			t.Errorf("SuggestFlag(%q) = %q, want %q", tt.unknown, got, tt.want)
		}
	}
}

func TestParseArgs_Suggestions(t *testing.T) {
	args := []string{"--verboze"}
	_, _, err := ParseArgs(args)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	
	if !strings.Contains(err.Error(), "Did you mean --verbose?") {
		t.Errorf("expected suggestion in error, got: %v", err)
	}
}
