package checkpoint

import (
	"context"
	"os"
	"testing"
)

func TestNewQualityEngine(t *testing.T) {
	engine := NewQualityEngine()
	if engine == nil {
		t.Fatal("NewQualityEngine returned nil")
	}
}

func TestQualityEngine_Name(t *testing.T) {
	engine := NewQualityEngine()
	if name := engine.Name(); name != "QualityEngine" {
		t.Errorf("expected QualityEngine, got %s", name)
	}
}

func TestQualityEngine_Analyze(t *testing.T) {
	engine := NewQualityEngine()
	ctx := context.Background()
	_, err := engine.Analyze(ctx, "../../")
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}
}

func TestCheckGoStandards(t *testing.T) {
	engine := NewQualityEngine()
	ctx := context.Background()

	content := "package test\nfunc LongFunc() {\n"
	for i := 0; i < 60; i++ {
		content += "\t_ = 1\n"
	}
	content += "}\n"

	tmpFile, err := os.CreateTemp("", "standards*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	issues, err := engine.CheckGoStandards(ctx, "../../")
	if err != nil {
		t.Errorf("CheckGoStandards failed: %v", err)
	}
	_ = issues
}

func TestEvaluateErrorHandling(t *testing.T) {
	engine := NewQualityEngine()
	ctx := context.Background()

	issues, err := engine.EvaluateErrorHandling(ctx, "../../")
	if err != nil {
		t.Errorf("EvaluateErrorHandling failed: %v", err)
	}
	_ = issues
}

func TestAnalyzePerformance(t *testing.T) {
	engine := NewQualityEngine()
	ctx := context.Background()

	issues, err := engine.AnalyzePerformance(ctx, "../../")
	if err != nil {
		t.Errorf("AnalyzePerformance failed: %v", err)
	}
	_ = issues
}

func TestAssessCLIDesign(t *testing.T) {
	engine := NewQualityEngine()
	ctx := context.Background()

	// In this repo, internal/config/config.go exists.
	issues, err := engine.AssessCLIDesign(ctx, "../../")
	if err != nil {
		t.Errorf("AssessCLIDesign failed: %v", err)
	}
	_ = issues
}
