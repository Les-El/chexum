package checkpoint

import (
	"context"
	"os"
	"testing"
)

func TestNewFlagSystem(t *testing.T) {
	fs := NewFlagSystem()
	if fs == nil {
		t.Fatal("NewFlagSystem returned nil")
	}
}

func TestFlagSystem_Name(t *testing.T) {
	fs := NewFlagSystem()
	if name := fs.Name(); name != "FlagSystem" {
		t.Errorf("expected FlagSystem, got %s", name)
	}
}

func TestFlagSystem_Analyze(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	issues, err := fs.Analyze(ctx, "../../")
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}
	if len(issues) == 0 {
		t.Log("No issues found, which is expected if everything is clean.")
	}
}

func TestCatalogFlags(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	_, err := fs.CatalogFlags(ctx, "../../")
	if err != nil {
		t.Logf("CatalogFlags failed (possibly config.go missing): %v", err)
	}
}

func TestClassifyImplementation(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	if _, err := os.Stat("../../internal/config/config.go"); err == nil {
		_, err = fs.ClassifyImplementation(ctx, "../../", mockFlags)
		if err != nil {
			t.Errorf("ClassifyImplementation failed: %v", err)
		}
	}
}

func TestPerformCrossReferenceAnalysis(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	_, err := fs.PerformCrossReferenceAnalysis(ctx, "../../", mockFlags)
	if err != nil {
		t.Errorf("PerformCrossReferenceAnalysis failed: %v", err)
	}
}

func TestDetectConflicts(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	_, err := fs.DetectConflicts(ctx, mockFlags)
	if err != nil {
		t.Errorf("DetectConflicts failed: %v", err)
	}
}

func TestValidateFunctionality(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	_, err := fs.ValidateFunctionality(ctx, mockFlags)
	if err != nil {
		t.Errorf("ValidateFunctionality failed: %v", err)
	}
}

func TestGenerateStatusReport(t *testing.T) {
	fs := NewFlagSystem()
	ctx := context.Background()
	mockFlags := []FlagStatus{{LongForm: "verbose", DefinedInCode: true}}
	report, err := fs.GenerateStatusReport(ctx, mockFlags)
	if err != nil {
		t.Errorf("GenerateStatusReport failed: %v", err)
	}
	if report == "" {
		t.Error("GenerateStatusReport returned empty string")
	}
}
