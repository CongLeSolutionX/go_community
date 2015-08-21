// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_print "runtime/internal/print"
	"unsafe"
)

var indexError = error(_base.ErrorString("index out of range"))

func panicindex() {
	panic(indexError)
}

var sliceError = error(_base.ErrorString("slice bounds out of range"))

func panicslice() {
	panic(sliceError)
}

func throwreturn() {
	_base.Throw("no return at end of a typed function - compiler is broken")
}

func throwinit() {
	_base.Throw("recursive call during initialization - linker skew")
}

// Create a new deferred function fn with siz bytes of arguments.
// The compiler turns a defer statement into a call to this.
//go:nosplit
func deferproc(siz int32, fn *_base.Funcval) { // arguments of fn follow fn
	if _base.Getg().M.Curg != _base.Getg() {
		// go code on the system stack can't defer
		_base.Throw("defer on system stack")
	}

	// the arguments of fn are in a perilous state.  The stack map
	// for deferproc does not describe them.  So we can't let garbage
	// collection or stack copying trigger until we've copied them out
	// to somewhere safe.  The memmove below does that.
	// Until the copy completes, we can only call nosplit routines.
	sp := _base.Getcallersp(unsafe.Pointer(&siz))
	argp := uintptr(unsafe.Pointer(&fn)) + unsafe.Sizeof(fn)
	callerpc := _base.Getcallerpc(unsafe.Pointer(&siz))

	_base.Systemstack(func() {
		d := newdefer(siz)
		if d.Panic != nil {
			_base.Throw("deferproc: d.panic != nil after newdefer")
		}
		d.Fn = fn
		d.Pc = callerpc
		d.Sp = sp
		_base.Memmove(_base.Add(unsafe.Pointer(d), unsafe.Sizeof(*d)), unsafe.Pointer(argp), uintptr(siz))
	})

	// deferproc returns 0 normally.
	// a deferred func that stops a panic
	// makes the deferproc return 1.
	// the code the compiler generates always
	// checks the return value and jumps to the
	// end of the function if deferproc returns != 0.
	return0()
	// No code can go here - the C return register has
	// been set and must not be clobbered.
}

// Small malloc size classes >= 16 are the multiples of 16: 16, 32, 48, 64, 80, 96, 112, 128, 144, ...
// Each P holds a pool for defers with small arg sizes.
// Assign defer allocations to pools by rounding to 16, to match malloc size classes.

const (
	deferHeaderSize = unsafe.Sizeof(_base.Defer{})
	minDeferAlloc   = (deferHeaderSize + 15) &^ 15
	minDeferArgs    = minDeferAlloc - deferHeaderSize
)

// defer size class for arg size sz
//go:nosplit
func deferclass(siz uintptr) uintptr {
	if siz <= minDeferArgs {
		return 0
	}
	return (siz - minDeferArgs + 15) / 16
}

// total size of memory block for defer with arg size sz
func totaldefersize(siz uintptr) uintptr {
	if siz <= minDeferArgs {
		return minDeferAlloc
	}
	return deferHeaderSize + siz
}

// Ensure that defer arg sizes that map to the same defer size class
// also map to the same malloc size class.
func testdefersizes() {
	var m [len(_base.P{}.Deferpool)]int32

	for i := range m {
		m[i] = -1
	}
	for i := uintptr(0); ; i++ {
		defersc := deferclass(i)
		if defersc >= uintptr(len(m)) {
			break
		}
		siz := roundupsize(totaldefersize(i))
		if m[defersc] < 0 {
			m[defersc] = int32(siz)
			continue
		}
		if m[defersc] != int32(siz) {
			print("bad defer size class: i=", i, " siz=", siz, " defersc=", defersc, "\n")
			_base.Throw("bad defer size class")
		}
	}
}

func init() {
	var x interface{}
	x = (*_base.Defer)(nil)
	_iface.DeferType = (*(**_gc.Ptrtype)(unsafe.Pointer(&x))).Elem
}

// Allocate a Defer, usually using per-P pool.
// Each defer must be released with freedefer.
// Note: runs on g0 stack
func newdefer(siz int32) *_base.Defer {
	var d *_base.Defer
	sc := deferclass(uintptr(siz))
	mp := _base.Acquirem()
	if sc < uintptr(len(_base.P{}.Deferpool)) {
		pp := mp.P.Ptr()
		if len(pp.Deferpool[sc]) == 0 && _base.Sched.Deferpool[sc] != nil {
			_base.Lock(&_base.Sched.Deferlock)
			for len(pp.Deferpool[sc]) < cap(pp.Deferpool[sc])/2 && _base.Sched.Deferpool[sc] != nil {
				d := _base.Sched.Deferpool[sc]
				_base.Sched.Deferpool[sc] = d.Link
				d.Link = nil
				pp.Deferpool[sc] = append(pp.Deferpool[sc], d)
			}
			_base.Unlock(&_base.Sched.Deferlock)
		}
		if n := len(pp.Deferpool[sc]); n > 0 {
			d = pp.Deferpool[sc][n-1]
			pp.Deferpool[sc][n-1] = nil
			pp.Deferpool[sc] = pp.Deferpool[sc][:n-1]
		}
	}
	if d == nil {
		// Allocate new defer+args.
		total := roundupsize(totaldefersize(uintptr(siz)))
		d = (*_base.Defer)(_iface.Mallocgc(total, _iface.DeferType, 0))
	}
	d.Siz = siz
	gp := mp.Curg
	d.Link = gp.Defer
	gp.Defer = d
	_base.Releasem(mp)
	return d
}

// Free the given defer.
// The defer cannot be used after this call.
func freedefer(d *_base.Defer) {
	if d.Panic != nil {
		freedeferpanic()
	}
	if d.Fn != nil {
		freedeferfn()
	}
	sc := deferclass(uintptr(d.Siz))
	if sc < uintptr(len(_base.P{}.Deferpool)) {
		mp := _base.Acquirem()
		pp := mp.P.Ptr()
		if len(pp.Deferpool[sc]) == cap(pp.Deferpool[sc]) {
			// Transfer half of local cache to the central cache.
			var first, last *_base.Defer
			for len(pp.Deferpool[sc]) > cap(pp.Deferpool[sc])/2 {
				n := len(pp.Deferpool[sc])
				d := pp.Deferpool[sc][n-1]
				pp.Deferpool[sc][n-1] = nil
				pp.Deferpool[sc] = pp.Deferpool[sc][:n-1]
				if first == nil {
					first = d
				} else {
					last.Link = d
				}
				last = d
			}
			_base.Lock(&_base.Sched.Deferlock)
			last.Link = _base.Sched.Deferpool[sc]
			_base.Sched.Deferpool[sc] = first
			_base.Unlock(&_base.Sched.Deferlock)
		}
		*d = _base.Defer{}
		pp.Deferpool[sc] = append(pp.Deferpool[sc], d)
		_base.Releasem(mp)
	}
}

// Separate function so that it can split stack.
// Windows otherwise runs out of stack space.
func freedeferpanic() {
	// _panic must be cleared before d is unlinked from gp.
	_base.Throw("freedefer with d._panic != nil")
}

func freedeferfn() {
	// fn must be cleared before d is unlinked from gp.
	_base.Throw("freedefer with d.fn != nil")
}

// Run a deferred function if there is one.
// The compiler inserts a call to this at the end of any
// function which calls defer.
// If there is a deferred function, this will call runtime·jmpdefer,
// which will jump to the deferred function such that it appears
// to have been called by the caller of deferreturn at the point
// just before deferreturn was called.  The effect is that deferreturn
// is called again and again until there are no more deferred functions.
// Cannot split the stack because we reuse the caller's frame to
// call the deferred function.

// The single argument isn't actually used - it just has its address
// taken so it can be matched against pending defers.
//go:nosplit
func deferreturn(arg0 uintptr) {
	gp := _base.Getg()
	d := gp.Defer
	if d == nil {
		return
	}
	sp := _base.Getcallersp(unsafe.Pointer(&arg0))
	if d.Sp != sp {
		return
	}

	// Moving arguments around.
	// Do not allow preemption here, because the garbage collector
	// won't know the form of the arguments until the jmpdefer can
	// flip the PC over to fn.
	mp := _base.Acquirem()
	_base.Memmove(unsafe.Pointer(&arg0), _gc.DeferArgs(d), uintptr(d.Siz))
	fn := d.Fn
	d.Fn = nil
	gp.Defer = d.Link
	// Switch to systemstack merely to save nosplit stack space.
	_base.Systemstack(func() {
		freedefer(d)
	})
	_base.Releasem(mp)
	jmpdefer(fn, uintptr(unsafe.Pointer(&arg0)))
}

// Goexit terminates the goroutine that calls it.  No other goroutine is affected.
// Goexit runs all deferred calls before terminating the goroutine.  Because Goexit
// is not panic, however, any recover calls in those deferred functions will return nil.
//
// Calling Goexit from the main goroutine terminates that goroutine
// without func main returning. Since func main has not returned,
// the program continues execution of other goroutines.
// If all other goroutines exit, the program crashes.
func Goexit() {
	// Run all deferred functions for the current goroutine.
	// This code is similar to gopanic, see that implementation
	// for detailed comments.
	gp := _base.Getg()
	for {
		d := gp.Defer
		if d == nil {
			break
		}
		if d.Started {
			if d.Panic != nil {
				d.Panic.Aborted = true
				d.Panic = nil
			}
			d.Fn = nil
			gp.Defer = d.Link
			freedefer(d)
			continue
		}
		d.Started = true
		reflectcall(nil, unsafe.Pointer(d.Fn), _gc.DeferArgs(d), uint32(d.Siz), uint32(d.Siz))
		if gp.Defer != d {
			_base.Throw("bad defer entry in Goexit")
		}
		d.Panic = nil
		d.Fn = nil
		gp.Defer = d.Link
		freedefer(d)
		// Note: we ignore recovers here because Goexit isn't a panic
	}
	goexit1()
}

// The implementation of the predeclared function panic.
func gopanic(e interface{}) {
	gp := _base.Getg()
	if gp.M.Curg != gp {
		print("panic: ")
		_print.Printany(e)
		print("\n")
		_base.Throw("panic on system stack")
	}

	// m.softfloat is set during software floating point.
	// It increments m.locks to avoid preemption.
	// We moved the memory loads out, so there shouldn't be
	// any reason for it to panic anymore.
	if gp.M.Softfloat != 0 {
		gp.M.Locks--
		gp.M.Softfloat = 0
		_base.Throw("panic during softfloat")
	}
	if gp.M.Mallocing != 0 {
		print("panic: ")
		_print.Printany(e)
		print("\n")
		_base.Throw("panic during malloc")
	}
	if gp.M.Preemptoff != "" {
		print("panic: ")
		_print.Printany(e)
		print("\n")
		print("preempt off reason: ")
		print(gp.M.Preemptoff)
		print("\n")
		_base.Throw("panic during preemptoff")
	}
	if gp.M.Locks != 0 {
		print("panic: ")
		_print.Printany(e)
		print("\n")
		_base.Throw("panic holding locks")
	}

	var p _base.Panic
	p.Arg = e
	p.Link = gp.Panic
	gp.Panic = (*_base.Panic)(_base.Noescape(unsafe.Pointer(&p)))

	for {
		d := gp.Defer
		if d == nil {
			break
		}

		// If defer was started by earlier panic or Goexit (and, since we're back here, that triggered a new panic),
		// take defer off list. The earlier panic or Goexit will not continue running.
		if d.Started {
			if d.Panic != nil {
				d.Panic.Aborted = true
			}
			d.Panic = nil
			d.Fn = nil
			gp.Defer = d.Link
			freedefer(d)
			continue
		}

		// Mark defer as started, but keep on list, so that traceback
		// can find and update the defer's argument frame if stack growth
		// or a garbage collection happens before reflectcall starts executing d.fn.
		d.Started = true

		// Record the panic that is running the defer.
		// If there is a new panic during the deferred call, that panic
		// will find d in the list and will mark d._panic (this panic) aborted.
		d.Panic = (*_base.Panic)(_base.Noescape((unsafe.Pointer)(&p)))

		p.Argp = unsafe.Pointer(getargp(0))
		reflectcall(nil, unsafe.Pointer(d.Fn), _gc.DeferArgs(d), uint32(d.Siz), uint32(d.Siz))
		p.Argp = nil

		// reflectcall did not panic. Remove d.
		if gp.Defer != d {
			_base.Throw("bad defer entry in panic")
		}
		d.Panic = nil
		d.Fn = nil
		gp.Defer = d.Link

		// trigger shrinkage to test stack copy.  See stack_test.go:TestStackPanic
		//GC()

		pc := d.Pc
		sp := unsafe.Pointer(d.Sp) // must be pointer so it gets adjusted during stack copy
		freedefer(d)
		if p.Recovered {
			gp.Panic = p.Link
			// Aborted panics are marked but remain on the g.panic list.
			// Remove them from the list.
			for gp.Panic != nil && gp.Panic.Aborted {
				gp.Panic = gp.Panic.Link
			}
			if gp.Panic == nil { // must be done with signal
				gp.Sig = 0
			}
			// Pass information about recovering frame to recovery.
			gp.Sigcode0 = uintptr(sp)
			gp.Sigcode1 = pc
			_base.Mcall(recovery)
			_base.Throw("recovery failed") // mcall should not return
		}
	}

	// ran out of deferred calls - old-school panic now
	_base.Startpanic()
	_print.Printpanics(gp.Panic)
	_base.Dopanic(0) // should not return
	*(*int)(nil) = 0 // not reached
}

// getargp returns the location where the caller
// writes outgoing function call arguments.
//go:nosplit
func getargp(x int) uintptr {
	// x is an argument mainly so that we can return its address.
	// However, we need to make the function complex enough
	// that it won't be inlined. We always pass x = 0, so this code
	// does nothing other than keep the compiler from thinking
	// the function is simple enough to inline.
	if x > 0 {
		return _base.Getcallersp(unsafe.Pointer(&x)) * 0
	}
	return uintptr(_base.Noescape(unsafe.Pointer(&x)))
}

// The implementation of the predeclared function recover.
// Cannot split the stack because it needs to reliably
// find the stack segment of its caller.
//
// TODO(rsc): Once we commit to CopyStackAlways,
// this doesn't need to be nosplit.
//go:nosplit
func gorecover(argp uintptr) interface{} {
	// Must be in a function running as part of a deferred call during the panic.
	// Must be called from the topmost function of the call
	// (the function used in the defer statement).
	// p.argp is the argument pointer of that topmost deferred function call.
	// Compare against argp reported by caller.
	// If they match, the caller is the one who can recover.
	gp := _base.Getg()
	p := gp.Panic
	if p != nil && !p.Recovered && argp == uintptr(p.Argp) {
		p.Recovered = true
		return p.Arg
	}
	return nil
}
