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
// a new piece by calling notesleep(&p.wait).  If the signal handler
// changes handoff from 0 to non-zero, it must call notewakeup(&p.wait)
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

package runtime

import (
	_base "runtime/internal/base"
	"unsafe"
)

var (
	cpuprofLock _base.Mutex
)

func setcpuprofilerate(hz int32) {
	_base.Systemstack(func() {
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

	_base.Lock(&cpuprofLock)
	if hz > 0 {
		if _base.Cpuprof == nil {
			_base.Cpuprof = (*_base.CpuProfile)(_base.SysAlloc(unsafe.Sizeof(_base.CpuProfile{}), &_base.Memstats.Other_sys))
			if _base.Cpuprof == nil {
				print("runtime: cpu profiling cannot allocate memory\n")
				_base.Unlock(&cpuprofLock)
				return
			}
		}
		if _base.Cpuprof.On || _base.Cpuprof.Handoff != 0 {
			print("runtime: cannot set cpu profile rate until previous profile has finished.\n")
			_base.Unlock(&cpuprofLock)
			return
		}

		_base.Cpuprof.On = true
		// pprof binary header format.
		// http://code.google.com/p/google-perftools/source/browse/trunk/src/profiledata.cc#117
		p := &_base.Cpuprof.Log[0]
		p[0] = 0                 // count for header
		p[1] = 3                 // depth for header
		p[2] = 0                 // version number
		p[3] = uintptr(1e6 / hz) // period (microseconds)
		p[4] = 0
		_base.Cpuprof.Nlog = 5
		_base.Cpuprof.Toggle = 0
		_base.Cpuprof.Wholding = false
		_base.Cpuprof.Wtoggle = 0
		_base.Cpuprof.Flushing = false
		_base.Cpuprof.EodSent = false
		_base.Noteclear(&_base.Cpuprof.Wait)

		setcpuprofilerate(int32(hz))
	} else if _base.Cpuprof != nil && _base.Cpuprof.On {
		setcpuprofilerate(0)
		_base.Cpuprof.On = false

		// Now add is not running anymore, and getprofile owns the entire log.
		// Set the high bit in cpuprof.handoff to tell getprofile.
		for {
			n := _base.Cpuprof.Handoff
			if n&0x80000000 != 0 {
				print("runtime: setcpuprofile(off) twice\n")
			}
			if _base.Cas(&_base.Cpuprof.Handoff, n, n|0x80000000) {
				if n == 0 {
					// we did the transition from 0 -> nonzero so we wake getprofile
					_base.Notewakeup(&_base.Cpuprof.Wait)
				}
				break
			}
		}
	}
	_base.Unlock(&cpuprofLock)
}

// CPUProfile returns the next chunk of binary CPU profiling stack trace data,
// blocking until data is available.  If profiling is turned off and all the profile
// data accumulated while it was on has been returned, CPUProfile returns nil.
// The caller must save the returned data before calling CPUProfile again.
//
// Most clients should use the runtime/pprof package or
// the testing package's -test.cpuprofile flag instead of calling
// CPUProfile directly.
func CPUProfile() []byte {
	return _base.Cpuprof.Getprofile()
}

//go:linkname runtime_pprof_runtime_cyclesPerSecond runtime/pprof.runtime_cyclesPerSecond
func runtime_pprof_runtime_cyclesPerSecond() int64 {
	return tickspersecond()
}
