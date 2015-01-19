package sched

import "unsafe"

func memmove(to, from unsafe.Pointer, n uintptr) {
	Memmove(to, from, n)
}

func printlock() {
	Printlock()
}

func printunlock() {
	Printunlock()
}
