// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import "unsafe"

// Swapper returns a function which swaps the elements in slice.
// Swapper panics if the provided interface is not a slice.
func Swapper(slice interface{}) func(i, j int) {
	v := ValueOf(slice)
	if v.kind() != Slice {
		panic(&ValueError{"reflect.Swapper", v.kind()})
	}
	s := (*sliceHeader)(v.ptr)
	tt := (*sliceType)(unsafe.Pointer(v.typ))
	typ := tt.elem
	size := typ.size
	hasPtr := typ.kind&kindNoPointers == 0

	// Some common & small cases, without using memmove:
	if hasPtr {
		if size == ptrSize {
			var ps []unsafe.Pointer
			*(*sliceHeader)(unsafe.Pointer(&ps)) = *s
			return func(i, j int) { ps[i], ps[j] = ps[j], ps[i] }
		}
		if Kind(typ.kind) == String {
			var ss []string
			*(*sliceHeader)(unsafe.Pointer(&ss)) = *s
			return func(i, j int) { ss[i], ss[j] = ss[j], ss[i] }
		}
	} else {
		switch size {
		case 8:
			var is []int64
			*(*sliceHeader)(unsafe.Pointer(&is)) = *s
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 4:
			var is []int32
			*(*sliceHeader)(unsafe.Pointer(&is)) = *s
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 2:
			var is []int16
			*(*sliceHeader)(unsafe.Pointer(&is)) = *s
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		case 1:
			var is []int8
			*(*sliceHeader)(unsafe.Pointer(&is)) = *s
			return func(i, j int) { is[i], is[j] = is[j], is[i] }
		}
	}

	// Allocate scratch space for swaps:
	tmpVal := New(typ)
	if tmpVal.flag&flagIndir != 0 {
		panic("unsupported scratch value")
	}
	tmp := tmpVal.ptr

	maxLen := uint(s.Len)

	// If no pointers, we don't require typedmemmove:
	if !hasPtr {
		return func(i, j int) {
			if uint(i) >= maxLen || uint(j) >= maxLen {
				panic("reflect: slice index out of range")
			}
			val1 := arrayAt(s.Data, i, size)
			val2 := arrayAt(s.Data, j, size)
			memmove(tmp, val1, size)
			memmove(val1, val2, size)
			memmove(val2, tmp, size)
		}
	}

	return func(i, j int) {
		if uint(i) >= maxLen || uint(j) >= maxLen {
			panic("reflect: slice index out of range")
		}
		val1 := arrayAt(s.Data, i, size)
		val2 := arrayAt(s.Data, j, size)
		typedmemmove(typ, tmp, val1)
		typedmemmove(typ, val1, val2)
		typedmemmove(typ, val2, tmp)
	}
}

// memmove copies size bytes from src to dst.
// The memory must not contain any pointers.
//go:noescape
func memmove(dst, src unsafe.Pointer, size uintptr)
