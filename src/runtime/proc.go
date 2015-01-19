// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

import (
	_cgo "runtime/internal/cgo"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
)

//go:linkname runtime_init runtime.init
func runtime_init()

//go:linkname main_init main.init
func main_init()

//go:linkname main_main main.main
func main_main()

// The main goroutine.
func main() {
	g := _core.Getg()

	// Racectx of m0->g0 is used only as the parent of the main goroutine.
	// It must not be used for anything else.
	g.M.G0.Racectx = 0

	// Max stack size is 1 GB on 64-bit, 250 MB on 32-bit.
	// Using decimal instead of binary GB and MB because
	// they look nicer in the stack overflow failure message.
	if _core.PtrSize == 8 {
		maxstacksize = 1000000000
	} else {
		maxstacksize = 250000000
	}

	_lock.Systemstack(newsysmon)

	// Lock the main goroutine onto this, the main OS thread,
	// during initialization.  Most programs won't care, but a few
	// do require certain calls to be made by the main thread.
	// Those can arrange for main.main to run in the main thread
	// by calling runtime.LockOSThread during initialization
	// to preserve the lock.
	_cgo.LockOSThread()

	if g.M != &_core.M0 {
		_lock.Gothrow("runtime.main not on m0")
	}

	runtime_init() // must be before defer

	// Defer unlock so that runtime.Goexit during init does the unlock too.
	needUnlock := true
	defer func() {
		if needUnlock {
			_cgo.UnlockOSThread()
		}
	}()

	_lock.Memstats.Enablegc = true // now that runtime is initialized, GC is okay

	if _sched.Iscgo {
		if _sched.Cgo_thread_start == nil {
			_lock.Gothrow("_cgo_thread_start missing")
		}
		if _cgo.Cgo_malloc == nil {
			_lock.Gothrow("_cgo_malloc missing")
		}
		if _cgo.Cgo_free == nil {
			_lock.Gothrow("_cgo_free missing")
		}
		if _lock.GOOS != "windows" {
			if _cgo_setenv == nil {
				_lock.Gothrow("_cgo_setenv missing")
			}
			if _cgo_unsetenv == nil {
				_lock.Gothrow("_cgo_unsetenv missing")
			}
		}
	}

	main_init()

	needUnlock = false
	_cgo.UnlockOSThread()

	main_main()
	if _sched.Raceenabled {
		racefini()
	}

	// Make racy client program work: if panicking on
	// another goroutine at the same time as main returns,
	// let the other goroutine finish printing the panic trace.
	// Once it does, it will exit. See issue 3934.
	if _lock.Panicking != 0 {
		_sched.Gopark(nil, nil, "panicwait")
	}

	_core.Exit(0)
	for {
		var x *int32
		*x = 0
	}
}

// start forcegc helper goroutine
func init() {
	go forcegchelper()
}

func forcegchelper() {
	forcegc.g = _core.Getg()
	forcegc.g.Issystem = true
	for {
		_lock.Lock(&forcegc.lock)
		if forcegc.idle != 0 {
			_lock.Gothrow("forcegc: phase error")
		}
		_lock.Atomicstore(&forcegc.idle, 1)
		_sched.Goparkunlock(&forcegc.lock, "force gc (idle)")
		// this goroutine is explicitly resumed by sysmon
		if _lock.Debug.Gctrace > 0 {
			println("GC forced")
		}
		_gc.Gogc(1)
	}
}

// called from assembly
func badmcall(fn func(*_core.G)) {
	_lock.Gothrow("runtime: mcall called on m->g0 stack")
}

func badmcall2(fn func(*_core.G)) {
	_lock.Gothrow("runtime: mcall function returned")
}

func badreflectcall() {
	panic("runtime: arg size to reflect.call more than 1GB")
}

func lockedOSThread() bool {
	gp := _core.Getg()
	return gp.Lockedm != nil && gp.M.Lockedg != nil
}
