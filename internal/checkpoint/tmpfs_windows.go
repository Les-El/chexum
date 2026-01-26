//go:build windows
// +build windows

package checkpoint

// getTmpfsUsage returns the current tmpfs usage percentage.
// On Windows, this is a stub as tmpfs is not applicable.
func (c *CleanupManager) getTmpfsUsage() float64 {
	return 0.0
}
