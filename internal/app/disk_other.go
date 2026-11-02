//go:build !windows

package app

// diskSpace is how much of the disk a folder is on is used and free. Only
// Windows answers it here, so elsewhere the storage page shows what the app
// holds without the share of the disk it is.
func diskSpace(dir string) (total, free int64) { return 0, 0 }
