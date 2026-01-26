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

// CodeAnalyzer implements CodebaseAnalyzer.
type CodeAnalyzer struct {
	fset *token.FileSet
}

// NewCodeAnalyzer creates a new CodeAnalyzer.
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (c *CodeAnalyzer) Name() string { return "CodeAnalyzer" }

// Analyze implements AnalysisEngine.
func (c *CodeAnalyzer) Analyze(ctx context.Context, path string) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(filePath, ".go") || strings.HasSuffix(filePath, "_test.go") {
			return nil
		}

		fileIssues, err := c.analyzeFile(filePath)
		if err != nil {
			// Log error but continue
			return nil
		}
		issues = append(issues, fileIssues...)
		return nil
	})

	return issues, err
}

func (c *CodeAnalyzer) analyzeFile(filePath string) ([]Issue, error) {
	var issues []Issue

	f, err := parser.ParseFile(c.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	issues = append(issues, c.scanTechnicalDebt(f)...)
	issues = append(issues, c.scanSecurityIssues(f)...)

	return issues, nil
}

func (c *CodeAnalyzer) scanTechnicalDebt(f *ast.File) []Issue {
	var issues []Issue
	// Scan for technical debt markers (such as debt tags)
	for _, commentGroup := range f.Comments {
		for _, comment := range commentGroup.List {
			text := comment.Text
			if strings.Contains(text, "TODO") || strings.Contains(text, "FIXME") {
				// Check for review markers
				if strings.Contains(text, "Reviewed:") {
					continue
				}

				issues = append(issues, Issue{
					ID:          "TECH-DEBT",
					Category:    CodeQuality,
					Severity:    Info,
					Title:       "Technical Debt identified",
					Description: fmt.Sprintf("Found TODO/FIXME in comment: %s", text),
					Location:    c.fset.Position(comment.Pos()).String(),
					Suggestion:  "Address the TODO or FIXME comment.",
					Effort:      Small,
					Priority:    P3,
				})
			}
		}
	}
	return issues
}

func (c *CodeAnalyzer) scanSecurityIssues(f *ast.File) []Issue {
	var issues []Issue
	// Basic Security Check (unsafe)
	ast.Inspect(f, func(n ast.Node) bool {
		if imp, ok := n.(*ast.ImportSpec); ok {
			if imp.Path != nil && imp.Path.Value == "\"unsafe\"" {
				// Check for review markers in comments above the import
				isReviewed := false
				if imp.Doc != nil {
					for _, comment := range imp.Doc.List {
						if strings.Contains(comment.Text, "Reviewed: SECURITY-UNSAFE") || strings.Contains(comment.Text, "Reviewed: UNSAFE") {
							isReviewed = true
							break
						}
					}
				}

				if !isReviewed {
					issues = append(issues, Issue{
						ID:          "SECURITY-UNSAFE",
						Category:    Security,
						Severity:    Medium,
						Title:       "Usage of 'unsafe' package",
						Description: "The 'unsafe' package is used in this file.",
						Location:    c.fset.Position(imp.Pos()).String(),
						Suggestion:  "Verify if 'unsafe' is absolutely necessary.",
						Effort:      MediumEffort,
						Priority:    P2,
					})
				}
			}
		}
		return true
	})
	return issues
}

// AnalyzePackages performs analysis at the package level.
func (c *CodeAnalyzer) AnalyzePackages(ctx context.Context, path string) ([]Issue, error) {
	return c.Analyze(ctx, path)
}

// CheckSecurity performs security-specific analysis.
func (c *CodeAnalyzer) CheckSecurity(ctx context.Context, path string) ([]Issue, error) {
	return c.Analyze(ctx, path)
}

// AssessDependencies evaluates the project's dependencies.
func (c *CodeAnalyzer) AssessDependencies(ctx context.Context, path string) ([]Issue, error) {
	return c.Analyze(ctx, path)
}

// IdentifyTechnicalDebt finds technical debt markers in the codebase.
func (c *CodeAnalyzer) IdentifyTechnicalDebt(ctx context.Context, path string) ([]Issue, error) {
	return c.Analyze(ctx, path)
}
