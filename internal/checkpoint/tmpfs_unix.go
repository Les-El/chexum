//go:build !windows
// +build !windows

package checkpoint

import (
	"syscall"
)

// getTmpfsUsage returns the current tmpfs usage percentage for /tmp.
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
