package testutil

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// RandomString generates a random string of the given length.
func RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// RandomHash generates a random hex string of the given length.
func RandomHash(length int) string {
	const hexBytes = "0123456789abcdef"
	b := make([]byte, length)
	for i := range b {
		b[i] = hexBytes[rand.Intn(len(hexBytes))]
	}
	return string(b)
}

// CreateRandomDirectoryStructure creates a random directory structure with files.
func CreateRandomDirectoryStructure(t *testing.T, root string, depth, maxDirs, maxFiles int) {
	t.Helper()
	if depth <= 0 {
		return
	}

	numDirs := rand.Intn(maxDirs + 1)
	numFiles := rand.Intn(maxFiles + 1)

	// Ensure at least one entry at the root depth
	if depth > 1 && numDirs == 0 && numFiles == 0 {
		if rand.Float32() < 0.5 {
			numDirs = 1
		} else {
			numFiles = 1
		}
	} else if depth == 1 && numDirs == 0 && numFiles == 0 {
		numFiles = 1
	}

	for i := 0; i < numFiles; i++ {
		fileName := fmt.Sprintf("file_%d_%d.txt", depth, i)
		content := RandomString(rand.Intn(100) + 10)
		CreateFile(t, root, fileName, content)
	}

	for i := 0; i < numDirs; i++ {
		dirName := fmt.Sprintf("dir_%d_%d", depth, i)
		dirPath := filepath.Join(root, dirName)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("failed to create random dir %s: %v", dirName, err)
		}
		CreateRandomDirectoryStructure(t, dirPath, depth-1, maxDirs, maxFiles)
	}
}

// GenerateMockGoFile creates a mock Go file with some content.
func GenerateMockGoFile(t *testing.T, dir, name string, hasTodo, hasUnsafe bool) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("package main\n")
	if hasUnsafe {
		sb.WriteString("import \"unsafe\"\n")
	}
	sb.WriteString("func main() {\n")
	if hasTodo {
		sb.WriteString("\t// TODO: implement this\n")
	}
	sb.WriteString("}\n")
	return CreateFile(t, dir, name, sb.String())
}
