//go:build windows

package app

import (
	"syscall"
	"unsafe"
)

// The pointer is asked for rather than waited on, so the panel that follows
// it needs to know where it is: a window is told only what happens on it,
// and nothing when the pointer goes elsewhere.

var procGetCursorPos = syscall.NewLazyDLL("user32.dll").NewProc("GetCursorPos")

// cursorPoint is where the pointer is on the desktop, in pixels.
func cursorPoint() (int, int, bool) {
	var p struct{ X, Y int32 }
	if r, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(&p))); r == 0 {
		return 0, 0, false
	}
	return int(p.X), int(p.Y), true
}
