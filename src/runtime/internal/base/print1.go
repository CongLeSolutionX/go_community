// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

// The compiler knows that a print of a value of this type
// should use printhex instead of printuint (decimal).
type Hex uint64

var debuglock Mutex

// The compiler emits calls to printlock and printunlock around
// the multiple calls that implement a single Go print or println
// statement. Some of the print helpers (printsp, for example)
// call print recursively. There is also the problem of a crash
// happening during the print routines and needing to acquire
// the print lock to print information about the crash.
// For both these reasons, let a thread acquire the printlock 'recursively'.

func Printlock() {
	mp := Getg().M
	mp.Locks++ // do not reschedule between printlock++ and lock(&debuglock).
	mp.printlock++
	if mp.printlock == 1 {
		Lock(&debuglock)
	}
	mp.Locks-- // now we know debuglock is held and holding up mp.locks for us.
}

func Printunlock() {
	mp := Getg().M
	mp.printlock--
	if mp.printlock == 0 {
		Unlock(&debuglock)
	}
}
