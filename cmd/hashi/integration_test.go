package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCLIBasicExecution(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to execute CLI: %v\nOutput: %s", err, output)
	}

	expectedOutput := "hashi [OPTIONS] [FILES...]"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}
}