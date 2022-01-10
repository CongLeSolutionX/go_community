//go:build !ppc64 && !ppc64le && !windows && !wasm && !gccgo

package main

func syncIcache(p uintptr) {
}
