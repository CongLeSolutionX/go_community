package core

import "unsafe"

// memclr clears n bytes starting at ptr.
// in memclr_*.s
func memclr(ptr unsafe.Pointer, n uintptr) {
	Memclr(ptr, n)
}
