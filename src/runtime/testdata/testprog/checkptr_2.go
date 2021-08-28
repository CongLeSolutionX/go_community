// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"unsafe"
)

func init() {
	register("CheckPtrAlignmentNested", CheckPtrAlignmentNested)
}

func CheckPtrAlignmentNested() {
	s := make([]int8, 100)
	p := unsafe.Pointer(&s[0])
	n := 9
	_ = ((*[10]int8)(unsafe.Pointer((*[10]int64)(unsafe.Pointer(&p)))))[:n:n]
}
