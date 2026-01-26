package checkpoint

// IssueCategory represents the type of issue identified.
type IssueCategory string

const (
	CodeQuality   IssueCategory = "code_quality"
	Documentation IssueCategory = "documentation"
	Testing       IssueCategory = "testing"
	Security      IssueCategory = "security"
	Performance   IssueCategory = "performance"
	Usability     IssueCategory = "usability"
)

// Severity represents the impact of an issue.
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
	Info     Severity = "info"
)

// EffortEstimate represents the estimated effort to fix an issue.
type EffortEstimate string

const (
	XSmall       EffortEstimate = "xs"
	Small        EffortEstimate = "small"
	MediumEffort EffortEstimate = "medium"
	Large        EffortEstimate = "large"
	XLarge       EffortEstimate = "xl"
)

// Priority represents the priority of fixing an issue.
type Priority string

const (
	P0 Priority = "p0"
	P1 Priority = "p1"
	P2 Priority = "p2"
	P3 Priority = "p3"
)

// IssueStatus represents the current state of an issue.
type IssueStatus string

const (
	Pending    IssueStatus = "pending"
	InProgress IssueStatus = "in_progress"
	Resolved   IssueStatus = "resolved"
	WontFix    IssueStatus = "wont_fix"
)

// Issue represents a single finding from an analysis engine.
type Issue struct {
	Status      IssueStatus    `json:"status"`
	ID          string         `json:"id"`
	Category    IssueCategory  `json:"category"`
	Severity    Severity       `json:"severity"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Location    string         `json:"location"`
	Suggestion  string         `json:"suggestion"`
	Effort      EffortEstimate `json:"effort"`
	Priority    Priority       `json:"priority"`
	Reviewed    bool           `json:"reviewed"`
	ReviewNote  string         `json:"review_note"`
}

// ImplementationStatus represents the state of a flag's implementation.
type ImplementationStatus string

const (
	FullyImplemented      ImplementationStatus = "fully_implemented"
	PartiallyImplemented  ImplementationStatus = "partially_implemented"
	NeedsRepair           ImplementationStatus = "needs_repair"
	NeedsRefactoring      ImplementationStatus = "needs_refactoring"
	PlannedNotImplemented ImplementationStatus = "planned_not_implemented"
	Deprecated            ImplementationStatus = "deprecated"
)

// FlagStatus records the state of a CLI flag.
type FlagStatus struct {
	Name             string               `json:"name"`
	ShortForm        string               `json:"short_form"`
	LongForm         string               `json:"long_form"`
	Status           ImplementationStatus `json:"status"`
	Description      string               `json:"description"`
	ExpectedBehavior string               `json:"expected_behavior"`
	ActualBehavior   string               `json:"actual_behavior"`
	TestCoverage     bool                 `json:"test_coverage"`
	Documentation    bool                 `json:"documentation"`
	Issues           []string             `json:"issues"`

	// Cross-reference tracking
	DefinedInCode     bool           `json:"defined_in_code"`
	DefinedInHelp     bool           `json:"defined_in_help"`
	DefinedInDocs     bool           `json:"defined_in_docs"`
	DefinedInPlanning bool           `json:"defined_in_planning"`
	ExamplesWork      bool           `json:"examples_work"`
	ConflictDetails   []FlagConflict `json:"conflict_details"`
}

// FlagConflict represents a discrepancy between flag definitions across sources.
type FlagConflict struct {
	Type        ConflictType     `json:"type"`
	Source1     string           `json:"source1"` // e.g., "code", "help_text", "user_docs"
	Source2     string           `json:"source2"`
	Description string           `json:"description"`
	Severity    ConflictSeverity `json:"severity"`
}

// ConflictType defines the category of flag conflict.
type ConflictType string

const (
	BehaviorMismatch    ConflictType = "behavior_mismatch"
	DescriptionConflict ConflictType = "description_conflict"
	ExampleFailure      ConflictType = "example_failure"
	OrphanedFlag        ConflictType = "orphaned_flag"
	GhostFlag           ConflictType = "ghost_flag"
	PlanningMismatch    ConflictType = "planning_mismatch"
)

// ConflictSeverity defines the impact of a conflict.
type ConflictSeverity string

const (
	ConflictCritical ConflictSeverity = "critical"
	ConflictHigh     ConflictSeverity = "high"
	ConflictMedium   ConflictSeverity = "medium"
	ConflictLow      ConflictSeverity = "low"
)

// CoverageReport provides details on test coverage for a package.
type CoverageReport struct {
	Package         string     `json:"package"`
	CurrentCoverage float64    `json:"current_coverage"`
	TargetCoverage  float64    `json:"target_coverage"`
	MissingTests    []string   `json:"missing_tests"`
	ExistingTests   []TestInfo `json:"existing_tests"`
	Recommendations []string   `json:"recommendations"`
}

// TestType represents the kind of test.
type TestType string

const (
	UnitTest      TestType = "unit"
	Integration   TestType = "integration"
	PropertyBased TestType = "property"
	Benchmark     TestType = "benchmark"
)

// TestQuality represents an assessment of test quality.
type TestQuality string

const (
	Excellent TestQuality = "excellent"
	Good      TestQuality = "good"
	Fair      TestQuality = "fair"
	Poor      TestQuality = "poor"
)

// TestInfo provides details about a specific test.
type TestInfo struct {
	Name     string      `json:"name"`
	Type     TestType    `json:"type"`
	Coverage []string    `json:"coverage"`
	Quality  TestQuality `json:"quality"`
}
