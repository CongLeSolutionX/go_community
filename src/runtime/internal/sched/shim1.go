package sched

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

// memmove copies n bytes from "from" to "to".
// in memmove_*.s
func memmove(to, from unsafe.Pointer, n uintptr) {
	Memmove(to, from, n)
}
