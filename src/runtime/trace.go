// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go execution tracer.
// The tracer captures a wide range of execution events like goroutine
// creation/blocking/unblocking, syscall enter/exit/block, GC-related events,
// changes of heap size, processor start/stop, etc and writes them to a buffer
// in a compact form. A precise nanosecond-precision timestamp and a stack
// trace is captured for most events.
// See https://golang.org/s/go15trace for more info.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

// StartTrace enables tracing for the current process.
// While tracing, the data will be buffered and available via ReadTrace.
// StartTrace returns an error if tracing is already enabled.
// Most clients should use the runtime/trace package or the testing package's
// -test.trace flag instead of calling StartTrace directly.
func StartTrace() error {
	// Stop the world, so that we can take a consistent snapshot
	// of all goroutines at the beginning of the trace.
	stopTheWorld("start tracing")

	// We are in stop-the-world, but syscalls can finish and write to trace concurrently.
	// Exitsyscall could check trace.enabled long before and then suddenly wake up
	// and decide to write to trace at a random point in time.
	// However, such syscall will use the global trace.buf buffer, because we've
	// acquired all p's by doing stop-the-world. So this protects us from such races.
	_base.Lock(&_base.Trace.BufLock)

	if _base.Trace.Enabled || _base.Trace.Shutdown {
		_base.Unlock(&_base.Trace.BufLock)
		startTheWorld()
		return _base.ErrorString("tracing is already enabled")
	}

	_base.Trace.SeqStart, _base.Trace.TicksStart = _base.Tracestamp()
	_base.Trace.TimeStart = _base.Nanotime()
	_base.Trace.HeaderWritten = false
	_base.Trace.FooterWritten = false

	// Can't set trace.enabled yet. While the world is stopped, exitsyscall could
	// already emit a delayed event (see exitTicks in exitsyscall) if we set trace.enabled here.
	// That would lead to an inconsistent trace:
	// - either GoSysExit appears before EvGoInSyscall,
	// - or GoSysExit appears for a goroutine for which we don't emit EvGoInSyscall below.
	// To instruct traceEvent that it must not ignore events below, we set startingtrace.
	// trace.enabled is set afterwards once we have emitted all preliminary events.
	_g_ := _base.Getg()
	_g_.M.Startingtrace = true
	for _, gp := range _base.Allgs {
		status := _base.Readgstatus(gp)
		if status != _base.Gdead {
			traceGoCreate(gp, gp.Startpc)
		}
		if status == _base.Gwaiting {
			_base.TraceEvent(_base.TraceEvGoWaiting, -1, uint64(gp.Goid))
		}
		if status == _base.Gsyscall {
			_base.TraceEvent(_base.TraceEvGoInSyscall, -1, uint64(gp.Goid))
		} else {
			gp.Sysblocktraced = false
		}
	}
	_base.TraceProcStart()
	_base.TraceGoStart()
	_g_.M.Startingtrace = false
	_base.Trace.Enabled = true

	_base.Unlock(&_base.Trace.BufLock)

	startTheWorld()
	return nil
}

// StopTrace stops tracing, if it was previously enabled.
// StopTrace only returns after all the reads for the trace have completed.
func StopTrace() {
	// Stop the world so that we can collect the trace buffers from all p's below,
	// and also to avoid races with traceEvent.
	stopTheWorld("stop tracing")

	// See the comment in StartTrace.
	_base.Lock(&_base.Trace.BufLock)

	if !_base.Trace.Enabled {
		_base.Unlock(&_base.Trace.BufLock)
		startTheWorld()
		return
	}

	_gc.TraceGoSched()

	for _, p := range &_base.Allp {
		if p == nil {
			break
		}
		buf := p.Tracebuf
		if buf != nil {
			_base.TraceFullQueue(buf)
			p.Tracebuf = nil
		}
	}
	if _base.Trace.Buf != nil && len(_base.Trace.Buf.Buf) != 0 {
		buf := _base.Trace.Buf
		_base.Trace.Buf = nil
		_base.TraceFullQueue(buf)
	}

	for {
		_base.Trace.TicksEnd = _base.Cputicks()
		_base.Trace.TimeEnd = _base.Nanotime()
		// Windows time can tick only every 15ms, wait for at least one tick.
		if _base.Trace.TimeEnd != _base.Trace.TimeStart {
			break
		}
		_base.Osyield()
	}

	_base.Trace.Enabled = false
	_base.Trace.Shutdown = true
	_base.Trace.StackTab.Dump()

	_base.Unlock(&_base.Trace.BufLock)

	startTheWorld()

	// The world is started but we've set trace.shutdown, so new tracing can't start.
	// Wait for the trace reader to flush pending buffers and stop.
	_gc.Semacquire(&_base.Trace.ShutdownSema, false)
	if _base.Raceenabled {
		_iface.Raceacquire(unsafe.Pointer(&_base.Trace.ShutdownSema))
	}

	// The lock protects us from races with StartTrace/StopTrace because they do stop-the-world.
	_base.Lock(&_base.Trace.Lock)
	for _, p := range &_base.Allp {
		if p == nil {
			break
		}
		if p.Tracebuf != nil {
			_base.Throw("trace: non-empty trace buffer in proc")
		}
	}
	if _base.Trace.Buf != nil {
		_base.Throw("trace: non-empty global trace buffer")
	}
	if _base.Trace.FullHead != nil || _base.Trace.FullTail != nil {
		_base.Throw("trace: non-empty full trace buffer")
	}
	if _base.Trace.Reading != nil || _base.Trace.Reader != nil {
		_base.Throw("trace: reading after shutdown")
	}
	for _base.Trace.Empty != nil {
		buf := _base.Trace.Empty
		_base.Trace.Empty = buf.Link
		_base.SysFree(unsafe.Pointer(buf), unsafe.Sizeof(*buf), &_base.Memstats.Other_sys)
	}
	_base.Trace.Shutdown = false
	_base.Unlock(&_base.Trace.Lock)
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
	_base.Lock(&_base.Trace.Lock)
	_base.Trace.LockOwner = _base.Getg()

	if _base.Trace.Reader != nil {
		// More than one goroutine reads trace. This is bad.
		// But we rather do not crash the program because of tracing,
		// because tracing can be enabled at runtime on prod servers.
		_base.Trace.LockOwner = nil
		_base.Unlock(&_base.Trace.Lock)
		println("runtime: ReadTrace called from multiple goroutines simultaneously")
		return nil
	}
	// Recycle the old buffer.
	if buf := _base.Trace.Reading; buf != nil {
		buf.Link = _base.Trace.Empty
		_base.Trace.Empty = buf
		_base.Trace.Reading = nil
	}
	// Write trace header.
	if !_base.Trace.HeaderWritten {
		_base.Trace.HeaderWritten = true
		_base.Trace.LockOwner = nil
		_base.Unlock(&_base.Trace.Lock)
		return []byte("go 1.5 trace\x00\x00\x00\x00")
	}
	// Wait for new data.
	if _base.Trace.FullHead == nil && !_base.Trace.Shutdown {
		_base.Trace.Reader = _base.Getg()
		_base.Goparkunlock(&_base.Trace.Lock, "trace reader (blocked)", _base.TraceEvGoBlock, 2)
		_base.Lock(&_base.Trace.Lock)
	}
	// Write a buffer.
	if _base.Trace.FullHead != nil {
		buf := traceFullDequeue()
		_base.Trace.Reading = buf
		_base.Trace.LockOwner = nil
		_base.Unlock(&_base.Trace.Lock)
		return buf.Buf
	}
	// Write footer with timer frequency.
	if !_base.Trace.FooterWritten {
		_base.Trace.FooterWritten = true
		// Use float64 because (trace.ticksEnd - trace.ticksStart) * 1e9 can overflow int64.
		freq := float64(_base.Trace.TicksEnd-_base.Trace.TicksStart) * 1e9 / float64(_base.Trace.TimeEnd-_base.Trace.TimeStart) / _base.TraceTickDiv
		_base.Trace.LockOwner = nil
		_base.Unlock(&_base.Trace.Lock)
		var data []byte
		data = append(data, _base.TraceEvFrequency|0<<_base.TraceArgCountShift)
		data = _base.TraceAppend(data, uint64(freq))
		data = _base.TraceAppend(data, 0)
		if _base.Timers.Gp != nil {
			data = append(data, _base.TraceEvTimerGoroutine|0<<_base.TraceArgCountShift)
			data = _base.TraceAppend(data, uint64(_base.Timers.Gp.Goid))
			data = _base.TraceAppend(data, 0)
		}
		return data
	}
	// Done.
	if _base.Trace.Shutdown {
		_base.Trace.LockOwner = nil
		_base.Unlock(&_base.Trace.Lock)
		if _base.Raceenabled {
			// Model synchronization on trace.shutdownSema, which race
			// detector does not see. This is required to avoid false
			// race reports on writer passed to trace.Start.
			_race.Racerelease(unsafe.Pointer(&_base.Trace.ShutdownSema))
		}
		// trace.enabled is already reset, so can call traceable functions.
		_gc.Semrelease(&_base.Trace.ShutdownSema)
		return nil
	}
	// Also bad, but see the comment above.
	_base.Trace.LockOwner = nil
	_base.Unlock(&_base.Trace.Lock)
	println("runtime: spurious wakeup of trace reader")
	return nil
}

// traceFullDequeue dequeues from queue of full buffers.
func traceFullDequeue() *_base.TraceBuf {
	buf := _base.Trace.FullHead
	if buf == nil {
		return nil
	}
	_base.Trace.FullHead = buf.Link
	if _base.Trace.FullHead == nil {
		_base.Trace.FullTail = nil
	}
	buf.Link = nil
	return buf
}

func traceGoCreate(newg *_base.G, pc uintptr) {
	_base.TraceEvent(_base.TraceEvGoCreate, 2, uint64(newg.Goid), uint64(pc))
}

func traceGoEnd() {
	_base.TraceEvent(_base.TraceEvGoEnd, -1)
}

func traceGoPreempt() {
	_base.TraceEvent(_base.TraceEvGoPreempt, 1)
}
