// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package finalize

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_ifacestuff "runtime/internal/ifacestuff"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

var fingCreate uint32

func Createfing() {
	// start the finalizer goroutine exactly once
	if fingCreate == 0 && _sched.Cas(&fingCreate, 0, 1) {
		go runfinq()
	}
}

// This is the goroutine that runs all of the finalizers
func runfinq() {
	var (
		frame    unsafe.Pointer
		framecap uintptr
	)

	for {
		_lock.Lock(&_sched.Finlock)
		fb := _gc.Finq
		_gc.Finq = nil
		if fb == nil {
			gp := _core.Getg()
			_core.Fing = gp
			_sched.Fingwait = true
			gp.Issystem = true
			_sched.Goparkunlock(&_sched.Finlock, "finalizer wait", _sched.TraceEvGoBlock)
			gp.Issystem = false
			continue
		}
		_lock.Unlock(&_sched.Finlock)
		if _sched.Raceenabled {
			racefingo()
		}
		for fb != nil {
			for i := fb.Cnt; i > 0; i-- {
				f := (*_gc.Finalizer)(_core.Add(unsafe.Pointer(&fb.Fin), uintptr(i-1)*unsafe.Sizeof(_gc.Finalizer{})))

				framesz := unsafe.Sizeof((interface{})(nil)) + uintptr(f.Nret)
				if framecap < framesz {
					// The frame does not contain pointers interesting for GC,
					// all not yet finalized objects are stored in finq.
					// If we do not mark it as FlagNoScan,
					// the last finalized object is not collected.
					frame = _maps.Mallocgc(framesz, nil, _sched.XFlagNoScan)
					framecap = framesz
				}

				if f.Fint == nil {
					_lock.Throw("missing type in runfinq")
				}
				switch f.Fint.Kind & _channels.KindMask {
				case _channels.KindPtr:
					// direct use of pointer
					*(*unsafe.Pointer)(frame) = f.Arg
				case _channels.KindInterface:
					ityp := (*_core.Interfacetype)(unsafe.Pointer(f.Fint))
					// set up with empty interface
					(*_core.Eface)(frame).Type = &f.Ot.Typ
					(*_core.Eface)(frame).Data = f.Arg
					if len(ityp.Mhdr) != 0 {
						// convert to interface with methods
						// this conversion is guaranteed to succeed - we checked in SetFinalizer
						_ifacestuff.AssertE2I(ityp, *(*interface{})(frame), (*_ifacestuff.FInterface)(frame))
					}
				default:
					_lock.Throw("bad kind in runfinq")
				}
				Reflectcall(nil, unsafe.Pointer(f.Fn), frame, uint32(framesz), uint32(framesz))

				// drop finalizer queue references to finalized object
				f.Fn = nil
				f.Arg = nil
				f.Ot = nil
				fb.Cnt = i - 1
			}
			next := fb.Next
			_lock.Lock(&_sched.Finlock)
			fb.Next = _gc.Finc
			_gc.Finc = fb
			_lock.Unlock(&_sched.Finlock)
			fb = next
		}
	}
}

// Look up pointer v in heap.  Return the span containing the object,
// the start of the object, and the size of the object.  If the object
// does not exist, return nil, nil, 0.
func FindObject(v unsafe.Pointer) (s *_core.Mspan, x unsafe.Pointer, n uintptr) {
	c := _sem.Gomcache()
	c.Local_nlookup++
	if _core.PtrSize == 4 && c.Local_nlookup >= 1<<30 {
		// purge cache stats to prevent overflow
		_lock.Lock(&_lock.Mheap_.Lock)
		_gc.Purgecachedstats(c)
		_lock.Unlock(&_lock.Mheap_.Lock)
	}

	// find span
	arena_start := uintptr(unsafe.Pointer(_lock.Mheap_.Arena_start))
	arena_used := uintptr(unsafe.Pointer(_lock.Mheap_.Arena_used))
	if uintptr(v) < arena_start || uintptr(v) >= arena_used {
		return
	}
	p := uintptr(v) >> _sched.PageShift
	q := p - arena_start>>_sched.PageShift
	s = *(**_core.Mspan)(_core.Add(unsafe.Pointer(_lock.Mheap_.Spans), q*_core.PtrSize))
	if s == nil {
		return
	}
	x = unsafe.Pointer(uintptr(s.Start) << _sched.PageShift)

	if uintptr(v) < uintptr(x) || uintptr(v) >= uintptr(unsafe.Pointer(s.Limit)) || s.State != _sched.XMSpanInUse {
		s = nil
		x = nil
		return
	}

	n = uintptr(s.Elemsize)
	if s.Sizeclass != 0 {
		x = _core.Add(x, (uintptr(v)-uintptr(x))/n*n)
	}
	return
}
