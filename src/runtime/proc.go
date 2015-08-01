// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import _ "unsafe"

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
)

//go:linkname runtime_init runtime.init
func runtime_init()

//go:linkname main_init main.init
func main_init()

// main_init_done is a signal used by cgocallbackg that initialization
// has been completed. It is made before _cgo_notify_runtime_init_done,
// so all cgo calls can rely on it existing. When main_init is complete,
// it is closed, meaning cgocallbackg can reliably receive from it.
var main_init_done chan bool

//go:linkname main_main main.main
func main_main()

// The main goroutine.
func main() {
	g := _base.Getg()

	// Racectx of m0->g0 is used only as the parent of the main goroutine.
	// It must not be used for anything else.
	g.M.G0.Racectx = 0

	// Max stack size is 1 GB on 64-bit, 250 MB on 32-bit.
	// Using decimal instead of binary GB and MB because
	// they look nicer in the stack overflow failure message.
	if _base.PtrSize == 8 {
		maxstacksize = 1000000000
	} else {
		maxstacksize = 250000000
	}

	// Record when the world started.
	_gc.RuntimeInitTime = _base.Nanotime()

	_base.Systemstack(func() {
		_base.Newm(sysmon, nil)
	})

	// Lock the main goroutine onto this, the main OS thread,
	// during initialization.  Most programs won't care, but a few
	// do require certain calls to be made by the main thread.
	// Those can arrange for main.main to run in the main thread
	// by calling runtime.LockOSThread during initialization
	// to preserve the lock.
	lockOSThread()

	if g.M != &_base.M0 {
		_base.Throw("runtime.main not on m0")
	}

	runtime_init() // must be before defer

	// Defer unlock so that runtime.Goexit during init does the unlock too.
	needUnlock := true
	defer func() {
		if needUnlock {
			unlockOSThread()
		}
	}()

	gcenable()

	main_init_done = make(chan bool)
	if _base.Iscgo {
		if _base.Cgo_thread_start == nil {
			_base.Throw("_cgo_thread_start missing")
		}
		if _cgo_malloc == nil {
			_base.Throw("_cgo_malloc missing")
		}
		if _cgo_free == nil {
			_base.Throw("_cgo_free missing")
		}
		if _base.GOOS != "windows" {
			if _cgo_setenv == nil {
				_base.Throw("_cgo_setenv missing")
			}
			if _cgo_unsetenv == nil {
				_base.Throw("_cgo_unsetenv missing")
			}
		}
		if _cgo_notify_runtime_init_done == nil {
			_base.Throw("_cgo_notify_runtime_init_done missing")
		}
		cgocall(_cgo_notify_runtime_init_done, nil)
	}

	main_init()
	close(main_init_done)

	needUnlock = false
	unlockOSThread()

	if _base.Isarchive || _base.Islibrary {
		// A program compiled with -buildmode=c-archive or c-shared
		// has a main, but it is not executed.
		return
	}
	main_main()
	if _base.Raceenabled {
		_race.Racefini()
	}

	// Make racy client program work: if panicking on
	// another goroutine at the same time as main returns,
	// let the other goroutine finish printing the panic trace.
	// Once it does, it will exit. See issue 3934.
	if _base.Panicking != 0 {
		_base.Gopark(nil, nil, "panicwait", _base.TraceEvGoStop, 1)
	}

	_base.Exit(0)
	for {
		var x *int32
		*x = 0
	}
}

// os_beforeExit is called from os.Exit(0).
//go:linkname os_beforeExit os.runtime_beforeExit
func os_beforeExit() {
	if _base.Raceenabled {
		_race.Racefini()
	}
}

// start forcegc helper goroutine
func init() {
	go forcegchelper()
}

func forcegchelper() {
	forcegc.g = _base.Getg()
	for {
		_base.Lock(&forcegc.lock)
		if forcegc.idle != 0 {
			_base.Throw("forcegc: phase error")
		}
		_base.Atomicstore(&forcegc.idle, 1)
		_base.Goparkunlock(&forcegc.lock, "force gc (idle)", _base.TraceEvGoBlock, 1)
		// this goroutine is explicitly resumed by sysmon
		if _base.Debug.Gctrace > 0 {
			println("GC forced")
		}
		_iface.StartGC(_gc.GcBackgroundMode)
	}
}

// called from assembly
func badmcall(fn func(*_base.G)) {
	_base.Throw("runtime: mcall called on m->g0 stack")
}

func badmcall2(fn func(*_base.G)) {
	_base.Throw("runtime: mcall function returned")
}

func badreflectcall() {
	panic("runtime: arg size to reflect.call more than 1GB")
}

func lockedOSThread() bool {
	gp := _base.Getg()
	return gp.Lockedm != nil && gp.M.Lockedg != nil
}
