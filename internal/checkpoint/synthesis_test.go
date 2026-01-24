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

func TestSortIssues(t *testing.T) {
	reporter := NewReporter()
	issues := []Issue{
		{ID: "LOW-P", Priority: P3},
		{ID: "HIGH-P", Priority: P0},
		{ID: "MED-P", Priority: P2},
	}
	reporter.Aggregate(issues, nil)
	reporter.SortIssues()

	if reporter.issues[0].ID != "HIGH-P" {
		t.Errorf("expected HIGH-P first, got %s", reporter.issues[0].ID)
	}
	if reporter.issues[2].ID != "LOW-P" {
		t.Errorf("expected LOW-P last, got %s", reporter.issues[2].ID)
	}
}