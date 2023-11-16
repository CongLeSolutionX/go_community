// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"
	"unsafe"
)

func init() {
	register("AllocAlign16", AllocAlign16)
}

type medObj struct {
	a *int
	_ [512]byte
}

var medObjSink any

// Verifies that GODEBUG=allocalign16 works.
func AllocAlign16() {
	if !strings.Contains(os.Getenv("GODEBUG"), "allocalign16=1") {
		println("allocalign16=1 not found in GODEBUG")
		return
	}
	for i := 0; i < 100; i++ {
		x := &medObj{a: new(int)}
		medObjSink = x

		if uintptr(unsafe.Pointer(x))%16 != 0 {
			println("found non-16-byte-aligned heap allocation")
			return
		}
	}
	println("OK")
}
