package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Les-El/hashi/internal/checkpoint"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	fmt.Println("Starting Major Checkpoint Analysis...")

	cleanup := checkpoint.NewCleanupManager(true)
	checkInitialResources(cleanup)

	engines := registerEngines()

	issues, flags, err := runAnalysis(ctx, engines)
	if err != nil {
		handleRunFailure(cleanup, "analysis", err)
		return fmt.Errorf("running analysis: %w", err)
	}

	reports := generateReports(ctx, issues, flags)

	if err := saveReports(reports); err != nil {
		handleRunFailure(cleanup, "report generation", err)
		return fmt.Errorf("saving reports: %w", err)
	}

	fmt.Println("Analysis complete. Reports generated in major_checkpoint/ directory.")

	// Perform cleanup at the end
	fmt.Println("Performing post-analysis cleanup...")
	if err := cleanup.CleanupOnExit(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Cleanup failed: %v\n", err)
	}
	return nil
}

func checkInitialResources(cleanup *checkpoint.CleanupManager) {
	// Check tmpfs usage before starting
	if needsCleanup, usage := cleanup.CheckTmpfsUsage(75.0); needsCleanup {
		fmt.Printf("Warning: Tmpfs usage is %.1f%%. Consider running cleanup before analysis.\n", usage)
	}
}

func registerEngines() []checkpoint.AnalysisEngine {
	return []checkpoint.AnalysisEngine{
		checkpoint.NewCodeAnalyzer(),
		checkpoint.NewDependencyAnalyzer(),
		checkpoint.NewDocAuditor(),
		checkpoint.NewTestingBattery(),
		checkpoint.NewFlagSystem(),
		checkpoint.NewQualityEngine(),
		checkpoint.NewCIEngine(85.0),
	}
}

func handleRunFailure(cleanup *checkpoint.CleanupManager, phase string, err error) {
	fmt.Printf("Attempting cleanup after %s failure...\n", phase)
	if cleanupErr := cleanup.CleanupOnExit(); cleanupErr != nil {
		fmt.Fprintf(os.Stderr, "Cleanup also failed: %v\n", cleanupErr)
	}
}

func runAnalysis(ctx context.Context, engines []checkpoint.AnalysisEngine) ([]checkpoint.Issue, []checkpoint.FlagStatus, error) {
	runner := checkpoint.NewRunner(engines)

	fmt.Println("Running comprehensive project analysis...")
	if err := runner.Run(ctx, "."); err != nil {
		return nil, nil, err
	}

	issues := runner.GetIssues()

	// Special handling for flags as they return FlagStatus
	flagSystem := checkpoint.NewFlagSystem()
	flags, err := flagSystem.CatalogFlags(ctx, ".")
	if err == nil {
		flags, _ = flagSystem.ClassifyImplementation(ctx, ".", flags)
		flags, _ = flagSystem.PerformCrossReferenceAnalysis(ctx, ".", flags)
		flags, _ = flagSystem.DetectConflicts(ctx, flags)
		flags, _ = flagSystem.ValidateFunctionality(ctx, flags)
	}

	return issues, flags, nil
}

type analysisReports struct {
	plan       string
	dashboard  string
	guide      string
	flagReport string
	jsonReport string
	csvReport  string
}

func generateReports(ctx context.Context, issues []checkpoint.Issue, flags []checkpoint.FlagStatus) analysisReports {
	reporter := checkpoint.NewReporter()
	reporter.Aggregate(issues, flags)
	reporter.SortIssues()

	plan, _ := reporter.GenerateRemediationPlan()
	dashboard, _ := reporter.GenerateStatusDashboard()
	guide, _ := reporter.GenerateOnboardingGuide()
	jsonReport, _ := reporter.GenerateJSONReport()
	csvReport, _ := reporter.GenerateCSVReport()

	flagSystem := checkpoint.NewFlagSystem()
	flagReport, _ := flagSystem.GenerateStatusReport(ctx, flags)

	return analysisReports{
		plan:       plan,
		dashboard:  dashboard,
		guide:      guide,
		flagReport: flagReport,
		jsonReport: jsonReport,
		csvReport:  csvReport,
	}
}

func saveReports(r analysisReports) error {
	const rootDir = "major_checkpoint"
	latestDir := filepath.Join(rootDir, "active", "latest")
	if err := os.MkdirAll(latestDir, 0755); err != nil {
		return err
	}

	files := map[string]string{
		filepath.Join(latestDir, "findings_remediation_plan.md"):   r.plan,
		filepath.Join(latestDir, "findings_status_dashboard.md"):   r.dashboard,
		filepath.Join(latestDir, "findings_onboarding_guide.md"):   r.guide,
		filepath.Join(latestDir, "findings_flag_report.md"):        r.flagReport,
		filepath.Join(latestDir, "findings_remediation_plan.json"): r.jsonReport,
		filepath.Join(latestDir, "findings_remediation_plan.csv"):  r.csvReport,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	fmt.Println("Organizing checkpoint artifacts...")
	organizer := checkpoint.NewOrganizer(rootDir)
	if err := organizer.CreateSnapshot(""); err != nil {
		return fmt.Errorf("creating snapshot: %w", err)
	}
	if err := organizer.ArchiveOldSnapshots(5); err != nil {
		fmt.Printf("Warning: Archival failed: %v\n", err)
	}

	return nil
}
