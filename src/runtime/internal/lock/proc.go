// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

// funcPC returns the entry PC of the function f.
// It assumes that f is a func value. Otherwise the behavior is undefined.
//go:nosplit
func FuncPC(f interface{}) uintptr {
	return **(**uintptr)(_core.Add(unsafe.Pointer(&f), _core.PtrSize))
}

var (
	Allgs    []*_core.G
	Allglock _core.Mutex
)
