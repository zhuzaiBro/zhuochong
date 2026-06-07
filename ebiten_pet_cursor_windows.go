//go:build ebitenpet && windows

package main

import (
	"syscall"
	"unsafe"
)

func screenCursorPos() (int, int, bool) {
	var pt struct{ x, y int32 }
	ret, _, _ := syscall.NewLazyDLL("user32.dll").
		NewProc("GetCursorPos").
		Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return 0, 0, false
	}
	return int(pt.x), int(pt.y), true
}
