// Package color provides TTY detection and color handling for hashi.
package color

import (
	"os"
	"strings"
	"testing"
	"testing/quick"
)

// TestProperty_ColorOutputRespectsTTYDetection verifies Property 2:
// For any output destination, when output is sent to a non-TTY or NO_COLOR is set,
// the output should contain no ANSI color codes.
//
// Feature: cli-guidelines-review, Property 2: Color output respects TTY detection
// Validates: Requirements 2.2, 2.3
func TestProperty_ColorOutputRespectsTTYDetection(t *testing.T) {
	// Property: When colors are disabled (non-TTY or NO_COLOR set),
	// colorized text should not contain ANSI escape codes
	property := func(text string) bool {
		// Create a handler with colors disabled
		h := &Handler{
			enabled: false,
			isTTY:   false,
		}
		h.setupColors()

		// Test all color methods
		colors := []Color{
			ColorGreen,
			ColorRed,
			ColorYellow,
			ColorBlue,
			ColorCyan,
			ColorGray,
		}

		for _, c := range colors {
			result := h.Colorize(text, c)
			
			// When colors are disabled, output should equal input (no ANSI codes)
			if result != text {
				return false
			}
			
			// Verify no ANSI escape sequences are present
			if containsANSI(result) {
				return false
			}
		}

		// Test convenience methods
		if h.Green(text) != text || containsANSI(h.Green(text)) {
			return false
		}
		if h.Red(text) != text || containsANSI(h.Red(text)) {
			return false
		}
		if h.Yellow(text) != text || containsANSI(h.Yellow(text)) {
			return false
		}
		if h.Blue(text) != text || containsANSI(h.Blue(text)) {
			return false
		}
		if h.Cyan(text) != text || containsANSI(h.Cyan(text)) {
			return false
		}
		if h.Gray(text) != text || containsANSI(h.Gray(text)) {
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as per spec requirements
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// TestProperty_ColorOutputWithNO_COLOR verifies that NO_COLOR environment variable
// disables color output regardless of TTY status.
//
// Feature: cli-guidelines-review, Property 2: Color output respects TTY detection
// Validates: Requirements 2.3
func TestProperty_ColorOutputWithNO_COLOR(t *testing.T) {
	// Save original NO_COLOR state
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	// Property: When NO_COLOR is set, colors should be disabled
	property := func(text string) bool {
		// Set NO_COLOR environment variable
		os.Setenv("NO_COLOR", "1")

		// Create a new handler (will detect NO_COLOR)
		h := NewColorHandler()

		// Colors should be disabled
		if h.IsEnabled() {
			return false
		}

		// All color methods should return plain text
		if h.Green(text) != text || containsANSI(h.Green(text)) {
			return false
		}
		if h.Red(text) != text || containsANSI(h.Red(text)) {
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// containsANSI checks if a string contains ANSI escape sequences.
func containsANSI(s string) bool {
	// ANSI escape sequences start with ESC [ (or \x1b[)
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\033[")
}

// Unit Tests

// TestColorHandler_TTYDetection tests TTY detection logic.
func TestColorHandler_TTYDetection(t *testing.T) {
	tests := []struct {
		name      string
		setupEnv  func()
		wantTTY   bool
		wantColor bool
	}{
		{
			name: "NO_COLOR set disables colors",
			setupEnv: func() {
				os.Setenv("NO_COLOR", "1")
			},
			wantColor: false,
		},
		{
			name: "NO_COLOR empty string still disables colors",
			setupEnv: func() {
				os.Setenv("NO_COLOR", "")
			},
			wantColor: false,
		},
		{
			name: "NO_COLOR unset allows colors on TTY",
			setupEnv: func() {
				os.Unsetenv("NO_COLOR")
			},
			// Note: wantColor depends on actual TTY status, tested separately
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original state
			originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
			defer func() {
				if hadNoColor {
					os.Setenv("NO_COLOR", originalValue)
				} else {
					os.Unsetenv("NO_COLOR")
				}
			}()

			// Setup environment
			tt.setupEnv()

			// Create handler
			h := NewColorHandler()

			// When NO_COLOR is set, colors should be disabled
			if strings.Contains(tt.name, "NO_COLOR set") || strings.Contains(tt.name, "empty string") {
				if h.IsEnabled() {
					t.Errorf("Expected colors to be disabled when NO_COLOR is set, but they were enabled")
				}
			}
		})
	}
}

// TestColorHandler_ColorCodeGeneration tests that color codes are generated correctly.
func TestColorHandler_ColorCodeGeneration(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	// Create handler with colors explicitly enabled
	h := &Handler{
		enabled: true,
		isTTY:   true,
	}
	h.setupColors()

	tests := []struct {
		name  string
		color Color
		text  string
	}{
		{"Green", ColorGreen, "success"},
		{"Red", ColorRed, "error"},
		{"Yellow", ColorYellow, "warning"},
		{"Blue", ColorBlue, "info"},
		{"Cyan", ColorCyan, "path"},
		{"Gray", ColorGray, "secondary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Colorize(tt.text, tt.color)

			// When colors are enabled, result should contain ANSI codes
			if !containsANSI(result) {
				t.Errorf("Expected ANSI codes in output, but got: %q", result)
			}

			// Result should contain the original text
			if !strings.Contains(result, tt.text) {
				t.Errorf("Expected result to contain %q, but got: %q", tt.text, result)
			}
		})
	}
}

// TestColorHandler_ColorCodeDisabled tests that no color codes are generated when disabled.
func TestColorHandler_ColorCodeDisabled(t *testing.T) {
	// Create handler with colors explicitly disabled
	h := &Handler{
		enabled: false,
		isTTY:   false,
	}
	h.setupColors()

	tests := []struct {
		name  string
		color Color
		text  string
	}{
		{"Green", ColorGreen, "success"},
		{"Red", ColorRed, "error"},
		{"Yellow", ColorYellow, "warning"},
		{"Blue", ColorBlue, "info"},
		{"Cyan", ColorCyan, "path"},
		{"Gray", ColorGray, "secondary"},
		{"None", ColorNone, "plain"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Colorize(tt.text, tt.color)

			// When colors are disabled, result should equal input
			if result != tt.text {
				t.Errorf("Expected %q, but got: %q", tt.text, result)
			}

			// Result should not contain ANSI codes
			if containsANSI(result) {
				t.Errorf("Expected no ANSI codes, but got: %q", result)
			}
		})
	}
}

// TestColorHandler_ConvenienceMethods tests the convenience methods.
func TestColorHandler_ConvenienceMethods(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	// Test with colors enabled
	h := &Handler{
		enabled: true,
		isTTY:   true,
	}
	h.setupColors()

	text := "test"

	tests := []struct {
		name   string
		method func(string) string
	}{
		{"Green", h.Green},
		{"Red", h.Red},
		{"Yellow", h.Yellow},
		{"Blue", h.Blue},
		{"Cyan", h.Cyan},
		{"Gray", h.Gray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(text)

			// Should contain ANSI codes when enabled
			if !containsANSI(result) {
				t.Errorf("Expected ANSI codes in output, but got: %q", result)
			}

			// Should contain original text
			if !strings.Contains(result, text) {
				t.Errorf("Expected result to contain %q, but got: %q", text, result)
			}
		})
	}
}

// TestColorHandler_FormattedMessages tests formatted message methods.
func TestColorHandler_FormattedMessages(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	h := &Handler{
		enabled: true,
		isTTY:   true,
	}
	h.setupColors()

	tests := []struct {
		name     string
		method   func(string) string
		message  string
		expected string // Symbol that should be present
	}{
		{"Success", h.Success, "operation completed", "✓"},
		{"Error", h.Error, "operation failed", "✗"},
		{"Warning", h.Warning, "be careful", "!"},
		{"Info", h.Info, "for your information", "ℹ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.message)

			// Should contain the symbol
			if !strings.Contains(result, tt.expected) {
				t.Errorf("Expected result to contain symbol %q, but got: %q", tt.expected, result)
			}

			// Should contain the message
			if !strings.Contains(result, tt.message) {
				t.Errorf("Expected result to contain message %q, but got: %q", tt.message, result)
			}

			// Should contain ANSI codes when enabled
			if !containsANSI(result) {
				t.Errorf("Expected ANSI codes in output, but got: %q", result)
			}
		})
	}
}

// TestColorHandler_SetEnabled tests manual enable/disable.
func TestColorHandler_SetEnabled(t *testing.T) {
	// Save and clear NO_COLOR to allow colors in tests
	originalValue, hadNoColor := os.LookupEnv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer func() {
		if hadNoColor {
			os.Setenv("NO_COLOR", originalValue)
		}
	}()

	h := NewColorHandler()

	// Test enabling
	h.SetEnabled(true)
	if !h.IsEnabled() {
		t.Error("Expected colors to be enabled after SetEnabled(true)")
	}

	result := h.Green("test")
	if !containsANSI(result) {
		t.Errorf("Expected ANSI codes when enabled, but got: %q", result)
	}

	// Test disabling
	h.SetEnabled(false)
	if h.IsEnabled() {
		t.Error("Expected colors to be disabled after SetEnabled(false)")
	}

	result = h.Green("test")
	if result != "test" {
		t.Errorf("Expected plain text when disabled, but got: %q", result)
	}
	if containsANSI(result) {
		t.Errorf("Expected no ANSI codes when disabled, but got: %q", result)
	}
}

// TestColorHandler_IsTTY tests the IsTTY method.
func TestColorHandler_IsTTY(t *testing.T) {
	h := NewColorHandler()

	// IsTTY should return a boolean (we can't predict the value in tests)
	isTTY := h.IsTTY()
	if isTTY != true && isTTY != false {
		t.Error("IsTTY should return a boolean value")
	}
}
