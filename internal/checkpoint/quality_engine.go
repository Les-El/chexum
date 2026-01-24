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

// QualityEngine implements QualityAssessmentEngine.
type QualityEngine struct {
	fset *token.FileSet
}

// NewQualityEngine initializes a new QualityEngine.
func NewQualityEngine() *QualityEngine {
	return &QualityEngine{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (q *QualityEngine) Name() string { return "QualityEngine" }

// Analyze performs comprehensive quality checks on the codebase.
func (q *QualityEngine) Analyze(ctx context.Context, path string) ([]Issue, error) {
	var allIssues []Issue

	standards, _ := q.CheckGoStandards(ctx, path)
	allIssues = append(allIssues, standards...)

	design, _ := q.AssessCLIDesign(ctx, path)
	allIssues = append(allIssues, design...)

	errors, _ := q.EvaluateErrorHandling(ctx, path)
	allIssues = append(allIssues, errors...)

	performance, _ := q.AnalyzePerformance(ctx, path)
	allIssues = append(allIssues, performance...)

	return allIssues, nil
}

// CheckGoStandards validates adherence to Go coding standards (e.g., function length).
func (q *QualityEngine) CheckGoStandards(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		f, err := parser.ParseFile(q.fset, path, nil, 0)
		if err != nil {
			return nil
		}

		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Metric: Function length
			start := q.fset.Position(fn.Pos()).Line
			end := q.fset.Position(fn.End()).Line
			length := end - start
			if length > 50 {
				issues = append(issues, Issue{
					ID:          "LONG-FUNCTION",
					Category:    CodeQuality,
					Severity:    Low,
					Title:       "Function is too long",
					Description: fmt.Sprintf("Function '%s' is %d lines long.", fn.Name.Name, length),
					Location:    fmt.Sprintf("%s:%d", path, start),
					Suggestion:  "Consider refactoring into smaller functions.",
					Effort:      MediumEffort,
					Priority:    P3,
				})
			}
			return true
		})
		return nil
	})

	return issues, err
}

// AssessCLIDesign reviews CLI flag definitions for best practices.
func (q *QualityEngine) AssessCLIDesign(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	// Focus on internal/config/config.go where flags are defined
	configPath := filepath.Join(rootPath, "internal/config/config.go")
	f, err := parser.ParseFile(q.fset, configPath, nil, 0)
	if err != nil {
		return nil, nil
	}

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if strings.HasSuffix(sel.Sel.Name, "Var") || strings.HasSuffix(sel.Sel.Name, "VarP") {
			// Usage/description is usually the last argument
			if len(call.Args) > 0 {
				lastArg := call.Args[len(call.Args)-1]
				if lit, ok := lastArg.(*ast.BasicLit); ok && lit.Value == "\"\"" {
					issues = append(issues, Issue{
						ID:          "CLI-DESIGN-MISSING-USAGE",
						Category:    Usability,
						Severity:    Medium,
						Title:       "Missing CLI flag description",
						Description: "A CLI flag is defined without a usage description.",
						Location:    q.fset.Position(call.Pos()).String(),
						Suggestion:  "Add a helpful description to the flag definition.",
						Effort:      Small,
						Priority:    P2,
					})
				}
			}
		}
		return true
	})

	return issues, nil
}

// EvaluateErrorHandling detects risky error handling patterns like panic usage.
func (q *QualityEngine) EvaluateErrorHandling(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		f, err := parser.ParseFile(q.fset, path, nil, 0)
		if err != nil {
			return nil
		}

		ast.Inspect(f, func(n ast.Node) bool {
			// Look for panic calls
			if call, ok := n.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "panic" {
					issues = append(issues, Issue{
						ID:          "USE-OF-PANIC",
						Category:    CodeQuality,
						Severity:    Medium,
						Title:       "Usage of panic detected",
						Description: "The 'panic' function is used instead of returning an error.",
						Location:    q.fset.Position(call.Pos()).String(),
						Suggestion:  "Use error return values instead of panic for better robustness.",
						Effort:      MediumEffort,
						Priority:    P2,
					})
				}
			}
			return true
		})
		return nil
	})

	return issues, err
}

// AnalyzePerformance scans for potential performance bottlenecks.
func (q *QualityEngine) AnalyzePerformance(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		f, err := parser.ParseFile(q.fset, path, nil, 0)
		if err != nil {
			return nil
		}

		ast.Inspect(f, func(n ast.Node) bool {
			if forStmt, ok := n.(*ast.ForStmt); ok {
				issues = append(issues, q.findNestedLoops(forStmt)...)
			}
			return true
		})
		return nil
	})

	return issues, err
}

func (q *QualityEngine) findNestedLoops(forStmt *ast.ForStmt) []Issue {
	var issues []Issue
	ast.Inspect(forStmt.Body, func(inner ast.Node) bool {
		switch x := inner.(type) {
		case *ast.ForStmt:
			issues = append(issues, Issue{
				ID:          "NESTED-LOOP",
				Category:    Performance,
				Severity:    Low,
				Title:       "Nested loop detected",
				Description: "A nested loop may indicate O(n^2) performance complexity.",
				Location:    q.fset.Position(x.Pos()).String(),
				Suggestion:  "Evaluate if a more efficient algorithm (e.g., using a map) is possible.",
				Effort:      Large,
				Priority:    P3,
			})
			return false // Don't report loops nested further as separate issues for now
		case *ast.RangeStmt:
			issues = append(issues, Issue{
				ID:          "NESTED-LOOP",
				Category:    Performance,
				Severity:    Low,
				Title:       "Nested loop detected",
				Description: "A nested range loop may indicate O(n^2) performance complexity.",
				Location:    q.fset.Position(x.Pos()).String(),
				Suggestion:  "Evaluate if a more efficient algorithm is possible.",
				Effort:      Large,
				Priority:    P3,
			})
			return false
		}
		return true
	})
	return issues
}
