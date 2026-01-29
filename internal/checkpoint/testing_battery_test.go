package checkpoint

import (
	"context"
	"go/ast"
	"os"
	"path/filepath"
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
	ws, _ := NewWorkspace(true)
	_, err := tb.Analyze(ctx, "../../", ws)
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
	ws, _ := NewWorkspace(true)

	_, err := tb.CheckTestReliability(ctx, "../../", ws)
	if err != nil {
		t.Logf("CheckTestReliability failed: %v", err)
	}
}

func TestIdentifyLowCoverage(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	_, err := tb.IdentifyLowCoverage(ctx, "../../", ws)
	if err != nil {
		t.Logf("IdentifyLowCoverage failed: %v", err)
	}
}

func TestCreateUnitTests(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	t.Run("Basic", func(t *testing.T) {
		_, err := tb.CreateUnitTests(ctx, "../../", ws)
		if err != nil {
			t.Logf("CreateUnitTests failed: %v", err)
		}
	})

	t.Run("NamingConventions", func(t *testing.T) {
		tmpDir := t.TempDir()
		code := `package test
type T struct{}
func (t *T) Method() {}
func (t T) ValueMethod() {}
func Exported() {}
`
		os.WriteFile(filepath.Join(tmpDir, "file.go"), []byte(code), 0644)

		testCode := `package test
import "testing"
func TestT_Method(t *testing.T) {}
func TestT_ValueMethod(t *testing.T) {}
func TestExported(t *testing.T) {}
`
		os.WriteFile(filepath.Join(tmpDir, "file_test.go"), []byte(testCode), 0644)

		issues, err := tb.CreateUnitTests(ctx, tmpDir, ws)
		if err != nil {
			t.Fatalf("CreateUnitTests failed: %v", err)
		}
		if len(issues) != 0 {
			t.Errorf("Expected 0 issues, got %d: %v", len(issues), issues)
		}
	})
}

func TestBuildIntegrationTests(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	_, err := tb.BuildIntegrationTests(ctx, "../../", ws)
	if err != nil {
		t.Logf("BuildIntegrationTests failed: %v", err)
	}
}

func TestIsExportedType(t *testing.T) {
	tb := NewTestingBattery()
	// Create mock AST types for testing
	if !tb.isExportedType(&ast.Ident{Name: "Exported"}) {
		t.Error("expected true for Exported")
	}
	if tb.isExportedType(&ast.Ident{Name: "unexported"}) {
		t.Error("expected false for unexported")
	}
	if !tb.isExportedType(&ast.StarExpr{X: &ast.Ident{Name: "Exported"}}) {
		t.Error("expected true for *Exported")
	}
}

func TestImplementPropertyTests(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	_, err := tb.ImplementPropertyTests(ctx, "../../", ws)
	if err != nil {
		t.Logf("ImplementPropertyTests failed: %v", err)
	}
}

func TestCreateBenchmarks(t *testing.T) {
	tb := NewTestingBattery()
	ctx := context.Background()
	ws, _ := NewWorkspace(true)

	_, err := tb.CreateBenchmarks(ctx, "../../", ws)
	if err != nil {
		t.Logf("CreateBenchmarks failed: %v", err)
	}
}
