//go:build !windows

package app

// cursorPoint is where the pointer is on the desktop. Only Windows reports
// it to this app, so elsewhere the panel stays where it is put.
func cursorPoint() (int, int, bool) { return 0, 0, false }
