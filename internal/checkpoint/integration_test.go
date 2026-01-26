package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Les-El/hashi/internal/testutil"
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
	testutil.CreateFile(t, tmpDir, "docs/user/README.md", "# User Docs\n")
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
	result, err := cleanupMgr.CleanupTemporaryFiles()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Since we didn't create any files matching cleanup patterns in tmpDir (other than the organizer files which are NOT in /tmp),
	// we don't expect much here unless we add some.
	t.Logf("Cleanup removed %d files and %d dirs", result.FilesRemoved, result.DirsRemoved)
}
