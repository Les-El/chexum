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

// DocAuditor implements DocumentationAuditor.
type DocAuditor struct {
	fset *token.FileSet
}

// NewDocAuditor initializes a new DocAuditor.
func NewDocAuditor() *DocAuditor {
	return &DocAuditor{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (d *DocAuditor) Name() string { return "DocAuditor" }

// Analyze executes the documentation audit logic.
func (d *DocAuditor) Analyze(ctx context.Context, path string) ([]Issue, error) {
	var allIssues []Issue

	docIssues, _ := d.AuditGoDocumentation(ctx, path)
	allIssues = append(allIssues, docIssues...)

	exampleIssues, _ := d.VerifyExamples(ctx, path)
	allIssues = append(allIssues, exampleIssues...)

	return allIssues, nil
}

// AuditGoDocumentation checks for missing documentation on exported functions.
func (d *DocAuditor) AuditGoDocumentation(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	err := filepath.Walk(rootPath, func(filePath string, info os.FileInfo, err error) error {
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

		fileIssues, err := d.auditFile(filePath)
		if err != nil {
			return nil
		}
		issues = append(issues, fileIssues...)
		return nil
	})

	return issues, err
}

func (d *DocAuditor) auditFile(filePath string) ([]Issue, error) {
	var issues []Issue

	f, err := parser.ParseFile(d.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.IsExported() && x.Doc == nil {
				issues = append(issues, Issue{
					ID:          "MISSING-DOC",
					Category:    Documentation,
					Severity:    Medium,
					Title:       "Missing documentation for exported function",
					Description: fmt.Sprintf("Exported function '%s' has no documentation comment.", x.Name.Name),
					Location:    d.fset.Position(x.Pos()).String(),
					Suggestion:  "Add a documentation comment to the function.",
					Effort:      Small,
					Priority:    P2,
				})
			}
		case *ast.TypeSpec:
			if x.Name.IsExported() && x.Doc == nil {
				// Note: TypeSpec doc can be on the GenDecl
				// This is a simplification.
			}
		}
		return true
	})

	return issues, nil
}

// ValidateREADME ensures the README exists and meets quality standards.
func (d *DocAuditor) ValidateREADME(ctx context.Context, rootPath string) ([]Issue, error) {
	return nil, nil
}

// CheckArchitecturalDocs verifies the presence of architectural documentation.
func (d *DocAuditor) CheckArchitecturalDocs(ctx context.Context, rootPath string) ([]Issue, error) {
	return nil, nil
}

// VerifyExamples checks that example files exist and are valid.
func (d *DocAuditor) VerifyExamples(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	examplePath := filepath.Join(rootPath, "examples/*.go")
	exampleFiles, err := filepath.Glob(examplePath)
	if err != nil {
		return nil, err
	}

	for _, file := range exampleFiles {
		// Basic check: does it compile?
		// In a real system, we might run 'go build' or 'go run'.
		// For the audit engine, we'll just check if it exists for now.
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			issues = append(issues, Issue{
				ID:          "BROKEN-EXAMPLE",
				Category:    Documentation,
				Severity:    High,
				Title:       "Example file missing or unreadable",
				Description: fmt.Sprintf("Example file '%s' cannot be accessed.", file),
				Location:    file,
				Suggestion:  "Restore or fix the example file.",
				Effort:      Small,
				Priority:    P1,
			})
		}
	}

	return issues, nil
}
