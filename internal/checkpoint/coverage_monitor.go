package checkpoint

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

// CoverageMonitor validates and tracks test coverage.
type CoverageMonitor struct {
	threshold float64
}

// NewCoverageMonitor creates a new coverage monitor with the given threshold.
func NewCoverageMonitor(threshold float64) *CoverageMonitor {
	if threshold <= 0 {
		threshold = 85.0
	}
	return &CoverageMonitor{
		threshold: threshold,
	}
}

// ParseCoverageOutput parses the output of 'go test -cover' to extract coverage percentages.
func (m *CoverageMonitor) ParseCoverageOutput(output string) (map[string]float64, error) {
	coverage := make(map[string]float64)
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		pkg, cov, ok := m.parseLine(scanner.Text())
		if ok {
			coverage[pkg] = cov
		}
	}
	return coverage, nil
}

func (m *CoverageMonitor) parseLine(line string) (string, float64, bool) {
	if !strings.Contains(line, "coverage:") {
		return "", 0, false
	}

	parts := strings.Fields(line)
	if len(parts) < 5 {
		return "", 0, false
	}

	pkg := parts[1]
	// Find the part that contains the percentage
	var covStr string
	for i, part := range parts {
		if part == "coverage:" && i+1 < len(parts) {
			covStr = strings.TrimSuffix(parts[i+1], "%")
			break
		}
	}

	if covStr == "" {
		return "", 0, false
	}

	cov, err := strconv.ParseFloat(covStr, 64)
	if err != nil {
		return "", 0, false
	}
	return pkg, cov, true
}

// ValidateThreshold checks if all packages meet the coverage threshold.
func (m *CoverageMonitor) ValidateThreshold(coverage map[string]float64) ([]string, bool) {
	var failures []string
	for pkg, cov := range coverage {
		if cov < m.threshold {
			failures = append(failures, fmt.Sprintf("package %s coverage %.1f%% is below threshold %.1f%%", pkg, cov, m.threshold))
		}
	}
	return failures, len(failures) == 0
}

// GenerateCoverageReport creates a markdown report of coverage status.
func (m *CoverageMonitor) GenerateCoverageReport(coverage map[string]float64) string {
	var sb strings.Builder
	sb.WriteString("# Test Coverage Report\n\n")
	sb.WriteString("| Package | Coverage | Status |\n")
	sb.WriteString("|---------|----------|--------|\n")

	for pkg, cov := range coverage {
		status := "✅"
		if cov < m.threshold {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %.1f%% | %s |\n", pkg, cov, status))
	}

	return sb.String()
}
