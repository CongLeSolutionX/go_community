// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cgotest

// Issue 52336: generate ELFv2 ABI save/restore functions from go linker.
//              These are calls are made when compiling with -Os on ppc64le.

/*
#cgo CFLAGS: -Os

int foo_fpr() {
        asm volatile("":::"fr31","fr30","fr29","fr28");
}
int foo_gpr0() {
        asm volatile("":::"r30","r29","r28");
}
int foo_gpr1() {
        asm volatile("":::"fr31", "fr30","fr29","fr28","r30","r29","r28");
}
int foo_vr() {
        asm volatile("":::"v31","v30","v29","v28");
}
*/
import "C"

import (
	"testing"
)

func test52336(t *testing.T) {
	C.foo_fpr()
	C.foo_gpr0()
	C.foo_gpr1()
	C.foo_vr()
	t.Log("success")
}
