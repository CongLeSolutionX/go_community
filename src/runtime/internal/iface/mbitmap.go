// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: type and heap bitmaps.
//
// Stack, data, and bss bitmaps
//
// Stack frames and global variables in the data and bss sections are described
// by 1-bit bitmaps in which 0 means uninteresting and 1 means live pointer
// to be visited during GC. The bits in each byte are consumed starting with
// the low bit: 1<<0, 1<<1, and so on.
//
// Heap bitmap
//
// The allocated heap comes from a subset of the memory in the range [start, used),
// where start == mheap_.arena_start and used == mheap_.arena_used.
// The heap bitmap comprises 2 bits for each pointer-sized word in that range,
// stored in bytes indexed backward in memory from start.
// That is, the byte at address start-1 holds the 2-bit entries for the four words
// start through start+3*ptrSize, the byte at start-2 holds the entries for
// start+4*ptrSize through start+7*ptrSize, and so on.
//
// In each 2-bit entry, the lower bit holds the same information as in the 1-bit
// bitmaps: 0 means uninteresting and 1 means live pointer to be visited during GC.
// The meaning of the high bit depends on the position of the word being described
// in its allocated object. In the first word, the high bit is the GC ``marked'' bit.
// In the second word, the high bit is the GC ``checkmarked'' bit (see below).
// In the third and later words, the high bit indicates that the object is still
// being described. In these words, if a bit pair with a high bit 0 is encountered,
// the low bit can also be assumed to be 0, and the object description is over.
// This 00 is called the ``dead'' encoding: it signals that the rest of the words
// in the object are uninteresting to the garbage collector.
//
// The 2-bit entries are split when written into the byte, so that the top half
// of the byte contains 4 mark bits and the bottom half contains 4 pointer bits.
// This form allows a copy from the 1-bit to the 4-bit form to keep the
// pointer bits contiguous, instead of having to space them out.
//
// The code makes use of the fact that the zero value for a heap bitmap
// has no live pointer bit set and is (depending on position), not marked,
// not checkmarked, and is the dead encoding.
// These properties must be preserved when modifying the encoding.
//
// Checkmarks
//
// In a concurrent garbage collector, one worries about failing to mark
// a live object due to mutations without write barriers or bugs in the
// collector implementation. As a sanity check, the GC has a 'checkmark'
// mode that retraverses the object graph with the world stopped, to make
// sure that everything that should be marked is marked.
// In checkmark mode, in the heap bitmap, the high bit of the 2-bit entry
// for the second word of the object holds the checkmark bit.
// When not in checkmark mode, this bit is set to 1.
//
// The smallest possible allocation is 8 bytes. On a 32-bit machine, that
// means every allocated object has two words, so there is room for the
// checkmark bit. On a 64-bit machine, however, the 8-byte allocation is
// just one word, so the second bit pair is not available for encoding the
// checkmark. However, because non-pointer allocations are combined
// into larger 16-byte (maxTinySize) allocations, a plain 8-byte allocation
// must be a pointer, so the type bit in the first word is not actually needed.
// It is still used in general, except in checkmark the type bit is repurposed
// as the checkmark bit and then reinitialized (to 1) as the type bit when
// finished.

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	"unsafe"
)

// heapBitsBulkBarrier executes writebarrierptr_nostore
// for every pointer slot in the memory range [p, p+size),
// using the heap bitmap to locate those pointer slots.
// This executes the write barriers necessary after a memmove.
// Both p and size must be pointer-aligned.
// The range [p, p+size) must lie within a single allocation.
//
// Callers should call heapBitsBulkBarrier immediately after
// calling memmove(p, src, size). This function is marked nosplit
// to avoid being preempted; the GC must not stop the goroutine
// between the memmove and the execution of the barriers.
//
// The heap bitmap is not maintained for allocations containing
// no pointers at all; any caller of heapBitsBulkBarrier must first
// make sure the underlying allocation contains pointers, usually
// by checking typ.kind&kindNoPointers.
//
//go:nosplit
func HeapBitsBulkBarrier(p, size uintptr) {
	if (p|size)&(_base.PtrSize-1) != 0 {
		_base.Throw("heapBitsBulkBarrier: unaligned arguments")
	}
	if !_base.WriteBarrierEnabled {
		return
	}
	if !_base.Inheap(p) {
		// If p is on the stack and in a higher frame than the
		// caller, we either need to execute write barriers on
		// it (which is what happens for normal stack writes
		// through pointers to higher frames), or we need to
		// force the mark termination stack scan to scan the
		// frame containing p.
		//
		// Executing write barriers on p is complicated in the
		// general case because we either need to unwind the
		// stack to get the stack map, or we need the type's
		// bitmap, which may be a GC program.
		//
		// Hence, we opt for forcing the re-scan to scan the
		// frame containing p, which we can do by simply
		// unwinding the stack barriers between the current SP
		// and p's frame.
		gp := _base.Getg().M.Curg
		if gp != nil && gp.Stack.Lo <= p && p < gp.Stack.Hi {
			// Run on the system stack to give it more
			// stack space.
			_base.Systemstack(func() {
				GcUnwindBarriers(gp, p)
			})
		}
		return
	}

	h := _base.HeapBitsForAddr(p)
	for i := uintptr(0); i < size; i += _base.PtrSize {
		if h.IsPointer() {
			x := (*uintptr)(unsafe.Pointer(p + i))
			_base.Writebarrierptr_nostore(x, *x)
		}
		h = h.Next()
	}
}

// heapBitsSetType records that the new allocation [x, x+size)
// holds in [x, x+dataSize) one or more values of type typ.
// (The number of values is given by dataSize / typ.size.)
// If dataSize < size, the fragment [x+dataSize, x+size) is
// recorded as non-pointer data.
// It is known that the type has pointers somewhere;
// malloc does not call heapBitsSetType when there are no pointers,
// because all free objects are marked as noscan during
// heapBitsSweepSpan.
// There can only be one allocation from a given span active at a time,
// so this code is not racing with other instances of itself,
// and we don't allocate from a span until it has been swept,
// so this code is not racing with heapBitsSweepSpan.
// It is, however, racing with the concurrent GC mark phase,
// which can be setting the mark bit in the leading 2-bit entry
// of an allocated block. The block we are modifying is not quite
// allocated yet, so the GC marker is not racing with updates to x's bits,
// but if the start or end of x shares a bitmap byte with an adjacent
// object, the GC marker is racing with updates to those object's mark bits.
func heapBitsSetType(x, size, dataSize uintptr, typ *_base.Type) {
	const doubleCheck = false // slow but helpful; enable to test modifications to this code

	// dataSize is always size rounded up to the next malloc size class,
	// except in the case of allocating a defer block, in which case
	// size is sizeof(_defer{}) (at least 6 words) and dataSize may be
	// arbitrarily larger.
	//
	// The checks for size == ptrSize and size == 2*ptrSize can therefore
	// assume that dataSize == size without checking it explicitly.

	if _base.PtrSize == 8 && size == _base.PtrSize {
		// It's one word and it has pointers, it must be a pointer.
		// In general we'd need an atomic update here if the
		// concurrent GC were marking objects in this span,
		// because each bitmap byte describes 3 other objects
		// in addition to the one being allocated.
		// However, since all allocated one-word objects are pointers
		// (non-pointers are aggregated into tinySize allocations),
		// initSpan sets the pointer bits for us. Nothing to do here.
		if doubleCheck {
			h := _base.HeapBitsForAddr(x)
			if !h.IsPointer() {
				_base.Throw("heapBitsSetType: pointer bit missing")
			}
		}
		return
	}

	h := _base.HeapBitsForAddr(x)
	ptrmask := typ.Gcdata // start of 1-bit pointer mask (or GC program, handled below)

	// Heap bitmap bits for 2-word object are only 4 bits,
	// so also shared with objects next to it; use atomic updates.
	// This is called out as a special case primarily for 32-bit systems,
	// so that on 32-bit systems the code below can assume all objects
	// are 4-word aligned (because they're all 16-byte aligned).
	if size == 2*_base.PtrSize {
		if typ.Size == _base.PtrSize {
			// We're allocating a block big enough to hold two pointers.
			// On 64-bit, that means the actual object must be two pointers,
			// or else we'd have used the one-pointer-sized block.
			// On 32-bit, however, this is the 8-byte block, the smallest one.
			// So it could be that we're allocating one pointer and this was
			// just the smallest block available. Distinguish by checking dataSize.
			// (In general the number of instances of typ being allocated is
			// dataSize/typ.size.)
			if _base.PtrSize == 4 && dataSize == _base.PtrSize {
				// 1 pointer.
				if _base.Gcphase == _base.GCoff {
					*h.Bitp |= _base.BitPointer << h.Shift
				} else {
					_base.Atomicor8(h.Bitp, _base.BitPointer<<h.Shift)
				}
			} else {
				// 2-element slice of pointer.
				if _base.Gcphase == _base.GCoff {
					*h.Bitp |= (_base.BitPointer | _base.BitPointer<<_base.HeapBitsShift) << h.Shift
				} else {
					_base.Atomicor8(h.Bitp, (_base.BitPointer|_base.BitPointer<<_base.HeapBitsShift)<<h.Shift)
				}
			}
			return
		}
		// Otherwise typ.size must be 2*ptrSize, and typ.kind&kindGCProg == 0.
		if doubleCheck {
			if typ.Size != 2*_base.PtrSize || typ.Kind&KindGCProg != 0 {
				print("runtime: heapBitsSetType size=", size, " but typ.size=", typ.Size, " gcprog=", typ.Kind&KindGCProg != 0, "\n")
				_base.Throw("heapBitsSetType")
			}
		}
		b := uint32(*ptrmask)
		hb := b & 3
		if _base.Gcphase == _base.GCoff {
			*h.Bitp |= uint8(hb << h.Shift)
		} else {
			_base.Atomicor8(h.Bitp, uint8(hb<<h.Shift))
		}
		return
	}

	// Copy from 1-bit ptrmask into 2-bit bitmap.
	// The basic approach is to use a single uintptr as a bit buffer,
	// alternating between reloading the buffer and writing bitmap bytes.
	// In general, one load can supply two bitmap byte writes.
	// This is a lot of lines of code, but it compiles into relatively few
	// machine instructions.

	var (
		// Ptrmask input.
		p     *byte   // last ptrmask byte read
		b     uintptr // ptrmask bits already loaded
		nb    uintptr // number of bits in b at next read
		endp  *byte   // final ptrmask byte to read (then repeat)
		endnb uintptr // number of valid bits in *endp
		pbits uintptr // alternate source of bits

		// Heap bitmap output.
		w     uintptr // words processed
		nw    uintptr // number of words to process
		hbitp *byte   // next heap bitmap byte to write
		hb    uintptr // bits being prepared for *hbitp
	)

	hbitp = h.Bitp

	// Handle GC program. Delayed until this part of the code
	// so that we can use the same double-checking mechanism
	// as the 1-bit case. Nothing above could have encountered
	// GC programs: the cases were all too small.
	if typ.Kind&KindGCProg != 0 {
		heapBitsSetTypeGCProg(h, typ.Ptrdata, typ.Size, dataSize, size, _gc.Addb(typ.Gcdata, 4))
		if doubleCheck {
			// Double-check the heap bits written by GC program
			// by running the GC program to create a 1-bit pointer mask
			// and then jumping to the double-check code below.
			// This doesn't catch bugs shared between the 1-bit and 4-bit
			// GC program execution, but it does catch mistakes specific
			// to just one of those and bugs in heapBitsSetTypeGCProg's
			// implementation of arrays.
			_base.Lock(&debugPtrmask.lock)
			if debugPtrmask.data == nil {
				debugPtrmask.data = (*byte)(_base.Persistentalloc(1<<20, 1, &_base.Memstats.Other_sys))
			}
			ptrmask = debugPtrmask.data
			_gc.RunGCProg(_gc.Addb(typ.Gcdata, 4), nil, ptrmask, 1)
			goto Phase4
		}
		return
	}

	// Note about sizes:
	//
	// typ.size is the number of words in the object,
	// and typ.ptrdata is the number of words in the prefix
	// of the object that contains pointers. That is, the final
	// typ.size - typ.ptrdata words contain no pointers.
	// This allows optimization of a common pattern where
	// an object has a small header followed by a large scalar
	// buffer. If we know the pointers are over, we don't have
	// to scan the buffer's heap bitmap at all.
	// The 1-bit ptrmasks are sized to contain only bits for
	// the typ.ptrdata prefix, zero padded out to a full byte
	// of bitmap. This code sets nw (below) so that heap bitmap
	// bits are only written for the typ.ptrdata prefix; if there is
	// more room in the allocated object, the next heap bitmap
	// entry is a 00, indicating that there are no more pointers
	// to scan. So only the ptrmask for the ptrdata bytes is needed.
	//
	// Replicated copies are not as nice: if there is an array of
	// objects with scalar tails, all but the last tail does have to
	// be initialized, because there is no way to say "skip forward".
	// However, because of the possibility of a repeated type with
	// size not a multiple of 4 pointers (one heap bitmap byte),
	// the code already must handle the last ptrmask byte specially
	// by treating it as containing only the bits for endnb pointers,
	// where endnb <= 4. We represent large scalar tails that must
	// be expanded in the replication by setting endnb larger than 4.
	// This will have the effect of reading many bits out of b,
	// but once the real bits are shifted out, b will supply as many
	// zero bits as we try to read, which is exactly what we need.

	p = ptrmask
	if typ.Size < dataSize {
		// Filling in bits for an array of typ.
		// Set up for repetition of ptrmask during main loop.
		// Note that ptrmask describes only a prefix of
		const maxBits = _base.PtrSize*8 - 7
		if typ.Ptrdata/_base.PtrSize <= maxBits {
			// Entire ptrmask fits in uintptr with room for a byte fragment.
			// Load into pbits and never read from ptrmask again.
			// This is especially important when the ptrmask has
			// fewer than 8 bits in it; otherwise the reload in the middle
			// of the Phase 2 loop would itself need to loop to gather
			// at least 8 bits.

			// Accumulate ptrmask into b.
			// ptrmask is sized to describe only typ.ptrdata, but we record
			// it as describing typ.size bytes, since all the high bits are zero.
			nb = typ.Ptrdata / _base.PtrSize
			for i := uintptr(0); i < nb; i += 8 {
				b |= uintptr(*p) << i
				p = _base.Add1(p)
			}
			nb = typ.Size / _base.PtrSize

			// Replicate ptrmask to fill entire pbits uintptr.
			// Doubling and truncating is fewer steps than
			// iterating by nb each time. (nb could be 1.)
			// Since we loaded typ.ptrdata/ptrSize bits
			// but are pretending to have typ.size/ptrSize,
			// there might be no replication necessary/possible.
			pbits = b
			endnb = nb
			if nb+nb <= maxBits {
				for endnb <= _base.PtrSize*8 {
					pbits |= pbits << endnb
					endnb += endnb
				}
				// Truncate to a multiple of original ptrmask.
				endnb = maxBits / nb * nb
				pbits &= 1<<endnb - 1
				b = pbits
				nb = endnb
			}

			// Clear p and endp as sentinel for using pbits.
			// Checked during Phase 2 loop.
			p = nil
			endp = nil
		} else {
			// Ptrmask is larger. Read it multiple times.
			n := (typ.Ptrdata/_base.PtrSize+7)/8 - 1
			endp = _gc.Addb(ptrmask, n)
			endnb = typ.Size/_base.PtrSize - n*8
		}
	}
	if p != nil {
		b = uintptr(*p)
		p = _base.Add1(p)
		nb = 8
	}

	if typ.Size == dataSize {
		// Single entry: can stop once we reach the non-pointer data.
		nw = typ.Ptrdata / _base.PtrSize
	} else {
		// Repeated instances of typ in an array.
		// Have to process first N-1 entries in full, but can stop
		// once we reach the non-pointer data in the final entry.
		nw = ((dataSize/typ.Size-1)*typ.Size + typ.Ptrdata) / _base.PtrSize
	}
	if nw == 0 {
		// No pointers! Caller was supposed to check.
		println("runtime: invalid type ", *typ.String)
		_base.Throw("heapBitsSetType: called with non-pointer type")
		return
	}
	if nw < 2 {
		// Must write at least 2 words, because the "no scan"
		// encoding doesn't take effect until the third word.
		nw = 2
	}

	// Phase 1: Special case for leading byte (shift==0) or half-byte (shift==4).
	// The leading byte is special because it contains the bits for words 0 and 1,
	// which do not have the marked bits set.
	// The leading half-byte is special because it's a half a byte and must be
	// manipulated atomically.
	switch {
	default:
		_base.Throw("heapBitsSetType: unexpected shift")

	case h.Shift == 0:
		// Ptrmask and heap bitmap are aligned.
		// Handle first byte of bitmap specially.
		// The first byte we write out contains the first two words of the object.
		// In those words, the mark bits are mark and checkmark, respectively,
		// and must not be set. In all following words, we want to set the mark bit
		// as a signal that the object continues to the next 2-bit entry in the bitmap.
		hb = b & _base.BitPointerAll
		hb |= _base.BitMarked<<(2*_base.HeapBitsShift) | _base.BitMarked<<(3*_base.HeapBitsShift)
		if w += 4; w >= nw {
			goto Phase3
		}
		*hbitp = uint8(hb)
		hbitp = _base.Subtract1(hbitp)
		b >>= 4
		nb -= 4

	case _base.PtrSize == 8 && h.Shift == 2:
		// Ptrmask and heap bitmap are misaligned.
		// The bits for the first two words are in a byte shared with another object
		// and must be updated atomically.
		// NOTE(rsc): The atomic here may not be necessary.
		// We took care of 1-word and 2-word objects above,
		// so this is at least a 6-word object, so our start bits
		// are shared only with the type bits of another object,
		// not with its mark bit. Since there is only one allocation
		// from a given span at a time, we should be able to set
		// these bits non-atomically. Not worth the risk right now.
		hb = (b & 3) << (2 * _base.HeapBitsShift)
		b >>= 2
		nb -= 2
		// Note: no bitMarker in hb because the first two words don't get markers from us.
		if _base.Gcphase == _base.GCoff {
			*hbitp |= uint8(hb)
		} else {
			_base.Atomicor8(hbitp, uint8(hb))
		}
		hbitp = _base.Subtract1(hbitp)
		if w += 2; w >= nw {
			// We know that there is more data, because we handled 2-word objects above.
			// This must be at least a 6-word object. If we're out of pointer words,
			// mark no scan in next bitmap byte and finish.
			hb = 0
			w += 4
			goto Phase3
		}
	}

	// Phase 2: Full bytes in bitmap, up to but not including write to last byte (full or partial) in bitmap.
	// The loop computes the bits for that last write but does not execute the write;
	// it leaves the bits in hb for processing by phase 3.
	// To avoid repeated adjustment of nb, we subtract out the 4 bits we're going to
	// use in the first half of the loop right now, and then we only adjust nb explicitly
	// if the 8 bits used by each iteration isn't balanced by 8 bits loaded mid-loop.
	nb -= 4
	for {
		// Emit bitmap byte.
		// b has at least nb+4 bits, with one exception:
		// if w+4 >= nw, then b has only nw-w bits,
		// but we'll stop at the break and then truncate
		// appropriately in Phase 3.
		hb = b & _base.BitPointerAll
		hb |= _base.BitMarkedAll
		if w += 4; w >= nw {
			break
		}
		*hbitp = uint8(hb)
		hbitp = _base.Subtract1(hbitp)
		b >>= 4

		// Load more bits. b has nb right now.
		if p != endp {
			// Fast path: keep reading from ptrmask.
			// nb unmodified: we just loaded 8 bits,
			// and the next iteration will consume 8 bits,
			// leaving us with the same nb the next time we're here.
			if nb < 8 {
				b |= uintptr(*p) << nb
				p = _base.Add1(p)
			} else {
				// Reduce the number of bits in b.
				// This is important if we skipped
				// over a scalar tail, since nb could
				// be larger than the bit width of b.
				nb -= 8
			}
		} else if p == nil {
			// Almost as fast path: track bit count and refill from pbits.
			// For short repetitions.
			if nb < 8 {
				b |= pbits << nb
				nb += endnb
			}
			nb -= 8 // for next iteration
		} else {
			// Slow path: reached end of ptrmask.
			// Process final partial byte and rewind to start.
			b |= uintptr(*p) << nb
			nb += endnb
			if nb < 8 {
				b |= uintptr(*ptrmask) << nb
				p = _base.Add1(ptrmask)
			} else {
				nb -= 8
				p = ptrmask
			}
		}

		// Emit bitmap byte.
		hb = b & _base.BitPointerAll
		hb |= _base.BitMarkedAll
		if w += 4; w >= nw {
			break
		}
		*hbitp = uint8(hb)
		hbitp = _base.Subtract1(hbitp)
		b >>= 4
	}

Phase3:
	// Phase 3: Write last byte or partial byte and zero the rest of the bitmap entries.
	if w > nw {
		// Counting the 4 entries in hb not yet written to memory,
		// there are more entries than possible pointer slots.
		// Discard the excess entries (can't be more than 3).
		mask := uintptr(1)<<(4-(w-nw)) - 1
		hb &= mask | mask<<4 // apply mask to both pointer bits and mark bits
	}

	// Change nw from counting possibly-pointer words to total words in allocation.
	nw = size / _base.PtrSize

	// Write whole bitmap bytes.
	// The first is hb, the rest are zero.
	if w <= nw {
		*hbitp = uint8(hb)
		hbitp = _base.Subtract1(hbitp)
		hb = 0 // for possible final half-byte below
		for w += 4; w <= nw; w += 4 {
			*hbitp = 0
			hbitp = _base.Subtract1(hbitp)
		}
	}

	// Write final partial bitmap byte if any.
	// We know w > nw, or else we'd still be in the loop above.
	// It can be bigger only due to the 4 entries in hb that it counts.
	// If w == nw+4 then there's nothing left to do: we wrote all nw entries
	// and can discard the 4 sitting in hb.
	// But if w == nw+2, we need to write first two in hb.
	// The byte is shared with the next object so we may need an atomic.
	if w == nw+2 {
		if _base.Gcphase == _base.GCoff {
			*hbitp = *hbitp&^(_base.BitPointer|_base.BitMarked|(_base.BitPointer|_base.BitMarked)<<_base.HeapBitsShift) | uint8(hb)
		} else {
			atomicand8(hbitp, ^uint8(_base.BitPointer|_base.BitMarked|(_base.BitPointer|_base.BitMarked)<<_base.HeapBitsShift))
			_base.Atomicor8(hbitp, uint8(hb))
		}
	}

Phase4:
	// Phase 4: all done, but perhaps double check.
	if doubleCheck {
		end := _base.HeapBitsForAddr(x + size)
		if typ.Kind&KindGCProg == 0 && (hbitp != end.Bitp || (w == nw+2) != (end.Shift == 2)) {
			println("ended at wrong bitmap byte for", *typ.String, "x", dataSize/typ.Size)
			print("typ.size=", typ.Size, " typ.ptrdata=", typ.Ptrdata, " dataSize=", dataSize, " size=", size, "\n")
			print("w=", w, " nw=", nw, " b=", _base.Hex(b), " nb=", nb, " hb=", _base.Hex(hb), "\n")
			h0 := _base.HeapBitsForAddr(x)
			print("initial bits h0.bitp=", h0.Bitp, " h0.shift=", h0.Shift, "\n")
			print("ended at hbitp=", hbitp, " but next starts at bitp=", end.Bitp, " shift=", end.Shift, "\n")
			_base.Throw("bad heapBitsSetType")
		}

		// Double-check that bits to be written were written correctly.
		// Does not check that other bits were not written, unfortunately.
		h := _base.HeapBitsForAddr(x)
		nptr := typ.Ptrdata / _base.PtrSize
		ndata := typ.Size / _base.PtrSize
		count := dataSize / typ.Size
		totalptr := ((count-1)*typ.Size + typ.Ptrdata) / _base.PtrSize
		for i := uintptr(0); i < size/_base.PtrSize; i++ {
			j := i % ndata
			var have, want uint8
			have = (*h.Bitp >> h.Shift) & (_base.BitPointer | _base.BitMarked)
			if i >= totalptr {
				want = 0 // deadmarker
				if typ.Kind&KindGCProg != 0 && i < (totalptr+3)/4*4 {
					want = _base.BitMarked
				}
			} else {
				if j < nptr && (*_gc.Addb(ptrmask, j/8)>>(j%8))&1 != 0 {
					want |= _base.BitPointer
				}
				if i >= 2 {
					want |= _base.BitMarked
				} else {
					have &^= _base.BitMarked
				}
			}
			if have != want {
				println("mismatch writing bits for", *typ.String, "x", dataSize/typ.Size)
				print("typ.size=", typ.Size, " typ.ptrdata=", typ.Ptrdata, " dataSize=", dataSize, " size=", size, "\n")
				print("kindGCProg=", typ.Kind&KindGCProg != 0, "\n")
				print("w=", w, " nw=", nw, " b=", _base.Hex(b), " nb=", nb, " hb=", _base.Hex(hb), "\n")
				h0 := _base.HeapBitsForAddr(x)
				print("initial bits h0.bitp=", h0.Bitp, " h0.shift=", h0.Shift, "\n")
				print("current bits h.bitp=", h.Bitp, " h.shift=", h.Shift, " *h.bitp=", _base.Hex(*h.Bitp), "\n")
				print("ptrmask=", ptrmask, " p=", p, " endp=", endp, " endnb=", endnb, " pbits=", _base.Hex(pbits), " b=", _base.Hex(b), " nb=", nb, "\n")
				println("at word", i, "offset", i*_base.PtrSize, "have", have, "want", want)
				if typ.Kind&KindGCProg != 0 {
					println("GC program:")
					dumpGCProg(_gc.Addb(typ.Gcdata, 4))
				}
				_base.Throw("bad heapBitsSetType")
			}
			h = h.Next()
		}
		if ptrmask == debugPtrmask.data {
			_base.Unlock(&debugPtrmask.lock)
		}
	}
}

var debugPtrmask struct {
	lock _base.Mutex
	data *byte
}

// heapBitsSetTypeGCProg implements heapBitsSetType using a GC program.
// progSize is the size of the memory described by the program.
// elemSize is the size of the element that the GC program describes (a prefix of).
// dataSize is the total size of the intended data, a multiple of elemSize.
// allocSize is the total size of the allocated memory.
//
// GC programs are only used for large allocations.
// heapBitsSetType requires that allocSize is a multiple of 4 words,
// so that the relevant bitmap bytes are not shared with surrounding
// objects and need not be accessed with atomic instructions.
func heapBitsSetTypeGCProg(h _base.HeapBits, progSize, elemSize, dataSize, allocSize uintptr, prog *byte) {
	if _base.PtrSize == 8 && allocSize%(4*_base.PtrSize) != 0 {
		// Alignment will be wrong.
		_base.Throw("heapBitsSetTypeGCProg: small allocation")
	}
	var totalBits uintptr
	if elemSize == dataSize {
		totalBits = _gc.RunGCProg(prog, nil, h.Bitp, 2)
		if totalBits*_base.PtrSize != progSize {
			println("runtime: heapBitsSetTypeGCProg: total bits", totalBits, "but progSize", progSize)
			_base.Throw("heapBitsSetTypeGCProg: unexpected bit count")
		}
	} else {
		count := dataSize / elemSize

		// Piece together program trailer to run after prog that does:
		//	literal(0)
		//	repeat(1, elemSize-progSize-1) // zeros to fill element size
		//	repeat(elemSize, count-1) // repeat that element for count
		// This zero-pads the data remaining in the first element and then
		// repeats that first element to fill the array.
		var trailer [40]byte // 3 varints (max 10 each) + some bytes
		i := 0
		if n := elemSize/_base.PtrSize - progSize/_base.PtrSize; n > 0 {
			// literal(0)
			trailer[i] = 0x01
			i++
			trailer[i] = 0
			i++
			if n > 1 {
				// repeat(1, n-1)
				trailer[i] = 0x81
				i++
				n--
				for ; n >= 0x80; n >>= 7 {
					trailer[i] = byte(n | 0x80)
					i++
				}
				trailer[i] = byte(n)
				i++
			}
		}
		// repeat(elemSize/ptrSize, count-1)
		trailer[i] = 0x80
		i++
		n := elemSize / _base.PtrSize
		for ; n >= 0x80; n >>= 7 {
			trailer[i] = byte(n | 0x80)
			i++
		}
		trailer[i] = byte(n)
		i++
		n = count - 1
		for ; n >= 0x80; n >>= 7 {
			trailer[i] = byte(n | 0x80)
			i++
		}
		trailer[i] = byte(n)
		i++
		trailer[i] = 0
		i++

		_gc.RunGCProg(prog, &trailer[0], h.Bitp, 2)

		// Even though we filled in the full array just now,
		// record that we only filled in up to the ptrdata of the
		// last element. This will cause the code below to
		// memclr the dead section of the final array element,
		// so that scanobject can stop early in the final element.
		totalBits = (elemSize*(count-1) + progSize) / _base.PtrSize
	}
	endProg := unsafe.Pointer(_base.Subtractb(h.Bitp, (totalBits+3)/4))
	endAlloc := unsafe.Pointer(_base.Subtractb(h.Bitp, allocSize/_base.HeapBitmapScale))
	_base.Memclr(_base.Add(endAlloc, 1), uintptr(endProg)-uintptr(endAlloc))
}

func dumpGCProg(p *byte) {
	nptr := 0
	for {
		x := *p
		p = _base.Add1(p)
		if x == 0 {
			print("\t", nptr, " end\n")
			break
		}
		if x&0x80 == 0 {
			print("\t", nptr, " lit ", x, ":")
			n := int(x+7) / 8
			for i := 0; i < n; i++ {
				print(" ", _base.Hex(*p))
				p = _base.Add1(p)
			}
			print("\n")
			nptr += int(x)
		} else {
			nbit := int(x &^ 0x80)
			if nbit == 0 {
				for nb := uint(0); ; nb += 7 {
					x := *p
					p = _base.Add1(p)
					nbit |= int(x&0x7f) << nb
					if x&0x80 == 0 {
						break
					}
				}
			}
			count := 0
			for nb := uint(0); ; nb += 7 {
				x := *p
				p = _base.Add1(p)
				count |= int(x&0x7f) << nb
				if x&0x80 == 0 {
					break
				}
			}
			print("\t", nptr, " repeat ", nbit, " × ", count, "\n")
			nptr += nbit * count
		}
	}
}
