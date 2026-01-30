package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Les-El/chexum/internal/testutil"
)

// Property 12: Comprehensive Integration Testing
// **Validates: Requirements 5.1, 5.4**
//
// Reviewed: LONG-FUNCTION - Integration test with full analysis pipeline.
func TestIntegration_CompleteWorkflow(t *testing.T) {
	// Feature: checkpoint-remediation, Property 12: Comprehensive integration testing

	tmpDir, cleanup := testutil.TempDir(t)
	defer cleanup()

	// 1. Setup project structure
	testutil.CreateFile(t, tmpDir, "internal/config/config.go", "package config\n")
	testutil.CreateFile(t, tmpDir, "README.md", "# User Docs\n")
	testutil.CreateFile(t, tmpDir, "major_checkpoint/design.md", "# Design\n")
	testutil.GenerateMockGoFile(t, tmpDir, "main.go", true, true)

	// 2. Run analysis
	ctx := context.Background()
	engines := []AnalysisEngine{
		NewCodeAnalyzer(),
		NewDocAuditor(),
	}
	runner := NewRunner(engines)
	if err := runner.Run(ctx, tmpDir); err != nil {
		t.Fatalf("Runner.Run failed: %v", err)
	}

	issues := runner.GetIssues()
	if len(issues) == 0 {
		t.Error("Expected issues to be found in mock project")
	}

	// 3. Generate reports
	reporter := NewReporter()
	reporter.Aggregate(issues, nil)

	plan, err := reporter.GenerateRemediationPlan()
	if err != nil || plan == "" {
		t.Fatalf("Failed to generate remediation plan: %v", err)
	}

	// 4. Organize artifacts
	organizer := NewOrganizer(tmpDir)
	latestDir := filepath.Join(tmpDir, "active", "latest")
	os.MkdirAll(latestDir, 0755)
	os.WriteFile(filepath.Join(latestDir, "plan.md"), []byte(plan), 0644)

	if err := organizer.CreateSnapshot("integration_test"); err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	// 5. Cleanup
	cleanupMgr := NewCleanupManager(false)
	cleanupMgr.baseDir = tmpDir

	// Register a mock workspace for cleanup test
	ws, _ := NewWorkspace(false)
	ws.Root = filepath.Join(tmpDir, "mock-workspace")
	os.MkdirAll(ws.Root, 0755)
	os.WriteFile(filepath.Join(ws.Root, "test.tmp"), []byte("data"), 0644)
	cleanupMgr.RegisterWorkspace(ws)

	result, err := cleanupMgr.CleanupTemporaryFiles()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if result.DirsRemoved == 0 {
		t.Error("Expected at least one directory (workspace) to be removed")
	}

	if _, err := os.Stat(ws.Root); !os.IsNotExist(err) {
		t.Errorf("Workspace root %s still exists after cleanup", ws.Root)
	}

	t.Logf("Cleanup removed %d files and %d dirs", result.FilesRemoved, result.DirsRemoved)
}
