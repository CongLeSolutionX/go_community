// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go execution tracer.
// The tracer captures a wide range of execution events like goroutine
// creation/blocking/unblocking, syscall enter/exit/block, GC-related events,
// changes of heap size, processor start/stop, etc and writes them to a buffer
// in a compact form. A precise nanosecond-precision timestamp and a stack
// trace is captured for most events.
// See http://golang.org/s/go15trace for more info.

package runtime

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// StartTrace enables tracing for the current process.
// While tracing, the data will be buffered and available via ReadTrace.
// StartTrace returns an error if tracing is already enabled.
// Most clients should use the runtime/pprof package or the testing package's
// -test.trace flag instead of calling StartTrace directly.
func StartTrace() error {
	// Stop the world, so that we can take a consistent snapshot
	// of all goroutines at the beginning of the trace.
	_sem.Semacquire(&_gc.Worldsema, false)
	_g_ := _core.Getg()
	_g_.M.Preemptoff = "start tracing"
	_lock.Systemstack(_gc.Stoptheworld)

	// We are in stop-the-world, but syscalls can finish and write to trace concurrently.
	// Exitsyscall could check trace.enabled long before and then suddenly wake up
	// and decide to write to trace at a random point in time.
	// However, such syscall will use the global trace.buf buffer, because we've
	// acquired all p's by doing stop-the-world. So this protects us from such races.
	_lock.Lock(&_sched.Trace.BufLock)

	if _sched.Trace.Enabled || _sched.Trace.Shutdown {
		_lock.Unlock(&_sched.Trace.BufLock)
		_g_.M.Preemptoff = ""
		_sem.Semrelease(&_gc.Worldsema)
		_lock.Systemstack(_gc.Starttheworld)
		return _sched.ErrorString("tracing is already enabled")
	}

	_sched.Trace.TicksStart = _sched.Cputicks()
	_sched.Trace.TimeStart = _lock.Nanotime()
	_sched.Trace.Enabled = true
	_sched.Trace.HeaderWritten = false
	_sched.Trace.FooterWritten = false

	for _, gp := range _lock.Allgs {
		status := _lock.Readgstatus(gp)
		if status != _lock.Gdead {
			traceGoCreate(gp, gp.Startpc)
		}
		if status == _lock.Gwaiting {
			_sched.TraceEvent(_sched.TraceEvGoWaiting, false, uint64(gp.Goid))
		}
		if status == _lock.Gsyscall {
			_sched.TraceEvent(_sched.TraceEvGoInSyscall, false, uint64(gp.Goid))
		}
	}
	_sched.TraceProcStart()
	_sched.TraceGoStart()

	_lock.Unlock(&_sched.Trace.BufLock)

	_g_.M.Preemptoff = ""
	_sem.Semrelease(&_gc.Worldsema)
	_lock.Systemstack(_gc.Starttheworld)
	return nil
}

// StopTrace stops tracing, if it was previously enabled.
// StopTrace only returns after all the reads for the trace have completed.
func StopTrace() {
	// Stop the world so that we can collect the trace buffers from all p's below,
	// and also to avoid races with traceEvent.
	_sem.Semacquire(&_gc.Worldsema, false)
	_g_ := _core.Getg()
	_g_.M.Preemptoff = "stop tracing"
	_lock.Systemstack(_gc.Stoptheworld)

	// See the comment in StartTrace.
	_lock.Lock(&_sched.Trace.BufLock)

	if !_sched.Trace.Enabled {
		_lock.Unlock(&_sched.Trace.BufLock)
		_g_.M.Preemptoff = ""
		_sem.Semrelease(&_gc.Worldsema)
		_lock.Systemstack(_gc.Starttheworld)
		return
	}

	_sched.TraceGoSched()
	_sched.TraceGoStart()

	for _, p := range &_lock.Allp {
		if p == nil {
			break
		}
		buf := p.Tracebuf
		if buf != nil {
			_sched.TraceFullQueue(buf)
			p.Tracebuf = nil
		}
	}
	if _sched.Trace.Buf != nil && len(_sched.Trace.Buf.Buf) != 0 {
		buf := _sched.Trace.Buf
		_sched.Trace.Buf = nil
		_sched.TraceFullQueue(buf)
	}

	for {
		_sched.Trace.TicksEnd = _sched.Cputicks()
		_sched.Trace.TimeEnd = _lock.Nanotime()
		// Windows time can tick only every 15ms, wait for at least one tick.
		if _sched.Trace.TimeEnd != _sched.Trace.TimeStart {
			break
		}
		_core.Osyield()
	}

	_sched.Trace.Enabled = false
	_sched.Trace.Shutdown = true
	_sched.Trace.StackTab.Dump()

	_lock.Unlock(&_sched.Trace.BufLock)

	_g_.M.Preemptoff = ""
	_sem.Semrelease(&_gc.Worldsema)
	_lock.Systemstack(_gc.Starttheworld)

	// The world is started but we've set trace.shutdown, so new tracing can't start.
	// Wait for the trace reader to flush pending buffers and stop.
	_sem.Semacquire(&_sched.Trace.ShutdownSema, false)
	if _sched.Raceenabled {
		_sched.Raceacquire(unsafe.Pointer(&_sched.Trace.ShutdownSema))
	}

	// The lock protects us from races with StartTrace/StopTrace because they do stop-the-world.
	_lock.Lock(&_sched.Trace.Lock)
	for _, p := range &_lock.Allp {
		if p == nil {
			break
		}
		if p.Tracebuf != nil {
			_lock.Throw("trace: non-empty trace buffer in proc")
		}
	}
	if _sched.Trace.Buf != nil {
		_lock.Throw("trace: non-empty global trace buffer")
	}
	if _sched.Trace.FullHead != nil || _sched.Trace.FullTail != nil {
		_lock.Throw("trace: non-empty full trace buffer")
	}
	if _sched.Trace.Reading != nil || _sched.Trace.Reader != nil {
		_lock.Throw("trace: reading after shutdown")
	}
	for _sched.Trace.Empty != nil {
		buf := _sched.Trace.Empty
		_sched.Trace.Empty = buf.Link
		_sched.SysFree(unsafe.Pointer(buf), unsafe.Sizeof(*buf), &_lock.Memstats.Other_sys)
	}
	_sched.Trace.Shutdown = false
	_lock.Unlock(&_sched.Trace.Lock)
}

// ReadTrace returns the next chunk of binary tracing data, blocking until data
// is available. If tracing is turned off and all the data accumulated while it
// was on has been returned, ReadTrace returns nil. The caller must copy the
// returned data before calling ReadTrace again.
// ReadTrace must be called from one goroutine at a time.
func ReadTrace() []byte {
	// This function may need to lock trace.lock recursively
	// (goparkunlock -> traceGoPark -> traceEvent -> traceFlush).
	// To allow this we use trace.lockOwner.
	// Also this function must not allocate while holding trace.lock:
	// allocation can call heap allocate, which will try to emit a trace
	// event while holding heap lock.
	_lock.Lock(&_sched.Trace.Lock)
	_sched.Trace.LockOwner = _core.Getg()

	if _sched.Trace.Reader != nil {
		// More than one goroutine reads trace. This is bad.
		// But we rather do not crash the program because of tracing,
		// because tracing can be enabled at runtime on prod servers.
		_sched.Trace.LockOwner = nil
		_lock.Unlock(&_sched.Trace.Lock)
		println("runtime: ReadTrace called from multiple goroutines simultaneously")
		return nil
	}
	// Recycle the old buffer.
	if buf := _sched.Trace.Reading; buf != nil {
		buf.Link = _sched.Trace.Empty
		_sched.Trace.Empty = buf
		_sched.Trace.Reading = nil
	}
	// Write trace header.
	if !_sched.Trace.HeaderWritten {
		_sched.Trace.HeaderWritten = true
		_sched.Trace.LockOwner = nil
		_lock.Unlock(&_sched.Trace.Lock)
		return []byte("gotrace\x00")
	}
	// Wait for new data.
	if _sched.Trace.FullHead == nil && !_sched.Trace.Shutdown {
		_sched.Trace.Reader = _core.Getg()
		_sched.Goparkunlock(&_sched.Trace.Lock, "trace reader (blocked)", _sched.TraceEvGoBlock)
		_lock.Lock(&_sched.Trace.Lock)
	}
	// Write a buffer.
	if _sched.Trace.FullHead != nil {
		buf := traceFullDequeue()
		_sched.Trace.Reading = buf
		_sched.Trace.LockOwner = nil
		_lock.Unlock(&_sched.Trace.Lock)
		return buf.Buf
	}
	// Write footer with timer frequency.
	if !_sched.Trace.FooterWritten {
		_sched.Trace.FooterWritten = true
		// Use float64 because (trace.ticksEnd - trace.ticksStart) * 1e9 can overflow int64.
		freq := float64(_sched.Trace.TicksEnd-_sched.Trace.TicksStart) * 1e9 / float64(_sched.Trace.TimeEnd-_sched.Trace.TimeStart) / _core.TraceTickDiv
		_sched.Trace.LockOwner = nil
		_lock.Unlock(&_sched.Trace.Lock)
		var data []byte
		data = append(data, _sched.TraceEvFrequency|0<<_core.TraceArgCountShift)
		data = _sched.TraceAppend(data, uint64(freq))
		if _sched.Timers.Gp != nil {
			data = append(data, _sched.TraceEvTimerGoroutine|0<<_core.TraceArgCountShift)
			data = _sched.TraceAppend(data, uint64(_sched.Timers.Gp.Goid))
		}
		return data
	}
	// Done.
	if _sched.Trace.Shutdown {
		_sched.Trace.LockOwner = nil
		_lock.Unlock(&_sched.Trace.Lock)
		if _sched.Raceenabled {
			// Model synchronization on trace.shutdownSema, which race
			// detector does not see. This is required to avoid false
			// race reports on writer passed to pprof.StartTrace.
			_channels.Racerelease(unsafe.Pointer(&_sched.Trace.ShutdownSema))
		}
		// trace.enabled is already reset, so can call traceable functions.
		_sem.Semrelease(&_sched.Trace.ShutdownSema)
		return nil
	}
	// Also bad, but see the comment above.
	_sched.Trace.LockOwner = nil
	_lock.Unlock(&_sched.Trace.Lock)
	println("runtime: spurious wakeup of trace reader")
	return nil
}

// traceFullDequeue dequeues from queue of full buffers.
func traceFullDequeue() *_core.TraceBuf {
	buf := _sched.Trace.FullHead
	if buf == nil {
		return nil
	}
	_sched.Trace.FullHead = buf.Link
	if _sched.Trace.FullHead == nil {
		_sched.Trace.FullTail = nil
	}
	buf.Link = nil
	return buf
}

func traceGoCreate(newg *_core.G, pc uintptr) {
	_sched.TraceEvent(_sched.TraceEvGoCreate, true, uint64(newg.Goid), uint64(pc))
}

func traceGoEnd() {
	_sched.TraceEvent(_sched.TraceEvGoEnd, false)
}

func traceGoPreempt() {
	_sched.TraceEvent(_sched.TraceEvGoPreempt, true)
}

func traceGoStop() {
	_sched.TraceEvent(_sched.TraceEvGoStop, true)
}
