// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// TODO(rsc): The code having to do with the heap bitmap needs very serious cleanup.
// It has gotten completely out of control.

// Garbage collector (GC).
//
// The GC runs concurrently with mutator threads, is type accurate (aka precise), allows multiple
// GC thread to run in parallel. It is a concurrent mark and sweep that uses a write barrier. It is
// non-generational and non-compacting. Allocation is done using size segregated per P allocation
// areas to minimize fragmentation while eliminating locks in the common case.
//
// The algorithm decomposes into several steps.
// This is a high level description of the algorithm being used. For an overview of GC a good
// place to start is Richard Jones' gchandbook.org.
//
// The algorithm's intellectual heritage includes Dijkstra's on-the-fly algorithm, see
// Edsger W. Dijkstra, Leslie Lamport, A. J. Martin, C. S. Scholten, and E. F. M. Steffens. 1978.
// On-the-fly garbage collection: an exercise in cooperation. Commun. ACM 21, 11 (November 1978),
// 966-975.
// For journal quality proofs that these steps are complete, correct, and terminate see
// Hudson, R., and Moss, J.E.B. Copying Garbage Collection without stopping the world.
// Concurrency and Computation: Practice and Experience 15(3-5), 2003.
//
//  0. Set phase = GCscan from GCoff.
//  1. Wait for all P's to acknowledge phase change.
//         At this point all goroutines have passed through a GC safepoint and
//         know we are in the GCscan phase.
//  2. GC scans all goroutine stacks, mark and enqueues all encountered pointers
//       (marking avoids most duplicate enqueuing but races may produce benign duplication).
//       Preempted goroutines are scanned before P schedules next goroutine.
//  3. Set phase = GCmark.
//  4. Wait for all P's to acknowledge phase change.
//  5. Now write barrier marks and enqueues black, grey, or white to white pointers.
//       Malloc still allocates white (non-marked) objects.
//  6. Meanwhile GC transitively walks the heap marking reachable objects.
//  7. When GC finishes marking heap, it preempts P's one-by-one and
//       retakes partial wbufs (filled by write barrier or during a stack scan of the goroutine
//       currently scheduled on the P).
//  8. Once the GC has exhausted all available marking work it sets phase = marktermination.
//  9. Wait for all P's to acknowledge phase change.
// 10. Malloc now allocates black objects, so number of unmarked reachable objects
//        monotonically decreases.
// 11. GC preempts P's one-by-one taking partial wbufs and marks all unmarked yet
//        reachable objects.
// 12. When GC completes a full cycle over P's and discovers no new grey
//         objects, (which means all reachable objects are marked) set phase = GCsweep.
// 13. Wait for all P's to acknowledge phase change.
// 14. Now malloc allocates white (but sweeps spans before use).
//         Write barrier becomes nop.
// 15. GC does background sweeping, see description below.
// 16. When sweeping is complete set phase to GCoff.
// 17. When sufficient allocation has taken place replay the sequence starting at 0 above,
//         see discussion of GC rate below.

// Changing phases.
// Phases are changed by setting the gcphase to the next phase and possibly calling ackgcphase.
// All phase action must be benign in the presence of a change.
// Starting with GCoff
// GCoff to GCscan
//     GSscan scans stacks and globals greying them and never marks an object black.
//     Once all the P's are aware of the new phase they will scan gs on preemption.
//     This means that the scanning of preempted gs can't start until all the Ps
//     have acknowledged.
// GCscan to GCmark
//     GCMark turns on the write barrier which also only greys objects. No scanning
//     of objects (making them black) can happen until all the Ps have acknowledged
//     the phase change.
// GCmark to GCmarktermination
//     The only change here is that we start allocating black so the Ps must acknowledge
//     the change before we begin the termination algorithm
// GCmarktermination to GSsweep
//     Object currently on the freelist must be marked black for this to work.
//     Are things on the free lists black or white? How does the sweep phase work?

// Concurrent sweep.
// The sweep phase proceeds concurrently with normal program execution.
// The heap is swept span-by-span both lazily (when a goroutine needs another span)
// and concurrently in a background goroutine (this helps programs that are not CPU bound).
// However, at the end of the stop-the-world GC phase we don't know the size of the live heap,
// and so next_gc calculation is tricky and happens as follows.
// At the end of the stop-the-world phase next_gc is conservatively set based on total
// heap size; all spans are marked as "needs sweeping".
// Whenever a span is swept, next_gc is decremented by GOGC*newly_freed_memory.
// The background sweeper goroutine simply sweeps spans one-by-one bringing next_gc
// closer to the target value. However, this is not enough to avoid over-allocating memory.
// Consider that a goroutine wants to allocate a new span for a large object and
// there are no free swept spans, but there are small-object unswept spans.
// If the goroutine naively allocates a new span, it can surpass the yet-unknown
// target next_gc value. In order to prevent such cases (1) when a goroutine needs
// to allocate a new small-object span, it sweeps small-object spans for the same
// object size until it frees at least one object; (2) when a goroutine needs to
// allocate large-object span from heap, it sweeps spans until it frees at least
// that many pages into heap. Together these two measures ensure that we don't surpass
// target next_gc value by a large margin. There is an exception: if a goroutine sweeps
// and frees two nonadjacent one-page spans to the heap, it will allocate a new two-page span,
// but there can still be other one-page unswept spans which could be combined into a
// two-page span.
// It's critical to ensure that no operations proceed on unswept spans (that would corrupt
// mark bits in GC bitmap). During GC all mcaches are flushed into the central cache,
// so they are empty. When a goroutine grabs a new span into mcache, it sweeps it.
// When a goroutine explicitly frees an object or sets a finalizer, it ensures that
// the span is swept (either by sweeping it, or by waiting for the concurrent sweep to finish).
// The finalizer goroutine is kicked off only when all spans are swept.
// When the next GC starts, it sweeps all not-yet-swept spans (if any).

// GC rate.
// Next GC is after we've allocated an extra amount of memory proportional to
// the amount already in use. The proportion is controlled by GOGC environment variable
// (100 by default). If GOGC=100 and we're using 4M, we'll GC again when we get to 8M
// (this mark is tracked in next_gc variable). This keeps the GC cost in linear
// proportion to the allocation cost. Adjusting GOGC just changes the linear constant
// (and also the amount of extra memory used).

package sched

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	"unsafe"
)

const (
	DebugGC         = 0
	DebugGCPtrs     = false // if true, print trace of every pointer load during GC
	ConcurrentSweep = true

	WorkbufSize     = 4 * 256
	FinBlockSize    = 4 * 1024
	RootData        = 0
	RootBss         = 1
	RootFinalizers  = 2
	RootSpans       = 3
	RootFlushCaches = 4
	RootCount       = 5
)

type Workbuf struct {
	node lfnode // must be first
	Nobj uintptr
	Obj  [(WorkbufSize - unsafe.Sizeof(lfnode{}) - _core.PtrSize) / _core.PtrSize]uintptr
}

type Workdata struct {
	Full    uint64                     // lock-free list of full blocks workbuf
	empty   uint64                     // lock-free list of empty blocks workbuf
	Partial uint64                     // lock-free list of partially filled blocks workbuf
	pad0    [_lock.CacheLineSize]uint8 // prevents false-sharing between full/empty and nproc/nwait
	Nproc   uint32
	Tstart  int64
	Nwait   uint32
	Ndone   uint32
	Alldone _core.Note
	Markfor *Parfor

	// Copy of mheap.allspans for marker or sweeper.
	Spans []*_core.Mspan
}

var Work Workdata

// To help debug the concurrent GC we remark with the world
// stopped ensuring that any object encountered has their normal
// mark bit set. To do this we use an orthogonal bit
// pattern to indicate the object is marked. The following pattern
// uses the upper two bits in the object's bounday nibble.
// 01: scalar  not marked
// 10: pointer not marked
// 11: pointer     marked
// 00: scalar      marked
// Xoring with 01 will flip the pattern from marked to unmarked and vica versa.
// The higher bit is 1 for pointers and 0 for scalars, whether the object
// is marked or not.
// The first nibble no longer holds the typeDead pattern indicating that the
// there are no more pointers in the object. This information is held
// in the second nibble.

// When marking an object if the bool checkmarkphase is true one uses the above
// encoding, otherwise one uses the bitMarked bit in the lower two bits
// of the nibble.
var Checkmarkphase = false

// inheap reports whether b is a pointer into a (potentially dead) heap object.
// It returns false for pointers into stack spans.
//go:nowritebarrier
func inheap(b uintptr) bool {
	if b == 0 || b < _lock.Mheap_.Arena_start || b >= _lock.Mheap_.Arena_used {
		return false
	}
	// Not a beginning of a block, consult span table to find the block beginning.
	k := b >> _core.PageShift
	x := k
	x -= _lock.Mheap_.Arena_start >> _core.PageShift
	s := H_spans[x]
	if s == nil || _core.PageID(k) < s.Start || b >= s.Limit || s.State != XMSpanInUse {
		return false
	}
	return true
}

// Slow for now as we serialize this, since this is on a debug path
// speed is not critical at this point.
var andlock _core.Mutex

//go:nowritebarrier
func atomicand8(src *byte, val byte) {
	_lock.Lock(&andlock)
	*src &= val
	_lock.Unlock(&andlock)
}

// obj is the start of an object with mark mbits.
// If it isn't already marked, mark it and enqueue into workbuf.
// Return possibly new workbuf to use.
// base and off are for debugging only and could be removed.
//go:nowritebarrier
func greyobject(obj, base, off uintptr, hbits HeapBits, wbuf *Workbuf) *Workbuf {
	// obj should be start of allocation, and so must be at least pointer-aligned.
	if obj&(_core.PtrSize-1) != 0 {
		_lock.Throw("greyobject: obj not pointer-aligned")
	}

	if Checkmarkphase {
		if !hbits.IsMarked() {
			print("runtime:greyobject: checkmarks finds unexpected unmarked object obj=", _core.Hex(obj), "\n")
			print("runtime: found obj at *(", _core.Hex(base), "+", _core.Hex(off), ")\n")

			// Dump the source (base) object

			kb := base >> _core.PageShift
			xb := kb
			xb -= _lock.Mheap_.Arena_start >> _core.PageShift
			sb := H_spans[xb]
			Printlock()
			print("runtime:greyobject Span: base=", _core.Hex(base), " kb=", _core.Hex(kb))
			if sb == nil {
				print(" sb=nil\n")
			} else {
				print(" sb.start*_PageSize=", _core.Hex(sb.Start*_core.PageSize), " sb.limit=", _core.Hex(sb.Limit), " sb.sizeclass=", sb.Sizeclass, " sb.elemsize=", sb.Elemsize, "\n")
				// base is (a pointer to) the source object holding the reference to object. Create a pointer to each of the fields
				// fields in base and print them out as hex values.
				for i := 0; i < int(sb.Elemsize/_core.PtrSize); i++ {
					print(" *(base+", i*_core.PtrSize, ") = ", _core.Hex(*(*uintptr)(unsafe.Pointer(base + uintptr(i)*_core.PtrSize))), "\n")
				}
			}

			// Dump the object

			k := obj >> _core.PageShift
			x := k
			x -= _lock.Mheap_.Arena_start >> _core.PageShift
			s := H_spans[x]
			print("runtime:greyobject Span: obj=", _core.Hex(obj), " k=", _core.Hex(k))
			if s == nil {
				print(" s=nil\n")
			} else {
				print(" s.start=", _core.Hex(s.Start*_core.PageSize), " s.limit=", _core.Hex(s.Limit), " s.sizeclass=", s.Sizeclass, " s.elemsize=", s.Elemsize, "\n")
				// NOTE(rsc): This code is using s.sizeclass as an approximation of the
				// number of pointer-sized words in an object. Perhaps not what was intended.
				for i := 0; i < int(s.Sizeclass); i++ {
					print(" *(obj+", i*_core.PtrSize, ") = ", _core.Hex(*(*uintptr)(unsafe.Pointer(obj + uintptr(i)*_core.PtrSize))), "\n")
				}
			}
			_lock.Throw("checkmark found unmarked object")
		}
		if !hbits.isCheckmarked() {
			return wbuf
		}
		hbits.setCheckmarked()
		if !hbits.isCheckmarked() {
			_lock.Throw("setCheckmarked and isCheckmarked disagree")
		}
	} else {
		// If marked we have nothing to do.
		if hbits.IsMarked() {
			return wbuf
		}

		// Each byte of GC bitmap holds info for two words.
		// Might be racing with other updates, so use atomic update always.
		// We used to be clever here and use a non-atomic update in certain
		// cases, but it's not worth the risk.
		hbits.SetMarked()
	}

	if !Checkmarkphase && hbits.TypeBits() == TypeDead {
		return wbuf // noscan object
	}

	// Queue the obj for scanning. The PREFETCH(obj) logic has been removed but
	// seems like a nice optimization that can be added back in.
	// There needs to be time between the PREFETCH and the use.
	// Previously we put the obj in an 8 element buffer that is drained at a rate
	// to give the PREFETCH time to do its work.
	// Use of PREFETCHNTA might be more appropriate than PREFETCH

	// If workbuf is full, obtain an empty one.
	if wbuf.Nobj >= uintptr(len(wbuf.Obj)) {
		wbuf = getempty(wbuf)
	}

	wbuf.Obj[wbuf.Nobj] = obj
	wbuf.Nobj++
	return wbuf
}

// Scan the object b of size n, adding pointers to wbuf.
// Return possibly new wbuf to use.
// If ptrmask != nil, it specifies where pointers are in b.
// If ptrmask == nil, the GC bitmap should be consulted.
// In this case, n may be an overestimate of the size; the GC bitmap
// must also be used to make sure the scan stops at the end of b.
//go:nowritebarrier
func Scanobject(b, n uintptr, ptrmask *uint8, wbuf *Workbuf) *Workbuf {
	arena_start := _lock.Mheap_.Arena_start
	arena_used := _lock.Mheap_.Arena_used

	// Find bits of the beginning of the object.
	var hbits HeapBits
	if ptrmask == nil {
		b, hbits = heapBitsForObject(b)
		if b == 0 {
			return wbuf
		}
		if n == 0 {
			n = _lock.Mheap_.Arena_used - b
		}
	}
	for i := uintptr(0); i < n; i += _core.PtrSize {
		// Find bits for this word.
		var bits uintptr
		if ptrmask != nil {
			// dense mask (stack or data)
			bits = (uintptr(*(*byte)(_core.Add(unsafe.Pointer(ptrmask), (i/_core.PtrSize)/4))) >> (((i / _core.PtrSize) % 4) * TypeBitsWidth)) & TypeMask
		} else {
			// Check if we have reached end of span.
			// n is an overestimate of the size of the object.
			if (b+i)%_core.PageSize == 0 && H_spans[(b-arena_start)>>_core.PageShift] != H_spans[(b+i-arena_start)>>_core.PageShift] {
				break
			}

			bits = uintptr(hbits.TypeBits())
			if i > 0 && (hbits.isBoundary() || bits == TypeDead) {
				break // reached beginning of the next object
			}
			hbits = hbits.Next()
		}

		if bits <= TypeScalar { // typeScalar, typeDead, typeScalarMarked
			continue
		}

		if bits&TypePointer != TypePointer {
			print("gc checkmarkphase=", Checkmarkphase, " b=", _core.Hex(b), " ptrmask=", ptrmask, "\n")
			_lock.Throw("unexpected garbage collection bits")
		}

		obj := *(*uintptr)(unsafe.Pointer(b + i))

		// At this point we have extracted the next potential pointer.
		// Check if it points into heap.
		if obj == 0 || obj < arena_start || obj >= arena_used {
			continue
		}

		if _lock.Mheap_.Shadow_enabled && _lock.Debug.Wbshadow >= 2 && _lock.Debug.Gccheckmark > 0 && Checkmarkphase {
			checkwbshadow((*uintptr)(unsafe.Pointer(b + i)))
		}

		// Mark the object.
		if obj, hbits := heapBitsForObject(obj); obj != 0 {
			wbuf = greyobject(obj, b, i, hbits, wbuf)
		}
	}
	return wbuf
}

// scanblock starts by scanning b as scanobject would.
// If the gcphase is GCscan, that's all scanblock does.
// Otherwise it traverses some fraction of the pointers it found in b, recursively.
// As a special case, scanblock(nil, 0, nil) means to scan previously queued work,
// stopping only when no work is left in the system.
//go:nowritebarrier
func Scanblock(b0, n0 uintptr, ptrmask *uint8) {
	// Use local copies of original parameters, so that a stack trace
	// due to one of the throws below shows the original block
	// base and extent.
	b := b0
	n := n0

	// ptrmask can have 2 possible values:
	// 1. nil - obtain pointer mask from GC bitmap.
	// 2. pointer to a compact mask (for stacks and data).

	wbuf := getpartialorempty()
	if b != 0 {
		wbuf = Scanobject(b, n, ptrmask, wbuf)
		if Gcphase == GCscan {
			if inheap(b) && ptrmask == nil {
				// b is in heap, we are in GCscan so there should be a ptrmask.
				_lock.Throw("scanblock: In GCscan phase and inheap is true.")
			}
			// GCscan only goes one level deep since mark wb not turned on.
			Putpartial(wbuf)
			return
		}
	}

	drainallwbufs := b == 0
	drainworkbuf(wbuf, drainallwbufs)
}

// Scan objects in wbuf until wbuf is empty.
// If drainallwbufs is true find all other available workbufs and repeat the process.
//go:nowritebarrier
func drainworkbuf(wbuf *Workbuf, drainallwbufs bool) {
	if Gcphase != GCmark && Gcphase != GCmarktermination {
		println("gcphase", Gcphase)
		_lock.Throw("scanblock phase")
	}

	for {
		if wbuf.Nobj == 0 {
			if !drainallwbufs {
				Putempty(wbuf)
				return
			}
			// Refill workbuf from global queue.
			wbuf = getfull(wbuf)
			if wbuf == nil { // nil means out of work barrier reached
				return
			}

			if wbuf.Nobj <= 0 {
				_lock.Throw("runtime:scanblock getfull returns empty buffer")
			}
		}

		// If another proc wants a pointer, give it some.
		if Work.Nwait > 0 && wbuf.Nobj > 4 && Work.Full == 0 {
			wbuf = handoff(wbuf)
		}

		// This might be a good place to add prefetch code...
		// if(wbuf->nobj > 4) {
		//         PREFETCH(wbuf->obj[wbuf->nobj - 3];
		//  }
		wbuf.Nobj--
		b := wbuf.Obj[wbuf.Nobj]
		wbuf = Scanobject(b, 0, nil, wbuf)
	}
}

// Get an empty work buffer off the work.empty list,
// allocating new buffers as needed.
//go:nowritebarrier
func getempty(b *Workbuf) *Workbuf {
	if b != nil {
		putfull(b)
		b = nil
	}
	if Work.empty != 0 {
		b = (*Workbuf)(Lfstackpop(&Work.empty))
	}
	if b != nil && b.Nobj != 0 {
		_g_ := _core.Getg()
		print("m", _g_.M.Id, ": getempty: popped b=", b, " with non-zero b.nobj=", b.Nobj, "\n")
		_lock.Throw("getempty: workbuffer not empty, b->nobj not 0")
	}
	if b == nil {
		b = (*Workbuf)(_lock.Persistentalloc(unsafe.Sizeof(*b), _lock.CacheLineSize, &_lock.Memstats.Gc_sys))
		b.Nobj = 0
	}
	return b
}

//go:nowritebarrier
func Putempty(b *Workbuf) {
	if b.Nobj != 0 {
		_lock.Throw("putempty: b->nobj not 0")
	}
	lfstackpush(&Work.empty, &b.node)
}

//go:nowritebarrier
func putfull(b *Workbuf) {
	if b.Nobj <= 0 {
		_lock.Throw("putfull: b->nobj <= 0")
	}
	lfstackpush(&Work.Full, &b.node)
}

// Get an partially empty work buffer
// if none are available get an empty one.
//go:nowritebarrier
func getpartialorempty() *Workbuf {
	b := (*Workbuf)(Lfstackpop(&Work.Partial))
	if b == nil {
		b = getempty(nil)
	}
	return b
}

//go:nowritebarrier
func Putpartial(b *Workbuf) {
	if b.Nobj == 0 {
		lfstackpush(&Work.empty, &b.node)
	} else if b.Nobj < uintptr(len(b.Obj)) {
		lfstackpush(&Work.Partial, &b.node)
	} else if b.Nobj == uintptr(len(b.Obj)) {
		lfstackpush(&Work.Full, &b.node)
	} else {
		print("b=", b, " b.nobj=", b.Nobj, " len(b.obj)=", len(b.Obj), "\n")
		_lock.Throw("putpartial: bad Workbuf b.nobj")
	}
}

// Get a full work buffer off the work.full or a partially
// filled one off the work.partial list. If nothing is available
// wait until all the other gc helpers have finished and then
// return nil.
// getfull acts as a barrier for work.nproc helpers. As long as one
// gchelper is actively marking objects it
// may create a workbuffer that the other helpers can work on.
// The for loop either exits when a work buffer is found
// or when _all_ of the work.nproc GC helpers are in the loop
// looking for work and thus not capable of creating new work.
// This is in fact the termination condition for the STW mark
// phase.
//go:nowritebarrier
func getfull(b *Workbuf) *Workbuf {
	if b != nil {
		Putempty(b)
	}

	b = (*Workbuf)(Lfstackpop(&Work.Full))
	if b == nil {
		b = (*Workbuf)(Lfstackpop(&Work.Partial))
	}
	if b != nil {
		return b
	}

	_lock.Xadd(&Work.Nwait, +1)
	for i := 0; ; i++ {
		if Work.Full != 0 {
			_lock.Xadd(&Work.Nwait, -1)
			b = (*Workbuf)(Lfstackpop(&Work.Full))
			if b == nil {
				b = (*Workbuf)(Lfstackpop(&Work.Partial))
			}
			if b != nil {
				return b
			}
			_lock.Xadd(&Work.Nwait, +1)
		}
		if Work.Nwait == Work.Nproc {
			return nil
		}
		_g_ := _core.Getg()
		if i < 10 {
			_g_.M.Gcstats.Nprocyield++
			_lock.Procyield(20)
		} else if i < 20 {
			_g_.M.Gcstats.Nosyield++
			_core.Osyield()
		} else {
			_g_.M.Gcstats.Nsleep++
			_core.Usleep(100)
		}
	}
}

//go:nowritebarrier
func handoff(b *Workbuf) *Workbuf {
	// Make new buffer with half of b's pointers.
	b1 := getempty(nil)
	n := b.Nobj / 2
	b.Nobj -= n
	b1.Nobj = n
	Memmove(unsafe.Pointer(&b1.Obj[0]), unsafe.Pointer(&b.Obj[b.Nobj]), n*unsafe.Sizeof(b1.Obj[0]))
	_g_ := _core.Getg()
	_g_.M.Gcstats.Nhandoff++
	_g_.M.Gcstats.Nhandoffcnt += uint64(n)

	// Put b on full list - let first half of b get stolen.
	lfstackpush(&Work.Full, &b.node)
	return b1
}

// Shade the object if it isn't already.
// The object is not nil and known to be in the heap.
//go:nowritebarrier
func shade(b uintptr) {
	if !inheap(b) {
		_lock.Throw("shade: passed an address not in the heap")
	}

	wbuf := getpartialorempty()

	if obj, hbits := heapBitsForObject(b); obj != 0 {
		wbuf = greyobject(obj, 0, 0, hbits, wbuf)
	}

	Putpartial(wbuf)
}

//go:nowritebarrier
func gchelper() {
	_g_ := _core.Getg()
	_g_.M.Traceback = 2
	Gchelperstart()

	if Trace.Enabled {
		TraceGCScanStart()
	}

	// parallel mark for over GC roots
	Parfordo(Work.Markfor)
	if Gcphase != GCscan {
		Scanblock(0, 0, nil) // blocks in getfull
	}

	if Trace.Enabled {
		TraceGCScanDone()
	}

	nproc := Work.Nproc // work.nproc can change right after we increment work.ndone
	if _lock.Xadd(&Work.Ndone, +1) == nproc-1 {
		Notewakeup(&Work.Alldone)
	}
	_g_.M.Traceback = 0
}

func Gchelperstart() {
	_g_ := _core.Getg()

	if _g_.M.Helpgc < 0 || _g_.M.Helpgc >= _core.MaxGcproc {
		_lock.Throw("gchelperstart: bad m->helpgc")
	}
	if _g_ != _g_.M.G0 {
		_lock.Throw("gchelper not running on g0 stack")
	}
}
