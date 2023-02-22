// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.swaplencap
// +build goexperiment.swaplencap

package runtime

import "unsafe"

type slice struct {
	_align [0][]byte // this ought to force alignment
	array  unsafe.Pointer
	cap    int
	len    int
}

// A notInHeapSlice is a slice backed by runtime/internal/sys.NotInHeap memory.
type notInHeapSlice struct {
	_align [0][]byte // this ought to force alignment
	array  *notInHeap
	cap    int
	len    int
}
