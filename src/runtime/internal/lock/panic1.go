// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
)

// Code related to defer, panic and recover.
// TODO: Merge into panic.go.

//uint32 runtime·panicking;
var paniclk _core.Mutex

func startpanic_m() {
	_g_ := _core.Getg()
	if Mheap_.Cachealloc.Size == 0 { // very early
		print("runtime: panic before malloc heap initialized\n")
		_g_.M.Mallocing = 1 // tell rest of panic not to try to malloc
	} else if _g_.M.Mcache == nil { // can happen if called from signal handler or throw
		_g_.M.Mcache = Allocmcache()
	}

	switch _g_.M.Dying {
	case 0:
		_g_.M.Dying = 1
		if _g_ != nil {
			_g_.Writebuf = nil
		}
		Xadd(&Panicking, 1)
		Lock(&paniclk)
		if Debug.Schedtrace > 0 || Debug.Scheddetail > 0 {
			Schedtrace(true)
		}
		freezetheworld()
		return
	case 1:
		// Something failed while panicing, probably the print of the
		// argument to panic().  Just print a stack trace and exit.
		_g_.M.Dying = 2
		print("panic during panic\n")
		Dopanic(0)
		_core.Exit(3)
		fallthrough
	case 2:
		// This is a genuine bug in the runtime, we couldn't even
		// print the stack trace successfully.
		_g_.M.Dying = 3
		print("stack trace unavailable\n")
		_core.Exit(4)
		fallthrough
	default:
		// Can't even print!  Just exit.
		_core.Exit(5)
	}
}

var didothers bool
var deadlock _core.Mutex

func dopanic_m(gp *_core.G, pc, sp uintptr) {
	if gp.Sig != 0 {
		print("[signal ", _core.Hex(gp.Sig), " code=", _core.Hex(gp.Sigcode0), " addr=", _core.Hex(gp.Sigcode1), " pc=", _core.Hex(gp.Sigpc), "]\n")
	}

	var docrash bool
	_g_ := _core.Getg()
	if t := Gotraceback(&docrash); t > 0 {
		if gp != gp.M.G0 {
			print("\n")
			Goroutineheader(gp)
			Traceback(pc, sp, 0, gp)
		} else if t >= 2 || _g_.M.Throwing > 0 {
			print("\nruntime stack:\n")
			Traceback(pc, sp, 0, gp)
		}
		if !didothers {
			didothers = true
			Tracebackothers(gp)
		}
	}
	Unlock(&paniclk)

	if Xadd(&Panicking, -1) != 0 {
		// Some other m is panicking too.
		// Let it print what it needs to print.
		// Wait forever without chewing up cpu.
		// It will exit when it's done.
		Lock(&deadlock)
		Lock(&deadlock)
	}

	if docrash {
		Crash()
	}

	_core.Exit(2)
}
