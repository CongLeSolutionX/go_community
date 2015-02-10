// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: finalizers and block profiling.

package runtime

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_finalize "runtime/internal/finalize"
	_gc "runtime/internal/gc"
	_ifacestuff "runtime/internal/ifacestuff"
	_lock "runtime/internal/lock"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	"unsafe"
)

//go:nowritebarrier
func iterate_finq(callback func(*_core.Funcval, unsafe.Pointer, uintptr, *_core.Type, *_gc.Ptrtype)) {
	for fb := _gc.Allfin; fb != nil; fb = fb.Alllink {
		for i := int32(0); i < fb.Cnt; i++ {
			f := &fb.Fin[i]
			callback(f.Fn, f.Arg, f.Nret, f.Fint, f.Ot)
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
	e := (*_core.Eface)(unsafe.Pointer(&obj))
	etyp := e.Type
	if etyp == nil {
		_lock.Throw("runtime.SetFinalizer: first argument is nil")
	}
	if etyp.Kind&_channels.KindMask != _channels.KindPtr {
		_lock.Throw("runtime.SetFinalizer: first argument is " + *etyp.String + ", not pointer")
	}
	ot := (*_gc.Ptrtype)(unsafe.Pointer(etyp))
	if ot.Elem == nil {
		_lock.Throw("nil elem type!")
	}

	// find the containing object
	_, base, _ := _finalize.FindObject(e.Data)

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
		if uintptr(unsafe.Pointer(&_schedinit.Noptrdata)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&_schedinit.Enoptrdata)) ||
			uintptr(unsafe.Pointer(&_gc.Data)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&_gc.Edata)) ||
			uintptr(unsafe.Pointer(&_gc.Bss)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&_gc.Ebss)) ||
			uintptr(unsafe.Pointer(&_schedinit.Noptrbss)) <= uintptr(e.Data) && uintptr(e.Data) < uintptr(unsafe.Pointer(&_schedinit.Enoptrbss)) {
			return
		}
		_lock.Throw("runtime.SetFinalizer: pointer not in allocated block")
	}

	if e.Data != base {
		// As an implementation detail we allow to set finalizers for an inner byte
		// of an object if it could come from tiny alloc (see mallocgc for details).
		if ot.Elem == nil || ot.Elem.Kind&_channels.KindNoPointers == 0 || ot.Elem.Size >= _sched.MaxTinySize {
			_lock.Throw("runtime.SetFinalizer: pointer not at beginning of allocated block")
		}
	}

	f := (*_core.Eface)(unsafe.Pointer(&finalizer))
	ftyp := f.Type
	if ftyp == nil {
		// switch to system stack and remove finalizer
		_lock.Systemstack(func() {
			_finalize.Removefinalizer(e.Data)
		})
		return
	}

	if ftyp.Kind&_channels.KindMask != _channels.KindFunc {
		_lock.Throw("runtime.SetFinalizer: second argument is " + *ftyp.String + ", not a function")
	}
	ft := (*_finalize.Functype)(unsafe.Pointer(ftyp))
	ins := *(*[]*_core.Type)(unsafe.Pointer(&ft.In))
	if ft.Dotdotdot || len(ins) != 1 {
		_lock.Throw("runtime.SetFinalizer: cannot pass " + *etyp.String + " to finalizer " + *ftyp.String)
	}
	fint := ins[0]
	switch {
	case fint == etyp:
		// ok - same type
		goto okarg
	case fint.Kind&_channels.KindMask == _channels.KindPtr:
		if (fint.X == nil || fint.X.Name == nil || etyp.X == nil || etyp.X.Name == nil) && (*_gc.Ptrtype)(unsafe.Pointer(fint)).Elem == ot.Elem {
			// ok - not same type, but both pointers,
			// one or the other is unnamed, and same element type, so assignable.
			goto okarg
		}
	case fint.Kind&_channels.KindMask == _channels.KindInterface:
		ityp := (*_core.Interfacetype)(unsafe.Pointer(fint))
		if len(ityp.Mhdr) == 0 {
			// ok - satisfies empty interface
			goto okarg
		}
		if _ifacestuff.AssertE2I2(ityp, obj, nil) {
			goto okarg
		}
	}
	_lock.Throw("runtime.SetFinalizer: cannot pass " + *etyp.String + " to finalizer " + *ftyp.String)
okarg:
	// compute size needed for return parameters
	nret := uintptr(0)
	for _, t := range *(*[]*_core.Type)(unsafe.Pointer(&ft.Out)) {
		nret = _lock.Round(nret, uintptr(t.Align)) + uintptr(t.Size)
	}
	nret = _lock.Round(nret, _core.PtrSize)

	// make sure we have a finalizer goroutine
	_finalize.Createfing()

	_lock.Systemstack(func() {
		if !_finalize.Addfinalizer(e.Data, (*_core.Funcval)(f.Data), nret, fint, ot) {
			_lock.Throw("runtime.SetFinalizer: finalizer already set")
		}
	})
}
