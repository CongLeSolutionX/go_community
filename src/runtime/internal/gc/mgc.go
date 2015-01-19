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

package gc

import (
	_core "runtime/internal/core"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

// ptrmask for an allocation containing a single pointer.
var oneptr = [...]uint8{_sched.XBitsPointer}

// Initialized from $GOGC.  GOGC=off means no GC.
var Gcpercent int32

// Holding worldsema grants an M the right to try to stop the world.
// The procedure is:
//
//	semacquire(&worldsema);
//	m.gcing = 1;
//	stoptheworld();
//
//	... do stuff ...
//
//	m.gcing = 0;
//	semrelease(&worldsema);
//	starttheworld();
//
var Worldsema uint32 = 1

var Data, Edata, Bss, Ebss, Gcdata, Gcbss struct{}
var Finq *Finblock // list of finalizers that are to be executed
var Finc *Finblock // cache of free blocks
var finptrmask [_sched.FinBlockSize / _core.PtrSize / _sched.XPointersPerByte]byte
var Allfin *Finblock // list of all blocks

var Gcdatamask _lock.Bitvector
var Gcbssmask _lock.Bitvector

var gclock _core.Mutex

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
// The first nibble no longer holds the bitsDead pattern indicating that the
// there are no more pointers in the object. This information is held
// in the second nibble.

// When marking an object if the bool checkmark is true one uses the above
// encoding, otherwise one uses the bitMarked bit in the lower two bits
// of the nibble.
var (
	Gccheckmarkenable = true
)

//go:nowritebarrier
func markroot(desc *_sched.Parfor, i uint32) {
	// Note: if you add a case here, please also update heapdump.c:dumproots.
	switch i {
	case _sched.RootData:
		_sched.Scanblock(uintptr(unsafe.Pointer(&Data)), uintptr(unsafe.Pointer(&Edata))-uintptr(unsafe.Pointer(&Data)), Gcdatamask.Bytedata)

	case _sched.RootBss:
		_sched.Scanblock(uintptr(unsafe.Pointer(&Bss)), uintptr(unsafe.Pointer(&Ebss))-uintptr(unsafe.Pointer(&Bss)), Gcbssmask.Bytedata)

	case _sched.RootFinalizers:
		for fb := Allfin; fb != nil; fb = fb.Alllink {
			_sched.Scanblock(uintptr(unsafe.Pointer(&fb.Fin[0])), uintptr(fb.Cnt)*unsafe.Sizeof(fb.Fin[0]), &finptrmask[0])
		}

	case _sched.RootSpans:
		// mark MSpan.specials
		sg := _lock.Mheap_.Sweepgen
		for spanidx := uint32(0); spanidx < uint32(len(_sched.Work.Spans)); spanidx++ {
			s := _sched.Work.Spans[spanidx]
			if s.State != _sched.MSpanInUse {
				continue
			}
			if !_sched.Checkmark && s.Sweepgen != sg {
				// sweepgen was updated (+2) during non-checkmark GC pass
				print("sweep ", s.Sweepgen, " ", sg, "\n")
				_lock.Throw("gc: unswept span")
			}
			for sp := s.Specials; sp != nil; sp = sp.Next {
				if sp.Kind != KindSpecialFinalizer {
					continue
				}
				// don't mark finalized object, but scan it so we
				// retain everything it points to.
				spf := (*Specialfinalizer)(unsafe.Pointer(sp))
				// A finalizer can be set for an inner byte of an object, find object beginning.
				p := uintptr(s.Start<<_core.PageShift) + uintptr(spf.Special.Offset)/s.Elemsize*s.Elemsize
				if _sched.Gcphase != _sched.GCscan {
					_sched.Scanblock(p, s.Elemsize, nil) // scanned during mark phase
				}
				_sched.Scanblock(uintptr(unsafe.Pointer(&spf.Fn)), _core.PtrSize, &oneptr[0])
			}
		}

	case _sched.RootFlushCaches:
		if _sched.Gcphase != _sched.GCscan { // Do not flush mcaches during GCscan phase.
			flushallmcaches()
		}

	default:
		// the rest is scanning goroutine stacks
		if uintptr(i-_sched.RootCount) >= Allglen {
			_lock.Throw("markroot: bad index")
		}
		gp := _lock.Allgs[i-_sched.RootCount]

		// remember when we've first observed the G blocked
		// needed only to output in traceback
		status := _lock.Readgstatus(gp) // We are not in a scan state
		if (status == _lock.Gwaiting || status == _lock.Gsyscall) && gp.Waitsince == 0 {
			gp.Waitsince = _sched.Work.Tstart
		}

		// Shrink a stack if not much of it is being used but not in the scan phase.
		if _sched.Gcphase != _sched.GCscan { // Do not shrink during GCscan phase.
			shrinkstack(gp)
		}
		if _lock.Readgstatus(gp) == _lock.Gdead {
			gp.Gcworkdone = true
		} else {
			gp.Gcworkdone = false
		}
		restart := Stopg(gp)

		// goroutine will scan its own stack when it stops running.
		// Wait until it has.
		for _lock.Readgstatus(gp) == _lock.Grunning && !gp.Gcworkdone {
		}

		// scanstack(gp) is done as part of gcphasework
		// But to make sure we finished we need to make sure that
		// the stack traps have all responded so drop into
		// this while loop until they respond.
		for !gp.Gcworkdone {
			status = _lock.Readgstatus(gp)
			if status == _lock.Gdead {
				gp.Gcworkdone = true // scan is a noop
				break
			}
			if status == _lock.Gwaiting || status == _lock.Grunnable {
				restart = Stopg(gp)
			}
		}
		if restart {
			Restartg(gp)
		}
	}
}

//go:nowritebarrier
func Stackmapdata(stkmap *Stackmap, n int32) _lock.Bitvector {
	if n < 0 || n >= stkmap.N {
		_lock.Throw("stackmapdata: index out of range")
	}
	return _lock.Bitvector{stkmap.nbit, (*byte)(_core.Add(unsafe.Pointer(&stkmap.bytedata), uintptr(n*((stkmap.nbit+31)/32*4))))}
}

// Scan a stack frame: local variables and function arguments/results.
//go:nowritebarrier
func scanframe(frame *_lock.Stkframe, unused unsafe.Pointer) bool {

	f := frame.Fn
	targetpc := frame.Continpc
	if targetpc == 0 {
		// Frame is dead.
		return true
	}
	if _sched.DebugGC > 1 {
		print("scanframe ", _lock.Gofuncname(f), "\n")
	}
	if targetpc != f.Entry {
		targetpc--
	}
	pcdata := Pcdatavalue(f, _lock.PCDATA_StackMapIndex, targetpc)
	if pcdata == -1 {
		// We do not have a valid pcdata value but there might be a
		// stackmap for this function.  It is likely that we are looking
		// at the function prologue, assume so and hope for the best.
		pcdata = 0
	}

	// Scan local variables if stack frame has been allocated.
	size := frame.Varp - frame.Sp
	var minsize uintptr
	if _lock.Thechar != '6' && _lock.Thechar != '8' {
		minsize = _core.PtrSize
	} else {
		minsize = 0
	}
	if size > minsize {
		stkmap := (*Stackmap)(Funcdata(f, _lock.FUNCDATA_LocalsPointerMaps))
		if stkmap == nil || stkmap.N <= 0 {
			print("runtime: frame ", _lock.Gofuncname(f), " untyped locals ", _core.Hex(frame.Varp-size), "+", _core.Hex(size), "\n")
			_lock.Throw("missing stackmap")
		}

		// Locals bitmap information, scan just the pointers in locals.
		if pcdata < 0 || pcdata >= stkmap.N {
			// don't know where we are
			print("runtime: pcdata is ", pcdata, " and ", stkmap.N, " locals stack map entries for ", _lock.Gofuncname(f), " (targetpc=", targetpc, ")\n")
			_lock.Throw("scanframe: bad symbol table")
		}
		bv := Stackmapdata(stkmap, pcdata)
		size = (uintptr(bv.N) * _core.PtrSize) / _sched.XBitsPerPointer
		_sched.Scanblock(frame.Varp-size, uintptr(bv.N)/_sched.XBitsPerPointer*_core.PtrSize, bv.Bytedata)
	}

	// Scan arguments.
	if frame.Arglen > 0 {
		var bv _lock.Bitvector
		if frame.Argmap != nil {
			bv = *frame.Argmap
		} else {
			stkmap := (*Stackmap)(Funcdata(f, _lock.FUNCDATA_ArgsPointerMaps))
			if stkmap == nil || stkmap.N <= 0 {
				print("runtime: frame ", _lock.Gofuncname(f), " untyped args ", _core.Hex(frame.Argp), "+", _core.Hex(frame.Arglen), "\n")
				_lock.Throw("missing stackmap")
			}
			if pcdata < 0 || pcdata >= stkmap.N {
				// don't know where we are
				print("runtime: pcdata is ", pcdata, " and ", stkmap.N, " args stack map entries for ", _lock.Gofuncname(f), " (targetpc=", targetpc, ")\n")
				_lock.Throw("scanframe: bad symbol table")
			}
			bv = Stackmapdata(stkmap, pcdata)
		}
		_sched.Scanblock(frame.Argp, uintptr(bv.N)/_sched.XBitsPerPointer*_core.PtrSize, bv.Bytedata)
	}
	return true
}

//go:nowritebarrier
func scanstack(gp *_core.G) {

	if _lock.Readgstatus(gp)&_lock.Gscan == 0 {
		print("runtime:scanstack: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _core.Hex(_lock.Readgstatus(gp)), "\n")
		_lock.Throw("scanstack - bad status")
	}

	switch _lock.Readgstatus(gp) &^ _lock.Gscan {
	default:
		print("runtime: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _lock.Readgstatus(gp), "\n")
		_lock.Throw("mark - bad status")
	case _lock.Gdead:
		return
	case _lock.Grunning:
		print("runtime: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _lock.Readgstatus(gp), "\n")
		_lock.Throw("scanstack: goroutine not stopped")
	case _lock.Grunnable, _lock.Gsyscall, _lock.Gwaiting:
		// ok
	}

	if gp == _core.Getg() {
		_lock.Throw("can't scan our own stack")
	}
	mp := gp.M
	if mp != nil && mp.Helpgc != 0 {
		_lock.Throw("can't scan gchelper stack")
	}

	_lock.Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, 0x7fffffff, scanframe, nil, 0)
	tracebackdefers(gp, scanframe, nil)
}

// The gp has been moved to a GC safepoint. GC phase specific
// work is done here.
//go:nowritebarrier
func Gcphasework(gp *_core.G) {
	switch _sched.Gcphase {
	default:
		_lock.Throw("gcphasework in bad gcphase")
	case _sched.GCoff, _sched.GCquiesce, _sched.GCstw, _sched.GCsweep:
		// No work.
	case _sched.GCscan:
		// scan the stack, mark the objects, put pointers in work buffers
		// hanging off the P where this is being run.
		scanstack(gp)
	case _sched.GCmark:
		// No work.
	case _sched.GCmarktermination:
		scanstack(gp)
		// All available mark work will be emptied before returning.
	}
	gp.Gcworkdone = true
}

var finalizer1 = [...]byte{
	// Each Finalizer is 5 words, ptr ptr uintptr ptr ptr.
	// Each byte describes 4 words.
	// Need 4 Finalizers described by 5 bytes before pattern repeats:
	//	ptr ptr uintptr ptr ptr
	//	ptr ptr uintptr ptr ptr
	//	ptr ptr uintptr ptr ptr
	//	ptr ptr uintptr ptr ptr
	// aka
	//	ptr ptr uintptr ptr
	//	ptr ptr ptr uintptr
	//	ptr ptr ptr ptr
	//	uintptr ptr ptr ptr
	//	ptr uintptr ptr ptr
	// Assumptions about Finalizer layout checked below.
	_sched.XBitsPointer | _sched.XBitsPointer<<2 | _sched.XBitsScalar<<4 | _sched.XBitsPointer<<6,
	_sched.XBitsPointer | _sched.XBitsPointer<<2 | _sched.XBitsPointer<<4 | _sched.XBitsScalar<<6,
	_sched.XBitsPointer | _sched.XBitsPointer<<2 | _sched.XBitsPointer<<4 | _sched.XBitsPointer<<6,
	_sched.XBitsScalar | _sched.XBitsPointer<<2 | _sched.XBitsPointer<<4 | _sched.XBitsPointer<<6,
	_sched.XBitsPointer | _sched.XBitsScalar<<2 | _sched.XBitsPointer<<4 | _sched.XBitsPointer<<6,
}

func queuefinalizer(p unsafe.Pointer, fn *_core.Funcval, nret uintptr, fint *_core.Type, ot *Ptrtype) {
	_lock.Lock(&_sched.Finlock)
	if Finq == nil || Finq.Cnt == int32(len(Finq.Fin)) {
		if Finc == nil {
			// Note: write barrier here, assigning to finc, but should be okay.
			Finc = (*Finblock)(_lock.Persistentalloc(_sched.FinBlockSize, 0, &_lock.Memstats.Gc_sys))
			Finc.Alllink = Allfin
			Allfin = Finc
			if finptrmask[0] == 0 {
				// Build pointer mask for Finalizer array in block.
				// Check assumptions made in finalizer1 array above.
				if (unsafe.Sizeof(Finalizer{}) != 5*_core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Fn) != 0 ||
					unsafe.Offsetof(Finalizer{}.Arg) != _core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Nret) != 2*_core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Fint) != 3*_core.PtrSize ||
					unsafe.Offsetof(Finalizer{}.Ot) != 4*_core.PtrSize ||
					_sched.XBitsPerPointer != 2) {
					_lock.Throw("finalizer out of sync")
				}
				for i := range finptrmask {
					finptrmask[i] = finalizer1[i%len(finalizer1)]
				}
			}
		}
		block := Finc
		Finc = block.Next
		block.Next = Finq
		Finq = block
	}
	f := &Finq.Fin[Finq.Cnt]
	Finq.Cnt++
	f.Fn = fn
	f.Nret = nret
	f.Fint = fint
	f.Ot = ot
	f.Arg = p
	_sched.Fingwake = true
	_lock.Unlock(&_sched.Finlock)
}

// Returns only when span s has been swept.
//go:nowritebarrier
func MSpan_EnsureSwept(s *_core.Mspan) {
	// Caller must disable preemption.
	// Otherwise when this function returns the span can become unswept again
	// (if GC is triggered on another goroutine).
	_g_ := _core.Getg()
	if _g_.M.Locks == 0 && _g_.M.Mallocing == 0 && _g_ != _g_.M.G0 {
		_lock.Throw("MSpan_EnsureSwept: m is not locked")
	}

	sg := _lock.Mheap_.Sweepgen
	if _lock.Atomicload(&s.Sweepgen) == sg {
		return
	}
	// The caller must be sure that the span is a MSpanInUse span.
	if _sched.Cas(&s.Sweepgen, sg-2, sg-1) {
		MSpan_Sweep(s, false)
		return
	}
	// unfortunate condition, and we don't have efficient means to wait
	for _lock.Atomicload(&s.Sweepgen) != sg {
		_core.Osyield()
	}
}

// Sweep frees or collects finalizers for blocks not marked in the mark phase.
// It clears the mark bits in preparation for the next GC round.
// Returns true if the span was returned to heap.
// If preserve=true, don't return it to heap nor relink in MCentral lists;
// caller takes care of it.
//TODO go:nowritebarrier
func MSpan_Sweep(s *_core.Mspan, preserve bool) bool {
	if _sched.Checkmark {
		_lock.Throw("MSpan_Sweep: checkmark only runs in STW and after the sweep")
	}

	// It's critical that we enter this function with preemption disabled,
	// GC must not start while we are in the middle of this function.
	_g_ := _core.Getg()
	if _g_.M.Locks == 0 && _g_.M.Mallocing == 0 && _g_ != _g_.M.G0 {
		_lock.Throw("MSpan_Sweep: m is not locked")
	}
	sweepgen := _lock.Mheap_.Sweepgen
	if s.State != _sched.MSpanInUse || s.Sweepgen != sweepgen-1 {
		print("MSpan_Sweep: state=", s.State, " sweepgen=", s.Sweepgen, " mheap.sweepgen=", sweepgen, "\n")
		_lock.Throw("MSpan_Sweep: bad span state")
	}
	arena_start := _lock.Mheap_.Arena_start
	cl := s.Sizeclass
	size := s.Elemsize
	var n int32
	var npages int32
	if cl == 0 {
		n = 1
	} else {
		// Chunk full of small blocks.
		npages = Class_to_allocnpages[cl]
		n = (npages << _core.PageShift) / int32(size)
	}
	res := false
	nfree := 0

	var head, end _core.Gclinkptr

	c := _g_.M.Mcache
	sweepgenset := false

	// Mark any free objects in this span so we don't collect them.
	for link := s.Freelist; link.Ptr() != nil; link = link.Ptr().Next {
		off := (uintptr(unsafe.Pointer(link)) - arena_start) / _core.PtrSize
		bitp := arena_start - off/_sched.WordsPerBitmapByte - 1
		shift := (off % _sched.WordsPerBitmapByte) * _sched.GcBits
		*(*byte)(unsafe.Pointer(bitp)) |= _sched.BitMarked << shift
	}

	// Unlink & free special records for any objects we're about to free.
	specialp := &s.Specials
	special := *specialp
	for special != nil {
		// A finalizer can be set for an inner byte of an object, find object beginning.
		p := uintptr(s.Start<<_core.PageShift) + uintptr(special.Offset)/size*size
		off := (p - arena_start) / _core.PtrSize
		bitp := arena_start - off/_sched.WordsPerBitmapByte - 1
		shift := (off % _sched.WordsPerBitmapByte) * _sched.GcBits
		bits := (*(*byte)(unsafe.Pointer(bitp)) >> shift) & _sched.BitMask
		if bits&_sched.BitMarked == 0 {
			// Find the exact byte for which the special was setup
			// (as opposed to object beginning).
			p := uintptr(s.Start<<_core.PageShift) + uintptr(special.Offset)
			// about to free object: splice out special record
			y := special
			special = special.Next
			*specialp = special
			if !freespecial(y, unsafe.Pointer(p), size, false) {
				// stop freeing of object if it has a finalizer
				*(*byte)(unsafe.Pointer(bitp)) |= _sched.BitMarked << shift
			}
		} else {
			// object is still live: keep special record
			specialp = &special.Next
			special = *specialp
		}
	}

	// Sweep through n objects of given size starting at p.
	// This thread owns the span now, so it can manipulate
	// the block bitmap without atomic operations.
	p := uintptr(s.Start << _core.PageShift)
	off := (p - arena_start) / _core.PtrSize
	bitp := arena_start - off/_sched.WordsPerBitmapByte - 1
	shift := uint(0)
	step := size / (_core.PtrSize * _sched.WordsPerBitmapByte)
	// Rewind to the previous quadruple as we move to the next
	// in the beginning of the loop.
	bitp += step
	if step == 0 {
		// 8-byte objects.
		bitp++
		shift = _sched.GcBits
	}
	for ; n > 0; n, p = n-1, p+size {
		bitp -= step
		if step == 0 {
			if shift != 0 {
				bitp--
			}
			shift = _sched.GcBits - shift
		}

		xbits := *(*byte)(unsafe.Pointer(bitp))
		bits := (xbits >> shift) & _sched.BitMask

		// Allocated and marked object, reset bits to allocated.
		if bits&_sched.BitMarked != 0 {
			*(*byte)(unsafe.Pointer(bitp)) &^= _sched.BitMarked << shift
			continue
		}

		// At this point we know that we are looking at garbage object
		// that needs to be collected.
		if _lock.Debug.Allocfreetrace != 0 {
			tracefree(unsafe.Pointer(p), size)
		}

		// Reset to allocated+noscan.
		*(*byte)(unsafe.Pointer(bitp)) = uint8(uintptr(xbits&^((_sched.BitMarked|_sched.XBitsMask<<2)<<shift)) | uintptr(_sched.XBitsDead)<<(shift+2))
		if cl == 0 {
			// Free large span.
			if preserve {
				_lock.Throw("can't preserve large span")
			}
			unmarkspan(p, s.Npages<<_core.PageShift)
			s.Needzero = 1

			// important to set sweepgen before returning it to heap
			_lock.Atomicstore(&s.Sweepgen, sweepgen)
			sweepgenset = true

			// NOTE(rsc,dvyukov): The original implementation of efence
			// in CL 22060046 used SysFree instead of SysFault, so that
			// the operating system would eventually give the memory
			// back to us again, so that an efence program could run
			// longer without running out of memory. Unfortunately,
			// calling SysFree here without any kind of adjustment of the
			// heap data structures means that when the memory does
			// come back to us, we have the wrong metadata for it, either in
			// the MSpan structures or in the garbage collection bitmap.
			// Using SysFault here means that the program will run out of
			// memory fairly quickly in efence mode, but at least it won't
			// have mysterious crashes due to confused memory reuse.
			// It should be possible to switch back to SysFree if we also
			// implement and then call some kind of MHeap_DeleteSpan.
			if _lock.Debug.Efence > 0 {
				s.Limit = 0 // prevent mlookup from finding this span
				sysFault(unsafe.Pointer(p), size)
			} else {
				mHeap_Free(&_lock.Mheap_, s, 1)
			}
			c.Local_nlargefree++
			c.Local_largefree += size
			_lock.Xadd64(&_lock.Memstats.Next_gc, -int64(size)*int64(Gcpercent+100)/100)
			res = true
		} else {
			// Free small object.
			if size > 2*_core.PtrSize {
				*(*uintptr)(unsafe.Pointer(p + _core.PtrSize)) = _lock.UintptrMask & 0xdeaddeaddeaddead // mark as "needs to be zeroed"
			} else if size > _core.PtrSize {
				*(*uintptr)(unsafe.Pointer(p + _core.PtrSize)) = 0
			}
			if head.Ptr() == nil {
				head = _core.Gclinkptr(p)
			} else {
				end.Ptr().Next = _core.Gclinkptr(p)
			}
			end = _core.Gclinkptr(p)
			end.Ptr().Next = _core.Gclinkptr(0x0bade5)
			nfree++
		}
	}

	// We need to set s.sweepgen = h.sweepgen only when all blocks are swept,
	// because of the potential for a concurrent free/SetFinalizer.
	// But we need to set it before we make the span available for allocation
	// (return it to heap or mcentral), because allocation code assumes that a
	// span is already swept if available for allocation.
	if !sweepgenset && nfree == 0 {
		// The span must be in our exclusive ownership until we update sweepgen,
		// check for potential races.
		if s.State != _sched.MSpanInUse || s.Sweepgen != sweepgen-1 {
			print("MSpan_Sweep: state=", s.State, " sweepgen=", s.Sweepgen, " mheap.sweepgen=", sweepgen, "\n")
			_lock.Throw("MSpan_Sweep: bad span state after sweep")
		}
		_lock.Atomicstore(&s.Sweepgen, sweepgen)
	}
	if nfree > 0 {
		c.Local_nsmallfree[cl] += uintptr(nfree)
		c.Local_cachealloc -= _core.Intptr(uintptr(nfree) * size)
		_lock.Xadd64(&_lock.Memstats.Next_gc, -int64(nfree)*int64(size)*int64(Gcpercent+100)/100)
		res = mCentral_FreeSpan(&_lock.Mheap_.Central[cl].Mcentral, s, int32(nfree), head, end, preserve)
		// MCentral_FreeSpan updates sweepgen
	}
	return res
}

// State of background sweep.
// Protected by gclock.
type sweepdata struct {
	g       *_core.G
	parked  bool
	started bool

	spanidx uint32 // background sweeper position

	nbgsweep    uint32
	npausesweep uint32
}

var sweep sweepdata

// sweeps one span
// returns number of pages returned to heap, or ^uintptr(0) if there is nothing to sweep
//go:nowritebarrier
func Sweepone() uintptr {
	_g_ := _core.Getg()

	// increment locks to ensure that the goroutine is not preempted
	// in the middle of sweep thus leaving the span in an inconsistent state for next GC
	_g_.M.Locks++
	sg := _lock.Mheap_.Sweepgen
	for {
		idx := _lock.Xadd(&sweep.spanidx, 1) - 1
		if idx >= uint32(len(_sched.Work.Spans)) {
			_lock.Mheap_.Sweepdone = 1
			_g_.M.Locks--
			return ^uintptr(0)
		}
		s := _sched.Work.Spans[idx]
		if s.State != _sched.MSpanInUse {
			s.Sweepgen = sg
			continue
		}
		if s.Sweepgen != sg-2 || !_sched.Cas(&s.Sweepgen, sg-2, sg-1) {
			continue
		}
		npages := s.Npages
		if !MSpan_Sweep(s, false) {
			npages = 0
		}
		_g_.M.Locks--
		return npages
	}
}

//go:nowritebarrier
func gosweepone() uintptr {
	var ret uintptr
	_lock.Systemstack(func() {
		ret = Sweepone()
	})
	return ret
}

//go:nowritebarrier
func gosweepdone() bool {
	return _lock.Mheap_.Sweepdone != 0
}

//go:nowritebarrier
func cachestats() {
	for i := 0; ; i++ {
		p := _lock.Allp[i]
		if p == nil {
			break
		}
		c := p.Mcache
		if c == nil {
			continue
		}
		Purgecachedstats(c)
	}
}

//go:nowritebarrier
func flushallmcaches() {
	for i := 0; ; i++ {
		p := _lock.Allp[i]
		if p == nil {
			break
		}
		c := p.Mcache
		if c == nil {
			continue
		}
		mCache_ReleaseAll(c)
		stackcache_clear(c)
	}
}

//go:nowritebarrier
func Updatememstats(stats *_core.Gcstats) {
	if stats != nil {
		*stats = _core.Gcstats{}
	}
	for mp := _lock.Allm; mp != nil; mp = mp.Alllink {
		if stats != nil {
			src := (*[unsafe.Sizeof(_core.Gcstats{}) / 8]uint64)(unsafe.Pointer(&mp.Gcstats))
			dst := (*[unsafe.Sizeof(_core.Gcstats{}) / 8]uint64)(unsafe.Pointer(stats))
			for i, v := range src {
				dst[i] += v
			}
			mp.Gcstats = _core.Gcstats{}
		}
	}

	_lock.Memstats.Mcache_inuse = uint64(_lock.Mheap_.Cachealloc.Inuse)
	_lock.Memstats.Mspan_inuse = uint64(_lock.Mheap_.Spanalloc.Inuse)
	_lock.Memstats.Sys = _lock.Memstats.Heap_sys + _lock.Memstats.Stacks_sys + _lock.Memstats.Mspan_sys +
		_lock.Memstats.Mcache_sys + _lock.Memstats.Buckhash_sys + _lock.Memstats.Gc_sys + _lock.Memstats.Other_sys

	// Calculate memory allocator stats.
	// During program execution we only count number of frees and amount of freed memory.
	// Current number of alive object in the heap and amount of alive heap memory
	// are calculated by scanning all spans.
	// Total number of mallocs is calculated as number of frees plus number of alive objects.
	// Similarly, total amount of allocated memory is calculated as amount of freed memory
	// plus amount of alive heap memory.
	_lock.Memstats.Alloc = 0
	_lock.Memstats.Total_alloc = 0
	_lock.Memstats.Nmalloc = 0
	_lock.Memstats.Nfree = 0
	for i := 0; i < len(_lock.Memstats.By_size); i++ {
		_lock.Memstats.By_size[i].Nmalloc = 0
		_lock.Memstats.By_size[i].Nfree = 0
	}

	// Flush MCache's to MCentral.
	_lock.Systemstack(flushallmcaches)

	// Aggregate local stats.
	cachestats()

	// Scan all spans and count number of alive objects.
	_lock.Lock(&_lock.Mheap_.Lock)
	for i := uint32(0); i < _lock.Mheap_.Nspan; i++ {
		s := H_allspans[i]
		if s.State != _sched.MSpanInUse {
			continue
		}
		if s.Sizeclass == 0 {
			_lock.Memstats.Nmalloc++
			_lock.Memstats.Alloc += uint64(s.Elemsize)
		} else {
			_lock.Memstats.Nmalloc += uint64(s.Ref)
			_lock.Memstats.By_size[s.Sizeclass].Nmalloc += uint64(s.Ref)
			_lock.Memstats.Alloc += uint64(s.Ref) * uint64(s.Elemsize)
		}
	}
	_lock.Unlock(&_lock.Mheap_.Lock)

	// Aggregate by size class.
	smallfree := uint64(0)
	_lock.Memstats.Nfree = _lock.Mheap_.Nlargefree
	for i := 0; i < len(_lock.Memstats.By_size); i++ {
		_lock.Memstats.Nfree += _lock.Mheap_.Nsmallfree[i]
		_lock.Memstats.By_size[i].Nfree = _lock.Mheap_.Nsmallfree[i]
		_lock.Memstats.By_size[i].Nmalloc += _lock.Mheap_.Nsmallfree[i]
		smallfree += uint64(_lock.Mheap_.Nsmallfree[i]) * uint64(Class_to_size[i])
	}
	_lock.Memstats.Nfree += _lock.Memstats.Tinyallocs
	_lock.Memstats.Nmalloc += _lock.Memstats.Nfree

	// Calculate derived stats.
	_lock.Memstats.Total_alloc = uint64(_lock.Memstats.Alloc) + uint64(_lock.Mheap_.Largefree) + smallfree
	_lock.Memstats.Heap_alloc = _lock.Memstats.Alloc
	_lock.Memstats.Heap_objects = _lock.Memstats.Nmalloc - _lock.Memstats.Nfree
}

// Called from malloc.go using onM, stopping and starting the world handled in caller.
//go:nowritebarrier
func gc_m(start_time int64, eagersweep bool) {
	_g_ := _core.Getg()
	gp := _g_.M.Curg
	_sched.Casgstatus(gp, _lock.Grunning, _lock.Gwaiting)
	gp.Waitreason = "garbage collection"

	gc(start_time, eagersweep)
	_sched.Casgstatus(gp, _lock.Gwaiting, _lock.Grunning)
}

// Similar to clearcheckmarkbits but works on a single span.
// It preforms two tasks.
// 1. When used before the checkmark phase it converts BitsDead (00) to bitsScalar (01)
//    for nibbles with the BoundaryBit set.
// 2. When used after the checkmark phase it converts BitsPointerMark (11) to BitsPointer 10 and
//    BitsScalarMark (00) to BitsScalar (01), thus clearing the checkmark mark encoding.
// For the second case it is possible to restore the BitsDead pattern but since
// clearmark is a debug tool performance has a lower priority than simplicity.
// The span is MSpanInUse and the world is stopped.
//go:nowritebarrier
func clearcheckmarkbitsspan(s *_core.Mspan) {
	if s.State != _sched.XMSpanInUse {
		print("runtime:clearcheckmarkbitsspan: state=", s.State, "\n")
		_lock.Throw("clearcheckmarkbitsspan: bad span state")
	}

	arena_start := _lock.Mheap_.Arena_start
	cl := s.Sizeclass
	size := s.Elemsize
	var n int32
	if cl == 0 {
		n = 1
	} else {
		// Chunk full of small blocks
		npages := Class_to_allocnpages[cl]
		n = npages << _core.PageShift / int32(size)
	}

	// MSpan_Sweep has similar code but instead of overloading and
	// complicating that routine we do a simpler walk here.
	// Sweep through n objects of given size starting at p.
	// This thread owns the span now, so it can manipulate
	// the block bitmap without atomic operations.
	p := uintptr(s.Start) << _core.PageShift

	// Find bits for the beginning of the span.
	off := (p - arena_start) / _core.PtrSize
	bitp := (*byte)(unsafe.Pointer(arena_start - off/_sched.WordsPerBitmapByte - 1))
	step := size / (_core.PtrSize * _sched.WordsPerBitmapByte)

	// The type bit values are:
	//	00 - BitsDead, for us BitsScalarMarked
	//	01 - BitsScalar
	//	10 - BitsPointer
	//	11 - unused, for us BitsPointerMarked
	//
	// When called to prepare for the checkmark phase (checkmark==1),
	// we change BitsDead to BitsScalar, so that there are no BitsScalarMarked
	// type bits anywhere.
	//
	// The checkmark phase marks by changing BitsScalar to BitsScalarMarked
	// and BitsPointer to BitsPointerMarked.
	//
	// When called to clean up after the checkmark phase (checkmark==0),
	// we unmark by changing BitsScalarMarked back to BitsScalar and
	// BitsPointerMarked back to BitsPointer.
	//
	// There are two problems with the scheme as just described.
	// First, the setup rewrites BitsDead to BitsScalar, but the type bits
	// following a BitsDead are uninitialized and must not be used.
	// Second, objects that are free are expected to have their type
	// bits zeroed (BitsDead), so in the cleanup we need to restore
	// any BitsDeads that were there originally.
	//
	// In a one-word object (8-byte allocation on 64-bit system),
	// there is no difference between BitsScalar and BitsDead, because
	// neither is a pointer and there are no more words in the object,
	// so using BitsScalar during the checkmark is safe and mapping
	// both back to BitsDead during cleanup is also safe.
	//
	// In a larger object, we need to be more careful. During setup,
	// if the type of the first word is BitsDead, we change it to BitsScalar
	// (as we must) but also initialize the type of the second
	// word to BitsDead, so that a scan during the checkmark phase
	// will still stop before seeing the uninitialized type bits in the
	// rest of the object. The sequence 'BitsScalar BitsDead' never
	// happens in real type bitmaps - BitsDead is always as early
	// as possible, so immediately after the last BitsPointer.
	// During cleanup, if we see a BitsScalar, we can check to see if it
	// is followed by BitsDead. If so, it was originally BitsDead and
	// we can change it back.

	if step == 0 {
		// updating top and bottom nibbles, all boundaries
		for i := int32(0); i < n/2; i, bitp = i+1, Addb(bitp, _lock.UintptrMask&-1) {
			if *bitp&_sched.BitBoundary == 0 {
				_lock.Throw("missing bitBoundary")
			}
			b := (*bitp & _sched.BitPtrMask) >> 2
			if !_sched.Checkmark && (b == _sched.BitsScalar || b == _sched.BitsScalarMarked) {
				*bitp &^= 0x0c // convert to _BitsDead
			} else if b == _sched.BitsScalarMarked || b == _sched.BitsPointerMarked {
				*bitp &^= _sched.BitsCheckMarkXor << 2
			}

			if (*bitp>>_sched.GcBits)&_sched.BitBoundary == 0 {
				_lock.Throw("missing bitBoundary")
			}
			b = ((*bitp >> _sched.GcBits) & _sched.BitPtrMask) >> 2
			if !_sched.Checkmark && (b == _sched.BitsScalar || b == _sched.BitsScalarMarked) {
				*bitp &^= 0xc0 // convert to _BitsDead
			} else if b == _sched.BitsScalarMarked || b == _sched.BitsPointerMarked {
				*bitp &^= _sched.BitsCheckMarkXor << (2 + _sched.GcBits)
			}
		}
	} else {
		// updating bottom nibble for first word of each object
		for i := int32(0); i < n; i, bitp = i+1, Addb(bitp, -step) {
			if *bitp&_sched.BitBoundary == 0 {
				_lock.Throw("missing bitBoundary")
			}
			b := (*bitp & _sched.BitPtrMask) >> 2

			if _sched.Checkmark && b == _sched.BitsDead {
				// move BitsDead into second word.
				// set bits to BitsScalar in preparation for checkmark phase.
				*bitp &^= 0xc0
				*bitp |= _sched.BitsScalar << 2
			} else if !_sched.Checkmark && (b == _sched.BitsScalar || b == _sched.BitsScalarMarked) && *bitp&0xc0 == 0 {
				// Cleaning up after checkmark phase.
				// First word is scalar or dead (we forgot)
				// and second word is dead.
				// First word might as well be dead too.
				*bitp &^= 0x0c
			} else if b == _sched.BitsScalarMarked || b == _sched.BitsPointerMarked {
				*bitp ^= _sched.BitsCheckMarkXor << 2
			}
		}
	}
}

// clearcheckmarkbits preforms two tasks.
// 1. When used before the checkmark phase it converts BitsDead (00) to bitsScalar (01)
//    for nibbles with the BoundaryBit set.
// 2. When used after the checkmark phase it converts BitsPointerMark (11) to BitsPointer 10 and
//    BitsScalarMark (00) to BitsScalar (01), thus clearing the checkmark mark encoding.
// This is a bit expensive but preserves the BitsDead encoding during the normal marking.
// BitsDead remains valid for every nibble except the ones with BitsBoundary set.
//go:nowritebarrier
func clearcheckmarkbits() {
	for _, s := range _sched.Work.Spans {
		if s.State == _sched.XMSpanInUse {
			clearcheckmarkbitsspan(s)
		}
	}
}

// Called from malloc.go using onM.
// The world is stopped. Rerun the scan and mark phases
// using the bitMarkedCheck bit instead of the
// bitMarked bit. If the marking encounters an
// bitMarked bit that is not set then we throw.
//go:nowritebarrier
func gccheckmark_m(startTime int64, eagersweep bool) {
	if !Gccheckmarkenable {
		return
	}

	if _sched.Checkmark {
		_lock.Throw("gccheckmark_m, entered with checkmark already true")
	}

	_sched.Checkmark = true
	clearcheckmarkbits()        // Converts BitsDead to BitsScalar.
	gc_m(startTime, eagersweep) // turns off checkmark
	// Work done, fixed up the GC bitmap to remove the checkmark bits.
	clearcheckmarkbits()
}

//go:nowritebarrier
func finishsweep_m() {
	// The world is stopped so we should be able to complete the sweeps
	// quickly.
	for Sweepone() != ^uintptr(0) {
		sweep.npausesweep++
	}

	// There may be some other spans being swept concurrently that
	// we need to wait for. If finishsweep_m is done with the world stopped
	// this code is not required.
	sg := _lock.Mheap_.Sweepgen
	for _, s := range _sched.Work.Spans {
		if s.Sweepgen != sg && s.State == _sched.XMSpanInUse {
			MSpan_EnsureSwept(s)
		}
	}
}

// Scan all of the stacks, greying (or graying if in America) the referents
// but not blackening them since the mark write barrier isn't installed.
//go:nowritebarrier
func gcscan_m() {
	_g_ := _core.Getg()

	// Grab the g that called us and potentially allow rescheduling.
	// This allows it to be scanned like other goroutines.
	mastergp := _g_.M.Curg
	_sched.Casgstatus(mastergp, _lock.Grunning, _lock.Gwaiting)
	mastergp.Waitreason = "garbage collection scan"

	// Span sweeping has been done by finishsweep_m.
	// Long term we will want to make this goroutine runnable
	// by placing it onto a scanenqueue state and then calling
	// runtime·restartg(mastergp) to make it Grunnable.
	// At the bottom we will want to return this p back to the scheduler.
	oldphase := _sched.Gcphase

	// Prepare flag indicating that the scan has not been completed.
	_lock.Lock(&_lock.Allglock)
	local_allglen := Allglen
	for i := uintptr(0); i < local_allglen; i++ {
		gp := _lock.Allgs[i]
		gp.Gcworkdone = false // set to true in gcphasework
	}
	_lock.Unlock(&_lock.Allglock)

	_sched.Work.Nwait = 0
	_sched.Work.Ndone = 0
	_sched.Work.Nproc = 1 // For now do not do this in parallel.
	_sched.Gcphase = _sched.GCscan
	//	ackgcphase is not needed since we are not scanning running goroutines.
	parforsetup(_sched.Work.Markfor, _sched.Work.Nproc, uint32(_sched.RootCount+local_allglen), nil, false, markroot)
	_sched.Parfordo(_sched.Work.Markfor)

	_lock.Lock(&_lock.Allglock)
	// Check that gc work is done.
	for i := uintptr(0); i < local_allglen; i++ {
		gp := _lock.Allgs[i]
		if !gp.Gcworkdone {
			_lock.Throw("scan missed a g")
		}
	}
	_lock.Unlock(&_lock.Allglock)

	_sched.Gcphase = oldphase
	_sched.Casgstatus(mastergp, _lock.Gwaiting, _lock.Grunning)
	// Let the g that called us continue to run.
}

// Mark all objects that are known about.
//go:nowritebarrier
func gcmark_m() {
	_sched.Scanblock(0, 0, nil)
}

// For now this must be bracketed with a stoptheworld and a starttheworld to ensure
// all go routines see the new barrier.
//go:nowritebarrier
func gcinstalloffwb_m() {
	_sched.Gcphase = _sched.GCoff
}

//TODO go:nowritebarrier
func gc(start_time int64, eagersweep bool) {
	if _sched.DebugGCPtrs {
		print("GC start\n")
	}

	if _lock.Debug.Allocfreetrace > 0 {
		tracegc()
	}

	_g_ := _core.Getg()
	_g_.M.Traceback = 2
	t0 := start_time
	_sched.Work.Tstart = start_time

	var t1 int64
	if _lock.Debug.Gctrace > 0 {
		t1 = _lock.Nanotime()
	}

	if !_sched.Checkmark {
		finishsweep_m() // skip during checkmark debug phase.
	}

	// Cache runtime.mheap_.allspans in work.spans to avoid conflicts with
	// resizing/freeing allspans.
	// New spans can be created while GC progresses, but they are not garbage for
	// this round:
	//  - new stack spans can be created even while the world is stopped.
	//  - new malloc spans can be created during the concurrent sweep

	// Even if this is stop-the-world, a concurrent exitsyscall can allocate a stack from heap.
	_lock.Lock(&_lock.Mheap_.Lock)
	// Free the old cached sweep array if necessary.
	if _sched.Work.Spans != nil && &_sched.Work.Spans[0] != &H_allspans[0] {
		_sched.SysFree(unsafe.Pointer(&_sched.Work.Spans[0]), uintptr(len(_sched.Work.Spans))*unsafe.Sizeof(_sched.Work.Spans[0]), &_lock.Memstats.Other_sys)
	}
	// Cache the current array for marking.
	_lock.Mheap_.Gcspans = _lock.Mheap_.Allspans
	_sched.Work.Spans = H_allspans
	_lock.Unlock(&_lock.Mheap_.Lock)
	oldphase := _sched.Gcphase

	_sched.Work.Nwait = 0
	_sched.Work.Ndone = 0
	_sched.Work.Nproc = uint32(gcprocs())
	_sched.Gcphase = _sched.GCmarktermination

	// World is stopped so allglen will not change.
	for i := uintptr(0); i < Allglen; i++ {
		gp := _lock.Allgs[i]
		gp.Gcworkdone = false // set to true in gcphasework
	}

	parforsetup(_sched.Work.Markfor, _sched.Work.Nproc, uint32(_sched.RootCount+Allglen), nil, false, markroot)
	if _sched.Work.Nproc > 1 {
		_sched.Noteclear(&_sched.Work.Alldone)
		helpgc(int32(_sched.Work.Nproc))
	}

	var t2 int64
	if _lock.Debug.Gctrace > 0 {
		t2 = _lock.Nanotime()
	}

	_sched.Gchelperstart()
	_sched.Parfordo(_sched.Work.Markfor)
	_sched.Scanblock(0, 0, nil)

	if _sched.Work.Full != 0 {
		_lock.Throw("work.full != 0")
	}
	if _sched.Work.Partial != 0 {
		_lock.Throw("work.partial != 0")
	}

	_sched.Gcphase = oldphase
	var t3 int64
	if _lock.Debug.Gctrace > 0 {
		t3 = _lock.Nanotime()
	}

	if _sched.Work.Nproc > 1 {
		_sched.Notesleep(&_sched.Work.Alldone)
	}

	shrinkfinish()

	cachestats()
	// next_gc calculation is tricky with concurrent sweep since we don't know size of live heap
	// estimate what was live heap size after previous GC (for printing only)
	heap0 := _lock.Memstats.Next_gc * 100 / (uint64(Gcpercent) + 100)
	// conservatively set next_gc to high value assuming that everything is live
	// concurrent/lazy sweep will reduce this number while discovering new garbage
	_lock.Memstats.Next_gc = _lock.Memstats.Heap_alloc + _lock.Memstats.Heap_alloc*uint64(Gcpercent)/100

	t4 := _lock.Nanotime()
	_sched.Atomicstore64(&_lock.Memstats.Last_gc, uint64(Unixnanotime())) // must be Unix time to make sense to user
	_lock.Memstats.Pause_ns[_lock.Memstats.Numgc%uint32(len(_lock.Memstats.Pause_ns))] = uint64(t4 - t0)
	_lock.Memstats.Pause_end[_lock.Memstats.Numgc%uint32(len(_lock.Memstats.Pause_end))] = uint64(t4)
	_lock.Memstats.Pause_total_ns += uint64(t4 - t0)
	_lock.Memstats.Numgc++
	if _lock.Memstats.Debuggc {
		print("pause ", t4-t0, "\n")
	}

	if _lock.Debug.Gctrace > 0 {
		heap1 := _lock.Memstats.Heap_alloc
		var stats _core.Gcstats
		Updatememstats(&stats)
		if heap1 != _lock.Memstats.Heap_alloc {
			print("runtime: mstats skew: heap=", heap1, "/", _lock.Memstats.Heap_alloc, "\n")
			_lock.Throw("mstats skew")
		}
		obj := _lock.Memstats.Nmalloc - _lock.Memstats.Nfree

		stats.Nprocyield += _sched.Work.Markfor.Nprocyield
		stats.Nosyield += _sched.Work.Markfor.Nosyield
		stats.Nsleep += _sched.Work.Markfor.Nsleep

		print("gc", _lock.Memstats.Numgc, "(", _sched.Work.Nproc, "): ",
			(t1-t0)/1000, "+", (t2-t1)/1000, "+", (t3-t2)/1000, "+", (t4-t3)/1000, " us, ",
			heap0>>20, " -> ", heap1>>20, " MB, ",
			obj, " (", _lock.Memstats.Nmalloc, "-", _lock.Memstats.Nfree, ") objects, ",
			Gcount(), " goroutines, ",
			len(_sched.Work.Spans), "/", sweep.nbgsweep, "/", sweep.npausesweep, " sweeps, ",
			stats.Nhandoff, "(", stats.Nhandoffcnt, ") handoff, ",
			_sched.Work.Markfor.Nsteal, "(", _sched.Work.Markfor.Nstealcnt, ") steal, ",
			stats.Nprocyield, "/", stats.Nosyield, "/", stats.Nsleep, " yields\n")
		sweep.nbgsweep = 0
		sweep.npausesweep = 0
	}

	// See the comment in the beginning of this function as to why we need the following.
	// Even if this is still stop-the-world, a concurrent exitsyscall can allocate a stack from heap.
	_lock.Lock(&_lock.Mheap_.Lock)
	// Free the old cached mark array if necessary.
	if _sched.Work.Spans != nil && &_sched.Work.Spans[0] != &H_allspans[0] {
		_sched.SysFree(unsafe.Pointer(&_sched.Work.Spans[0]), uintptr(len(_sched.Work.Spans))*unsafe.Sizeof(_sched.Work.Spans[0]), &_lock.Memstats.Other_sys)
	}

	if Gccheckmarkenable {
		if !_sched.Checkmark {
			// first half of two-pass; don't set up sweep
			_lock.Unlock(&_lock.Mheap_.Lock)
			return
		}
		_sched.Checkmark = false // done checking marks
	}

	// Cache the current array for sweeping.
	_lock.Mheap_.Gcspans = _lock.Mheap_.Allspans
	_lock.Mheap_.Sweepgen += 2
	_lock.Mheap_.Sweepdone = 0
	_sched.Work.Spans = H_allspans
	sweep.spanidx = 0
	_lock.Unlock(&_lock.Mheap_.Lock)

	if _sched.XConcurrentSweep && !eagersweep {
		_lock.Lock(&gclock)
		if !sweep.started {
			go bgsweep()
			sweep.started = true
		} else if sweep.parked {
			sweep.parked = false
			_sched.Ready(sweep.g)
		}
		_lock.Unlock(&gclock)
	} else {
		// Sweep all spans eagerly.
		for Sweepone() != ^uintptr(0) {
			sweep.npausesweep++
		}
		// Do an additional mProf_GC, because all 'free' events are now real as well.
		mProf_GC()
	}

	mProf_GC()
	_g_.M.Traceback = 0

	if _sched.DebugGCPtrs {
		print("GC end\n")
	}
}

//go:nowritebarrier
func Addb(p *byte, n uintptr) *byte {
	return (*byte)(_core.Add(unsafe.Pointer(p), n))
}

// unmark the span of memory at v of length n bytes.
//go:nowritebarrier
func unmarkspan(v, n uintptr) {
	if v+n > _lock.Mheap_.Arena_used || v < _lock.Mheap_.Arena_start {
		_lock.Throw("markspan: bad pointer")
	}

	off := (v - _lock.Mheap_.Arena_start) / _core.PtrSize // word offset
	if off%(_core.PtrSize*_sched.WordsPerBitmapByte) != 0 {
		_lock.Throw("markspan: unaligned pointer")
	}

	b := _lock.Mheap_.Arena_start - off/_sched.WordsPerBitmapByte - 1
	n /= _core.PtrSize
	if n%(_core.PtrSize*_sched.WordsPerBitmapByte) != 0 {
		_lock.Throw("unmarkspan: unaligned length")
	}

	// Okay to use non-atomic ops here, because we control
	// the entire span, and each bitmap word has bits for only
	// one span, so no other goroutines are changing these
	// bitmap words.
	n /= _sched.WordsPerBitmapByte
	_core.Memclr(unsafe.Pointer(b-n+1), n)
}

func Unixnanotime() int64 {
	var now int64
	gc_unixnanotime(&now)
	return now
}
