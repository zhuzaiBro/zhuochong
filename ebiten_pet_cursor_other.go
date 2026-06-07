//go:build ebitenpet && !darwin && !windows

package main

func screenCursorPos() (int, int, bool) {
	return 0, 0, false
}
