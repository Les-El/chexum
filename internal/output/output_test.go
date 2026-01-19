package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/Les-El/hashi/internal/hash"
)

// Feature: cli-guidelines-review, Property 19: Default output groups by matches
// **Validates: Requirements 2.5**
func TestProperty_DefaultOutputGroupsByMatches(t *testing.T) {
	f := func(numGroups uint8, filesPerGroup uint8) bool {
		// Limit to reasonable sizes for testing
		if numGroups == 0 || numGroups > 10 {
			return true // Skip invalid inputs
		}
		if filesPerGroup == 0 || filesPerGroup > 10 {
			return true // Skip invalid inputs
		}

		// Create test data with match groups
		result := &hash.Result{
			Matches:        make([]hash.MatchGroup, 0),
			Unmatched:      make([]hash.Entry, 0),
			FilesProcessed: 0,
			Duration:       time.Second,
		}

		// Create match groups
		for i := uint8(0); i < numGroups; i++ {
			entries := make([]hash.Entry, 0)
			hashValue := strings.Repeat(string('a'+i), 64) // Unique hash per group

			for j := uint8(0); j < filesPerGroup; j++ {
				entries = append(entries, hash.Entry{
					Original: string('a'+i) + "_file_" + string('0'+j) + ".txt",
					Hash:     hashValue,
					IsFile:   true,
				})
			}

			result.Matches = append(result.Matches, hash.MatchGroup{
				Hash:    hashValue,
				Entries: entries,
				Count:   int(filesPerGroup),
			})
			result.FilesProcessed += int(filesPerGroup)
		}

		// Format with default formatter
		formatter := &DefaultFormatter{}
		output := formatter.Format(result)

		// Property: Output should have blank lines between groups
		lines := strings.Split(output, "\n")
		
		// Count blank lines (should be numGroups - 1)
		blankLines := 0
		for _, line := range lines {
			if line == "" {
				blankLines++
			}
		}

		// For match groups, we expect numGroups - 1 blank lines between them
		expectedBlankLines := int(numGroups) - 1
		if blankLines != expectedBlankLines {
			return false
		}

		// Property: Each group should be contiguous (no blank lines within a group)
		currentGroup := 0
		filesInCurrentGroup := 0
		for _, line := range lines {
			if line == "" {
				// Blank line means we're moving to next group
				if filesInCurrentGroup != int(filesPerGroup) {
					return false // Group wasn't complete
				}
				currentGroup++
				filesInCurrentGroup = 0
			} else if line != "" {
				filesInCurrentGroup++
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Feature: cli-guidelines-review, Property 20: Preserve-order flag maintains input order
// **Validates: Requirements 2.5**
func TestProperty_PreserveOrderMaintainsInputOrder(t *testing.T) {
	f := func(numFiles uint8) bool {
		// Limit to reasonable sizes
		if numFiles == 0 || numFiles > 20 {
			return true
		}

		// Create test data with entries in specific order
		result := &hash.Result{
			Entries:        make([]hash.Entry, 0),
			FilesProcessed: int(numFiles),
			Duration:       time.Second,
		}

		// Create entries with predictable names and hashes
		for i := uint8(0); i < numFiles; i++ {
			result.Entries = append(result.Entries, hash.Entry{
				Original: "file_" + string('0'+i) + ".txt",
				Hash:     strings.Repeat(string('a'+(i%5)), 64), // Some will match
				IsFile:   true,
			})
		}

		// Format with preserve order formatter
		formatter := &PreserveOrderFormatter{}
		output := formatter.Format(result)

		// Property: Output order should match input order
		lines := strings.Split(output, "\n")
		
		// Filter out empty lines
		nonEmptyLines := make([]string, 0)
		for _, line := range lines {
			if line != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}

		if len(nonEmptyLines) != int(numFiles) {
			return false
		}

		// Check that each line corresponds to the correct input entry
		for i, line := range nonEmptyLines {
			expectedPrefix := "file_" + string('0'+uint8(i)) + ".txt"
			if !strings.HasPrefix(line, expectedPrefix) {
				return false
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Feature: cli-guidelines-review, Property 10: JSON output is valid
// **Validates: Requirements 7.1**
func TestProperty_JSONOutputIsValid(t *testing.T) {
	f := func(numMatches uint8, numUnmatched uint8) bool {
		// Limit to reasonable sizes
		if numMatches > 10 || numUnmatched > 10 {
			return true
		}

		// Create test data
		result := &hash.Result{
			Matches:        make([]hash.MatchGroup, 0),
			Unmatched:      make([]hash.Entry, 0),
			Errors:         make([]error, 0),
			FilesProcessed: int(numMatches + numUnmatched),
			Duration:       time.Second,
		}

		// Add match groups
		for i := uint8(0); i < numMatches; i++ {
			entries := []hash.Entry{
				{Original: "match_" + string('0'+i) + "_a.txt", Hash: strings.Repeat(string('a'+i), 64)},
				{Original: "match_" + string('0'+i) + "_b.txt", Hash: strings.Repeat(string('a'+i), 64)},
			}
			result.Matches = append(result.Matches, hash.MatchGroup{
				Hash:    strings.Repeat(string('a'+i), 64),
				Entries: entries,
				Count:   2,
			})
		}

		// Add unmatched entries
		for i := uint8(0); i < numUnmatched; i++ {
			result.Unmatched = append(result.Unmatched, hash.Entry{
				Original: "unmatched_" + string('0'+i) + ".txt",
				Hash:     strings.Repeat(string('z'-i), 64),
			})
		}

		// Format with JSON formatter
		formatter := &JSONFormatter{}
		output := formatter.Format(result)

		// Property: Output should be valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			return false
		}

		// Property: JSON should contain expected fields
		if _, ok := parsed["processed"]; !ok {
			return false
		}
		if _, ok := parsed["duration_ms"]; !ok {
			return false
		}
		if _, ok := parsed["match_groups"]; !ok {
			return false
		}
		if _, ok := parsed["unmatched"]; !ok {
			return false
		}
		if _, ok := parsed["errors"]; !ok {
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Feature: cli-guidelines-review, Property 11: Plain output is line-based
// **Validates: Requirements 7.2**
func TestProperty_PlainOutputIsLineBased(t *testing.T) {
	f := func(numFiles uint8) bool {
		// Limit to reasonable sizes
		if numFiles == 0 || numFiles > 20 {
			return true
		}

		// Create test data
		result := &hash.Result{
			Entries:        make([]hash.Entry, 0),
			FilesProcessed: int(numFiles),
			Duration:       time.Second,
		}

		for i := uint8(0); i < numFiles; i++ {
			result.Entries = append(result.Entries, hash.Entry{
				Original: "file_" + string('0'+i) + ".txt",
				Hash:     strings.Repeat(string('a'+(i%5)), 64),
				IsFile:   true,
			})
		}

		// Format with plain formatter
		formatter := &PlainFormatter{}
		output := formatter.Format(result)

		// Property: Each line should have exactly one tab character
		lines := strings.Split(output, "\n")
		
		// Filter out empty lines
		nonEmptyLines := make([]string, 0)
		for _, line := range lines {
			if line != "" {
				nonEmptyLines = append(nonEmptyLines, line)
			}
		}

		if len(nonEmptyLines) != int(numFiles) {
			return false
		}

		for _, line := range nonEmptyLines {
			// Each line should have exactly one tab
			if strings.Count(line, "\t") != 1 {
				return false
			}

			// Line should have two parts: filename and hash
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				return false
			}

			// Both parts should be non-empty
			if parts[0] == "" || parts[1] == "" {
				return false
			}
		}

		// Property: No blank lines in plain output
		for _, line := range lines {
			if line == "" && len(lines) > 1 {
				// Only allow empty line at the very end
				if line != lines[len(lines)-1] {
					return false
				}
			}
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(f, config); err != nil {
		t.Error(err)
	}
}

// Unit tests for all formatters

func TestDefaultFormatter_EmptyResult(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		Matches:   []hash.MatchGroup{},
		Unmatched: []hash.Entry{},
	}

	output := formatter.Format(result)
	if output != "" {
		t.Errorf("Expected empty output for empty result, got: %q", output)
	}
}

func TestDefaultFormatter_SingleFile(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		Matches: []hash.MatchGroup{},
		Unmatched: []hash.Entry{
			{Original: "file.txt", Hash: "abc123"},
		},
	}

	output := formatter.Format(result)
	expected := "file.txt    abc123"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestDefaultFormatter_ManyMatches(t *testing.T) {
	formatter := &DefaultFormatter{}
	result := &hash.Result{
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "file2.txt", Hash: "hash1"},
				},
				Count: 2,
			},
			{
				Hash: "hash2",
				Entries: []hash.Entry{
					{Original: "file3.txt", Hash: "hash2"},
					{Original: "file4.txt", Hash: "hash2"},
				},
				Count: 2,
			},
		},
		Unmatched: []hash.Entry{},
	}

	output := formatter.Format(result)
	
	// Should have blank line between groups
	if !strings.Contains(output, "\n\n") {
		t.Error("Expected blank line between match groups")
	}

	// Should contain all files
	for _, group := range result.Matches {
		for _, entry := range group.Entries {
			if !strings.Contains(output, entry.Original) {
				t.Errorf("Expected output to contain %s", entry.Original)
			}
		}
	}
}

func TestPreserveOrderFormatter_MaintainsOrder(t *testing.T) {
	formatter := &PreserveOrderFormatter{}
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1"},
			{Original: "file2.txt", Hash: "hash2"},
			{Original: "file3.txt", Hash: "hash1"}, // Matches file1
		},
	}

	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	// Should maintain input order
	if !strings.HasPrefix(lines[0], "file1.txt") {
		t.Error("First line should be file1.txt")
	}
	if !strings.HasPrefix(lines[1], "file2.txt") {
		t.Error("Second line should be file2.txt")
	}
	if !strings.HasPrefix(lines[2], "file3.txt") {
		t.Error("Third line should be file3.txt")
	}
}

func TestVerboseFormatter_IncludesSummary(t *testing.T) {
	formatter := &VerboseFormatter{}
	result := &hash.Result{
		FilesProcessed: 5,
		Duration:       123 * time.Millisecond,
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "file2.txt", Hash: "hash1"},
				},
				Count: 2,
			},
		},
		Unmatched: []hash.Entry{
			{Original: "file3.txt", Hash: "hash3"},
		},
	}

	output := formatter.Format(result)

	// Should include processing stats
	if !strings.Contains(output, "Processed 5 files") {
		t.Error("Expected processing stats")
	}

	// Should include summary
	if !strings.Contains(output, "Summary:") {
		t.Error("Expected summary section")
	}

	// Should mention match groups
	if !strings.Contains(output, "Match Groups:") {
		t.Error("Expected match groups section")
	}

	// Should mention unmatched files
	if !strings.Contains(output, "Unmatched Files:") {
		t.Error("Expected unmatched files section")
	}
}

func TestJSONFormatter_ValidStructure(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &hash.Result{
		FilesProcessed: 3,
		Duration:       100 * time.Millisecond,
		Matches: []hash.MatchGroup{
			{
				Hash: "hash1",
				Entries: []hash.Entry{
					{Original: "file1.txt", Hash: "hash1"},
					{Original: "file2.txt", Hash: "hash1"},
				},
				Count: 2,
			},
		},
		Unmatched: []hash.Entry{
			{Original: "file3.txt", Hash: "hash3"},
		},
		Errors: []error{},
	}

	output := formatter.Format(result)

	// Parse JSON
	var parsed jsonOutput
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify structure
	if parsed.Processed != 3 {
		t.Errorf("Expected processed=3, got %d", parsed.Processed)
	}

	if len(parsed.MatchGroups) != 1 {
		t.Errorf("Expected 1 match group, got %d", len(parsed.MatchGroups))
	}

	if len(parsed.Unmatched) != 1 {
		t.Errorf("Expected 1 unmatched, got %d", len(parsed.Unmatched))
	}

	if parsed.MatchGroups[0].Count != 2 {
		t.Errorf("Expected match group count=2, got %d", parsed.MatchGroups[0].Count)
	}
}

func TestPlainFormatter_TabSeparated(t *testing.T) {
	formatter := &PlainFormatter{}
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1"},
			{Original: "file2.txt", Hash: "hash2"},
		},
	}

	output := formatter.Format(result)
	lines := strings.Split(output, "\n")

	// Each line should be tab-separated
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "\t") {
			t.Errorf("Expected tab-separated line, got: %q", line)
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			t.Errorf("Expected 2 parts, got %d in line: %q", len(parts), line)
		}
	}
}

func TestNewFormatter_SelectsCorrectFormatter(t *testing.T) {
	tests := []struct {
		format        string
		preserveOrder bool
		expectedType  string
	}{
		{"verbose", false, "*output.VerboseFormatter"},
		{"json", false, "*output.JSONFormatter"},
		{"plain", false, "*output.PlainFormatter"},
		{"default", false, "*output.DefaultFormatter"},
		{"default", true, "*output.PreserveOrderFormatter"},
		{"", false, "*output.DefaultFormatter"},
		{"", true, "*output.PreserveOrderFormatter"},
	}

	for _, tt := range tests {
		formatter := NewFormatter(tt.format, tt.preserveOrder)
		typeName := fmt.Sprintf("%T", formatter)
		if typeName != tt.expectedType {
			t.Errorf("NewFormatter(%q, %v) = %s, want %s",
				tt.format, tt.preserveOrder, typeName, tt.expectedType)
		}
	}
}

func TestFormatters_HandleErrors(t *testing.T) {
	result := &hash.Result{
		Entries: []hash.Entry{
			{Original: "file1.txt", Hash: "hash1", Error: nil},
			{Original: "file2.txt", Hash: "", Error: fmt.Errorf("read error")},
		},
		FilesProcessed: 1,
		Duration:       time.Second,
	}

	// Plain and PreserveOrder formatters should skip entries with errors
	plainFormatter := &PlainFormatter{}
	plainOutput := plainFormatter.Format(result)
	if strings.Contains(plainOutput, "file2.txt") {
		t.Error("Plain formatter should skip entries with errors")
	}

	preserveFormatter := &PreserveOrderFormatter{}
	preserveOutput := preserveFormatter.Format(result)
	if strings.Contains(preserveOutput, "file2.txt") {
		t.Error("PreserveOrder formatter should skip entries with errors")
	}
}
