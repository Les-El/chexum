package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDocAuditor(t *testing.T) {
	auditor := NewDocAuditor()
	if auditor == nil {
		t.Fatal("NewDocAuditor returned nil")
	}
}

func TestDocAuditor_Name(t *testing.T) {
	auditor := NewDocAuditor()
	if name := auditor.Name(); name != "DocAuditor" {
		t.Errorf("expected DocAuditor, got %s", name)
	}
}

func TestDocAuditor_Analyze(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()

	// Analyze calls AuditGoDocumentation
	_, err := auditor.Analyze(ctx, "../../")
	if err != nil {
		t.Errorf("Analyze failed: %v", err)
	}
}

func TestAuditGoDocumentation(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()

	// Create a temp file with missing doc
	tmpDir, err := os.MkdirTemp("", "doc_auditor_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	content := `package test
func ExportedWithoutDoc() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	issues, err := auditor.AuditGoDocumentation(ctx, tmpDir)
	if err != nil {
		t.Fatalf("AuditGoDocumentation failed: %v", err)
	}

	foundMissingDoc := false
	for _, issue := range issues {
		if issue.ID == "MISSING-DOC" {
			foundMissingDoc = true
			break
		}
	}

	if !foundMissingDoc {
		t.Error("expected to find MISSING-DOC issue")
	}
}

func TestValidateREADME(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()

	issues, err := auditor.ValidateREADME(ctx, "../../")
	if err != nil {
		t.Errorf("ValidateREADME failed: %v", err)
	}
	_ = issues
}

func TestCheckArchitecturalDocs(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()

	issues, err := auditor.CheckArchitecturalDocs(ctx, "../../")
	if err != nil {
		t.Errorf("CheckArchitecturalDocs failed: %v", err)
	}
	_ = issues
}

func TestVerifyExamples(t *testing.T) {
	auditor := NewDocAuditor()
	ctx := context.Background()

	// Create a dummy example directory in tmpDir
	tmpDir, _ := os.MkdirTemp("", "examples_test")
	defer os.RemoveAll(tmpDir)
	os.Mkdir(filepath.Join(tmpDir, "examples"), 0755)

	content := `package main
func main() {}
`
	os.WriteFile(filepath.Join(tmpDir, "examples/test.go"), []byte(content), 0644)

	issues, err := auditor.VerifyExamples(ctx, tmpDir)
	if err != nil {
		t.Fatalf("VerifyExamples failed: %v", err)
	}

	// Should be 0 issues if example exists
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}
