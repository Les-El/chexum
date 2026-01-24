package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Les-El/hashi/internal/checkpoint"
)

func main() {
	ctx := context.Background()

	fmt.Println("Starting Major Checkpoint Analysis...")

	// Check tmpfs usage before starting
	cleanup := checkpoint.NewCleanupManager(true)
	needsCleanup, usage := cleanup.CheckTmpfsUsage(75.0)
	if needsCleanup {
		fmt.Printf("Warning: Tmpfs usage is %.1f%%. Consider running cleanup before analysis.\n", usage)
	}

	issues, flags, err := runAnalysis(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running analysis: %v\n", err)
		
		// Attempt cleanup on failure
		fmt.Println("Attempting cleanup after analysis failure...")
		if cleanupErr := cleanup.CleanupOnExit(); cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "Cleanup also failed: %v\n", cleanupErr)
		}
		os.Exit(1)
	}

	reports := generateReports(ctx, issues, flags)

	if err := saveReports(reports); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving reports: %v\n", err)
		
		// Attempt cleanup on failure
		fmt.Println("Attempting cleanup after report generation failure...")
		if cleanupErr := cleanup.CleanupOnExit(); cleanupErr != nil {
			fmt.Fprintf(os.Stderr, "Cleanup also failed: %v\n", cleanupErr)
		}
		os.Exit(1)
	}

	fmt.Println("Analysis complete. Reports generated in major_checkpoint/ directory.")
	
	// Perform cleanup at the end
	fmt.Println("Performing post-analysis cleanup...")
	if err := cleanup.CleanupOnExit(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Cleanup failed: %v\n", err)
	}
}

func runAnalysis(ctx context.Context) ([]checkpoint.Issue, []checkpoint.FlagStatus, error) {
	// Register all engines
	engines := []checkpoint.AnalysisEngine{
		checkpoint.NewCodeAnalyzer(),
		checkpoint.NewDependencyAnalyzer(),
		checkpoint.NewDocAuditor(),
		checkpoint.NewTestingBattery(),
		checkpoint.NewFlagSystem(),
		checkpoint.NewQualityEngine(),
	}

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
	plan        string
	dashboard   string
	guide       string
	flagReport  string
	jsonReport  string
	csvReport   string
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
		plan:        plan,
		dashboard:   dashboard,
		guide:       guide,
		flagReport:  flagReport,
		jsonReport:  jsonReport,
		csvReport:   csvReport,
	}
}

func saveReports(r analysisReports) error {
	const dir = "major_checkpoint"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	files := map[string]string{
		dir + "/findings_remediation_plan.md":   r.plan,
		dir + "/findings_status_dashboard.md":   r.dashboard,
		dir + "/findings_onboarding_guide.md":    r.guide,
		dir + "/findings_flag_report.md":        r.flagReport,
		dir + "/findings_remediation_plan.json": r.jsonReport,
		dir + "/findings_remediation_plan.csv":  r.csvReport,
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

	