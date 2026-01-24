package checkpoint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// CleanupManager handles temporary file cleanup operations
type CleanupManager struct {
	verbose bool
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(verbose bool) *CleanupManager {
	return &CleanupManager{
		verbose: verbose,
	}
}

// CleanupResult contains information about the cleanup operation
type CleanupResult struct {
	FilesRemoved     int
	DirsRemoved      int
	SpaceFreed       int64
	Errors           []string
	Duration         time.Duration
	TmpfsUsageBefore float64
	TmpfsUsageAfter  float64
}

// CleanupTemporaryFiles removes Go build artifacts and other temporary files
func (c *CleanupManager) CleanupTemporaryFiles() (*CleanupResult, error) {
	start := time.Now()
	result := &CleanupResult{}
	
	// Get tmpfs usage before cleanup
	result.TmpfsUsageBefore = c.getTmpfsUsage()
	
	if c.verbose {
		fmt.Printf("Starting temporary file cleanup...\n")
		fmt.Printf("Tmpfs usage before cleanup: %.1f%%\n", result.TmpfsUsageBefore)
	}
	
	// Clean Go build artifacts
	if err := c.cleanGoBuildArtifacts(result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Go build cleanup error: %v", err))
	}
	
	// Clean other temporary files
	if err := c.cleanOtherTempFiles(result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Other temp cleanup error: %v", err))
	}
	
	// Get tmpfs usage after cleanup
	result.TmpfsUsageAfter = c.getTmpfsUsage()
	result.Duration = time.Since(start)
	
	if c.verbose {
		fmt.Printf("Cleanup completed in %v\n", result.Duration)
		fmt.Printf("Files removed: %d, Directories removed: %d\n", result.FilesRemoved, result.DirsRemoved)
		fmt.Printf("Space freed: %s\n", c.formatBytes(result.SpaceFreed))
		fmt.Printf("Tmpfs usage after cleanup: %.1f%%\n", result.TmpfsUsageAfter)
		if len(result.Errors) > 0 {
			fmt.Printf("Errors encountered: %d\n", len(result.Errors))
		}
	}
	
	return result, nil
}

// cleanGoBuildArtifacts removes Go build temporary directories
func (c *CleanupManager) cleanGoBuildArtifacts(result *CleanupResult) error {
	tmpDir := "/tmp"
	
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read /tmp directory: %w", err)
	}
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		if strings.HasPrefix(name, "go-build") {
			dirPath := filepath.Join(tmpDir, name)
			
			// Get size before removal
			size, err := c.getDirSize(dirPath)
			if err == nil {
				result.SpaceFreed += size
			}
			
			if c.verbose {
				fmt.Printf("Removing Go build directory: %s\n", dirPath)
			}
			
			if err := os.RemoveAll(dirPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove %s: %v", dirPath, err))
			} else {
				result.DirsRemoved++
			}
		}
	}
	
	return nil
}

// cleanOtherTempFiles removes other temporary files that might accumulate
func (c *CleanupManager) cleanOtherTempFiles(result *CleanupResult) error {
	tmpDir := "/tmp"
	patterns := []string{"hashi-*", "checkpoint-*", "test-*", "*.tmp"}
	
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read /tmp directory: %w", err)
	}
	
	for _, entry := range entries {
		c.processTempEntry(tmpDir, entry, patterns, result)
	}
	
	return nil
}

func (c *CleanupManager) processTempEntry(tmpDir string, entry os.DirEntry, patterns []string, result *CleanupResult) {
	name := entry.Name()
	filePath := filepath.Join(tmpDir, name)
	
	shouldClean := false
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, name); matched {
			shouldClean = true
			break
		}
	}
	
	if !shouldClean {
		return
	}
	
	if info, err := entry.Info(); err == nil {
		if entry.IsDir() {
			if size, err := c.getDirSize(filePath); err == nil {
				result.SpaceFreed += size
			}
		} else {
			result.SpaceFreed += info.Size()
		}
	}
	
	if c.verbose {
		fmt.Printf("Removing temporary file/directory: %s\n", filePath)
	}
	
	if err := os.RemoveAll(filePath); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove %s: %v", filePath, err))
	} else {
		if entry.IsDir() {
			result.DirsRemoved++
		} else {
			result.FilesRemoved++
		}
	}
}

// getDirSize calculates the total size of a directory
func (c *CleanupManager) getDirSize(path string) (int64, error) {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	
	return size, err
}

// getTmpfsUsage returns the current tmpfs usage percentage
func (c *CleanupManager) getTmpfsUsage() float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/tmp", &stat); err != nil {
		return 0.0
	}
	
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	
	if total == 0 {
		return 0.0
	}
	
	return float64(used) / float64(total) * 100.0
}

// formatBytes formats bytes into human-readable format
func (c *CleanupManager) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CleanupOnExit performs cleanup and reports results
func (c *CleanupManager) CleanupOnExit() error {
	result, err := c.CleanupTemporaryFiles()
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}
	
	// Always report summary, even in non-verbose mode
	fmt.Printf("\n=== Cleanup Summary ===\n")
	fmt.Printf("Files removed: %d\n", result.FilesRemoved)
	fmt.Printf("Directories removed: %d\n", result.DirsRemoved)
	fmt.Printf("Space freed: %s\n", c.formatBytes(result.SpaceFreed))
	fmt.Printf("Tmpfs usage: %.1f%% → %.1f%%\n", result.TmpfsUsageBefore, result.TmpfsUsageAfter)
	fmt.Printf("Duration: %v\n", result.Duration)
	
	if len(result.Errors) > 0 {
		fmt.Printf("Errors: %d\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
	
	return nil
}

// CheckTmpfsUsage checks if tmpfs usage is above a threshold and suggests cleanup
func (c *CleanupManager) CheckTmpfsUsage(threshold float64) (bool, float64) {
	usage := c.getTmpfsUsage()
	return usage > threshold, usage
}