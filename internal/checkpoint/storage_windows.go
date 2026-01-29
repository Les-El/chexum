//go:build windows
// +build windows

package checkpoint

import (
	"golang.org/x/sys/windows"
)

// getStorageUsage returns the current storage usage percentage for the base directory.
// On Windows, it uses GetDiskFreeSpaceEx to retrieve disk statistics.
func (c *CleanupManager) getStorageUsage() float64 {
	pathPtr, err := windows.UTF16PtrFromString(c.baseDir)
	if err != nil {
		return 0.0
	}

	var freeBytes, totalBytes, availBytes uint64
	err = windows.GetDiskFreeSpaceEx(pathPtr, &freeBytes, &totalBytes, &availBytes)
	if err != nil {
		return 0.0
	}

	if totalBytes == 0 {
		return 0.0
	}

	used := totalBytes - freeBytes
	return float64(used) / float64(totalBytes) * 100.0
}
