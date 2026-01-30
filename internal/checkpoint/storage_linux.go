//go:build linux
// +build linux

package checkpoint

import (
	"syscall"
)

// getStorageUsage returns the current storage usage percentage for the base directory.
func (c *CleanupManager) getStorageUsage() float64 {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(c.baseDir, &stat); err != nil {
		return 0.0
	}

	total := stat.Blocks * uint64(stat.Frsize)
	free := stat.Bavail * uint64(stat.Frsize)
	used := total - free

	if total == 0 {
		return 0.0
	}

	return float64(used) / float64(total) * 100.0
}
