//go:build !linux && !darwin && !windows
// +build !linux,!darwin,!windows

package checkpoint

// getStorageUsage returns 0.0 for unsupported platforms.
func (c *CleanupManager) getStorageUsage() float64 {
	return 0.0
}
