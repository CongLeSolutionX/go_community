// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.swaplencap
// +build goexperiment.swaplencap

package runtime

import "unsafe"

type slice struct {
	array unsafe.Pointer
	cap   int
	len   int
	_     int
}

// A notInHeapSlice is a slice backed by runtime/internal/sys.NotInHeap memory.
type notInHeapSlice struct {
	array *notInHeap
	cap   int
	len   int
	_     int
}
