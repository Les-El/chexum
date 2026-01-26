package testutil

import (
	"testing"
)

func TestRunPropertyTest(t *testing.T) {
	// Smoke test for RunPropertyTest
	RunPropertyTest(t, "test", 1, "test property", func(s string) bool {
		return len(s) >= 0
	}, nil)
}

func TestCheckProperty(t *testing.T) {
	// Smoke test for CheckProperty
	CheckProperty(t, func(s string) bool {
		return len(s) >= 0
	})
}
