// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_race "runtime/internal/race"
	"unsafe"
)

//go:nowritebarrier
func iterate_finq(callback func(*_base.Funcval, unsafe.Pointer, uintptr, *_base.Type, *_gc.Ptrtype)) {
	for fb := _gc.Allfin; fb != nil; fb = fb.Alllink {
		for i := int32(0); i < fb.Cnt; i++ {
			f := &fb.Fin[i]
			callback(f.Fn, f.Arg, f.Nret, f.Fint, f.Ot)
		}
	}
}

var (
	fingCreate uint32
)

func createfing() {
	// start the finalizer goroutine exactly once
	if fingCreate == 0 && _base.Cas(&fingCreate, 0, 1) {
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
		_base.Lock(&_base.Finlock)
		fb := _gc.Finq
		_gc.Finq = nil
		if fb == nil {
			gp := _base.Getg()
			_base.Fing = gp
			_base.Fingwait = true
			_base.Goparkunlock(&_base.Finlock, "finalizer wait", _base.TraceEvGoBlock, 1)
			continue
		}
		_base.Unlock(&_base.Finlock)
		if _base.Raceenabled {
			_race.Racefingo()
		}
		for fb != nil {
			for i := fb.Cnt; i > 0; i-- {
				f := (*_gc.Finalizer)(_base.Add(unsafe.Pointer(&fb.Fin), uintptr(i-1)*unsafe.Sizeof(_gc.Finalizer{})))

				framesz := unsafe.Sizeof((interface{})(nil)) + uintptr(f.Nret)
				if framecap < framesz {
					// The frame does not contain pointers interesting for GC,
					// all not yet finalized objects are stored in finq.
					// If we do not mark it as FlagNoScan,
					// the last finalized object is not collected.
					frame = _iface.Mallocgc(framesz, nil, _base.FlagNoScan)
					framecap = framesz
				}

				if f.Fint == nil {
					_base.Throw("missing type in runfinq")
				}
				switch f.Fint.Kind & _iface.KindMask {
				case _iface.KindPtr:
					// direct use of pointer
					*(*unsafe.Pointer)(frame) = f.Arg
				case _iface.KindInterface:
					ityp := (*_iface.Interfacetype)(unsafe.Pointer(f.Fint))
					// set up with empty interface
					(*_iface.Eface)(frame).Type = &f.Ot.Typ
					(*_iface.Eface)(frame).Data = f.Arg
					if len(ityp.Mhdr) != 0 {
						// convert to interface with methods
						// this conversion is guaranteed to succeed - we checked in SetFinalizer
						_iface.AssertE2I(ityp, *(*interface{})(frame), (*_iface.FInterface)(frame))
					}
				default:
					_base.Throw("bad kind in runfinq")
				}
				_base.FingRunning = true
				reflectcall(nil, unsafe.Pointer(f.Fn), frame, uint32(framesz), uint32(framesz))
				_base.FingRunning = false

				// drop finalizer queue references to finalized object
				f.Fn = nil
				f.Arg = nil
				f.Ot = nil
				fb.Cnt = i - 1
			}
			next := fb.Next
			_base.Lock(&_base.Finlock)
			fb.Next = _gc.Finc
			_gc.Finc = fb
			_base.Unlock(&_base.Finlock)
			fb = next
		}
	}
}

// SetFinalizer sets the finalizer associated with x to f.
// When the garbage collector finds an unreachable block
// with an associated finalizer, it clears the association and runs
// f(x) in a separate goroutine.  This makes x reachable again, but
// now without an associated finalizer.  Assuming that SetFinalizer
// is not called again, the next time the garbage collector sees
// that x is unreachable, it will free x.
//
// SetFinalizer(x, nil) clears any finalizer associated with x.
//
// The argument x must be a pointer to an object allocated by
// calling new or by taking the address of a composite literal.
// The argument f must be a function that takes a single argument
// to which x's type can be assigned, and can have arbitrary ignored return
// values. If either of these is not true, SetFinalizer aborts the
// program.
//
// Finalizers are run in dependency order: if A points at B, both have
// finalizers, and they are otherwise unreachable, only the finalizer
// for A runs; once A is freed, the finalizer for B can run.
// If a cyclic structure includes a block with a finalizer, that
// cycle is not guaranteed to be garbage collected and the finalizer
// is not guaranteed to run, because there is no ordering that
// respects the dependencies.
//
// The finalizer for x is scheduled to run at some arbitrary time after
// x becomes unreachable.
// There is no guarantee that finalizers will run before a program exits,
// so typically they are useful only for releasing non-memory resources
// associated with an object during a long-running program.
// For example, an os.File object could use a finalizer to close the
// associated operating system file descriptor when a program discards
// an os.File without calling Close, but it would be a mistake
// to depend on a finalizer to flush an in-memory I/O buffer such as a
// bufio.Writer, because the buffer would not be flushed at program exit.
//
// It is not guaranteed that a finalizer will run if the size of *x is
// zero bytes.
//
// It is not guaranteed that a finalizer will run for objects allocated
// in initializers for package-level variables. Such objects may be
// linker-allocated, not heap-allocated.
//
// A single goroutine runs all finalizers for a program, sequentially.
// If a finalizer must run for a long time, it should do so by starting
// a new goroutine.
func SetFinalizer(obj interface{}, finalizer interface{}) {
	if _base.Debug.Sbrk != 0 {
		// debug.sbrk never frees memory, so no finalizers run
		// (and we don't have the data structures to record them).
		return
	}
	e := (*_iface.Eface)(unsafe.Pointer(&obj))
	etyp := e.Type
	if etyp == nil {
		_base.Throw("runtime.SetFinalizer: first argument is nil")
	}
	if etyp.Kind&_iface.KindMask != _iface.KindPtr {
		_base.Throw("runtime.SetFinalizer: first argument is " + *etyp.String + ", not pointer")
	}
	ot := (*_gc.Ptrtype)(unsafe.Pointer(etyp))
	if ot.Elem == nil {
		_base.Throw("nil elem type!")
	}

	// find the containing object
	_, base, _ := findObject(e.Data)

	if base == nil {
		// 0-length objects are okay.
		if e.Data == unsafe.Pointer(&_iface.Zerobase) {
			return
		}

		// Global initializers might be linker-allocated.
		//	var Foo = &Object{}
		//	func main() {
		//		runtime.SetFinalizer(Foo, nil)
		//	}
		// The relevant segments are: noptrdata, data, bss, noptrbss.
		// We cannot assume they are in any order or even contiguous,
		// due to external linking.
		for datap := &_base.Firstmoduledata; datap != nil; datap = datap.Next {
			if datap.Noptrdata <= uintptr(e.Data) && uintptr(e.Data) < datap.Enoptrdata ||
				datap.Data <= uintptr(e.Data) && uintptr(e.Data) < datap.Edata ||
				datap.Bss <= uintptr(e.Data) && uintptr(e.Data) < datap.Ebss ||
				datap.Noptrbss <= uintptr(e.Data) && uintptr(e.Data) < datap.Enoptrbss {
				return
			}
		}
		_base.Throw("runtime.SetFinalizer: pointer not in allocated block")
	}

	if e.Data != base {
		// As an implementation detail we allow to set finalizers for an inner byte
		// of an object if it could come from tiny alloc (see mallocgc for details).
		if ot.Elem == nil || ot.Elem.Kind&_iface.KindNoPointers == 0 || ot.Elem.Size >= _base.MaxTinySize {
			_base.Throw("runtime.SetFinalizer: pointer not at beginning of allocated block")
		}
	}

	f := (*_iface.Eface)(unsafe.Pointer(&finalizer))
	ftyp := f.Type
	if ftyp == nil {
		// switch to system stack and remove finalizer
		_base.Systemstack(func() {
			removefinalizer(e.Data)
		})
		return
	}

	if ftyp.Kind&_iface.KindMask != _iface.KindFunc {
		_base.Throw("runtime.SetFinalizer: second argument is " + *ftyp.String + ", not a function")
	}
	ft := (*functype)(unsafe.Pointer(ftyp))
	ins := *(*[]*_base.Type)(unsafe.Pointer(&ft.in))
	if ft.dotdotdot || len(ins) != 1 {
		_base.Throw("runtime.SetFinalizer: cannot pass " + *etyp.String + " to finalizer " + *ftyp.String)
	}
	fint := ins[0]
	switch {
	case fint == etyp:
		// ok - same type
		goto okarg
	case fint.Kind&_iface.KindMask == _iface.KindPtr:
		if (fint.X == nil || fint.X.Name == nil || etyp.X == nil || etyp.X.Name == nil) && (*_gc.Ptrtype)(unsafe.Pointer(fint)).Elem == ot.Elem {
			// ok - not same type, but both pointers,
			// one or the other is unnamed, and same element type, so assignable.
			goto okarg
		}
	case fint.Kind&_iface.KindMask == _iface.KindInterface:
		ityp := (*_iface.Interfacetype)(unsafe.Pointer(fint))
		if len(ityp.Mhdr) == 0 {
			// ok - satisfies empty interface
			goto okarg
		}
		if _iface.AssertE2I2(ityp, obj, nil) {
			goto okarg
		}
	}
	_base.Throw("runtime.SetFinalizer: cannot pass " + *etyp.String + " to finalizer " + *ftyp.String)
okarg:
	// compute size needed for return parameters
	nret := uintptr(0)
	for _, t := range *(*[]*_base.Type)(unsafe.Pointer(&ft.out)) {
		nret = _base.Round(nret, uintptr(t.Align)) + uintptr(t.Size)
	}
	nret = _base.Round(nret, _base.PtrSize)

	// make sure we have a finalizer goroutine
	createfing()

	_base.Systemstack(func() {
		if !addfinalizer(e.Data, (*_base.Funcval)(f.Data), nret, fint, ot) {
			_base.Throw("runtime.SetFinalizer: finalizer already set")
		}
	})
}

// Look up pointer v in heap.  Return the span containing the object,
// the start of the object, and the size of the object.  If the object
// does not exist, return nil, nil, 0.
func findObject(v unsafe.Pointer) (s *_base.Mspan, x unsafe.Pointer, n uintptr) {
	c := _iface.Gomcache()
	c.Local_nlookup++
	if _base.PtrSize == 4 && c.Local_nlookup >= 1<<30 {
		// purge cache stats to prevent overflow
		_base.Lock(&_base.Mheap_.Lock)
		_gc.Purgecachedstats(c)
		_base.Unlock(&_base.Mheap_.Lock)
	}

	// find span
	arena_start := uintptr(unsafe.Pointer(_base.Mheap_.Arena_start))
	arena_used := uintptr(unsafe.Pointer(_base.Mheap_.Arena_used))
	if uintptr(v) < arena_start || uintptr(v) >= arena_used {
		return
	}
	p := uintptr(v) >> _base.XPageShift
	q := p - arena_start>>_base.XPageShift
	s = *(**_base.Mspan)(_base.Add(unsafe.Pointer(_base.Mheap_.Spans), q*_base.PtrSize))
	if s == nil {
		return
	}
	x = unsafe.Pointer(uintptr(s.Start) << _base.XPageShift)

	if uintptr(v) < uintptr(x) || uintptr(v) >= uintptr(unsafe.Pointer(s.Limit)) || s.State != _base.MSpanInUse {
		s = nil
		x = nil
		return
	}

	n = uintptr(s.Elemsize)
	if s.Sizeclass != 0 {
		x = _base.Add(x, (uintptr(v)-uintptr(x))/n*n)
	}
	return
}
