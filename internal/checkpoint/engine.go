package checkpoint

import (
	"context"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// discoverPackageFiles returns the absolute paths of all Go files in a package relative to root.
func discoverPackageFiles(root, pkgPath string) ([]string, error) {
	ctx := build.Default
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	pkg, err := ctx.ImportDir(filepath.Join(absRoot, pkgPath), build.ImportComment)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, f := range pkg.GoFiles {
		files = append(files, filepath.Join(absRoot, pkgPath, f))
	}
	return files, nil
}

// discoverCorePackages scans the internal directory to find all available packages.
func discoverCorePackages(root string) ([]string, error) {
	var packages []string
	internalDir := filepath.Join(root, "internal")

	if _, err := os.Stat(internalDir); os.IsNotExist(err) {
		return nil, nil
	}

	err := filepath.Walk(internalDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != internalDir {
			// Get relative path from root
			rel, err := filepath.Rel(root, path)
			if err == nil {
				packages = append(packages, rel)
			}
		}
		return nil
	})

	return packages, err
}

// discoverPackageByName finds a package path by its suffix (e.g., "config").
func discoverPackageByName(root, name string) (string, error) {
	pkgs, err := discoverCorePackages(root)
	if err != nil {
		return "", err
	}
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg, "/"+name) || pkg == "internal/"+name {
			return pkg, nil
		}
	}
	return "", fmt.Errorf("package %s not found", name)
}

// AnalysisEngine is the base interface for all analysis components.
type AnalysisEngine interface {
	Name() string
	Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
}

// CodebaseAnalyzer examines Go source code.
type CodebaseAnalyzer interface {
	AnalysisEngine
	AnalyzePackages(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	CheckSecurity(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	AssessDependencies(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	IdentifyTechnicalDebt(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
}

// DocumentationAuditor validates project documentation.
type DocumentationAuditor interface {
	AnalysisEngine
	AuditGoDocumentation(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	ValidateREADME(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	CheckArchitecturalDocs(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	VerifyExamples(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
}

// TestingBatteryManager manages the testing suite.
type TestingBatteryManager interface {
	AnalysisEngine
	CreateUnitTests(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	BuildIntegrationTests(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	ImplementPropertyTests(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	CreateBenchmarks(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	CheckTestReliability(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	IdentifyLowCoverage(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
}

// FlagDocumentationSystem catalogs and documents CLI flags.
type FlagDocumentationSystem interface {
	AnalysisEngine
	CatalogFlags(ctx context.Context, path string, ws *Workspace) ([]FlagStatus, error)
	ClassifyImplementation(ctx context.Context, path string, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error)
	PerformCrossReferenceAnalysis(ctx context.Context, path string, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error)
	DetectConflicts(ctx context.Context, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error)
	ValidateFunctionality(ctx context.Context, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error)
	GenerateStatusReport(ctx context.Context, ws *Workspace, flags []FlagStatus) (string, error)
}

// QualityAssessmentEngine evaluates code quality and standards.
type QualityAssessmentEngine interface {
	AnalysisEngine
	CheckGoStandards(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	AssessCLIDesign(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	EvaluateErrorHandling(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
	AnalyzePerformance(ctx context.Context, path string, ws *Workspace) ([]Issue, error)
}

// SynthesisEngine aggregates findings into actionable reports.
type SynthesisEngine interface {
	Aggregate(issues []Issue, flagStatuses []FlagStatus) error
	GenerateRemediationPlan() (string, error)
	GenerateStatusDashboard() (string, error)
	GenerateJSONReport() (string, error)
	GenerateCSVReport() (string, error)
}
