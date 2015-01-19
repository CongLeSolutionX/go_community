// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
	"unsafe"
)

//go:nosplit
func Startpanic() {
	Systemstack(startpanic_m)
}

//go:nosplit
func Dopanic(unused int) {
	pc := Getcallerpc(unsafe.Pointer(&unused))
	sp := Getcallersp(unsafe.Pointer(&unused))
	gp := _core.Getg()
	Systemstack(func() {
		dopanic_m(gp, pc, sp) // should never return
	})
	*(*int)(nil) = 0
}

//go:nosplit
func throw(s *byte) {
	gp := _core.Getg()
	if gp.M.Throwing == 0 {
		gp.M.Throwing = 1
	}
	Startpanic()
	print("fatal error: ", Gostringnocopy(s), "\n")
	Dopanic(0)
	*(*int)(nil) = 0 // not reached
}

//go:nosplit
func Gothrow(s string) {
	print("fatal error: ", s, "\n")
	gp := _core.Getg()
	if gp.M.Throwing == 0 {
		gp.M.Throwing = 1
	}
	Startpanic()
	Dopanic(0)
	*(*int)(nil) = 0 // not reached
}
