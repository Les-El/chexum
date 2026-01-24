package checkpoint

import "context"

// AnalysisEngine is the base interface for all analysis components.
type AnalysisEngine interface {
	Name() string
	Analyze(ctx context.Context, path string) ([]Issue, error)
}

// CodebaseAnalyzer examines Go source code.
type CodebaseAnalyzer interface {
	AnalysisEngine
	AnalyzePackages(ctx context.Context, path string) ([]Issue, error)
	CheckSecurity(ctx context.Context, path string) ([]Issue, error)
	AssessDependencies(ctx context.Context, path string) ([]Issue, error)
	IdentifyTechnicalDebt(ctx context.Context, path string) ([]Issue, error)
}

// DocumentationAuditor validates project documentation.
type DocumentationAuditor interface {
	AnalysisEngine
	AuditGoDocumentation(ctx context.Context, path string) ([]Issue, error)
	ValidateREADME(ctx context.Context, path string) ([]Issue, error)
	CheckArchitecturalDocs(ctx context.Context, path string) ([]Issue, error)
	VerifyExamples(ctx context.Context, path string) ([]Issue, error)
}

// TestingBatteryManager manages the testing suite.
type TestingBatteryManager interface {
	AnalysisEngine
	CreateUnitTests(ctx context.Context, path string) ([]Issue, error)
	BuildIntegrationTests(ctx context.Context, path string) ([]Issue, error)
	ImplementPropertyTests(ctx context.Context, path string) ([]Issue, error)
	CreateBenchmarks(ctx context.Context, path string) ([]Issue, error)
	CheckTestReliability(ctx context.Context, path string) ([]Issue, error)
	IdentifyLowCoverage(ctx context.Context, path string) ([]Issue, error)
}

// FlagDocumentationSystem catalogs and documents CLI flags.
type FlagDocumentationSystem interface {
	AnalysisEngine
	CatalogFlags(ctx context.Context, path string) ([]FlagStatus, error)
	ClassifyImplementation(ctx context.Context, path string, flags []FlagStatus) ([]FlagStatus, error)
	PerformCrossReferenceAnalysis(ctx context.Context, path string, flags []FlagStatus) ([]FlagStatus, error)
	DetectConflicts(ctx context.Context, flags []FlagStatus) ([]FlagStatus, error)
	ValidateFunctionality(ctx context.Context, flags []FlagStatus) ([]FlagStatus, error)
	GenerateStatusReport(ctx context.Context, flags []FlagStatus) (string, error)
}

// QualityAssessmentEngine evaluates code quality and standards.
type QualityAssessmentEngine interface {
	AnalysisEngine
	CheckGoStandards(ctx context.Context, path string) ([]Issue, error)
	AssessCLIDesign(ctx context.Context, path string) ([]Issue, error)
	EvaluateErrorHandling(ctx context.Context, path string) ([]Issue, error)
	AnalyzePerformance(ctx context.Context, path string) ([]Issue, error)
}

// SynthesisEngine aggregates findings into actionable reports.
type SynthesisEngine interface {
	Aggregate(issues []Issue, flagStatuses []FlagStatus) error
	GenerateRemediationPlan() (string, error)
	GenerateStatusDashboard() (string, error)
	GenerateJSONReport() (string, error)
	GenerateCSVReport() (string, error)
}
