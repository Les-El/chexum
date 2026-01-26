package checkpoint

import (
	"strings"
	"testing"
)

func TestNewReporter(t *testing.T) {
	reporter := NewReporter()
	if reporter == nil {
		t.Fatal("NewReporter returned nil")
	}
}

func TestAggregate(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{{ID: "1"}}
	flags := []FlagStatus{{Name: "test"}}
	err := reporter.Aggregate(issues, flags)
	if err != nil {
		t.Errorf("Aggregate failed: %v", err)
	}
	if len(reporter.issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(reporter.issues))
	}
	if len(reporter.flagStatuses) != 1 {
		t.Errorf("expected 1 flag, got %d", len(reporter.flagStatuses))
	}
}

func TestGenerateJSONReport(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{{ID: "ISSUE-1"}}
	reporter.Aggregate(issues, nil)
	json, err := reporter.GenerateJSONReport()
	if err != nil {
		t.Errorf("GenerateJSONReport failed: %v", err)
	}
	if !strings.Contains(json, "ISSUE-1") {
		t.Error("JSON report missing issue ID")
	}
}

func TestGenerateCSVReport(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{{ID: "ISSUE-1"}}
	reporter.Aggregate(issues, nil)
	csv, err := reporter.GenerateCSVReport()
	if err != nil {
		t.Errorf("GenerateCSVReport failed: %v", err)
	}
	if !strings.Contains(csv, "ISSUE-1") {
		t.Error("CSV report missing issue ID")
	}
}

func TestGenerateRemediationPlan(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{{ID: "ISSUE-1", Priority: P2}}
	reporter.Aggregate(issues, nil)
	plan, err := reporter.GenerateRemediationPlan()
	if err != nil {
		t.Errorf("GenerateRemediationPlan failed: %v", err)
	}
	if !strings.Contains(plan, "Priority P2 Tasks") {
		t.Error("Remediation plan missing priority section")
	}
}

func TestGenerateStatusDashboard(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{{ID: "ISSUE-1"}}
	reporter.Aggregate(issues, nil)
	dashboard, err := reporter.GenerateStatusDashboard()
	if err != nil {
		t.Errorf("GenerateStatusDashboard failed: %v", err)
	}
	if !strings.Contains(dashboard, "**Total Issues**: 1") {
		t.Error("Dashboard missing total issues count")
	}
}

func TestGenerateOnboardingGuide(t *testing.T) {
	reporter := NewReporter()
	guide, err := reporter.GenerateOnboardingGuide()
	if err != nil {
		t.Errorf("GenerateOnboardingGuide failed: %v", err)
	}
	if !strings.Contains(guide, "Developer Onboarding Guide") {
		t.Error("Onboarding guide missing title")
	}
}

func TestReporter_Counts(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{
		{Category: CodeQuality, Severity: Critical},
		{Category: Testing, Severity: High},
		{Category: Documentation, Severity: Medium},
	}
	reporter.Aggregate(issues, nil)

	if count := reporter.countByCategory(CodeQuality); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
	if count := reporter.countByCategory(Security); count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
	if count := reporter.countHighSeverity(); count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestReporter_CountImplementedFlags(t *testing.T) {
	reporter := NewReporter()
	flags := []FlagStatus{
		{Status: FullyImplemented},
		{Status: PartiallyImplemented},
	}
	reporter.Aggregate(nil, flags)

	if count := reporter.countImplementedFlags(); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

// Reviewed: LONG-FUNCTION - Table-driven test with many sorting cases.
func TestGenerateRemediationPlan_NoIssues(t *testing.T) {
	reporter := NewReporter()
	reporter.Aggregate(nil, nil)
	plan, err := reporter.GenerateRemediationPlan()
	if err != nil {
		t.Fatalf("GenerateRemediationPlan failed: %v", err)
	}
	if !strings.Contains(plan, "identified 0 issues") {
		t.Error("Expected 0 issues in plan")
	}
}

func TestGenerateJSONReport_NoIssues(t *testing.T) {
	reporter := NewReporter()
	reporter.Aggregate(nil, nil)
	json, err := reporter.GenerateJSONReport()
	if err != nil {
		t.Fatalf("GenerateJSONReport failed: %v", err)
	}
	if !strings.Contains(json, "issues") {
		t.Error("JSON report missing issues key")
	}
}

func TestGenerateCSVReport_NoIssues(t *testing.T) {
	reporter := NewReporter()
	reporter.Aggregate(nil, nil)
	csv, err := reporter.GenerateCSVReport()
	if err != nil {
		t.Fatalf("GenerateCSVReport failed: %v", err)
	}
	if !strings.Contains(csv, "Status,ID,Category") {
		t.Error("CSV report missing header")
	}
}

// Reviewed: LONG-FUNCTION - Table-driven test with many sorting cases.
func TestSortIssues(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{
		{ID: "LOW-P", Priority: P3, Severity: Medium},
		{ID: "HIGH-P", Priority: P0, Severity: High},
		{ID: "MED-P", Priority: P2, Severity: Critical},
		{ID: "SAME-P-LOW-S", Priority: P1, Severity: Info},
		{ID: "SAME-P-HIGH-S", Priority: P1, Severity: Critical},
		{ID: "UNKNOWN-P", Priority: "unknown", Severity: "unknown"},
	}
	reporter.Aggregate(issues, nil)
	reporter.SortIssues()

	if reporter.issues[0].ID != "HIGH-P" {
		t.Errorf("expected HIGH-P first, got %s", reporter.issues[0].ID)
	}
	
	// Test same priority, different severity
	if reporter.issues[1].ID != "SAME-P-HIGH-S" {
		t.Errorf("expected SAME-P-HIGH-S second, got %s", reporter.issues[1].ID)
	}
	if reporter.issues[2].ID != "SAME-P-LOW-S" {
		t.Errorf("expected SAME-P-LOW-S third, got %s", reporter.issues[2].ID)
	}

	if reporter.issues[4].ID != "LOW-P" {
		t.Errorf("expected LOW-P fifth, got %s", reporter.issues[4].ID)
	}

	t.Run("PriorityValueDefault", func(t *testing.T) {
		if val := reporter.priorityValue("invalid"); val != 4 {
			t.Errorf("expected 4, got %d", val)
		}
	})

	t.Run("SeverityValueDefault", func(t *testing.T) {
		if val := reporter.severityValue("invalid"); val != 5 {
			t.Errorf("expected 5, got %d", val)
		}
		if val := reporter.severityValue(Critical); val != 0 {
			t.Errorf("expected 0, got %d", val)
		}
		if val := reporter.severityValue(High); val != 1 {
			t.Errorf("expected 1, got %d", val)
		}
		if val := reporter.severityValue(Medium); val != 2 {
			t.Errorf("expected 2, got %d", val)
		}
		if val := reporter.severityValue(Low); val != 3 {
			t.Errorf("expected 3, got %d", val)
		}
		if val := reporter.severityValue(Info); val != 4 {
			t.Errorf("expected 4, got %d", val)
		}
	})
}
