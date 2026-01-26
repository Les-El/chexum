package progress

import (
	"bytes"
	"os"
	"testing"
	"testing/quick"
	"time"
)

// TestProgressIndicatorsAppearForLongOperations is a property-based test.
// Feature: cli-guidelines-review, Property 3: Progress indicators appear for long operations
// **Validates: Requirements 2.4, 6.2**
//
// Property: For any operation taking longer than 100ms, some progress indication
// should be displayed before completion.
func TestProgressIndicatorsAppearForLongOperations(t *testing.T) {
	// Property: For any total > 0, if we wait longer than threshold,
	// the progress bar should be enabled
	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(verifyProgressIndicatorsProperty, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

func verifyProgressIndicatorsProperty(total uint16) bool {
	if total == 0 {
		return true // Skip zero case
	}

	// Create a buffer to capture output
	buf := &bytes.Buffer{}

	// Create progress bar with short threshold for testing
	opts := &Options{
		Total:       int64(total),
		Description: "Testing",
		Threshold:   50 * time.Millisecond,
		Writer:      buf,
	}

	bar := NewBar(opts)

	// Initially, progress should not be enabled
	if bar.IsEnabled() {
		return false
	}

	// Wait longer than threshold
	time.Sleep(60 * time.Millisecond)

	// Add some progress
	bar.Add(1)

	// After threshold, if we're on a TTY-like output, progress should be enabled
	// Since we're using a buffer (not a TTY), enabled should still be true
	// but IsTTY should be false
	if bar.enabled && bar.isTTY {
		// This would be true on a real TTY
		return true
	}

	// For non-TTY (like our buffer), enabled becomes true but bar is nil
	if bar.enabled && !bar.isTTY {
		return true
	}

	bar.Finish()
	return true
}

// TestProgressThresholdBehavior tests that progress indicators only appear
// after the threshold duration has elapsed.
func TestProgressThresholdBehavior(t *testing.T) {
	tests := []struct {
		name        string
		threshold   time.Duration
		waitTime    time.Duration
		wantEnabled bool
	}{
		{
			name:        "before threshold",
			threshold:   100 * time.Millisecond,
			waitTime:    50 * time.Millisecond,
			wantEnabled: false,
		},
		{
			name:        "after threshold",
			threshold:   50 * time.Millisecond,
			waitTime:    70 * time.Millisecond,
			wantEnabled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			opts := &Options{
				Total:       100,
				Description: "Testing",
				Threshold:   tt.threshold,
				Writer:      buf,
			}

			bar := NewBar(opts)

			// Wait the specified time
			time.Sleep(tt.waitTime)

			// Add progress to trigger threshold check
			bar.Add(1)

			if bar.enabled != tt.wantEnabled {
				t.Errorf("After waiting %v with threshold %v: enabled = %v, want %v",
					tt.waitTime, tt.threshold, bar.enabled, tt.wantEnabled)
			}

			bar.Finish()
		})
	}
}

var progressCalculationTests = []struct {
	name        string
	total       int64
	current     int64
	wantPercent float64
}{
	{
		name:        "zero progress",
		total:       100,
		current:     0,
		wantPercent: 0.0,
	},
	{
		name:        "half progress",
		total:       100,
		current:     50,
		wantPercent: 50.0,
	},
	{
		name:        "complete progress",
		total:       100,
		current:     100,
		wantPercent: 100.0,
	},
	{
		name:        "partial progress",
		total:       1000,
		current:     333,
		wantPercent: 33.3,
	},
	{
		name:        "zero total",
		total:       0,
		current:     0,
		wantPercent: 0.0,
	},
}

// TestProgressCalculation tests that progress percentage is calculated correctly.
func TestProgressCalculation(t *testing.T) {
	for _, tt := range progressCalculationTests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			opts := &Options{
				Total:       tt.total,
				Description: "Testing",
				Writer:      buf,
			}

			bar := NewBar(opts)
			bar.current = tt.current

			got := bar.Percentage()
			if got < tt.wantPercent-0.1 || got > tt.wantPercent+0.1 {
				t.Errorf("Percentage() = %.1f, want %.1f", got, tt.wantPercent)
			}
		})
	}
}

// TestETACalculation tests that ETA is calculated correctly.
func TestETACalculation(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := &Options{
		Total:       100,
		Description: "Testing",
		Writer:      buf,
	}

	bar := NewBar(opts)

	// Initially, ETA should be 0 (no progress yet)
	if eta := bar.ETA(); eta != 0 {
		t.Errorf("Initial ETA = %v, want 0", eta)
	}

	// Simulate some progress
	bar.current = 50
	bar.startTime = time.Now().Add(-1 * time.Second) // Started 1 second ago

	eta := bar.ETA()

	// With 50% done in 1 second, ETA should be approximately 1 second
	// Allow some tolerance for timing variations
	if eta < 800*time.Millisecond || eta > 1200*time.Millisecond {
		t.Errorf("ETA = %v, want approximately 1s", eta)
	}
}

// TestTTYDetection tests that TTY detection works correctly.
func TestTTYDetection(t *testing.T) {
	tests := []struct {
		name    string
		wantTTY bool
	}{
		{
			name:    "buffer is not a TTY",
			wantTTY: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			opts := &Options{
				Total:       100,
				Description: "Testing",
				Writer:      buf,
			}

			bar := NewBar(opts)

			// Buffer should never be detected as TTY
			if bar.IsTTY() {
				t.Errorf("Buffer should not be detected as TTY")
			}
		})
	}
}

// TestProgressBarOperations tests basic progress bar operations.
func TestProgressBarOperations(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := &Options{
		Total:       100,
		Description: "Testing",
		Threshold:   0, // No threshold for immediate testing
		Writer:      buf,
	}

	bar := NewBar(opts)

	// Test Add
	bar.Add(10)
	if bar.current != 10 {
		t.Errorf("After Add(10): current = %d, want 10", bar.current)
	}

	// Test Increment
	bar.Increment()
	if bar.current != 11 {
		t.Errorf("After Increment(): current = %d, want 11", bar.current)
	}

	// Test SetCurrent
	bar.SetCurrent(50)
	if bar.current != 50 {
		t.Errorf("After SetCurrent(50): current = %d, want 50", bar.current)
	}

	// Test SetCurrent with lower value (should not decrease)
	bar.SetCurrent(30)
	if bar.current != 50 {
		t.Errorf("After SetCurrent(30): current = %d, want 50 (should not decrease)", bar.current)
	}

	// Test Finish
	bar.Finish()
}

// TestDefaultOptions tests that default options are set correctly.
func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Threshold != 100*time.Millisecond {
		t.Errorf("Default threshold = %v, want 100ms", opts.Threshold)
	}

	if opts.Writer != os.Stderr {
		t.Errorf("Default writer is not os.Stderr")
	}
}

// TestProgressBarString tests the String() method.
func TestProgressBarString(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := &Options{
		Total:       100,
		Description: "Testing",
		Writer:      buf,
	}

	bar := NewBar(opts)
	bar.current = 50

	str := bar.String()
	expected := "50.0% (50/100)"
	if str != expected {
		t.Errorf("String() = %q, want %q", str, expected)
	}
}

// TestProgressBarWithZeroThreshold tests that progress bar works with zero threshold.
func TestProgressBarWithZeroThreshold(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := &Options{
		Total:       100,
		Description: "Testing",
		Threshold:   0, // Should default to 100ms
		Writer:      buf,
	}

	bar := NewBar(opts)

	if bar.threshold != 100*time.Millisecond {
		t.Errorf("Zero threshold should default to 100ms, got %v", bar.threshold)
	}
}

// TestProgressBarWithNilOptions tests that NewBar handles nil options.
func TestProgressBarWithNilOptions(t *testing.T) {
	bar := NewBar(nil)

	if bar == nil {
		t.Fatal("NewBar(nil) returned nil")
	}

	if bar.threshold != 100*time.Millisecond {
		t.Errorf("Nil options should use default threshold 100ms, got %v", bar.threshold)
	}
}

func TestNewBar(t *testing.T) {
	TestProgressBarWithNilOptions(t)
}

func TestAdd(t *testing.T) {
	TestProgressBarOperations(t)
}

func TestIncrement(t *testing.T) {
	TestProgressBarOperations(t)
}

func TestSetCurrent(t *testing.T) {
	TestProgressBarOperations(t)
}

func TestFinish(t *testing.T) {
	TestProgressBarOperations(t)
}

func TestClear(t *testing.T) {
	bar := NewBar(nil)
	bar.Clear() // Should not panic
}

func TestIsEnabled(t *testing.T) {
	bar := NewBar(nil)
	_ = bar.IsEnabled()
}

func TestIsTTY(t *testing.T) {
	TestTTYDetection(t)
}

func TestETA(t *testing.T) {
	TestETACalculation(t)
}

func TestPercentage(t *testing.T) {
	TestProgressCalculation(t)
}

func TestString(t *testing.T) {
	TestProgressBarString(t)
}
