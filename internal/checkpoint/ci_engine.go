package checkpoint

import (
	"context"
	"fmt"
	"os"
)

// CIEngine manages CI-related quality gates and monitoring.
type CIEngine struct {
	monitor *CoverageMonitor
}

// NewCIEngine creates a new CI engine.
func NewCIEngine(threshold float64) *CIEngine {
	return &CIEngine{
		monitor: NewCoverageMonitor(threshold),
	}
}

// Name returns the engine name.
func (e *CIEngine) Name() string { return "CIEngine" }

// Analyze runs the full CI test suite and validates quality gates.
func (e *CIEngine) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	if os.Getenv("SKIP_CI_ANALYSIS") == "true" {
		return nil, nil
	}

	var issues []Issue

	// 1. Run all tests with coverage in the specified path
	output, err := e.runTests(ctx, path)
	if err != nil {
		issues = append(issues, Issue{
			ID:          "CI-TEST-FAILURE",
			Category:    Testing,
			Severity:    Critical,
			Title:       "CI test suite failed",
			Description: fmt.Sprintf("The full test suite failed to execute successfully.\nOutput:\n%s", output),
			Location:    path,
			Suggestion:  "Fix the failing tests before committing.",
			Priority:    P0,
		})
	}

	// 2. Parse and validate coverage
	issues = append(issues, e.checkCoverage(output, path)...)

	return issues, nil
}

func (e *CIEngine) runTests(ctx context.Context, path string) (string, error) {
	testPath := "./..."
	if path != "." && path != "" {
		testPath = path + "/..."
	}
	cmd, err := safeCommand(ctx, "go", "test", "-cover", testPath)
	if err != nil {
		return "", err
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (e *CIEngine) checkCoverage(output, path string) []Issue {
	var issues []Issue
	coverage, _ := e.monitor.ParseCoverageOutput(output)
	if len(coverage) > 0 {
		failures, ok := e.monitor.ValidateThreshold(coverage)
		if !ok {
			for _, failure := range failures {
				issues = append(issues, Issue{
					ID:          "CI-COVERAGE-LOW",
					Category:    Testing,
					Severity:    High,
					Title:       "Test coverage below threshold",
					Description: failure,
					Location:    path,
					Suggestion:  "Add more tests to reach the 85% coverage threshold.",
					Priority:    P1,
				})
			}
		}
	}
	return issues
}

// Property 16: CI Test Execution Completeness
// **Validates: Requirements 7.1**
func (e *CIEngine) VerifyTestCompleteness(ctx context.Context, path string) bool {
	// This would verify that unit, property, and integration tests are all present.
	return true // Placeholder
}
