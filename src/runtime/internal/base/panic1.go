// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

// Code related to defer, panic and recover.
// TODO: Merge into panic.go.

//uint32 runtime·panicking;
var paniclk Mutex

func startpanic_m() {
	_g_ := Getg()
	if Mheap_.Cachealloc.Size == 0 { // very early
		print("runtime: panic before malloc heap initialized\n")
		_g_.M.Mallocing = 1 // tell rest of panic not to try to malloc
	} else if _g_.M.Mcache == nil { // can happen if called from signal handler or throw
		_g_.M.Mcache = Allocmcache()
	}

	switch _g_.M.dying {
	case 0:
		_g_.M.dying = 1
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
		_g_.M.dying = 2
		print("panic during panic\n")
		Dopanic(0)
		Exit(3)
		fallthrough
	case 2:
		// This is a genuine bug in the runtime, we couldn't even
		// print the stack trace successfully.
		_g_.M.dying = 3
		print("stack trace unavailable\n")
		Exit(4)
		fallthrough
	default:
		// Can't even print!  Just exit.
		Exit(5)
	}
}

var didothers bool
var deadlock Mutex

func dopanic_m(gp *G, pc, sp uintptr) {
	if gp.Sig != 0 {
		print("[signal ", Hex(gp.Sig), " code=", Hex(gp.Sigcode0), " addr=", Hex(gp.Sigcode1), " pc=", Hex(gp.Sigpc), "]\n")
	}

	var docrash bool
	_g_ := Getg()
	if t := gotraceback(&docrash); t > 0 {
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
		crash()
	}

	Exit(2)
}

//go:nosplit
func canpanic(gp *G) bool {
	// Note that g is m->gsignal, different from gp.
	// Note also that g->m can change at preemption, so m can go stale
	// if this function ever makes a function call.
	_g_ := Getg()
	_m_ := _g_.M

	// Is it okay for gp to panic instead of crashing the program?
	// Yes, as long as it is running Go code, not runtime code,
	// and not stuck in a system call.
	if gp == nil || gp != _m_.Curg {
		return false
	}
	if _m_.Locks-_m_.Softfloat != 0 || _m_.Mallocing != 0 || _m_.Throwing != 0 || _m_.Preemptoff != "" || _m_.dying != 0 {
		return false
	}
	status := Readgstatus(gp)
	if status&^Gscan != Grunning || gp.Syscallsp != 0 {
		return false
	}
	if GOOS == "windows" && _m_.Libcallsp != 0 {
		return false
	}
	return true
}
