package checkpoint

import (
	"os"
	"testing"
)

// TestMain runs after all tests in the checkpoint package
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}


