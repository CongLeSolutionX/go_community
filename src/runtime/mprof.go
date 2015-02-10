// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc profiling.
// Patterned after tcmalloc's algorithms; shorter code.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_prof "runtime/internal/prof"
	_sem "runtime/internal/sem"
	"unsafe"
)

func iterate_memprof(fn func(*_sem.Bucket, uintptr, *uintptr, uintptr, uintptr, uintptr)) {
	_lock.Lock(&_sem.Proflock)
	for b := _sem.Mbuckets; b != nil; b = b.Allnext {
		mp := b.Mp()
		fn(b, uintptr(b.Nstk), &b.Stk()[0], b.Size, mp.Allocs, mp.Frees)
	}
	_lock.Unlock(&_sem.Proflock)
}

// GoroutineProfile returns n, the number of records in the active goroutine stack profile.
// If len(p) >= n, GoroutineProfile copies the profile into p and returns n, true.
// If len(p) < n, GoroutineProfile does not change p and returns n, false.
//
// Most clients should use the runtime/pprof package instead
// of calling GoroutineProfile directly.
func GoroutineProfile(p []_prof.StackRecord) (n int, ok bool) {

	n = NumGoroutine()
	if n <= len(p) {
		gp := _core.Getg()
		_sem.Semacquire(&_gc.Worldsema, false)
		gp.M.Preemptoff = "profile"
		_lock.Systemstack(_gc.Stoptheworld)

		n = NumGoroutine()
		if n <= len(p) {
			ok = true
			r := p
			sp := _lock.Getcallersp(unsafe.Pointer(&p))
			pc := _lock.Getcallerpc(unsafe.Pointer(&p))
			_lock.Systemstack(func() {
				saveg(pc, sp, gp, &r[0])
			})
			r = r[1:]
			for _, gp1 := range _lock.Allgs {
				if gp1 == gp || _lock.Readgstatus(gp1) == _lock.Gdead {
					continue
				}
				saveg(^uintptr(0), ^uintptr(0), gp1, &r[0])
				r = r[1:]
			}
		}

		gp.M.Preemptoff = ""
		_sem.Semrelease(&_gc.Worldsema)
		_lock.Systemstack(_gc.Starttheworld)
	}

	return n, ok
}

func saveg(pc, sp uintptr, gp *_core.G, r *_prof.StackRecord) {
	n := _lock.Gentraceback(pc, sp, 0, gp, 0, &r.Stack0[0], len(r.Stack0), nil, nil, 0)
	if n < len(r.Stack0) {
		r.Stack0[n] = 0
	}
}

// Stack formats a stack trace of the calling goroutine into buf
// and returns the number of bytes written to buf.
// If all is true, Stack formats stack traces of all other goroutines
// into buf after the trace for the current goroutine.
func Stack(buf []byte, all bool) int {
	if all {
		_sem.Semacquire(&_gc.Worldsema, false)
		gp := _core.Getg()
		gp.M.Preemptoff = "stack trace"
		_lock.Systemstack(_gc.Stoptheworld)
	}

	n := 0
	if len(buf) > 0 {
		gp := _core.Getg()
		sp := _lock.Getcallersp(unsafe.Pointer(&buf))
		pc := _lock.Getcallerpc(unsafe.Pointer(&buf))
		_lock.Systemstack(func() {
			g0 := _core.Getg()
			g0.Writebuf = buf[0:0:len(buf)]
			_lock.Goroutineheader(gp)
			_lock.Traceback(pc, sp, 0, gp)
			if all {
				_lock.Tracebackothers(gp)
			}
			n = len(g0.Writebuf)
			g0.Writebuf = nil
		})
	}

	if all {
		gp := _core.Getg()
		gp.M.Preemptoff = ""
		_sem.Semrelease(&_gc.Worldsema)
		_lock.Systemstack(_gc.Starttheworld)
	}
	return n
}
