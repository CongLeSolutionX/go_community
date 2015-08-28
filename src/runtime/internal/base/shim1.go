package base

import "unsafe"

func memhash(p unsafe.Pointer, seed, s uintptr) uintptr {
	return Memhash(p, seed, s)
}

func printlock() {
	Printlock()
}

func printunlock() {
	Printunlock()
}

// memclr clears n bytes starting at ptr.
// in memclr_*.s
func memclr(ptr unsafe.Pointer, n uintptr) {
	Memclr(ptr, n)
}

// memmove copies n bytes from "from" to "to".
// in memmove_*.s
func memmove(to, from unsafe.Pointer, n uintptr) {
	Memmove(to, from, n)
}
