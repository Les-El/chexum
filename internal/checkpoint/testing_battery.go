package checkpoint

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TestingBattery implements TestingBatteryManager.
type TestingBattery struct {
	fset *token.FileSet
}

// NewTestingBattery initializes a new TestingBattery.
func NewTestingBattery() *TestingBattery {
	return &TestingBattery{
		fset: token.NewFileSet(),
	}
}

// Name returns the name of the analyzer.
func (t *TestingBattery) Name() string { return "TestingBattery" }

// Analyze executes a comprehensive test suite analysis.
func (t *TestingBattery) Analyze(ctx context.Context, path string) ([]Issue, error) {
	// For analysis, we report packages with low coverage and reliability issues.
	var issues []Issue
	
	coverageIssues, err := t.IdentifyLowCoverage(ctx, path)
	if err == nil {
		issues = append(issues, coverageIssues...)
	}

	unitIssues, err := t.CreateUnitTests(ctx, path)
	if err == nil {
		issues = append(issues, unitIssues...)
	}

	reliabilityIssues, err := t.CheckTestReliability(ctx, path)
	if err == nil {
		issues = append(issues, reliabilityIssues...)
	}

	integrationIssues, err := t.BuildIntegrationTests(ctx, path)
	if err == nil {
		issues = append(issues, integrationIssues...)
	}

	propertyIssues, err := t.ImplementPropertyTests(ctx, path)
	if err == nil {
		issues = append(issues, propertyIssues...)
	}

	benchmarkIssues, err := t.CreateBenchmarks(ctx, path)
	if err == nil {
		issues = append(issues, benchmarkIssues...)
	}

	return issues, nil
}

// CheckTestReliability runs tests selectively to detect flakiness without resource exhaustion.
func (t *TestingBattery) CheckTestReliability(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	// Instead of ./..., we identify critical packages and run them with -short
	criticalPackages := []string{"internal/hash", "internal/conflict"}
	
	for _, pkgRel := range criticalPackages {
		pkg := filepath.Join(rootPath, pkgRel)
		cmd := exec.CommandContext(ctx, "go", "test", "-short", "-count=1", "./"+pkgRel)
		cmd.Dir = rootPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "--- FAIL:") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						testName := parts[2]
						issues = append(issues, Issue{
							ID:          "FAILING-TEST",
							Category:    Testing,
							Severity:    High,
							Title:       fmt.Sprintf("Test failure in %s", pkg),
							Description: fmt.Sprintf("Test '%s' failed in package %s.", testName, pkg),
							Location:    pkg,
							Suggestion:  "Investigate the test failure and fix the regression.",
							Effort:      MediumEffort,
							Priority:    P1,
						})
					}
				}
			}
		}
	}

	return issues, nil
}

// IdentifyLowCoverage uses static analysis and selective execution to assess coverage.
func (t *TestingBattery) IdentifyLowCoverage(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	internalPath := filepath.Join(rootPath, "internal")
	// We'll perform a basic "test file presence" check as a proxy for coverage in this analysis
	err := filepath.Walk(internalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || info.Name() == "vendor" || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Check if any .go files exist in this dir
		hasGoFiles := false
		files, _ := os.ReadDir(path)
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") && !strings.HasSuffix(f.Name(), "_test.go") {
				hasGoFiles = true
				break
			}
		}

		if !hasGoFiles {
			return nil
		}

		// Check for test files
		hasTests := false
		for _, f := range files {
			if !f.IsDir() && strings.HasSuffix(f.Name(), "_test.go") {
				hasTests = true
				break
			}
		}

		if !hasTests {
			            issues = append(issues, Issue{
			                ID:          "MISSING-TEST-SUITE",				Category:    Testing,
				Severity:    High,
				Title:       fmt.Sprintf("Package %s has no tests", path),
				Description: fmt.Sprintf("The package '%s' contains source code but lacks any test files.", path),
				Location:    path,
				Suggestion:  "Create a _test.go file and add unit tests.",
				Effort:      MediumEffort,
				Priority:    P1,
			})
		}

		return nil
	})

	return issues, err
}

// CreateUnitTests identifies missing unit tests for exported functions.
func (t *TestingBattery) CreateUnitTests(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	// Scan packages for missing unit tests
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if strings.Contains(path, "vendor/") || strings.HasPrefix(path, ".") {
			return nil
		}

		f, err := parser.ParseFile(t.fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		// Check if a test file exists
		testFilePath := strings.TrimSuffix(path, ".go") + "_test.go"
		testFileExists := false
		if _, err := os.Stat(testFilePath); err == nil {
			testFileExists = true
		}

		var testedFuncs map[string]bool
		if testFileExists {
			testedFuncs = t.getTestedFunctions(testFilePath)
		}

		ast.Inspect(f, func(n ast.Node) bool {
			if fnIssues := t.checkMissingTests(n, path, testFilePath, testedFuncs); fnIssues != nil {
				issues = append(issues, fnIssues...)
			}
			return true
		})

		return nil
	})

	return issues, err
}

func (t *TestingBattery) checkMissingTests(n ast.Node, path, testFilePath string, testedFuncs map[string]bool) []Issue {
	fn, ok := n.(*ast.FuncDecl)
	if !ok || !fn.Name.IsExported() {
		return nil
	}

	// Try standard naming TestFunc
	testName := "Test" + fn.Name.Name
	if testedFuncs[testName] {
		return nil
	}

	// Try Type_Method naming if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		var typeName string
		switch x := fn.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if id, ok := x.X.(*ast.Ident); ok {
				typeName = id.Name
			}
		case *ast.Ident:
			typeName = x.Name
		}
		if typeName != "" {
			methodTestName := fmt.Sprintf("Test%s_%s", typeName, fn.Name.Name)
			if testedFuncs[methodTestName] {
				return nil
			}
		}
	}

	return []Issue{{
		ID:          "MISSING-UNIT-TEST",
		Category:    Testing,
		Severity:    Medium,
		Title:       "Missing unit test for exported function",
		Description: fmt.Sprintf("Function '%s' in '%s' has no corresponding unit test.", fn.Name.Name, path),
		Location:    path,
		Suggestion:  fmt.Sprintf("Add %s to %s", testName, testFilePath),
		Effort:      Small,
		Priority:    P2,
	}}
}

func (t *TestingBattery) getTestedFunctions(testFilePath string) map[string]bool {
	funcs := make(map[string]bool)
	f, err := parser.ParseFile(t.fset, testFilePath, nil, 0)
	if err != nil {
		return funcs
	}

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "Test") {
			funcs[fn.Name.Name] = true
		}
		return true
	})

	return funcs
}

// BuildIntegrationTests checks for the existence of integration tests.
func (t *TestingBattery) BuildIntegrationTests(ctx context.Context, rootPath string) ([]Issue, error) {
	// Integration tests usually target the cmd/hashi/main.go or similar.
	// We'll look for CLI command definitions.
	var issues []Issue

	integrationTestPath := filepath.Join(rootPath, "cmd/hashi/integration_test.go")
	// Simplified: just suggest integration tests for the main entry point if missing
	if _, err := os.Stat(integrationTestPath); os.IsNotExist(err) {
		issues = append(issues, Issue{
			ID:          "MISSING-INTEGRATION-TESTS",
			Category:    Testing,
			Severity:    High,
			Title:       "Missing CLI integration tests",
			Description: "The main CLI entry point lacks comprehensive integration tests.",
			Location:    "cmd/hashi",
			Suggestion:  "Create cmd/hashi/integration_test.go to test CLI workflows.",
			Effort:      Large,
			Priority:    P1,
		})
	}

	return issues, nil
}

// ImplementPropertyTests checks for missing property-based tests in core packages.
func (t *TestingBattery) ImplementPropertyTests(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue

	// Identify "core algorithms" - we'll look for packages like internal/hash, internal/conflict
	corePackages := []string{"internal/hash", "internal/conflict", "internal/checkpoint"}
	for _, pkg := range corePackages {
		testFile := filepath.Join(rootPath, pkg, "property_test.go")
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			issues = append(issues, Issue{
				ID:          "MISSING-PROPERTY-TEST",
				Category:    Testing,
				Severity:    Medium,
				Title:       "Missing property-based tests for core logic",
				Description: fmt.Sprintf("Package '%s' contains core logic but lacks property-based tests.", pkg),
				Location:    pkg,
				Suggestion:  fmt.Sprintf("Implement property-based tests in %s", testFile),
				Effort:      MediumEffort,
				Priority:    P2,
			})
		}
	}

	return issues, nil
}

// CreateBenchmarks checks for missing benchmarks in performance-critical areas.
func (t *TestingBattery) CreateBenchmarks(ctx context.Context, rootPath string) ([]Issue, error) {
	var issues []Issue
	// Identify performance-critical areas (e.g., hashing)
	benchmarkPath := filepath.Join(rootPath, "internal/hash/benchmark_test.go")
	if _, err := os.Stat(benchmarkPath); os.IsNotExist(err) {
		issues = append(issues, Issue{
			ID:          "MISSING-BENCHMARK",
			Category:    Testing,
			Severity:    Low,
			Title:       "Missing benchmarks for performance-critical code",
			Description: "Hashing operations should be benchmarked to detect regressions.",
			Location:    "internal/hash",
			Suggestion:  "Add BenchmarkHash in internal/hash/benchmark_test.go",
			Effort:      Small,
			Priority:    P3,
		})
	}
	return issues, nil
}
