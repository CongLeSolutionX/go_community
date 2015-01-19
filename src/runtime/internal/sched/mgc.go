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

	WorkbufSize     = 4 * 1024
	FinBlockSize    = 4 * 1024
	RootData        = 0
	RootBss         = 1
	RootFinalizers  = 2
	RootSpans       = 3
	RootFlushCaches = 4
	RootCount       = 5
)

// It is a bug if bits does not have bitBoundary set but
// there are still some cases where this happens related
// to stack spans.
type Markbits struct {
	Bitp  *byte   // pointer to the byte holding xbits
	Shift uintptr // bits xbits needs to be shifted to get bits
	Xbits byte    // byte holding all the bits from *bitp
	Bits  byte    // mark and boundary bits relevant to corresponding slot.
	tbits byte    // pointer||scalar bits relevant to corresponding slot.
}

type Workbuf struct {
	node lfnode // must be first
	nobj uintptr
	obj  [(WorkbufSize - unsafe.Sizeof(lfnode{}) - _core.PtrSize) / _core.PtrSize]uintptr
}

var Finlock _core.Mutex // protects the following variables
var Fingwait bool
var Fingwake bool

type Workdata struct {
	Full    uint64                     // lock-free list of full blocks
	empty   uint64                     // lock-free list of empty blocks
	Partial uint64                     // lock-free list of partially filled blocks
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
// The first nibble no longer holds the bitsDead pattern indicating that the
// there are no more pointers in the object. This information is held
// in the second nibble.

// When marking an object if the bool checkmark is true one uses the above
// encoding, otherwise one uses the bitMarked bit in the lower two bits
// of the nibble.
var (
	Checkmark = false
)

// Is address b in the known heap. If it doesn't have a valid gcmap
// returns false. For example pointers into stacks will return false.
//go:nowritebarrier
func Inheap(b uintptr) bool {
	if b == 0 || b < _lock.Mheap_.Arena_start || b >= _lock.Mheap_.Arena_used {
		return false
	}
	// Not a beginning of a block, consult span table to find the block beginning.
	k := b >> _core.PageShift
	x := k
	x -= _lock.Mheap_.Arena_start >> _core.PageShift
	s := H_spans[x]
	if s == nil || _core.PageID(k) < s.Start || b >= s.Limit || s.State != MSpanInUse {
		return false
	}
	return true
}

// Given an address in the heap return the relevant byte from the gcmap. This routine
// can be used on addresses to the start of an object or to the interior of the an object.
//go:nowritebarrier
func Slottombits(obj uintptr, mbits *Markbits) {
	off := (obj&^(_core.PtrSize-1) - _lock.Mheap_.Arena_start) / _core.PtrSize
	*(*uintptr)(unsafe.Pointer(&mbits.Bitp)) = _lock.Mheap_.Arena_start - off/WordsPerBitmapByte - 1
	mbits.Shift = off % WordsPerBitmapByte * GcBits
	mbits.Xbits = *mbits.Bitp
	mbits.Bits = (mbits.Xbits >> mbits.Shift) & BitMask
	mbits.tbits = ((mbits.Xbits >> mbits.Shift) & BitPtrMask) >> 2
}

// b is a pointer into the heap.
// Find the start of the object refered to by b.
// Set mbits to the associated bits from the bit map.
// If b is not a valid heap object return nil and
// undefined values in mbits.
//go:nowritebarrier
func Objectstart(b uintptr, mbits *Markbits) uintptr {
	obj := b &^ (_core.PtrSize - 1)
	for {
		Slottombits(obj, mbits)
		if mbits.Bits&BitBoundary == BitBoundary {
			break
		}

		// Not a beginning of a block, consult span table to find the block beginning.
		k := b >> _core.PageShift
		x := k
		x -= _lock.Mheap_.Arena_start >> _core.PageShift
		s := H_spans[x]
		if s == nil || _core.PageID(k) < s.Start || b >= s.Limit || s.State != MSpanInUse {
			if s != nil && s.State == MSpanStack {
				return 0 // This is legit.
			}

			// The following ensures that we are rigorous about what data
			// structures hold valid pointers
			if false {
				// Still happens sometimes. We don't know why.
				Printlock()
				print("runtime:objectstart Span weird: obj=", _core.Hex(obj), " k=", _core.Hex(k))
				if s == nil {
					print(" s=nil\n")
				} else {
					print(" s.start=", _core.Hex(s.Start<<_core.PageShift), " s.limit=", _core.Hex(s.Limit), " s.state=", s.State, "\n")
				}
				Printunlock()
				_lock.Gothrow("objectstart: bad pointer in unexpected span")
			}
			return 0
		}

		p := uintptr(s.Start) << _core.PageShift
		if s.Sizeclass != 0 {
			size := s.Elemsize
			idx := (obj - p) / size
			p = p + idx*size
		}
		if p == obj {
			print("runtime: failed to find block beginning for ", _core.Hex(p), " s=", _core.Hex(s.Start*_core.PageSize), " s.limit=", _core.Hex(s.Limit), "\n")
			_lock.Gothrow("failed to find block beginning")
		}
		obj = p
	}

	// if size(obj.firstfield) < PtrSize, the &obj.secondfield could map to the boundary bit
	// Clear any low bits to get to the start of the object.
	// greyobject depends on this.
	return obj
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

// Mark using the checkmark scheme.
//go:nowritebarrier
func docheckmark(mbits *Markbits) {
	// xor 01 moves 01(scalar unmarked) to 00(scalar marked)
	// and 10(pointer unmarked) to 11(pointer marked)
	if mbits.tbits == BitsScalar {
		atomicand8(mbits.Bitp, ^byte(BitsCheckMarkXor<<mbits.Shift<<2))
	} else if mbits.tbits == BitsPointer {
		Atomicor8(mbits.Bitp, byte(BitsCheckMarkXor<<mbits.Shift<<2))
	}

	// reload bits for ischeckmarked
	mbits.Xbits = *mbits.Bitp
	mbits.Bits = (mbits.Xbits >> mbits.Shift) & BitMask
	mbits.tbits = ((mbits.Xbits >> mbits.Shift) & BitPtrMask) >> 2
}

// In the default scheme does mbits refer to a marked object.
//go:nowritebarrier
func ismarked(mbits *Markbits) bool {
	if mbits.Bits&BitBoundary != BitBoundary {
		_lock.Gothrow("ismarked: bits should have boundary bit set")
	}
	return mbits.Bits&BitMarked == BitMarked
}

// In the checkmark scheme does mbits refer to a marked object.
//go:nowritebarrier
func Ischeckmarked(mbits *Markbits) bool {
	if mbits.Bits&BitBoundary != BitBoundary {
		_lock.Gothrow("ischeckmarked: bits should have boundary bit set")
	}
	return mbits.tbits == BitsScalarMarked || mbits.tbits == BitsPointerMarked
}

// obj is the start of an object with mark mbits.
// If it isn't already marked, mark it and enqueue into workbuf.
// Return possibly new workbuf to use.
//go:nowritebarrier
func Greyobject(obj uintptr, mbits *Markbits, wbuf *Workbuf) *Workbuf {
	// obj should be start of allocation, and so must be at least pointer-aligned.
	if obj&(_core.PtrSize-1) != 0 {
		_lock.Gothrow("greyobject: obj not pointer-aligned")
	}

	if Checkmark {
		if !ismarked(mbits) {
			print("runtime:greyobject: checkmarks finds unexpected unmarked object obj=", _core.Hex(obj), ", mbits->bits=", _core.Hex(mbits.Bits), " *mbits->bitp=", _core.Hex(*mbits.Bitp), "\n")

			k := obj >> _core.PageShift
			x := k
			x -= _lock.Mheap_.Arena_start >> _core.PageShift
			s := H_spans[x]
			Printlock()
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
			_lock.Gothrow("checkmark found unmarked object")
		}
		if Ischeckmarked(mbits) {
			return wbuf
		}
		docheckmark(mbits)
		if !Ischeckmarked(mbits) {
			print("mbits xbits=", _core.Hex(mbits.Xbits), " bits=", _core.Hex(mbits.Bits), " tbits=", _core.Hex(mbits.tbits), " shift=", mbits.Shift, "\n")
			_lock.Gothrow("docheckmark and ischeckmarked disagree")
		}
	} else {
		// If marked we have nothing to do.
		if mbits.Bits&BitMarked != 0 {
			return wbuf
		}

		// Each byte of GC bitmap holds info for two words.
		// If the current object is larger than two words, or if the object is one word
		// but the object it shares the byte with is already marked,
		// then all the possible concurrent updates are trying to set the same bit,
		// so we can use a non-atomic update.
		if mbits.Xbits&(BitMask|BitMask<<GcBits) != BitBoundary|BitBoundary<<GcBits || Work.Nproc == 1 {
			*mbits.Bitp = mbits.Xbits | BitMarked<<mbits.Shift
		} else {
			Atomicor8(mbits.Bitp, BitMarked<<mbits.Shift)
		}
	}

	if !Checkmark && (mbits.Xbits>>(mbits.Shift+2))&BitsMask == BitsDead {
		return wbuf // noscan object
	}

	// Queue the obj for scanning. The PREFETCH(obj) logic has been removed but
	// seems like a nice optimization that can be added back in.
	// There needs to be time between the PREFETCH and the use.
	// Previously we put the obj in an 8 element buffer that is drained at a rate
	// to give the PREFETCH time to do its work.
	// Use of PREFETCHNTA might be more appropriate than PREFETCH

	// If workbuf is full, obtain an empty one.
	if wbuf.nobj >= uintptr(len(wbuf.obj)) {
		wbuf = getempty(wbuf)
	}

	wbuf.obj[wbuf.nobj] = obj
	wbuf.nobj++
	return wbuf
}

// Scan the object b of size n, adding pointers to wbuf.
// Return possibly new wbuf to use.
// If ptrmask != nil, it specifies where pointers are in b.
// If ptrmask == nil, the GC bitmap should be consulted.
// In this case, n may be an overestimate of the size; the GC bitmap
// must also be used to make sure the scan stops at the end of b.
//go:nowritebarrier
func scanobject(b, n uintptr, ptrmask *uint8, wbuf *Workbuf) *Workbuf {
	arena_start := _lock.Mheap_.Arena_start
	arena_used := _lock.Mheap_.Arena_used

	// Find bits of the beginning of the object.
	var ptrbitp unsafe.Pointer
	var mbits Markbits
	if ptrmask == nil {
		b = Objectstart(b, &mbits)
		if b == 0 {
			return wbuf
		}
		ptrbitp = unsafe.Pointer(mbits.Bitp)
	}
	for i := uintptr(0); i < n; i += _core.PtrSize {
		// Find bits for this word.
		var bits uintptr
		if ptrmask != nil {
			// dense mask (stack or data)
			bits = (uintptr(*(*byte)(_core.Add(unsafe.Pointer(ptrmask), (i/_core.PtrSize)/4))) >> (((i / _core.PtrSize) % 4) * XBitsPerPointer)) & XBitsMask
		} else {
			// Check if we have reached end of span.
			// n is an overestimate of the size of the object.
			if (b+i)%_core.PageSize == 0 && H_spans[(b-arena_start)>>_core.PageShift] != H_spans[(b+i-arena_start)>>_core.PageShift] {
				break
			}

			// Consult GC bitmap.
			bits = uintptr(*(*byte)(ptrbitp))
			if WordsPerBitmapByte != 2 {
				_lock.Gothrow("alg doesn't work for wordsPerBitmapByte != 2")
			}
			j := (uintptr(b) + i) / _core.PtrSize & 1 // j indicates upper nibble or lower nibble
			bits >>= GcBits * j
			if i == 0 {
				bits &^= BitBoundary
			}
			ptrbitp = _core.Add(ptrbitp, -j)

			if bits&BitBoundary != 0 && i != 0 {
				break // reached beginning of the next object
			}
			bits = (bits & BitPtrMask) >> 2 // bits refer to the type bits.

			if i != 0 && bits == XBitsDead { // BitsDead in first nibble not valid during checkmark
				break // reached no-scan part of the object
			}
		}

		if bits <= BitsScalar { // _BitsScalar, _BitsDead, _BitsScalarMarked
			continue
		}

		if bits&BitsPointer != BitsPointer {
			print("gc checkmark=", Checkmark, " b=", _core.Hex(b), " ptrmask=", ptrmask, " mbits.bitp=", mbits.Bitp, " mbits.xbits=", _core.Hex(mbits.Xbits), " bits=", _core.Hex(bits), "\n")
			_lock.Gothrow("unexpected garbage collection bits")
		}

		obj := *(*uintptr)(unsafe.Pointer(b + i))

		// At this point we have extracted the next potential pointer.
		// Check if it points into heap.
		if obj == 0 || obj < arena_start || obj >= arena_used {
			continue
		}

		// Mark the object. return some important bits.
		// We we combine the following two rotines we don't have to pass mbits or obj around.
		var mbits Markbits
		obj = Objectstart(obj, &mbits)
		if obj == 0 {
			continue
		}
		wbuf = Greyobject(obj, &mbits, wbuf)
	}
	return wbuf
}

// scanblock starts by scanning b as scanobject would.
// If the gcphase is GCscan, that's all scanblock does.
// Otherwise it traverses some fraction of the pointers it found in b, recursively.
// As a special case, scanblock(nil, 0, nil) means to scan previously queued work,
// stopping only when no work is left in the system.
//go:nowritebarrier
func Scanblock(b, n uintptr, ptrmask *uint8) {
	wbuf := Getpartialorempty()
	if b != 0 {
		wbuf = scanobject(b, n, ptrmask, wbuf)
		if Gcphase == GCscan {
			if Inheap(b) && ptrmask == nil {
				// b is in heap, we are in GCscan so there should be a ptrmask.
				_lock.Gothrow("scanblock: In GCscan phase and inheap is true.")
			}
			// GCscan only goes one level deep since mark wb not turned on.
			Putpartial(wbuf)
			return
		}
	}
	if Gcphase == GCscan {
		_lock.Gothrow("scanblock: In GCscan phase but no b passed in.")
	}

	keepworking := b == 0

	// ptrmask can have 2 possible values:
	// 1. nil - obtain pointer mask from GC bitmap.
	// 2. pointer to a compact mask (for stacks and data).
	for {
		if wbuf.nobj == 0 {
			if !keepworking {
				putempty(wbuf)
				return
			}
			// Refill workbuf from global queue.
			wbuf = getfull(wbuf)
			if wbuf == nil { // nil means out of work barrier reached
				return
			}

			if wbuf.nobj <= 0 {
				_lock.Gothrow("runtime:scanblock getfull returns empty buffer")
			}
		}

		// If another proc wants a pointer, give it some.
		if Work.Nwait > 0 && wbuf.nobj > 4 && Work.Full == 0 {
			wbuf = handoff(wbuf)
		}

		// This might be a good place to add prefetch code...
		// if(wbuf->nobj > 4) {
		//         PREFETCH(wbuf->obj[wbuf->nobj - 3];
		//  }
		wbuf.nobj--
		b = wbuf.obj[wbuf.nobj]
		wbuf = scanobject(b, _lock.Mheap_.Arena_used-b, nil, wbuf)
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
		b = (*Workbuf)(lfstackpop(&Work.empty))
	}
	if b != nil && b.nobj != 0 {
		_g_ := _core.Getg()
		print("m", _g_.M.Id, ": getempty: popped b=", b, " with non-zero b.nobj=", b.nobj, "\n")
		_lock.Gothrow("getempty: workbuffer not empty, b->nobj not 0")
	}
	if b == nil {
		b = (*Workbuf)(_lock.Persistentalloc(unsafe.Sizeof(*b), _lock.CacheLineSize, &_lock.Memstats.Gc_sys))
		b.nobj = 0
	}
	return b
}

//go:nowritebarrier
func putempty(b *Workbuf) {
	if b.nobj != 0 {
		_lock.Gothrow("putempty: b->nobj not 0")
	}
	lfstackpush(&Work.empty, &b.node)
}

//go:nowritebarrier
func putfull(b *Workbuf) {
	if b.nobj <= 0 {
		_lock.Gothrow("putfull: b->nobj <= 0")
	}
	lfstackpush(&Work.Full, &b.node)
}

// Get an partially empty work buffer
// if none are available get an empty one.
//go:nowritebarrier
func Getpartialorempty() *Workbuf {
	b := (*Workbuf)(lfstackpop(&Work.Partial))
	if b == nil {
		b = getempty(nil)
	}
	return b
}

//go:nowritebarrier
func Putpartial(b *Workbuf) {
	if b.nobj == 0 {
		lfstackpush(&Work.empty, &b.node)
	} else if b.nobj < uintptr(len(b.obj)) {
		lfstackpush(&Work.Partial, &b.node)
	} else if b.nobj == uintptr(len(b.obj)) {
		lfstackpush(&Work.Full, &b.node)
	} else {
		print("b=", b, " b.nobj=", b.nobj, " len(b.obj)=", len(b.obj), "\n")
		_lock.Gothrow("putpartial: bad Workbuf b.nobj")
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
		putempty(b)
	}

	b = (*Workbuf)(lfstackpop(&Work.Full))
	if b == nil {
		b = (*Workbuf)(lfstackpop(&Work.Partial))
	}
	if b != nil || Work.Nproc == 1 {
		return b
	}

	_lock.Xadd(&Work.Nwait, +1)
	for i := 0; ; i++ {
		if Work.Full != 0 {
			_lock.Xadd(&Work.Nwait, -1)
			b = (*Workbuf)(lfstackpop(&Work.Full))
			if b == nil {
				b = (*Workbuf)(lfstackpop(&Work.Partial))
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
	n := b.nobj / 2
	b.nobj -= n
	b1.nobj = n
	Memmove(unsafe.Pointer(&b1.obj[0]), unsafe.Pointer(&b.obj[b.nobj]), n*unsafe.Sizeof(b1.obj[0]))
	_g_ := _core.Getg()
	_g_.M.Gcstats.Nhandoff++
	_g_.M.Gcstats.Nhandoffcnt += uint64(n)

	// Put b on full list - let first half of b get stolen.
	lfstackpush(&Work.Full, &b.node)
	return b1
}

//go:nowritebarrier
func gchelper() {
	_g_ := _core.Getg()
	_g_.M.Traceback = 2
	Gchelperstart()

	// parallel mark for over GC roots
	Parfordo(Work.Markfor)
	if Gcphase != GCscan {
		Scanblock(0, 0, nil) // blocks in getfull
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
		_lock.Gothrow("gchelperstart: bad m->helpgc")
	}
	if _g_ != _g_.M.G0 {
		_lock.Gothrow("gchelper not running on g0 stack")
	}
}

func wakefing() *_core.G {
	var res *_core.G
	_lock.Lock(&Finlock)
	if Fingwait && Fingwake {
		Fingwait = false
		Fingwake = false
		res = _core.Fing
	}
	_lock.Unlock(&Finlock)
	return res
}

//go:nowritebarrier
func mHeap_MapBits(h *_lock.Mheap) {
	// Caller has added extra mappings to the arena.
	// Add extra mappings of bitmap words as needed.
	// We allocate extra bitmap pieces in chunks of bitmapChunk.
	const bitmapChunk = 8192

	n := (h.Arena_used - h.Arena_start) / (_core.PtrSize * WordsPerBitmapByte)
	n = Round(n, bitmapChunk)
	n = Round(n, _lock.PhysPageSize)
	if h.Bitmap_mapped >= n {
		return
	}

	sysMap(unsafe.Pointer(h.Arena_start-n), n-h.Bitmap_mapped, h.Arena_reserved, &_lock.Memstats.Gc_sys)
	h.Bitmap_mapped = n
}
