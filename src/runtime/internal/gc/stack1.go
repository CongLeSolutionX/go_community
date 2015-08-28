// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

// List of stack spans to be freed at the end of GC. Protected by
// stackpoolmu.
var StackFreeQueue _base.Mspan

// Adds stack x to the free pool.  Must be called with stackpoolmu held.
func stackpoolfree(x _base.Gclinkptr, order uint8) {
	s := mHeap_Lookup(&_base.Mheap_, (unsafe.Pointer)(x))
	if s.State != _base.MSpanStack {
		_base.Throw("freeing stack not in a stack span")
	}
	if s.Freelist.Ptr() == nil {
		// s will now have a free stack
		_base.MSpanList_Insert(&_base.Stackpool[order], s)
	}
	x.Ptr().Next = s.Freelist
	s.Freelist = x
	s.Ref--
	if _base.Gcphase == _base.GCoff && s.Ref == 0 {
		// Span is completely free. Return it to the heap
		// immediately if we're sweeping.
		//
		// If GC is active, we delay the free until the end of
		// GC to avoid the following type of situation:
		//
		// 1) GC starts, scans a SudoG but does not yet mark the SudoG.elem pointer
		// 2) The stack that pointer points to is copied
		// 3) The old stack is freed
		// 4) The containing span is marked free
		// 5) GC attempts to mark the SudoG.elem pointer. The
		//    marking fails because the pointer looks like a
		//    pointer into a free span.
		//
		// By not freeing, we prevent step #4 until GC is done.
		_base.MSpanList_Remove(s)
		s.Freelist = 0
		mHeap_FreeStack(&_base.Mheap_, s)
	}
}

func stackcacherelease(c *_base.Mcache, order uint8) {
	if _base.StackDebug >= 1 {
		print("stackcacherelease order=", order, "\n")
	}
	x := c.Stackcache[order].List
	size := c.Stackcache[order].Size
	_base.Lock(&_base.Stackpoolmu)
	for size > _base.StackCacheSize/2 {
		y := x.Ptr().Next
		stackpoolfree(x, order)
		x = y
		size -= _base.FixedStack << order
	}
	_base.Unlock(&_base.Stackpoolmu)
	c.Stackcache[order].List = x
	c.Stackcache[order].Size = size
}

func stackcache_clear(c *_base.Mcache) {
	if _base.StackDebug >= 1 {
		print("stackcache clear\n")
	}
	_base.Lock(&_base.Stackpoolmu)
	for order := uint8(0); order < _base.NumStackOrders; order++ {
		x := c.Stackcache[order].List
		for x.Ptr() != nil {
			y := x.Ptr().Next
			stackpoolfree(x, order)
			x = y
		}
		c.Stackcache[order].List = 0
		c.Stackcache[order].Size = 0
	}
	_base.Unlock(&_base.Stackpoolmu)
}

func Stackfree(stk _base.Stack, n uintptr) {
	gp := _base.Getg()
	v := (unsafe.Pointer)(stk.Lo)
	if n&(n-1) != 0 {
		_base.Throw("stack not a power of 2")
	}
	if stk.Lo+n < stk.Hi {
		_base.Throw("bad stack size")
	}
	if _base.StackDebug >= 1 {
		println("stackfree", v, n)
		_base.Memclr(v, n) // for testing, clobber stack data
	}
	if _base.Debug.Efence != 0 || _base.StackFromSystem != 0 {
		if _base.Debug.Efence != 0 || _base.StackFaultOnFree != 0 {
			sysFault(v, n)
		} else {
			_base.SysFree(v, n, &_base.Memstats.Stacks_sys)
		}
		return
	}
	if _base.StackCache != 0 && n < _base.FixedStack<<_base.NumStackOrders && n < _base.StackCacheSize {
		order := uint8(0)
		n2 := n
		for n2 > _base.FixedStack {
			order++
			n2 >>= 1
		}
		x := _base.Gclinkptr(v)
		c := gp.M.Mcache
		if c == nil || gp.M.Preemptoff != "" || gp.M.Helpgc != 0 {
			_base.Lock(&_base.Stackpoolmu)
			stackpoolfree(x, order)
			_base.Unlock(&_base.Stackpoolmu)
		} else {
			if c.Stackcache[order].Size >= _base.StackCacheSize {
				stackcacherelease(c, order)
			}
			x.Ptr().Next = c.Stackcache[order].List
			c.Stackcache[order].List = x
			c.Stackcache[order].Size += n
		}
	} else {
		s := mHeap_Lookup(&_base.Mheap_, v)
		if s.State != _base.MSpanStack {
			println(_base.Hex(s.Start<<_base.PageShift), v)
			_base.Throw("bad span state")
		}
		if _base.Gcphase == _base.GCoff {
			// Free the stack immediately if we're
			// sweeping.
			mHeap_FreeStack(&_base.Mheap_, s)
		} else {
			// Otherwise, add it to a list of stack spans
			// to be freed at the end of GC.
			//
			// TODO(austin): Make it possible to re-use
			// these spans as stacks, like we do for small
			// stack spans. (See issue #11466.)
			_base.Lock(&_base.Stackpoolmu)
			_base.MSpanList_Insert(&StackFreeQueue, s)
			_base.Unlock(&_base.Stackpoolmu)
		}
	}
}

var ptrnames = []string{
	0: "scalar",
	1: "ptr",
}

// Stack frame layout
//
// (x86)
// +------------------+
// | args from caller |
// +------------------+ <- frame->argp
// |  return address  |
// +------------------+
// |  caller's BP (*) | (*) if framepointer_enabled && varp < sp
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
	old   _base.Stack
	delta uintptr // ptr distance from old to new stack (newbase - oldbase)
}

// Adjustpointer checks whether *vpp is in the old stack described by adjinfo.
// If so, it rewrites *vpp to point into the new stack.
func adjustpointer(adjinfo *adjustinfo, vpp unsafe.Pointer) {
	pp := (*unsafe.Pointer)(vpp)
	p := *pp
	if _base.StackDebug >= 4 {
		print("        ", pp, ":", p, "\n")
	}
	if adjinfo.old.Lo <= uintptr(p) && uintptr(p) < adjinfo.old.Hi {
		*pp = _base.Add(p, adjinfo.delta)
		if _base.StackDebug >= 3 {
			print("        adjust ptr ", pp, ":", p, " -> ", *pp, "\n")
		}
	}
}

type Gobitvector struct {
	N        uintptr
	Bytedata []uint8
}

func Gobv(bv _base.Bitvector) Gobitvector {
	return Gobitvector{
		uintptr(bv.N),
		(*[1 << 30]byte)(unsafe.Pointer(bv.Bytedata))[:(bv.N+7)/8],
	}
}

func ptrbit(bv *Gobitvector, i uintptr) uint8 {
	return (bv.Bytedata[i/8] >> (i % 8)) & 1
}

// bv describes the memory starting at address scanp.
// Adjust any pointers contained therein.
func adjustpointers(scanp unsafe.Pointer, cbv *_base.Bitvector, adjinfo *adjustinfo, f *_base.Func) {
	bv := Gobv(*cbv)
	minp := adjinfo.old.Lo
	maxp := adjinfo.old.Hi
	delta := adjinfo.delta
	num := uintptr(bv.N)
	for i := uintptr(0); i < num; i++ {
		if _base.StackDebug >= 4 {
			print("        ", _base.Add(scanp, i*_base.PtrSize), ":", ptrnames[ptrbit(&bv, i)], ":", _base.Hex(*(*uintptr)(_base.Add(scanp, i*_base.PtrSize))), " # ", i, " ", bv.Bytedata[i/8], "\n")
		}
		if ptrbit(&bv, i) == 1 {
			pp := (*uintptr)(_base.Add(scanp, i*_base.PtrSize))
			p := *pp
			if f != nil && 0 < p && p < _base.PageSize && _base.Debug.Invalidptr != 0 || p == _base.PoisonStack {
				// Looks like a junk value in a pointer slot.
				// Live analysis wrong?
				_base.Getg().M.Traceback = 2
				print("runtime: bad pointer in frame ", _base.Funcname(f), " at ", pp, ": ", _base.Hex(p), "\n")
				_base.Throw("invalid stack pointer")
			}
			if minp <= p && p < maxp {
				if _base.StackDebug >= 3 {
					print("adjust ptr ", p, " ", _base.Funcname(f), "\n")
				}
				*pp = p + delta
			}
		}
	}
}

// Note: the argument/return area is adjusted by the callee.
func adjustframe(frame *_base.Stkframe, arg unsafe.Pointer) bool {
	adjinfo := (*adjustinfo)(arg)
	targetpc := frame.Continpc
	if targetpc == 0 {
		// Frame is dead.
		return true
	}
	f := frame.Fn
	if _base.StackDebug >= 2 {
		print("    adjusting ", _base.Funcname(f), " frame=[", _base.Hex(frame.Sp), ",", _base.Hex(frame.Fp), "] pc=", _base.Hex(frame.Pc), " continpc=", _base.Hex(frame.Continpc), "\n")
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
	pcdata := Pcdatavalue(f, _base.PCDATA_StackMapIndex, targetpc)
	if pcdata == -1 {
		pcdata = 0 // in prologue
	}

	// Adjust local variables if stack frame has been allocated.
	size := frame.Varp - frame.Sp
	var minsize uintptr
	switch _base.Thechar {
	case '6', '8':
		minsize = 0
	case '7':
		minsize = SpAlign
	default:
		minsize = _base.PtrSize
	}
	if size > minsize {
		var bv _base.Bitvector
		stackmap := (*Stackmap)(Funcdata(f, _base.FUNCDATA_LocalsPointerMaps))
		if stackmap == nil || stackmap.N <= 0 {
			print("runtime: frame ", _base.Funcname(f), " untyped locals ", _base.Hex(frame.Varp-size), "+", _base.Hex(size), "\n")
			_base.Throw("missing stackmap")
		}
		// Locals bitmap information, scan just the pointers in locals.
		if pcdata < 0 || pcdata >= stackmap.N {
			// don't know where we are
			print("runtime: pcdata is ", pcdata, " and ", stackmap.N, " locals stack map entries for ", _base.Funcname(f), " (targetpc=", targetpc, ")\n")
			_base.Throw("bad symbol table")
		}
		bv = Stackmapdata(stackmap, pcdata)
		size = uintptr(bv.N) * _base.PtrSize
		if _base.StackDebug >= 3 {
			print("      locals ", pcdata, "/", stackmap.N, " ", size/_base.PtrSize, " words ", bv.Bytedata, "\n")
		}
		adjustpointers(unsafe.Pointer(frame.Varp-size), &bv, adjinfo, f)
	}

	// Adjust saved base pointer if there is one.
	if _base.Thechar == '6' && frame.Argp-frame.Varp == 2*_base.RegSize {
		if !_base.Framepointer_enabled {
			print("runtime: found space for saved base pointer, but no framepointer experiment\n")
			print("argp=", _base.Hex(frame.Argp), " varp=", _base.Hex(frame.Varp), "\n")
			_base.Throw("bad frame layout")
		}
		if _base.StackDebug >= 3 {
			print("      saved bp\n")
		}
		adjustpointer(adjinfo, unsafe.Pointer(frame.Varp))
	}

	// Adjust arguments.
	if frame.Arglen > 0 {
		var bv _base.Bitvector
		if frame.Argmap != nil {
			bv = *frame.Argmap
		} else {
			stackmap := (*Stackmap)(Funcdata(f, _base.FUNCDATA_ArgsPointerMaps))
			if stackmap == nil || stackmap.N <= 0 {
				print("runtime: frame ", _base.Funcname(f), " untyped args ", frame.Argp, "+", uintptr(frame.Arglen), "\n")
				_base.Throw("missing stackmap")
			}
			if pcdata < 0 || pcdata >= stackmap.N {
				// don't know where we are
				print("runtime: pcdata is ", pcdata, " and ", stackmap.N, " args stack map entries for ", _base.Funcname(f), " (targetpc=", targetpc, ")\n")
				_base.Throw("bad symbol table")
			}
			bv = Stackmapdata(stackmap, pcdata)
		}
		if _base.StackDebug >= 3 {
			print("      args\n")
		}
		adjustpointers(unsafe.Pointer(frame.Argp), &bv, adjinfo, nil)
	}
	return true
}

func adjustctxt(gp *_base.G, adjinfo *adjustinfo) {
	adjustpointer(adjinfo, (unsafe.Pointer)(&gp.Sched.Ctxt))
}

func adjustdefers(gp *_base.G, adjinfo *adjustinfo) {
	// Adjust defer argument blocks the same way we adjust active stack frames.
	tracebackdefers(gp, adjustframe, _base.Noescape(unsafe.Pointer(adjinfo)))

	// Adjust pointers in the Defer structs.
	// Defer structs themselves are never on the stack.
	for d := gp.Defer; d != nil; d = d.Link {
		adjustpointer(adjinfo, (unsafe.Pointer)(&d.Fn))
		adjustpointer(adjinfo, (unsafe.Pointer)(&d.Sp))
		adjustpointer(adjinfo, (unsafe.Pointer)(&d.Panic))
	}
}

func adjustpanics(gp *_base.G, adjinfo *adjustinfo) {
	// Panics are on stack and already adjusted.
	// Update pointer to head of list in G.
	adjustpointer(adjinfo, (unsafe.Pointer)(&gp.Panic))
}

func adjustsudogs(gp *_base.G, adjinfo *adjustinfo) {
	// the data elements pointed to by a SudoG structure
	// might be in the stack.
	for s := gp.Waiting; s != nil; s = s.Waitlink {
		adjustpointer(adjinfo, (unsafe.Pointer)(&s.Elem))
		adjustpointer(adjinfo, (unsafe.Pointer)(&s.Selectdone))
	}
}

func adjuststkbar(gp *_base.G, adjinfo *adjustinfo) {
	for i := int(gp.StkbarPos); i < len(gp.Stkbar); i++ {
		adjustpointer(adjinfo, (unsafe.Pointer)(&gp.Stkbar[i].SavedLRPtr))
	}
}

func fillstack(stk _base.Stack, b byte) {
	for p := stk.Lo; p < stk.Hi; p++ {
		*(*byte)(unsafe.Pointer(p)) = b
	}
}

// Copies gp's stack to a new stack of a different size.
// Caller must have changed gp status to Gcopystack.
func Copystack(gp *_base.G, newsize uintptr) {
	if gp.Syscallsp != 0 {
		_base.Throw("stack growth not allowed in system call")
	}
	old := gp.Stack
	if old.Lo == 0 {
		_base.Throw("nil stackbase")
	}
	used := old.Hi - gp.Sched.Sp

	// allocate new stack
	new, newstkbar := _base.Stackalloc(uint32(newsize))
	if _base.StackPoisonCopy != 0 {
		fillstack(new, 0xfd)
	}
	if _base.StackDebug >= 1 {
		print("copystack gp=", gp, " [", _base.Hex(old.Lo), " ", _base.Hex(old.Hi-used), " ", _base.Hex(old.Hi), "]/", gp.StackAlloc, " -> [", _base.Hex(new.Lo), " ", _base.Hex(new.Hi-used), " ", _base.Hex(new.Hi), "]/", newsize, "\n")
	}

	// adjust pointers in the to-be-copied frames
	var adjinfo adjustinfo
	adjinfo.old = old
	adjinfo.delta = new.Hi - old.Hi
	_base.Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, 0x7fffffff, adjustframe, _base.Noescape(unsafe.Pointer(&adjinfo)), 0)

	// adjust other miscellaneous things that have pointers into stacks.
	adjustctxt(gp, &adjinfo)
	adjustdefers(gp, &adjinfo)
	adjustpanics(gp, &adjinfo)
	adjustsudogs(gp, &adjinfo)
	adjuststkbar(gp, &adjinfo)

	// copy the stack to the new location
	if _base.StackPoisonCopy != 0 {
		fillstack(new, 0xfb)
	}
	_base.Memmove(unsafe.Pointer(new.Hi-used), unsafe.Pointer(old.Hi-used), used)

	// copy old stack barriers to new stack barrier array
	newstkbar = newstkbar[:len(gp.Stkbar)]
	copy(newstkbar, gp.Stkbar)

	// Swap out old stack for new one
	gp.Stack = new
	gp.Stackguard0 = new.Lo + _base.StackGuard // NOTE: might clobber a preempt request
	gp.Sched.Sp = new.Hi - used
	oldsize := gp.StackAlloc
	gp.StackAlloc = newsize
	gp.Stkbar = newstkbar

	// free old stack
	if _base.StackPoisonCopy != 0 {
		fillstack(old, 0xfc)
	}
	Stackfree(old, oldsize)
}

// Maybe shrink the stack being used by gp.
// Called at garbage collection time.
func shrinkstack(gp *_base.G) {
	if _base.Readgstatus(gp) == _base.Gdead {
		if gp.Stack.Lo != 0 {
			// Free whole stack - it will get reallocated
			// if G is used again.
			Stackfree(gp.Stack, gp.StackAlloc)
			gp.Stack.Lo = 0
			gp.Stack.Hi = 0
			gp.Stkbar = nil
			gp.StkbarPos = 0
		}
		return
	}
	if gp.Stack.Lo == 0 {
		_base.Throw("missing stack in shrinkstack")
	}

	if _base.Debug.Gcshrinkstackoff > 0 {
		return
	}

	oldsize := gp.StackAlloc
	newsize := oldsize / 2
	// Don't shrink the allocation below the minimum-sized stack
	// allocation.
	if newsize < _base.FixedStack {
		return
	}
	// Compute how much of the stack is currently in use and only
	// shrink the stack if gp is using less than a quarter of its
	// current stack. The currently used stack includes everything
	// down to the SP plus the stack guard space that ensures
	// there's room for nosplit functions.
	avail := gp.Stack.Hi - gp.Stack.Lo
	if used := gp.Stack.Hi - gp.Sched.Sp + _base.StackLimit; used >= avail/4 {
		return
	}

	// We can't copy the stack if we're in a syscall.
	// The syscall might have pointers into the stack.
	if gp.Syscallsp != 0 {
		return
	}
	if _base.Goos_windows != 0 && gp.M != nil && gp.M.Libcallsp != 0 {
		return
	}

	if _base.StackDebug > 0 {
		print("shrinking stack ", oldsize, "->", newsize, "\n")
	}

	oldstatus := casgcopystack(gp)
	Copystack(gp, newsize)
	_base.Casgstatus(gp, _base.Gcopystack, oldstatus)
}

// freeStackSpans frees unused stack spans at the end of GC.
func freeStackSpans() {
	_base.Lock(&_base.Stackpoolmu)

	// Scan stack pools for empty stack spans.
	for order := range _base.Stackpool {
		list := &_base.Stackpool[order]
		for s := list.Next; s != list; {
			next := s.Next
			if s.Ref == 0 {
				_base.MSpanList_Remove(s)
				s.Freelist = 0
				mHeap_FreeStack(&_base.Mheap_, s)
			}
			s = next
		}
	}

	// Free queued stack spans.
	for StackFreeQueue.Next != &StackFreeQueue {
		s := StackFreeQueue.Next
		_base.MSpanList_Remove(s)
		mHeap_FreeStack(&_base.Mheap_, s)
	}

	_base.Unlock(&_base.Stackpoolmu)
}
