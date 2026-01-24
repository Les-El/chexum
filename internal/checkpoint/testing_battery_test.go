package checkpoint

import (
	"context"
	"testing"
)

func TestNewTestingBattery(t *testing.T) {
	tb := NewTestingBattery()
	if tb == nil {
		t.Fatal("NewTestingBattery returned nil")
	}
}

func TestTestingBattery_Name(t *testing.T) {
	tb := NewTestingBattery()
	if name := tb.Name(); name != "TestingBattery" {
		t.Errorf("expected TestingBattery, got %s", name)
	}
}

func TestTestingBattery_Analyze(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode as it runs go test")
	}

	tb := NewTestingBattery()
	ctx := context.Background()
	_, err := tb.Analyze(ctx, "../../")
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}
}

func TestCheckTestReliability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tb := NewTestingBattery()
	ctx := context.Background()

	_, err := tb.CheckTestReliability(ctx, "../../")
	if err != nil {
		t.Logf("CheckTestReliability failed: %v", err)
	}
}

func TestIdentifyLowCoverage(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()

	_, err := tb.IdentifyLowCoverage(ctx, "../../")
	if err != nil {
		t.Logf("IdentifyLowCoverage failed: %v", err)
	}
}

func TestCreateUnitTests(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()

	_, err := tb.CreateUnitTests(ctx, "../../")
	if err != nil {
		t.Logf("CreateUnitTests failed: %v", err)
	}
}

func TestBuildIntegrationTests(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()

	_, err := tb.BuildIntegrationTests(ctx, "../../")
	if err != nil {
		t.Logf("BuildIntegrationTests failed: %v", err)
	}
}

func TestImplementPropertyTests(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()

	_, err := tb.ImplementPropertyTests(ctx, "../../")
	if err != nil {
		t.Logf("ImplementPropertyTests failed: %v", err)
	}
}

func TestCreateBenchmarks(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()

	_, err := tb.CreateBenchmarks(ctx, "../../")
	if err != nil {
		t.Logf("CreateBenchmarks failed: %v", err)
	}
}
