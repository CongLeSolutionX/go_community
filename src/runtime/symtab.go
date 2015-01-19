// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_lock "runtime/internal/lock"
	"unsafe"
)

// NOTE: Func does not expose the actual unexported fields, because we return *Func
// values to users, and we want to keep them from being able to overwrite the data
// with (say) *f = Func{}.
// All code operating on a *Func must call raw to get the *_func instead.

// A Func represents a Go function in the running binary.
type Func struct {
	opaque struct{} // unexported field to disallow conversions
}

func (f *Func) raw() *_lock.Func {
	return (*_lock.Func)(unsafe.Pointer(f))
}

// FuncForPC returns a *Func describing the function that contains the
// given program counter address, or else nil.
func FuncForPC(pc uintptr) *Func {
	return (*Func)(unsafe.Pointer(_lock.Findfunc(pc)))
}

// Name returns the name of the function.
func (f *Func) Name() string {
	return _lock.Gofuncname(f.raw())
}

// Entry returns the entry address of the function.
func (f *Func) Entry() uintptr {
	return f.raw().Entry
}

// FileLine returns the file name and line number of the
// source code corresponding to the program counter pc.
// The result will not be accurate if pc is not a program
// counter within f.
func (f *Func) FileLine(pc uintptr) (file string, line int) {
	// Pass strict=false here, because anyone can call this function,
	// and they might just be wrong about targetpc belonging to f.
	file, line32 := _lock.Funcline1(f.raw(), pc, false)
	return file, int(line32)
}
