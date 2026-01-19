// Package output provides formatters for hashi output.
//
// It supports multiple output formats:
// - default: Groups files by matching hash with blank lines between groups
// - verbose: Detailed output with summaries and statistics
// - json: Machine-readable JSON format
// - plain: Tab-separated format for scripting (grep/awk/cut)
package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Les-El/hashi/internal/hash"
)

// Formatter is the interface for output formatters.
type Formatter interface {
	// Format formats the hash result for output.
	Format(result *hash.Result) string
}

// DefaultFormatter groups files by matching hash with blank lines between groups.
type DefaultFormatter struct{}

// Format implements Formatter for DefaultFormatter.
func (f *DefaultFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	// Output match groups first
	for i, group := range result.Matches {
		if i > 0 {
			sb.WriteString("\n")
		}
		for _, entry := range group.Entries {
			sb.WriteString(fmt.Sprintf("%s    %s\n", entry.Original, entry.Hash))
		}
	}

	// Add blank line before unmatched if there were matches
	if len(result.Matches) > 0 && len(result.Unmatched) > 0 {
		sb.WriteString("\n")
	}

	// Output unmatched files
	for i, entry := range result.Unmatched {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s    %s\n", entry.Original, entry.Hash))
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// PreserveOrderFormatter maintains input order without grouping.
type PreserveOrderFormatter struct{}

// Format implements Formatter for PreserveOrderFormatter.
func (f *PreserveOrderFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	for _, entry := range result.Entries {
		if entry.Error == nil {
			sb.WriteString(fmt.Sprintf("%s    %s\n", entry.Original, entry.Hash))
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// VerboseFormatter provides detailed output with summaries.
type VerboseFormatter struct{}

// Format implements Formatter for VerboseFormatter.
func (f *VerboseFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	// Header with processing stats
	sb.WriteString(fmt.Sprintf("Processed %d files in %s\n\n",
		result.FilesProcessed, result.Duration.Round(time.Millisecond)))

	// Match groups
	if len(result.Matches) > 0 {
		sb.WriteString("Match Groups:\n")
		for i, group := range result.Matches {
			sb.WriteString(fmt.Sprintf("  Group %d (%d files):\n", i+1, group.Count))
			for _, entry := range group.Entries {
				sb.WriteString(fmt.Sprintf("    %s    %s\n", entry.Original, entry.Hash))
			}
			sb.WriteString("\n")
		}
	}

	// Unmatched files
	if len(result.Unmatched) > 0 {
		sb.WriteString("Unmatched Files:\n")
		for _, entry := range result.Unmatched {
			sb.WriteString(fmt.Sprintf("  %s    %s\n", entry.Original, entry.Hash))
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString(fmt.Sprintf("Summary: %d match groups, %d unmatched files",
		len(result.Matches), len(result.Unmatched)))

	return sb.String()
}

// JSONFormatter outputs results in machine-readable JSON format.
type JSONFormatter struct{}

// jsonOutput is the structure for JSON output.
type jsonOutput struct {
	Processed   int              `json:"processed"`
	DurationMS  int64            `json:"duration_ms"`
	MatchGroups []jsonMatchGroup `json:"match_groups"`
	Unmatched   []jsonEntry      `json:"unmatched"`
	Errors      []string         `json:"errors"`
}

type jsonMatchGroup struct {
	Hash  string   `json:"hash"`
	Count int      `json:"count"`
	Files []string `json:"files"`
}

type jsonEntry struct {
	File string `json:"file"`
	Hash string `json:"hash"`
}

// Format implements Formatter for JSONFormatter.
func (f *JSONFormatter) Format(result *hash.Result) string {
	output := jsonOutput{
		Processed:   result.FilesProcessed,
		DurationMS:  result.Duration.Milliseconds(),
		MatchGroups: make([]jsonMatchGroup, 0, len(result.Matches)),
		Unmatched:   make([]jsonEntry, 0, len(result.Unmatched)),
		Errors:      make([]string, 0, len(result.Errors)),
	}

	// Convert match groups
	for _, group := range result.Matches {
		files := make([]string, 0, len(group.Entries))
		for _, entry := range group.Entries {
			files = append(files, entry.Original)
		}
		output.MatchGroups = append(output.MatchGroups, jsonMatchGroup{
			Hash:  group.Hash,
			Count: group.Count,
			Files: files,
		})
	}

	// Convert unmatched entries
	for _, entry := range result.Unmatched {
		output.Unmatched = append(output.Unmatched, jsonEntry{
			File: entry.Original,
			Hash: entry.Hash,
		})
	}

	// Convert errors
	for _, err := range result.Errors {
		output.Errors = append(output.Errors, err.Error())
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal JSON: %s"}`, err.Error())
	}

	return string(data)
}

// PlainFormatter outputs tab-separated results for scripting.
type PlainFormatter struct{}

// Format implements Formatter for PlainFormatter.
func (f *PlainFormatter) Format(result *hash.Result) string {
	var sb strings.Builder

	// Output all entries in input order, tab-separated
	for _, entry := range result.Entries {
		if entry.Error == nil {
			sb.WriteString(fmt.Sprintf("%s\t%s\n", entry.Original, entry.Hash))
		}
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// NewFormatter creates a formatter based on the format name.
func NewFormatter(format string, preserveOrder bool) Formatter {
	switch format {
	case "verbose":
		return &VerboseFormatter{}
	case "json":
		return &JSONFormatter{}
	case "plain":
		return &PlainFormatter{}
	default:
		if preserveOrder {
			return &PreserveOrderFormatter{}
		}
		return &DefaultFormatter{}
	}
}
