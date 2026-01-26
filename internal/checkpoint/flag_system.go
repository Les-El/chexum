package checkpoint

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
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
func (f *FlagSystem) Analyze(ctx context.Context, path string) ([]Issue, error) {
	var issues []Issue

	flags, err := f.CatalogFlags(ctx, path)
	if err != nil {
		return nil, err
	}

	flags, _ = f.ClassifyImplementation(ctx, path, flags)
	flags, _ = f.PerformCrossReferenceAnalysis(ctx, path, flags)
	flags, _ = f.DetectConflicts(ctx, flags)

	for _, flag := range flags {
		if flag.Status == PartiallyImplemented {
			issues = append(issues, Issue{
				ID:          "FLAG-PARTIAL-IMPLEMENTATION",
				Category:    Usability,
				Severity:    Medium,
				Title:       fmt.Sprintf("Flag '--%s' is partially implemented", flag.LongForm),
				Description: fmt.Sprintf("The flag '--%s' is defined but not fully integrated into the configuration system.", flag.LongForm),
				Location:    filepath.Join(path, "internal/config/config.go"),
				Suggestion:  fmt.Sprintf("Complete the implementation of '--%s' in internal/config/config.go.", flag.LongForm),
				Effort:      MediumEffort,
				Priority:    P2,
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
				Location:    filepath.Join(path, "internal/config/config.go"),
				Suggestion:  "Resolve the discrepancy between the flag sources.",
				Effort:      Small,
				Priority:    P1,
			})
		}
	}

	return issues, nil
}

// CatalogFlags parses the codebase to discover CLI flag definitions.
func (f *FlagSystem) CatalogFlags(ctx context.Context, rootPath string) ([]FlagStatus, error) {
	var statuses []FlagStatus

	configPath := filepath.Join(rootPath, "internal/config/config.go")
	node, err := parser.ParseFile(f.fset, configPath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if status, ok := f.parseFlagCall(call); ok {
			statuses = append(statuses, status)
		}
		return true
	})

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

// ClassifyImplementation analyzes the implementation status of each flag.
func (f *FlagSystem) ClassifyImplementation(ctx context.Context, rootPath string, flags []FlagStatus) ([]FlagStatus, error) {
	// Read config.go again to see if flags are used in ApplyConfigFile and ApplyEnvConfig
	configPath := filepath.Join(rootPath, "internal/config/config.go")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	code := string(content)

	for i := range flags {
		flag := &flags[i]
		inConfigFile := strings.Contains(code, "flagSet.Changed(\""+flag.LongForm+"\")")

		// This is a heuristic.
		if inConfigFile {
			flag.Status = FullyImplemented
		} else {
			flag.Status = PartiallyImplemented
		}

		// Check for documentation in HelpText
		if strings.Contains(code, "--"+flag.LongForm) {
			flag.Documentation = true
			flag.DefinedInHelp = true
		}
		flag.DefinedInCode = true
	}

	return flags, nil
}

// PerformCrossReferenceAnalysis compares flag definitions across multiple sources.
func (f *FlagSystem) PerformCrossReferenceAnalysis(ctx context.Context, rootPath string, flags []FlagStatus) ([]FlagStatus, error) {
	// 1. Check user documentation (docs/user/README.md or similar)
	userDocPath := filepath.Join(rootPath, "docs/user/README.md")
	docData, _ := os.ReadFile(userDocPath)
	docContent := string(docData)

	// 2. Check planning documents (major_checkpoint/design.md or requirements.md)
	planPath := filepath.Join(rootPath, "major_checkpoint/design.md")
	planData, _ := os.ReadFile(planPath)
	planContent := string(planData)

	for i := range flags {
		flag := &flags[i]
		if strings.Contains(docContent, "--"+flag.LongForm) {
			flag.DefinedInDocs = true
		}
		if strings.Contains(planContent, flag.LongForm) {
			flag.DefinedInPlanning = true
		}
	}

	// Identify Ghost Flags (documented but not in code)
	// This would require parsing docs for ALL mentioned flags and comparing with 'flags' list.
	// For now, let's just implement the conflict detection logic on existing flags.

	return flags, nil
}

// DetectConflicts identifies discrepancies between documented and implemented behavior.
func (f *FlagSystem) DetectConflicts(ctx context.Context, flags []FlagStatus) ([]FlagStatus, error) {
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
			flag.Status = "orphaned_implemented"
		}

		// Planning Mismatch: In planning but not in code
		if flag.DefinedInPlanning && !flag.DefinedInCode {
			flag.ConflictDetails = append(flag.ConflictDetails, FlagConflict{
				Type:        PlanningMismatch,
				Source1:     "planning",
				Source2:     "code",
				Description: "Flag mentioned in planning documents but not implemented in code.",
				Severity:    ConflictMedium,
			})
		}
	}
	return flags, nil
}

// ValidateFunctionality verifies that flags work as expected (placeholder).
func (f *FlagSystem) ValidateFunctionality(ctx context.Context, flags []FlagStatus) ([]FlagStatus, error) {
	// In a real system, we'd run 'hashi --help' or similar.
	for i := range flags {
		flags[i].ActualBehavior = "Matches expected" // Placeholder
		flags[i].TestCoverage = true                 // Placeholder
	}
	return flags, nil
}

// GenerateStatusReport produces a markdown report of flag statuses.
func (f *FlagSystem) GenerateStatusReport(ctx context.Context, flags []FlagStatus) (string, error) {
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
