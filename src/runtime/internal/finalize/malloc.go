// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package finalize

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_hash "runtime/internal/hash"
	_ifacestuff "runtime/internal/ifacestuff"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_sem "runtime/internal/sem"
	"unsafe"
)

// linker-provided
var noptrdata struct{}
var enoptrdata struct{}
var noptrbss struct{}
var enoptrbss struct{}

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
	e := (*_core.Eface)(unsafe.Pointer(&obj))
	etyp := e.Type
	if etyp == nil {
		_lock.Gothrow("runtime.SetFinalizer: first argument is nil")
	}
	if etyp.Kind&_hash.KindMask != _hash.KindPtr {
		_lock.Gothrow("runtime.SetFinalizer: first argument is " + *etyp.String + ", not pointer")
	}
	ot := (*_gc.Ptrtype)(unsafe.Pointer(etyp))
	if ot.Elem == nil {
		_lock.Gothrow("nil elem type!")
	}

	// find the containing object
	_, base, _ := findObject(e.Data)

	if base == nil {
		// 0-length objects are okay.
		if e.Data == unsafe.Pointer(&_maps.Zerobase) {
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
		if uintptr(unsafe.Pointer(&noptrdata)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&enoptrdata)) ||
			uintptr(unsafe.Pointer(&_gc.Data)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&_gc.Edata)) ||
			uintptr(unsafe.Pointer(&_gc.Bss)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&_gc.Ebss)) ||
			uintptr(unsafe.Pointer(&noptrbss)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&enoptrbss)) {
			return
		}
		_lock.Gothrow("runtime.SetFinalizer: pointer not in allocated block")
	}

	if e.Data != base {
		// As an implementation detail we allow to set finalizers for an inner byte
		// of an object if it could come from tiny alloc (see mallocgc for details).
		if ot.Elem == nil || ot.Elem.Kind&_hash.KindNoPointers == 0 || ot.Elem.Size >= _sched.MaxTinySize {
			_lock.Gothrow("runtime.SetFinalizer: pointer not at beginning of allocated block")
		}
	}

	f := (*_core.Eface)(unsafe.Pointer(&finalizer))
	ftyp := f.Type
	if ftyp == nil {
		// switch to system stack and remove finalizer
		_lock.Systemstack(func() {
			removefinalizer(e.Data)
		})
		return
	}

	if ftyp.Kind&_hash.KindMask != _hash.KindFunc {
		_lock.Gothrow("runtime.SetFinalizer: second argument is " + *ftyp.String + ", not a function")
	}
	ft := (*functype)(unsafe.Pointer(ftyp))
	ins := *(*[]*_core.Type)(unsafe.Pointer(&ft.in))
	if ft.dotdotdot || len(ins) != 1 {
		_lock.Gothrow("runtime.SetFinalizer: cannot pass " + *etyp.String + " to finalizer " + *ftyp.String)
	}
	fint := ins[0]
	switch {
	case fint == etyp:
		// ok - same type
		goto okarg
	case fint.Kind&_hash.KindMask == _hash.KindPtr:
		if (fint.X == nil || fint.X.Name == nil || etyp.X == nil || etyp.X.Name == nil) && (*_gc.Ptrtype)(unsafe.Pointer(fint)).Elem == ot.Elem {
			// ok - not same type, but both pointers,
			// one or the other is unnamed, and same element type, so assignable.
			goto okarg
		}
	case fint.Kind&_hash.KindMask == _hash.KindInterface:
		ityp := (*_core.Interfacetype)(unsafe.Pointer(fint))
		if len(ityp.Mhdr) == 0 {
			// ok - satisfies empty interface
			goto okarg
		}
		if _, ok := _ifacestuff.AssertE2I2(ityp, obj); ok {
			goto okarg
		}
	}
	_lock.Gothrow("runtime.SetFinalizer: cannot pass " + *etyp.String + " to finalizer " + *ftyp.String)
okarg:
	// compute size needed for return parameters
	nret := uintptr(0)
	for _, t := range *(*[]*_core.Type)(unsafe.Pointer(&ft.out)) {
		nret = _sched.Round(nret, uintptr(t.Align)) + uintptr(t.Size)
	}
	nret = _sched.Round(nret, _core.PtrSize)

	// make sure we have a finalizer goroutine
	createfing()

	_lock.Systemstack(func() {
		if !addfinalizer(e.Data, (*_core.Funcval)(f.Data), nret, fint, ot) {
			_lock.Gothrow("runtime.SetFinalizer: finalizer already set")
		}
	})
}

// Look up pointer v in heap.  Return the span containing the object,
// the start of the object, and the size of the object.  If the object
// does not exist, return nil, nil, 0.
func findObject(v unsafe.Pointer) (s *_core.Mspan, x unsafe.Pointer, n uintptr) {
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

	if uintptr(v) < uintptr(x) || uintptr(v) >= uintptr(unsafe.Pointer(s.Limit)) || s.State != _sched.MSpanInUse {
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

var fingCreate uint32

func createfing() {
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
			_sched.Goparkunlock(&_sched.Finlock, "finalizer wait")
			gp.Issystem = false
			continue
		}
		_lock.Unlock(&_sched.Finlock)
		if _sched.Raceenabled {
			racefingo()
		}
		for fb != nil {
			for i := int32(0); i < fb.Cnt; i++ {
				f := (*_gc.Finalizer)(_core.Add(unsafe.Pointer(&fb.Fin), uintptr(i)*unsafe.Sizeof(_gc.Finalizer{})))

				framesz := unsafe.Sizeof((interface{})(nil)) + uintptr(f.Nret)
				if framecap < framesz {
					// The frame does not contain pointers interesting for GC,
					// all not yet finalized objects are stored in finq.
					// If we do not mark it as FlagNoScan,
					// the last finalized object is not collected.
					frame = _maps.Mallocgc(framesz, nil, _sched.FlagNoScan)
					framecap = framesz
				}

				if f.Fint == nil {
					_lock.Gothrow("missing type in runfinq")
				}
				switch f.Fint.Kind & _hash.KindMask {
				case _hash.KindPtr:
					// direct use of pointer
					*(*unsafe.Pointer)(frame) = f.Arg
				case _hash.KindInterface:
					ityp := (*_core.Interfacetype)(unsafe.Pointer(f.Fint))
					// set up with empty interface
					(*_core.Eface)(frame).Type = &f.Ot.Typ
					(*_core.Eface)(frame).Data = f.Arg
					if len(ityp.Mhdr) != 0 {
						// convert to interface with methods
						// this conversion is guaranteed to succeed - we checked in SetFinalizer
						*(*_ifacestuff.FInterface)(frame) = _ifacestuff.AssertE2I(ityp, *(*interface{})(frame))
					}
				default:
					_lock.Gothrow("bad kind in runfinq")
				}
				Reflectcall(unsafe.Pointer(f.Fn), frame, uint32(framesz), uint32(framesz))

				// drop finalizer queue references to finalized object
				f.Fn = nil
				f.Arg = nil
				f.Ot = nil
			}
			fb.Cnt = 0
			next := fb.Next
			_lock.Lock(&_sched.Finlock)
			fb.Next = _gc.Finc
			_gc.Finc = fb
			_lock.Unlock(&_sched.Finlock)
			fb = next
		}
	}
}
