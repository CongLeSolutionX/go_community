// errorcheck

// Verify that the Go compiler catches invalid pointer literals.
// Does not compile.

package main

import (
	"unsafe"
)

func main() {
	c := (*int)(unsafe.Pointer(uintptr(1)))	// ERROR "bad pointer literal 1"
	_ = c
}
