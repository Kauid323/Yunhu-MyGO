//go:build windows

package app

import (
	"syscall"
	"unsafe"
)

var procGetDiskFreeSpaceEx = syscall.NewLazyDLL("kernel32.dll").
	NewProc("GetDiskFreeSpaceExW")

// diskSpace is how much of the disk a folder is on is used and free, in
// bytes: what the storage space page shows the app's share of.
func diskSpace(dir string) (total, free int64) {
	if dir == "" {
		return 0, 0
	}
	p, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return 0, 0
	}
	var avail, all, freeBytes uint64
	r, _, _ := procGetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&avail)),
		uintptr(unsafe.Pointer(&all)),
		uintptr(unsafe.Pointer(&freeBytes)),
	)
	if r == 0 {
		return 0, 0
	}
	return int64(all), int64(freeBytes)
}
