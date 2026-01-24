package progress

import (
	"os"
	"testing"
	"time"
)

func TestConfigureBar_Coverage(t *testing.T) {
	opts := &Options{
		Description: "test",
		ShowBytes:   true,
		Writer:      os.Stderr,
	}
	barOpts := configureBar(opts)
	if len(barOpts) == 0 {
		t.Error("Expected bar options")
	}
}

func TestBar_Methods(t *testing.T) {
	b := NewBar(&Options{Total: 100, Threshold: 0})
	b.Increment()
	if b.current != 1 {
		t.Errorf("Expected 1, got %d", b.current)
	}
	
	b.SetCurrent(10)
	if b.current != 10 {
		t.Errorf("Expected 10, got %d", b.current)
	}

	b.Finish()
	b.Clear()

	if b.IsEnabled() {
		t.Error("Expected disabled initially")
	}
	if b.IsTTY() != b.isTTY {
		t.Error("IsTTY() mismatch")
	}
}

func TestBar_ETA_Coverage(t *testing.T) {
	b := &Bar{
		total:     100,
		current:   50,
		startTime: time.Now().Add(-1 * time.Second),
	}
	eta := b.ETA()
	if eta < 0 {
		t.Errorf("Expected positive ETA, got %v", eta)
	}

	b.current = 0
	if b.ETA() != 0 {
		t.Error("Expected 0 ETA for 0 progress")
	}
}

func TestBar_Percentage_Coverage(t *testing.T) {
	b := &Bar{total: 100, current: 50}
	if b.Percentage() != 50.0 {
		t.Errorf("Expected 50.0, got %f", b.Percentage())
	}
	
	b.total = 0
	if b.Percentage() != 0 {
		t.Error("Expected 0 percentage for 0 total")
	}
}

func TestBar_String_Coverage(t *testing.T) {
	b := &Bar{total: 100, current: 50}
	s := b.String()
	if s == "" {
		t.Error("Expected non-empty string")
	}
}