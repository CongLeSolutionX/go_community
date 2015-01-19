// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

var stackfreequeue _core.Stack

// Adds stack x to the free pool.  Must be called with stackpoolmu held.
func stackpoolfree(x _core.Gclinkptr, order uint8) {
	s := mHeap_Lookup(&_lock.Mheap_, (unsafe.Pointer)(x))
	if s.State != _sched.MSpanStack {
		_lock.Gothrow("freeing stack not in a stack span")
	}
	if s.Freelist.Ptr() == nil {
		// s will now have a free stack
		_sched.MSpanList_Insert(&_sched.Stackpool[order], s)
	}
	x.Ptr().Next = s.Freelist
	s.Freelist = x
	s.Ref--
	if s.Ref == 0 {
		// span is completely free - return to heap
		_sched.MSpanList_Remove(s)
		s.Freelist = 0
		mHeap_FreeStack(&_lock.Mheap_, s)
	}
}

func stackcacherelease(c *_core.Mcache, order uint8) {
	if _sched.StackDebug >= 1 {
		print("stackcacherelease order=", order, "\n")
	}
	x := c.Stackcache[order].List
	size := c.Stackcache[order].Size
	_lock.Lock(&_sched.Stackpoolmu)
	for size > _core.StackCacheSize/2 {
		y := x.Ptr().Next
		stackpoolfree(x, order)
		x = y
		size -= _core.FixedStack << order
	}
	_lock.Unlock(&_sched.Stackpoolmu)
	c.Stackcache[order].List = x
	c.Stackcache[order].Size = size
}

func stackcache_clear(c *_core.Mcache) {
	if _sched.StackDebug >= 1 {
		print("stackcache clear\n")
	}
	_lock.Lock(&_sched.Stackpoolmu)
	for order := uint8(0); order < _core.NumStackOrders; order++ {
		x := c.Stackcache[order].List
		for x.Ptr() != nil {
			y := x.Ptr().Next
			stackpoolfree(x, order)
			x = y
		}
		c.Stackcache[order].List = 0
		c.Stackcache[order].Size = 0
	}
	_lock.Unlock(&_sched.Stackpoolmu)
}

func Stackfree(stk _core.Stack) {
	gp := _core.Getg()
	n := stk.Hi - stk.Lo
	v := (unsafe.Pointer)(stk.Lo)
	if n&(n-1) != 0 {
		_lock.Gothrow("stack not a power of 2")
	}
	if _sched.StackDebug >= 1 {
		println("stackfree", v, n)
		_core.Memclr(v, n) // for testing, clobber stack data
	}
	if _lock.Debug.Efence != 0 || _sched.StackFromSystem != 0 {
		if _lock.Debug.Efence != 0 || _sched.StackFaultOnFree != 0 {
			sysFault(v, n)
		} else {
			_sched.SysFree(v, n, &_lock.Memstats.Stacks_sys)
		}
		return
	}
	if _sched.StackCache != 0 && n < _core.FixedStack<<_core.NumStackOrders && n < _core.StackCacheSize {
		order := uint8(0)
		n2 := n
		for n2 > _core.FixedStack {
			order++
			n2 >>= 1
		}
		x := _core.Gclinkptr(v)
		c := gp.M.Mcache
		if c == nil || gp.M.Gcing != 0 || gp.M.Helpgc != 0 {
			_lock.Lock(&_sched.Stackpoolmu)
			stackpoolfree(x, order)
			_lock.Unlock(&_sched.Stackpoolmu)
		} else {
			if c.Stackcache[order].Size >= _core.StackCacheSize {
				stackcacherelease(c, order)
			}
			x.Ptr().Next = c.Stackcache[order].List
			c.Stackcache[order].List = x
			c.Stackcache[order].Size += n
		}
	} else {
		s := mHeap_Lookup(&_lock.Mheap_, v)
		if s.State != _sched.MSpanStack {
			println(_core.Hex(s.Start<<_core.PageShift), v)
			_lock.Gothrow("bad span state")
		}
		mHeap_FreeStack(&_lock.Mheap_, s)
	}
}

var mapnames = []string{
	_sched.BitsDead:    "---",
	_sched.BitsScalar:  "scalar",
	_sched.BitsPointer: "ptr",
}

// Stack frame layout
//
// (x86)
// +------------------+
// | args from caller |
// +------------------+ <- frame->argp
// |  return address  |
// +------------------+ <- frame->varp
// |     locals       |
// +------------------+
// |  args to callee  |
// +------------------+ <- frame->sp
//
// (arm)
// +------------------+
// | args from caller |
// +------------------+ <- frame->argp
// | caller's retaddr |
// +------------------+ <- frame->varp
// |     locals       |
// +------------------+
// |  args to callee  |
// +------------------+
// |  return address  |
// +------------------+ <- frame->sp

type adjustinfo struct {
	old   _core.Stack
	delta uintptr // ptr distance from old to new stack (newbase - oldbase)
}

// Adjustpointer checks whether *vpp is in the old stack described by adjinfo.
// If so, it rewrites *vpp to point into the new stack.
func adjustpointer(adjinfo *adjustinfo, vpp unsafe.Pointer) {
	pp := (*unsafe.Pointer)(vpp)
	p := *pp
	if _sched.StackDebug >= 4 {
		print("        ", pp, ":", p, "\n")
	}
	if adjinfo.old.Lo <= uintptr(p) && uintptr(p) < adjinfo.old.Hi {
		*pp = _core.Add(p, adjinfo.delta)
		if _sched.StackDebug >= 3 {
			print("        adjust ptr ", pp, ":", p, " -> ", *pp, "\n")
		}
	}
}

type Gobitvector struct {
	N        uintptr
	Bytedata []uint8
}

func Gobv(bv _lock.Bitvector) Gobitvector {
	return Gobitvector{
		uintptr(bv.N),
		(*[1 << 30]byte)(unsafe.Pointer(bv.Bytedata))[:(bv.N+7)/8],
	}
}

func ptrbits(bv *Gobitvector, i uintptr) uint8 {
	return (bv.Bytedata[i/4] >> ((i & 3) * 2)) & 3
}

// bv describes the memory starting at address scanp.
// Adjust any pointers contained therein.
func adjustpointers(scanp unsafe.Pointer, cbv *_lock.Bitvector, adjinfo *adjustinfo, f *_lock.Func) {
	bv := Gobv(*cbv)
	minp := adjinfo.old.Lo
	maxp := adjinfo.old.Hi
	delta := adjinfo.delta
	num := uintptr(bv.N / _sched.BitsPerPointer)
	for i := uintptr(0); i < num; i++ {
		if _sched.StackDebug >= 4 {
			print("        ", _core.Add(scanp, i*_core.PtrSize), ":", mapnames[ptrbits(&bv, i)], ":", _core.Hex(*(*uintptr)(_core.Add(scanp, i*_core.PtrSize))), " # ", i, " ", bv.Bytedata[i/4], "\n")
		}
		switch ptrbits(&bv, i) {
		default:
			_lock.Gothrow("unexpected pointer bits")
		case _sched.BitsDead:
			if _lock.Debug.Gcdead != 0 {
				*(*unsafe.Pointer)(_core.Add(scanp, i*_core.PtrSize)) = unsafe.Pointer(uintptr(_lock.PoisonStack))
			}
		case _sched.BitsScalar:
			// ok
		case _sched.BitsPointer:
			p := *(*unsafe.Pointer)(_core.Add(scanp, i*_core.PtrSize))
			up := uintptr(p)
			if f != nil && 0 < up && up < _core.PageSize && Invalidptr != 0 || up == _lock.PoisonGC || up == _lock.PoisonStack {
				// Looks like a junk value in a pointer slot.
				// Live analysis wrong?
				_core.Getg().M.Traceback = 2
				print("runtime: bad pointer in frame ", _lock.Gofuncname(f), " at ", _core.Add(scanp, i*_core.PtrSize), ": ", p, "\n")
				_lock.Gothrow("invalid stack pointer")
			}
			if minp <= up && up < maxp {
				if _sched.StackDebug >= 3 {
					print("adjust ptr ", p, " ", _lock.Gofuncname(f), "\n")
				}
				*(*unsafe.Pointer)(_core.Add(scanp, i*_core.PtrSize)) = unsafe.Pointer(up + delta)
			}
		}
	}
}

// Note: the argument/return area is adjusted by the callee.
func adjustframe(frame *_lock.Stkframe, arg unsafe.Pointer) bool {
	adjinfo := (*adjustinfo)(arg)
	targetpc := frame.Continpc
	if targetpc == 0 {
		// Frame is dead.
		return true
	}
	f := frame.Fn
	if _sched.StackDebug >= 2 {
		print("    adjusting ", _lock.Funcname(f), " frame=[", _core.Hex(frame.Sp), ",", _core.Hex(frame.Fp), "] pc=", _core.Hex(frame.Pc), " continpc=", _core.Hex(frame.Continpc), "\n")
	}
	if f.Entry == Systemstack_switchPC {
		// A special routine at the bottom of stack of a goroutine that does an systemstack call.
		// We will allow it to be copied even though we don't
		// have full GC info for it (because it is written in asm).
		return true
	}
	if targetpc != f.Entry {
		targetpc--
	}
	pcdata := Pcdatavalue(f, _lock.PCDATA_StackMapIndex, targetpc)
	if pcdata == -1 {
		pcdata = 0 // in prologue
	}

	// Adjust local variables if stack frame has been allocated.
	size := frame.Varp - frame.Sp
	var minsize uintptr
	if _lock.Thechar != '6' && _lock.Thechar != '8' {
		minsize = _core.PtrSize
	} else {
		minsize = 0
	}
	if size > minsize {
		var bv _lock.Bitvector
		stackmap := (*Stackmap)(Funcdata(f, _lock.FUNCDATA_LocalsPointerMaps))
		if stackmap == nil || stackmap.N <= 0 {
			print("runtime: frame ", _lock.Funcname(f), " untyped locals ", _core.Hex(frame.Varp-size), "+", _core.Hex(size), "\n")
			_lock.Gothrow("missing stackmap")
		}
		// Locals bitmap information, scan just the pointers in locals.
		if pcdata < 0 || pcdata >= stackmap.N {
			// don't know where we are
			print("runtime: pcdata is ", pcdata, " and ", stackmap.N, " locals stack map entries for ", _lock.Funcname(f), " (targetpc=", targetpc, ")\n")
			_lock.Gothrow("bad symbol table")
		}
		bv = Stackmapdata(stackmap, pcdata)
		size = (uintptr(bv.N) * _core.PtrSize) / _sched.BitsPerPointer
		if _sched.StackDebug >= 3 {
			print("      locals ", pcdata, "/", stackmap.N, " ", size/_core.PtrSize, " words ", bv.Bytedata, "\n")
		}
		adjustpointers(unsafe.Pointer(frame.Varp-size), &bv, adjinfo, f)
	}

	// Adjust arguments.
	if frame.Arglen > 0 {
		var bv _lock.Bitvector
		if frame.Argmap != nil {
			bv = *frame.Argmap
		} else {
			stackmap := (*Stackmap)(Funcdata(f, _lock.FUNCDATA_ArgsPointerMaps))
			if stackmap == nil || stackmap.N <= 0 {
				print("runtime: frame ", _lock.Funcname(f), " untyped args ", frame.Argp, "+", uintptr(frame.Arglen), "\n")
				_lock.Gothrow("missing stackmap")
			}
			if pcdata < 0 || pcdata >= stackmap.N {
				// don't know where we are
				print("runtime: pcdata is ", pcdata, " and ", stackmap.N, " args stack map entries for ", _lock.Funcname(f), " (targetpc=", targetpc, ")\n")
				_lock.Gothrow("bad symbol table")
			}
			bv = Stackmapdata(stackmap, pcdata)
		}
		if _sched.StackDebug >= 3 {
			print("      args\n")
		}
		adjustpointers(unsafe.Pointer(frame.Argp), &bv, adjinfo, nil)
	}
	return true
}

func adjustctxt(gp *_core.G, adjinfo *adjustinfo) {
	adjustpointer(adjinfo, (unsafe.Pointer)(&gp.Sched.Ctxt))
}

func adjustdefers(gp *_core.G, adjinfo *adjustinfo) {
	// Adjust defer argument blocks the same way we adjust active stack frames.
	tracebackdefers(gp, adjustframe, _core.Noescape(unsafe.Pointer(adjinfo)))

	// Adjust pointers in the Defer structs.
	// Defer structs themselves are never on the stack.
	for d := gp.Defer; d != nil; d = d.Link {
		adjustpointer(adjinfo, (unsafe.Pointer)(&d.Fn))
		adjustpointer(adjinfo, (unsafe.Pointer)(&d.Sp))
		adjustpointer(adjinfo, (unsafe.Pointer)(&d.Panic))
	}
}

func adjustpanics(gp *_core.G, adjinfo *adjustinfo) {
	// Panics are on stack and already adjusted.
	// Update pointer to head of list in G.
	adjustpointer(adjinfo, (unsafe.Pointer)(&gp.Panic))
}

func adjustsudogs(gp *_core.G, adjinfo *adjustinfo) {
	// the data elements pointed to by a SudoG structure
	// might be in the stack.
	for s := gp.Waiting; s != nil; s = s.Waitlink {
		adjustpointer(adjinfo, (unsafe.Pointer)(&s.Elem))
		adjustpointer(adjinfo, (unsafe.Pointer)(&s.Selectdone))
	}
}

func fillstack(stk _core.Stack, b byte) {
	for p := stk.Lo; p < stk.Hi; p++ {
		*(*byte)(unsafe.Pointer(p)) = b
	}
}

// Copies gp's stack to a new stack of a different size.
// Caller must have changed gp status to Gcopystack.
func Copystack(gp *_core.G, newsize uintptr) {
	if gp.Syscallsp != 0 {
		_lock.Gothrow("stack growth not allowed in system call")
	}
	old := gp.Stack
	if old.Lo == 0 {
		_lock.Gothrow("nil stackbase")
	}
	used := old.Hi - gp.Sched.Sp

	// allocate new stack
	new := _sched.Stackalloc(uint32(newsize))
	if _sched.StackPoisonCopy != 0 {
		fillstack(new, 0xfd)
	}
	if _sched.StackDebug >= 1 {
		print("copystack gp=", gp, " [", _core.Hex(old.Lo), " ", _core.Hex(old.Hi-used), " ", _core.Hex(old.Hi), "]/", old.Hi-old.Lo, " -> [", _core.Hex(new.Lo), " ", _core.Hex(new.Hi-used), " ", _core.Hex(new.Hi), "]/", newsize, "\n")
	}

	// adjust pointers in the to-be-copied frames
	var adjinfo adjustinfo
	adjinfo.old = old
	adjinfo.delta = new.Hi - old.Hi
	_lock.Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, 0x7fffffff, adjustframe, _core.Noescape(unsafe.Pointer(&adjinfo)), 0)

	// adjust other miscellaneous things that have pointers into stacks.
	adjustctxt(gp, &adjinfo)
	adjustdefers(gp, &adjinfo)
	adjustpanics(gp, &adjinfo)
	adjustsudogs(gp, &adjinfo)

	// copy the stack to the new location
	if _sched.StackPoisonCopy != 0 {
		fillstack(new, 0xfb)
	}
	_sched.Memmove(unsafe.Pointer(new.Hi-used), unsafe.Pointer(old.Hi-used), used)

	// Swap out old stack for new one
	gp.Stack = new
	gp.Stackguard0 = new.Lo + _core.StackGuard // NOTE: might clobber a preempt request
	gp.Sched.Sp = new.Hi - used

	// free old stack
	if _sched.StackPoisonCopy != 0 {
		fillstack(old, 0xfc)
	}
	if newsize > old.Hi-old.Lo {
		// growing, free stack immediately
		Stackfree(old)
	} else {
		// shrinking, queue up free operation.  We can't actually free the stack
		// just yet because we might run into the following situation:
		// 1) GC starts, scans a SudoG but does not yet mark the SudoG.elem pointer
		// 2) The stack that pointer points to is shrunk
		// 3) The old stack is freed
		// 4) The containing span is marked free
		// 5) GC attempts to mark the SudoG.elem pointer.  The marking fails because
		//    the pointer looks like a pointer into a free span.
		// By not freeing, we prevent step #4 until GC is done.
		_lock.Lock(&_sched.Stackpoolmu)
		*(*_core.Stack)(unsafe.Pointer(old.Lo)) = stackfreequeue
		stackfreequeue = old
		_lock.Unlock(&_sched.Stackpoolmu)
	}
}

// Maybe shrink the stack being used by gp.
// Called at garbage collection time.
func shrinkstack(gp *_core.G) {
	if _lock.Readgstatus(gp) == _lock.Gdead {
		if gp.Stack.Lo != 0 {
			// Free whole stack - it will get reallocated
			// if G is used again.
			Stackfree(gp.Stack)
			gp.Stack.Lo = 0
			gp.Stack.Hi = 0
		}
		return
	}
	if gp.Stack.Lo == 0 {
		_lock.Gothrow("missing stack in shrinkstack")
	}

	oldsize := gp.Stack.Hi - gp.Stack.Lo
	newsize := oldsize / 2
	if newsize < _core.FixedStack {
		return // don't shrink below the minimum-sized stack
	}
	used := gp.Stack.Hi - gp.Sched.Sp
	if used >= oldsize/4 {
		return // still using at least 1/4 of the segment.
	}

	// We can't copy the stack if we're in a syscall.
	// The syscall might have pointers into the stack.
	if gp.Syscallsp != 0 {
		return
	}
	if _core.Goos_windows != 0 && gp.M != nil && gp.M.Libcallsp != 0 {
		return
	}

	if _sched.StackDebug > 0 {
		print("shrinking stack ", oldsize, "->", newsize, "\n")
	}

	oldstatus := casgcopystack(gp)
	Copystack(gp, newsize)
	_sched.Casgstatus(gp, _lock.Gcopystack, oldstatus)
}

// Do any delayed stack freeing that was queued up during GC.
func shrinkfinish() {
	_lock.Lock(&_sched.Stackpoolmu)
	s := stackfreequeue
	stackfreequeue = _core.Stack{}
	_lock.Unlock(&_sched.Stackpoolmu)
	for s.Lo != 0 {
		t := *(*_core.Stack)(unsafe.Pointer(s.Lo))
		Stackfree(s)
		s = t
	}
}
