// Package config handles configuration and argument parsing.
package config

import (
	"strings"
)

// KnownFlags is a list of all supported long flag names.
var KnownFlags = []string{
	"help",
	"version",
	"verbose",
	"quiet",
	"bool",
	"recursive",
	"hidden",
	"algorithm",
	"preserve-order",
	"match-required",
	"format",
	"output",
	"append",
	"force",
	"log-file",
	"log-json",
	"include",
	"exclude",
	"min-size",
	"max-size",
	"modified-after",
	"modified-before",
	"config",
	"h",
	"V",
	"v",
	"q",
	"b",
	"r",
	"a",
	"f",
	"o",
	"c",
	"i",
	"e",
}

// SuggestFlag suggests a similar flag name if a typo is detected.
func SuggestFlag(unknown string) string {
	unknownLower := strings.ToLower(strings.TrimPrefix(unknown, "--"))
	unknownLower = strings.TrimPrefix(unknownLower, "-")

	bestMatch := ""
	minDist := 4 // Increased distance for longer flag names

	// First check for substring matches (e.g. "algo" matches "algorithm")
	if len(unknownLower) >= 3 {
		for _, known := range KnownFlags {
			if strings.HasPrefix(known, unknownLower) && len(known) > 1 {
				return "--" + known
			}
		}
	}

	for _, known := range KnownFlags {
		knownLower := strings.ToLower(known)
		// Optimization: length difference is a lower bound for Levenshtein distance
		lenDiff := len(unknownLower) - len(knownLower)
		if lenDiff < 0 {
			lenDiff = -lenDiff
		}
		if lenDiff >= minDist {
			continue
		}

		dist := levenshtein(unknownLower, knownLower)
		if dist < minDist {
			minDist = dist
			bestMatch = known
		}
	}

	if bestMatch != "" {
		if len(bestMatch) == 1 {
			return "-" + bestMatch
		}
		return "--" + bestMatch
	}

	return ""
}

// levenshtein calculates the Levenshtein distance between two strings.
func levenshtein(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	row := make([]int, len(s2)+1)
	for i := 0; i < len(row); i++ {
		row[i] = i
	}

	for i := 0; i < len(s1); i++ {
		row[len(s2)] = updateLevenshteinRow(s1[i], s2, row)
	}
	return row[len(s2)]
}

func updateLevenshteinRow(char1 byte, s2 string, row []int) int {
	prev := row[0] + 1
	for j := 0; j < len(s2); j++ {
		cur := row[j]
		if char1 != s2[j] {
			cur++
		}
		if cur > prev+1 {
			cur = prev + 1
		}
		if cur > row[j+1]+1 {
			cur = row[j+1] + 1
		}
		row[j], prev = prev, cur
	}
	return prev
}
