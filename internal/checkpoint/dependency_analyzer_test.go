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

	// Analyze currently just calls AssessDependencies which looks for go.mod

	_, err := analyzer.Analyze(ctx, "../../")

	if err != nil {

		t.Errorf("Analyze failed: %v", err)

	}

}

func TestDependencyAnalyzer_AssessDependencies(t *testing.T) {

	analyzer := NewDependencyAnalyzer()

	ctx := context.Background()

	// In the real environment, go.mod should exist.

	_, err := analyzer.AssessDependencies(ctx, "../../")

	if err != nil {

		t.Errorf("AssessDependencies failed: %v", err)

	}

}
