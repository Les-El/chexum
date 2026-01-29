package checkpoint

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Les-El/hashi/internal/config"
)

// FlagSystem implements FlagDocumentationSystem.
type FlagSystem struct {
	fset *token.FileSet
}

// NewFlagSystem initializes a new FlagSystem.
func NewFlagSystem() *FlagSystem {
	return &FlagSystem{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (f *FlagSystem) Name() string { return "FlagSystem" }

// Analyze performs a comprehensive analysis of the flag system and returns identified issues.
func (f *FlagSystem) Analyze(ctx context.Context, path string, ws *Workspace) ([]Issue, error) {
	var issues []Issue

	configPkg, _ := discoverPackageByName(path, "config")

	flags, err := f.CatalogFlags(ctx, path, ws)
	if err != nil {
		return nil, err
	}

	flags, _ = f.ClassifyImplementation(ctx, path, ws, flags)
	flags, _ = f.PerformCrossReferenceAnalysis(ctx, path, ws, flags)
	flags, _ = f.DetectConflicts(ctx, ws, flags)
	flags, _ = f.ValidateFunctionality(ctx, ws, flags)

	for _, flag := range flags {
		if flag.Status == PartiallyImplemented {
			issues = append(issues, Issue{
				ID:          "FLAG-PARTIAL-IMPLEMENTATION",
				Category:    Usability,
				Severity:    Medium,
				Title:       fmt.Sprintf("Flag '--%s' is partially implemented", flag.LongForm),
				Description: fmt.Sprintf("The flag '--%s' is defined but not fully integrated into the configuration system.", flag.LongForm),
				Location:    filepath.Join(path, configPkg, "cli.go"),
				Suggestion:  fmt.Sprintf("Complete the implementation of '--%s' in %s/cli.go.", flag.LongForm, configPkg),
				Effort:      MediumEffort,
				Priority:    P2,
			})
		}

		if flag.ActualBehavior == "Not found in --help" {
			issues = append(issues, Issue{
				ID:          "FLAG-MISSING-FROM-HELP-OUTPUT",
				Category:    Usability,
				Severity:    High,
				Title:       fmt.Sprintf("Flag '--%s' missing from CLI help output", flag.LongForm),
				Description: fmt.Sprintf("The flag '--%s' is defined in code but does not appear when running 'hashi --help'.", flag.LongForm),
				Location:    filepath.Join(path, configPkg, "cli.go"),
				Suggestion:  "Ensure the flag is correctly added to the flagset used by the CLI.",
				Effort:      Small,
				Priority:    P1,
			})
		}

		for _, conflict := range flag.ConflictDetails {
			severity := Medium
			if conflict.Severity == ConflictCritical || conflict.Severity == ConflictHigh {
				severity = High
			}

			issues = append(issues, Issue{
				ID:          fmt.Sprintf("FLAG-CONFLICT-%s", strings.ToUpper(string(conflict.Type))),
				Category:    Usability,
				Severity:    severity,
				Title:       fmt.Sprintf("Conflict detected for flag '--%s'", flag.LongForm),
				Description: conflict.Description,
				Location:    filepath.Join(path, configPkg, "cli.go"),
				Suggestion:  "Resolve the discrepancy between the flag sources.",
				Effort:      Small,
				Priority:    P1,
			})
		}
	}

	return issues, nil
}

// CatalogFlags parses the codebase to discover CLI flag definitions.
func (f *FlagSystem) CatalogFlags(ctx context.Context, rootPath string, ws *Workspace) ([]FlagStatus, error) {
	var statuses []FlagStatus

	configPkg, err := discoverPackageByName(rootPath, "config")
	if err != nil {
		return nil, err
	}

	files, err := discoverPackageFiles(rootPath, configPkg)
	if err != nil {
		return nil, err
	}

	for _, configPath := range files {
		node, err := parser.ParseFile(f.fset, configPath, nil, parser.ParseComments)
		if err != nil {
			continue
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if status, ok := f.parseFlagCall(call); ok {
				status.DefinedInCode = true
				statuses = append(statuses, status)
			}
			return true
		})
	}

	return statuses, nil
}

func (f *FlagSystem) parseFlagCall(call *ast.CallExpr) (FlagStatus, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return FlagStatus{}, false
	}

	// Look for calls like fs.BoolVarP, fs.StringVar, fs.StringSliceVarP, etc.
	if !strings.HasSuffix(sel.Sel.Name, "Var") && !strings.HasSuffix(sel.Sel.Name, "VarP") {
		return FlagStatus{}, false
	}

	status := FlagStatus{}
	isP := strings.HasSuffix(sel.Sel.Name, "P")

	// Args are usually: (&var, name, [short], default, usage)
	if len(call.Args) < 3 {
		return FlagStatus{}, false
	}

	// Long name is arg[1]
	if lit, ok := call.Args[1].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		status.LongForm = strings.Trim(lit.Value, "\"")
		status.Name = status.LongForm
	}

	if isP && len(call.Args) >= 4 {
		// Short name is arg[2]
		if lit, ok := call.Args[2].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			status.ShortForm = strings.Trim(lit.Value, "\"")
		}
		// Description is arg[4]
		if len(call.Args) >= 5 {
			if lit, ok := call.Args[4].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				status.Description = strings.Trim(lit.Value, "\"")
			}
		}
	} else if len(call.Args) >= 4 {
		// Description is arg[3]
		if lit, ok := call.Args[3].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			status.Description = strings.Trim(lit.Value, "\"")
		}
	}

	return status, status.Name != ""
}

// ClassifyImplementation analyzes the implementation status of each flag by tracing variable usage.
func (f *FlagSystem) ClassifyImplementation(ctx context.Context, rootPath string, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error) {
	// Read config files to see if flags are used
	configPkg, err := discoverPackageByName(rootPath, "config")
	if err != nil {
		return nil, err
	}

	files, err := discoverPackageFiles(rootPath, configPkg)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	for _, configPath := range files {
		content, err := os.ReadFile(configPath)
		if err == nil {
			sb.Write(content)
		}
	}
	code := sb.String()

	// Also check cmd/hashi/main.go for usage of the Config struct
	mainPath := filepath.Join(rootPath, "cmd/hashi/main.go")
	mainContent, _ := os.ReadFile(mainPath)
	mainCode := string(mainContent)

	for i := range flags {
		flag := &flags[i]

		// Static analysis: check for usage of flag in config parsing or in main
		// We look for .Name, .JSON, .Recursive etc. usage on a config object.
		// This is still a heuristic but better than just checking flagSet.Changed().

		// Map flag names to field names (simplified heuristic)
		fieldName := strings.ReplaceAll(strings.Title(strings.ReplaceAll(flag.LongForm, "-", " ")), " ", "")
		// Special cases
		if flag.LongForm == "json" {
			fieldName = "JSON"
		}
		if flag.LongForm == "jsonl" {
			fieldName = "JSONL"
		}

		isUsedInConfig := strings.Contains(code, "cfg."+fieldName) || strings.Contains(code, "flagSet.Changed(\""+flag.LongForm+"\")")
		isUsedInMain := strings.Contains(mainCode, "cfg."+fieldName)

		if isUsedInConfig && isUsedInMain {
			flag.Status = FullyImplemented
		} else if isUsedInConfig || isUsedInMain {
			flag.Status = PartiallyImplemented
		} else {
			flag.Status = PlannedNotImplemented
		}

		// Check for documentation in HelpText (usually in help.go now)
		if strings.Contains(code, "--"+flag.LongForm) {
			flag.Documentation = true
			flag.DefinedInHelp = true
		}
	}

	return flags, nil
}

func (f *FlagSystem) detectGhostFlags(docContent, planContent string, flags []FlagStatus, ws *Workspace) {
	// Extract potential flags from docs using regex
	// This is a simplified implementation
	// In a real system, we'd look for --[a-z-]+
}

// PerformCrossReferenceAnalysis compares flag definitions across multiple sources.
func (f *FlagSystem) PerformCrossReferenceAnalysis(ctx context.Context, rootPath string, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error) {
	// 1. Check user documentation
	userDocs := []string{
		filepath.Join(rootPath, "docs/user/README.md"),
		filepath.Join(rootPath, "docs/user/dry-run.md"),
		filepath.Join(rootPath, "docs/user/examples.md"),
		filepath.Join(rootPath, "docs/user/filtering.md"),
		filepath.Join(rootPath, "docs/user/incremental.md"),
		filepath.Join(rootPath, "docs/user/command-reference.md"),
	}

	// 2. Check planning documents
	planDocs := []string{
		filepath.Join(rootPath, "docs/checkpoint/checkpoint_design.md"),
		filepath.Join(rootPath, "docs/checkpoint/chekpoint_requirements.md"),
		filepath.Join(rootPath, "docs/remediation/audit_remediation_plan.md"),
		filepath.Join(rootPath, "docs/remediation/remediation_tasks.md"),
		filepath.Join(rootPath, "docs/dev/flag_conflicts.md"),
		filepath.Join(rootPath, "docs/design/new_conflict_resolution.md"),
	}

	docContent := f.readFilesCombined(userDocs)
	planContent := f.readFilesCombined(planDocs)

	for i := range flags {
		flag := &flags[i]
		if strings.Contains(docContent, "--"+flag.LongForm) {
			flag.DefinedInDocs = true
		}
		// Match flag name or field name in planning docs
		if strings.Contains(planContent, "--"+flag.LongForm) || strings.Contains(planContent, flag.LongForm) {
			flag.DefinedInPlanning = true
		}
	}

	// Identify Ghost Flags (planned but not in code)
	// We scan the planning content for --[a-z-]+ and add them if not already present
	ghostFlags := f.extractPotentialFlags(planContent)
	for _, ghost := range ghostFlags {
		found := false
		for _, f := range flags {
			if f.LongForm == ghost {
				found = true
				break
			}
		}
		if !found {
			flags = append(flags, FlagStatus{
				Name:              ghost,
				LongForm:          ghost,
				DefinedInPlanning: true,
				Status:            PlannedNotImplemented,
			})
		}
	}

	return flags, nil
}

func (f *FlagSystem) extractPotentialFlags(content string) []string {
	var results []string
	// Use regex to find all flags starting with -- followed by lowercase letters, numbers, or hyphens
	re := regexp.MustCompile(`--([a-z][a-z0-9-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			flag := match[1]
			if !seen[flag] {
				results = append(results, flag)
				seen[flag] = true
			}
		}
	}
	return results
}

func (f *FlagSystem) readFilesCombined(paths []string) string {
	var sb strings.Builder
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			sb.Write(data)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// DetectConflicts identifies discrepancies between documented and implemented behavior.
func (f *FlagSystem) DetectConflicts(ctx context.Context, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error) {
	for i := range flags {
		flag := &flags[i]

		// Orphaned Flag: Implemented but not documented in help or user docs
		if flag.DefinedInCode && !flag.DefinedInHelp && !flag.DefinedInDocs {
			flag.ConflictDetails = append(flag.ConflictDetails, FlagConflict{
				Type:        OrphanedFlag,
				Source1:     "code",
				Source2:     "documentation",
				Description: "Flag is implemented in code but missing from help text and user documentation.",
				Severity:    ConflictHigh,
			})
		}

		// Documentation Mismatch: Implemented but missing from user docs (even if in help)
		if flag.DefinedInCode && !flag.DefinedInDocs {
			found := false
			for _, c := range flag.ConflictDetails {
				if c.Type == OrphanedFlag {
					found = true
					break
				}
			}
			if !found {
				flag.ConflictDetails = append(flag.ConflictDetails, FlagConflict{
					Type:        DescriptionConflict,
					Source1:     "code",
					Source2:     "user_docs",
					Description: "Flag is implemented but missing from user-facing markdown documentation.",
					Severity:    ConflictMedium,
				})
			}
		}

		// Planning Mismatch: In planning but not in code
		if flag.DefinedInPlanning && !flag.DefinedInCode {
			flag.ConflictDetails = append(flag.ConflictDetails, FlagConflict{
				Type:        PlanningMismatch,
				Source1:     "planning",
				Source2:     "code",
				Description: "Flag mentioned in planning documents but not implemented in code.",
				Severity:    ConflictHigh,
			})
		}
	}
	return flags, nil
}

// ValidateFunctionality verifies that flags appear in hashi --help output.
func (f *FlagSystem) ValidateFunctionality(ctx context.Context, ws *Workspace, flags []FlagStatus) ([]FlagStatus, error) {
	// Instead of running the CLI, we call the internal HelpText function directly.
	// This is much faster and avoids dependency on 'go run'.
	helpOutput := config.HelpText()

	for i := range flags {
		flag := &flags[i]
		if flag.Status == PlannedNotImplemented {
			continue
		}

		if strings.Contains(helpOutput, "--"+flag.LongForm) {
			flag.ActualBehavior = "Present in HelpText()"
			flag.TestCoverage = true
		} else {
			flag.ActualBehavior = "Not found in HelpText()"
			flag.TestCoverage = false
		}
	}
	return flags, nil
}

// GenerateStatusReport produces a markdown report of flag statuses.
func (f *FlagSystem) GenerateStatusReport(ctx context.Context, ws *Workspace, flags []FlagStatus) (string, error) {
	var sb strings.Builder
	sb.WriteString("# CLI Flag Status and Conflict Report\n\n")
	sb.WriteString("| Flag | Status | Help | Docs | Plan | Conflicts |\n")
	sb.WriteString("|------|--------|------|------|------|-----------|\n")
	for _, flag := range flags {
		help := "❌"
		if flag.DefinedInHelp {
			help = "✅"
		}
		docs := "❌"
		if flag.DefinedInDocs {
			docs = "✅"
		}
		plan := "❌"
		if flag.DefinedInPlanning {
			plan = "✅"
		}

		conflictStr := "None"
		if len(flag.ConflictDetails) > 0 {
			var conflicts []string
			for _, c := range flag.ConflictDetails {
				conflicts = append(conflicts, string(c.Type))
			}
			conflictStr = strings.Join(conflicts, ", ")
		}

		sb.WriteString(fmt.Sprintf("| --%s | %s | %s | %s | %s | %s |\n",
			flag.LongForm, flag.Status, help, docs, plan, conflictStr))
	}
	return sb.String(), nil
}
