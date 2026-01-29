package checkpoint

import (
	"context"
	"testing"
)

func TestNewDependencyAnalyzer(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	if analyzer == nil {
		t.Fatal("NewDependencyAnalyzer returned nil")
	}
}

func TestDependencyAnalyzer_Name(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	if name := analyzer.Name(); name != "DependencyAnalyzer" {
		t.Errorf("expected DependencyAnalyzer, got %s", name)
	}
}

func TestDependencyAnalyzer_Analyze(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	// Analyze currently just calls AssessDependencies which looks for go.mod
	_, err := analyzer.Analyze(ctx, "../../", ws)
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}
}

func TestDependencyAnalyzer_AssessDependencies(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	// In the real environment, go.mod should exist.
	_, err := analyzer.AssessDependencies(ctx, "../../", ws)
	if err != nil {
		t.Errorf("AssessDependencies failed: %v", err)
	}
}

func TestDependencyAnalyzer_CheckVulnerabilities(t *testing.T) {
	analyzer := NewDependencyAnalyzer()
	ctx := context.Background()
	tmpDir := t.TempDir()

	// 1. Tool not found
	issues, err := analyzer.checkVulnerabilities(ctx, tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues when tool not found, got %d", len(issues))
	}

	// 2. Tool found but no vulnerabilities
	// (Hard to test without mocking safeCommand or PATH)
}

func TestDependencyAnalyzer_Helpers(t *testing.T) {
	d := NewDependencyAnalyzer()
	tests := []struct {
		current string
		update  string
		want    bool
	}{
		{"v1.0.0", "v1.2.0", false},
		{"v1.0.0", "v2.0.0", false},
		{"v1.0.0", "v3.0.0", true},
		{"v1.0.0", "v1.0.1", false},
		{"invalid", "v1.0.0", false},
	}

	for _, tt := range tests {
		if got := d.isMajorVersionBehind(tt.current, tt.update); got != tt.want {
			t.Errorf("isMajorVersionBehind(%s, %s) = %v, want %v", tt.current, tt.update, got, tt.want)
		}
	}
}
