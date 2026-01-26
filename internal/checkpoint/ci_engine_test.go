package checkpoint

import (
	"context"
	"os"
	"testing"
)

func TestNewCIEngine(t *testing.T) {
	engine := NewCIEngine(85.0)

	if engine.Name() != "CIEngine" {
		t.Errorf("expected CIEngine, got %s", engine.Name())
	}
}

func TestCIEngine_Name(t *testing.T) {
	engine := NewCIEngine(85.0)
	if engine.Name() != "CIEngine" {
		t.Errorf("expected CIEngine, got %s", engine.Name())
	}
}

func TestCIEngine_Analyze(t *testing.T) {
	ctx := context.Background()
	engine := NewCIEngine(85.0)

	t.Run("SkipAnalysis", func(t *testing.T) {
		os.Setenv("SKIP_CI_ANALYSIS", "true")
		defer os.Unsetenv("SKIP_CI_ANALYSIS")

		issues, err := engine.Analyze(ctx, ".")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("expected 0 issues, got %d", len(issues))
		}
	})

	t.Run("RealAnalysis", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping in short mode")
		}
		// Run on internal/color which is fast and has no dependencies on checkpoint
		issues, err := engine.Analyze(ctx, "../color")
		if err != nil {
			t.Fatalf("Analyze failed: %v", err)
		}
		t.Logf("Found %d issues", len(issues))
	})
}

func TestCIEngine_VerifyTestCompleteness(t *testing.T) {
	ctx := context.Background()
	engine := NewCIEngine(85.0)
	if !engine.VerifyTestCompleteness(ctx, ".") {
		t.Error("expected true")
	}
}
