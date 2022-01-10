//go:build !(aix && ppc64) && !windows && !wasm && !gccgo

package main

func syncIcache(p uintptr) {
}
