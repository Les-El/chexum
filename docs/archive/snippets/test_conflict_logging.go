// +build ignore

// This is a manual test file to demonstrate conflict logging.
// Run with: go run test_conflict_logging.go

package main

import (
	"fmt"
	"os"

	"hashi/internal/conflict"
)

func main() {
	fmt.Println("=== Conflict Detection with Logging Demo ===\n")

	resolver := conflict.NewResolver()
	resolver.SetLogger(os.Stdout)

	fmt.Println("Test 1: Valid flag combination (--verbose only)")
	fmt.Println("------------------------------------------------")
	flags1 := conflict.FlagSet{
		"--verbose": true,
		"--quiet":   false,
	}
	if err := resolver.Check(flags1); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	fmt.Println("\n\nTest 2: Conflicting flags (--quiet and --verbose)")
	fmt.Println("------------------------------------------------")
	flags2 := conflict.FlagSet{
		"--quiet":   true,
		"--verbose": true,
	}
	if err := resolver.Check(flags2); err != nil {
		fmt.Printf("\nERROR DETECTED: %v\n", err)
	}

	fmt.Println("\n\nTest 3: Short form conflict (-v with --quiet)")
	fmt.Println("------------------------------------------------")
	flags3 := conflict.FlagSet{
		"-v":      true,
		"--quiet": true,
	}
	if err := resolver.Check(flags3); err != nil {
		fmt.Printf("\nERROR DETECTED: %v\n", err)
	}
}
