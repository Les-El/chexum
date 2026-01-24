package checkpoint

import (
	"context"
	"testing"
)

type MockEngine struct {
	name   string
	issues []Issue
}

func (m *MockEngine) Name() string { return m.name }
func (m *MockEngine) Analyze(ctx context.Context, path string) ([]Issue, error) {
	return m.issues, nil
}

func TestIssueCollector(t *testing.T) {
	collector := NewIssueCollector()
	issues := []Issue{
		{ID: "1", Title: "Issue 1"},
		{ID: "2", Title: "Issue 2"},
	}
	collector.Collect(issues)

	collected := collector.Issues()
	if len(collected) != 2 {
		t.Errorf("expected 2 issues, got %d", len(collected))
	}
}

func TestSystemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	engines := []AnalysisEngine{
		NewCodeAnalyzer(),
		NewDependencyAnalyzer(),
		NewDocAuditor(),
		// NewTestingBattery(), // Skipping to avoid recursive go test calls
		NewFlagSystem(),
		NewQualityEngine(),
	}

	runner := NewRunner(engines)
	err := runner.Run(ctx, "../../")
	if err != nil {
		t.Fatalf("System run failed: %v", err)
	}

	issues := runner.GetIssues()
	if len(issues) == 0 {
		t.Log("No issues found, which is possible but unlikely in this repo.")
	}

	flagSystem := NewFlagSystem()
	flags, _ := flagSystem.CatalogFlags(ctx, "../../")
	flags, _ = flagSystem.ClassifyImplementation(ctx, "../../", flags)
	flags, _ = flagSystem.PerformCrossReferenceAnalysis(ctx, "../../", flags)
	flags, _ = flagSystem.DetectConflicts(ctx, flags)

	reporter := NewReporter()
	reporter.Aggregate(issues, flags)

	plan, err := reporter.GenerateRemediationPlan()
	if err != nil {
		t.Errorf("Failed to generate remediation plan: %v", err)
	}
	if plan == "" {
		t.Errorf("Remediation plan is empty")
	}

	dashboard, err := reporter.GenerateStatusDashboard()
	if err != nil {
		t.Errorf("Failed to generate status dashboard: %v", err)
	}
	if dashboard == "" {
		t.Errorf("Status dashboard is empty")
	}
}
