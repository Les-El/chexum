package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Les-El/hashi/internal/checkpoint"
)

func main() {
	var (
		verbose   = flag.Bool("verbose", false, "Enable verbose output")
		dryRun    = flag.Bool("dry-run", false, "Show what would be cleaned without actually removing files")
		threshold = flag.Float64("threshold", 80.0, "Tmpfs usage threshold percentage to trigger cleanup warning")
		force     = flag.Bool("force", false, "Force cleanup even if tmpfs usage is below threshold")
	)
	flag.Parse()

	cleanup := checkpoint.NewCleanupManager(*verbose)

	// Check current tmpfs usage
	needsCleanup, usage := cleanup.CheckTmpfsUsage(*threshold)
	
	fmt.Printf("Current tmpfs usage: %.1f%%\n", usage)
	
	if !needsCleanup && !*force {
		fmt.Printf("Tmpfs usage (%.1f%%) is below threshold (%.1f%%). Use -force to cleanup anyway.\n", usage, *threshold)
		return
	}
	
	if *dryRun {
		fmt.Println("DRY RUN MODE - No files will be actually removed")
		// In a real implementation, we'd add dry-run logic to the cleanup manager
		fmt.Println("Would clean:")
		fmt.Println("  - /tmp/go-build* directories")
		fmt.Println("  - /tmp/hashi-* files")
		fmt.Println("  - /tmp/checkpoint-* files")
		fmt.Println("  - /tmp/test-* files")
		fmt.Println("  - /tmp/*.tmp files")
		return
	}
	
	if needsCleanup {
		fmt.Printf("Tmpfs usage (%.1f%%) exceeds threshold (%.1f%%). Starting cleanup...\n", usage, *threshold)
	} else {
		fmt.Println("Force cleanup requested...")
	}
	
	if err := cleanup.CleanupOnExit(); err != nil {
		fmt.Fprintf(os.Stderr, "Cleanup failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Cleanup completed successfully!")
}