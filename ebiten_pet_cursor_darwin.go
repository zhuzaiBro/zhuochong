//go:build ebitenpet && darwin

package main

/*
#include <ApplicationServices/ApplicationServices.h>
*/
import "C"

func screenCursorPos() (int, int, bool) {
	ev := C.CGEventCreate(C.CGEventSourceRef(0))
	if ev == 0 {
		return 0, 0, false
	}
	defer C.CFRelease(C.CFTypeRef(ev))
	p := C.CGEventGetLocation(ev)
	return int(p.x), int(p.y), true
}
