// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// CPU profiling.
// Based on algorithms and data structures used in
// http://code.google.com/p/google-perftools/.
//
// The main difference between this code and the google-perftools
// code is that this code is written to allow copying the profile data
// to an arbitrary io.Writer, while the google-perftools code always
// writes to an operating system file.
//
// The signal handler for the profiling clock tick adds a new stack trace
// to a hash table tracking counts for recent traces.  Most clock ticks
// hit in the cache.  In the event of a cache miss, an entry must be
// evicted from the hash table, copied to a log that will eventually be
// written as profile data.  The google-perftools code flushed the
// log itself during the signal handler.  This code cannot do that, because
// the io.Writer might block or need system calls or locks that are not
// safe to use from within the signal handler.  Instead, we split the log
// into two halves and let the signal handler fill one half while a goroutine
// is writing out the other half.  When the signal handler fills its half, it
// offers to swap with the goroutine.  If the writer is not done with its half,
// we lose the stack trace for this clock tick (and record that loss).
// The goroutine interacts with the signal handler by calling getprofile() to
// get the next log piece to write, implicitly handing back the last log
// piece it obtained.
//
// The state of this dance between the signal handler and the goroutine
// is encoded in the Profile.handoff field.  If handoff == 0, then the goroutine
// is not using either log half and is waiting (or will soon be waiting) for
// a new piece by calling notesleep(&p->wait).  If the signal handler
// changes handoff from 0 to non-zero, it must call notewakeup(&p->wait)
// to wake the goroutine.  The value indicates the number of entries in the
// log half being handed off.  The goroutine leaves the non-zero value in
// place until it has finished processing the log half and then flips the number
// back to zero.  Setting the high bit in handoff means that the profiling is over,
// and the goroutine is now in charge of flushing the data left in the hash table
// to the log and returning that data.
//
// The handoff field is manipulated using atomic operations.
// For the most part, the manipulation of handoff is orderly: if handoff == 0
// then the signal handler owns it and can change it to non-zero.
// If handoff != 0 then the goroutine owns it and can change it to zero.
// If that were the end of the story then we would not need to manipulate
// handoff using atomic operations.  The operations are needed, however,
// in order to let the log closer set the high bit to indicate "EOF" safely
// in the situation when normally the goroutine "owns" handoff.

package prof

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

var (
	cpuprofLock _core.Mutex
)

func setcpuprofilerate(hz int32) {
	_lock.Systemstack(func() {
		setcpuprofilerate_m(hz)
	})
}

// SetCPUProfileRate sets the CPU profiling rate to hz samples per second.
// If hz <= 0, SetCPUProfileRate turns off profiling.
// If the profiler is on, the rate cannot be changed without first turning it off.
//
// Most clients should use the runtime/pprof package or
// the testing package's -test.cpuprofile flag instead of calling
// SetCPUProfileRate directly.
func SetCPUProfileRate(hz int) {
	// Clamp hz to something reasonable.
	if hz < 0 {
		hz = 0
	}
	if hz > 1000000 {
		hz = 1000000
	}

	_lock.Lock(&cpuprofLock)
	if hz > 0 {
		if _sched.Cpuprof == nil {
			_sched.Cpuprof = (*_sched.CpuProfile)(_lock.SysAlloc(unsafe.Sizeof(_sched.CpuProfile{}), &_lock.Memstats.Other_sys))
			if _sched.Cpuprof == nil {
				print("runtime: cpu profiling cannot allocate memory\n")
				_lock.Unlock(&cpuprofLock)
				return
			}
		}
		if _sched.Cpuprof.On || _sched.Cpuprof.Handoff != 0 {
			print("runtime: cannot set cpu profile rate until previous profile has finished.\n")
			_lock.Unlock(&cpuprofLock)
			return
		}

		_sched.Cpuprof.On = true
		// pprof binary header format.
		// http://code.google.com/p/google-perftools/source/browse/trunk/src/profiledata.cc#117
		p := &_sched.Cpuprof.Log[0]
		p[0] = 0                 // count for header
		p[1] = 3                 // depth for header
		p[2] = 0                 // version number
		p[3] = uintptr(1e6 / hz) // period (microseconds)
		p[4] = 0
		_sched.Cpuprof.Nlog = 5
		_sched.Cpuprof.Toggle = 0
		_sched.Cpuprof.Wholding = false
		_sched.Cpuprof.Wtoggle = 0
		_sched.Cpuprof.Flushing = false
		_sched.Cpuprof.EodSent = false
		_sched.Noteclear(&_sched.Cpuprof.Wait)

		setcpuprofilerate(int32(hz))
	} else if _sched.Cpuprof != nil && _sched.Cpuprof.On {
		setcpuprofilerate(0)
		_sched.Cpuprof.On = false

		// Now add is not running anymore, and getprofile owns the entire log.
		// Set the high bit in prof->handoff to tell getprofile.
		for {
			n := _sched.Cpuprof.Handoff
			if n&0x80000000 != 0 {
				print("runtime: setcpuprofile(off) twice\n")
			}
			if _sched.Cas(&_sched.Cpuprof.Handoff, n, n|0x80000000) {
				if n == 0 {
					// we did the transition from 0 -> nonzero so we wake getprofile
					_sched.Notewakeup(&_sched.Cpuprof.Wait)
				}
				break
			}
		}
	}
	_lock.Unlock(&cpuprofLock)
}
