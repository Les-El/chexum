package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Les-El/hashi/internal/checkpoint"
)

func main() {

	if err := run(os.Args[1:], nil); err != nil {

		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		os.Exit(1)

	}

}



func run(args []string, cm *checkpoint.CleanupManager) error {
	fs := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	var (
		verbose   = fs.Bool("verbose", false, "Enable verbose output")
		dryRun    = fs.Bool("dry-run", false, "Show what would be cleaned without actually removing files")
		threshold = fs.Float64("threshold", 80.0, "Tmpfs usage threshold percentage to trigger cleanup warning")
		force     = fs.Bool("force", false, "Force cleanup even if tmpfs usage is below threshold")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cm == nil {
		cm = checkpoint.NewCleanupManager(*verbose)
	}
	cm.SetDryRun(*dryRun)

	// Check current tmpfs usage
	needsCleanup, usage := cm.CheckTmpfsUsage(*threshold)
	fmt.Printf("Current tmpfs usage: %.1f%%\n", usage)

	if !needsCleanup && !*force && !*dryRun {
		fmt.Printf("Tmpfs usage (%.1f%%) is below threshold (%.1f%%). Use -force to cleanup anyway.\n", usage, *threshold)
		return nil
	}

	if *dryRun {
		showDryRunInfo()
		return nil
	}

	if needsCleanup {
		fmt.Printf("Tmpfs usage (%.1f%%) exceeds threshold (%.1f%%). Starting cleanup...\n", usage, *threshold)
	} else {
		fmt.Println("Force cleanup requested...")
	}

	if err := cm.CleanupOnExit(); err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	fmt.Println("Cleanup completed successfully!")
	return nil
}

func showDryRunInfo() {
	fmt.Println("DRY RUN MODE - No files will be actually removed")
	fmt.Println("Would clean:")
	fmt.Println("  - /tmp/go-build* directories")
	fmt.Println("  - /tmp/hashi-* files")
	fmt.Println("  - /tmp/checkpoint-* files")
	fmt.Println("  - /tmp/test-* files")
	fmt.Println("  - /tmp/*.tmp files")
}


