// asmcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

// This file contains code generation tests related to the handling of
// slice types.

// ------------------ //
//      Clear         //
// ------------------ //

// Issue #5373 optimize memset idiom

func SliceClear(s []int) []int {
	// amd64:`.*memclrNoHeapPointers`
	for i := range s {
		s[i] = 0
	}
	return s
}

func SliceClearPointers(s []*int) []*int {
	// amd64:`.*memclrHasPointers`
	for i := range s {
		s[i] = nil
	}
	return s
}

// ------------------ //
//      Extension     //
// ------------------ //

// Issue #21266 - avoid makeslice in append(x, make([]T, y)...)

func SliceExtensionConst(s []int) []int {
	// amd64:`.*runtime\.memclrNoHeapPointers`
	// amd64:-`.*runtime\.makeslice`
	// amd64:-`.*runtime\.panicmakeslicelen`
	return append(s, make([]int, 1<<2)...)
}

func SliceExtensionPointer(s []*int, l int) []*int {
	// amd64:`.*runtime\.memclrHasPointers`
	// amd64:-`.*runtime\.makeslice`
	return append(s, make([]*int, l)...)
}

func SliceExtensionVar(s []byte, l int) []byte {
	// amd64:`.*runtime\.memclrNoHeapPointers`
	// amd64:-`.*runtime\.makeslice`
	return append(s, make([]byte, l)...)
}

func SliceExtensionInt64(s []int, l64 int64) []int {
	// 386:`.*runtime\.makeslice`
	// 386:-`.*runtime\.memclr`
	return append(s, make([]int, l64)...)
}

// ------------------ //
//      Make+Copy     //
// ------------------ //

// Issue #26252 - avoid memclr for make+copy

func SliceMakeCopyLen(s []int) []int {
	// amd64:`.*runtime\.mallocgc`
	// amd64:`.*runtime\.memmove`
	// amd64:-`.*runtime\.makeslice`
	a := make([]int, len(s))
	copy(a, s)
	return a
}

func SliceMakeCopyLenPtr(s []*int) []*int {
	// amd64:`.*runtime\.makeslicecopy`
	// amd64:-`.*runtime\.makeslice$`
	// amd64:-`.*runtime\.typedslicecopy
	a := make([]*int, len(s))
	copy(a, s)
	return a
}

func SliceMakeCopyConst(s []int) []int {
	// amd64:`.*runtime\.makeslicecopy`
	// amd64:-`.*runtime\.makeslice$`
	// amd64:-`.*runtime\.memmove`
	a := make([]int, 4)
	copy(a, s)
	return a
}

func SliceMakeCopyConstPtr(s []*int) []*int {
	// amd64:`.*runtime\.makeslicecopy`
	// amd64:-`.*runtime\.makeslice$`
	// amd64:-`.*runtime\.typedslicecopy
	a := make([]*int, 4)
	copy(a, s)
	return a
}

// ---------------------- //
//   Nil check of &s[0]   //
// ---------------------- //
// See issue 30366
func SliceNilCheck(s []int) {
	p := &s[0]
	// amd64:-`TESTB`
	_ = *p
}
