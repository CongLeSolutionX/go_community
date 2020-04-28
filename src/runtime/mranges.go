// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Address range data structure.
//
// This file contains an implementation of a data structure which
// manages ordered address ranges.

package runtime

import (
	"runtime/internal/sys"
	"unsafe"
)

// addrRange represents a region of address space.
//
// An addrRange must never span a gap in the address space.
type addrRange struct {
	// base and limit together represent the region of address space
	// [base, limit], i.e. an inclusive range.
	//
	// These are address over an offset view of the address space on
	// platforms with a segmented address space, that is, on platforms
	// where arenaBaseOffset != 0.
	base, limit offAddr
}

var dummyAddrRange = addrRange{
	base:  1,
	limit: 0,
}

// makeAddrRange creates an addrRange from two real addresses.
func makeAddrRange(base, limit uintptr) addrRange {
	return addrRange{
		base:  offsetAddress(base),
		limit: offsetAddress(limit),
	}
}

// start is the inclusive lower bound of the range.
func (a addrRange) start() uintptr {
	return a.base.addr()
}

// end is the exclusive upper bound of the range.
func (a addrRange) end() uintptr {
	return a.limit.addr()
}

// size returns the size of the range represented in bytes.
func (a addrRange) size() uintptr {
	if a.limit < a.base {
		return 0
	}
	// Subtraction is safe because limit and base must bein the same
	// segment of the address space.
	return uintptr(a.limit - a.base + 1)
}

// contains returns whether or not the range contains a given address.
func (a addrRange) contains(addr uintptr) bool {
	return offsetAddress(addr) >= a.base && offsetAddress(addr) <= a.limit
}

// subtract takes the addrRange toPrune and cuts out any overlap with
// from, then returns the new range. subtract assumes that a and b
// either don't overlap at all, only overlap on one side, or are equal.
// If b is strictly contained in a, thus forcing a split, it will throw.
func (a addrRange) subtract(b addrRange) addrRange {
	if a.base >= b.base && a.limit <= b.limit {
		return dummyAddrRange
	} else if a.base < b.base && a.limit > b.limit {
		throw("bad prune")
	} else if a.limit > b.limit && a.base < b.limit {
		a.base = b.limit
	} else if a.base < b.base && a.limit > b.base {
		a.limit = b.base
	}
	return a
}

// offAddr represents an address in a contiguous view
// of the address space on systems where the address space is
// segmented. On other systems, it's just a normal address.
//
// Although arithmetic operations are supported on offAddrs,
// most operations between two offAddrs are not well-defined,
// except subtraction, and only if the users knows the two
// offAddrs' corresponding real addresses are in the same
// segment.
type offAddr uintptr

// offsetAddress generates a offAddr from a real address.
func offsetAddress(a uintptr) offAddr {
	return offAddr(a + arenaBaseOffset)
}

// add adds a uintptr offset to the offAddr.
func (l offAddr) add(bytes uintptr) offAddr {
	return offAddr(uintptr(l) + bytes)
}

// sub subtracts a uintptr offset from the offAddr.
func (l offAddr) sub(bytes uintptr) offAddr {
	return offAddr(uintptr(l) - bytes)
}

// addr returns the read address for this linearized address.
func (l offAddr) addr() uintptr {
	return uintptr(l) - arenaBaseOffset
}

// addrRanges is a data structure holding a collection of ranges of
// address space.
//
// The ranges are coalesced eagerly to reduce the
// number ranges it holds.
//
// The slice backing store for this field is persistentalloc'd
// and thus there is no way to free it.
//
// addrRanges is not thread-safe.
type addrRanges struct {
	// ranges is a slice of ranges sorted by base.
	ranges []addrRange

	// totalBytes is the total amount address space in bytes counted by
	// this addrRanges.
	totalBytes uintptr

	// sysStat is the stat to track allocations by this type
	sysStat *uint64
}

func (a *addrRanges) init(sysStat *uint64) {
	ranges := (*notInHeapSlice)(unsafe.Pointer(&a.ranges))
	ranges.len = 0
	ranges.cap = 16
	ranges.array = (*notInHeap)(persistentalloc(unsafe.Sizeof(addrRange{})*uintptr(ranges.cap), sys.PtrSize, sysStat))
	a.sysStat = sysStat
	a.totalBytes = 0
}

// findSucc returns the first index in a such that base is
// less than the base of the addrRange at that index.
func (a *addrRanges) findSucc(addr uintptr) int {
	// TODO(mknyszek): Consider a binary search for large arrays.
	// While iterating over these ranges is potentially expensive,
	// the expected number of ranges is small, ideally just 1,
	// since Go heaps are usually mostly contiguous.
	base := offsetAddress(addr)
	for i := range a.ranges {
		if base < a.ranges[i].base {
			return i
		}
	}
	return len(a.ranges)
}

// contains returns true if a covers the address addr.
func (a *addrRanges) contains(addr uintptr) bool {
	i := a.findSucc(addr)
	if i == 0 {
		return false
	}
	return a.ranges[i-1].contains(addr)
}

// add inserts a new address range to a.
//
// r must not overlap with any address range in a.
func (a *addrRanges) add(r addrRange) {
	// The copies in this function are potentially expensive, but this data
	// structure is meant to represent the Go heap. At worst, copying this
	// would take ~160µs assuming a conservative copying rate of 25 GiB/s (the
	// copy will almost never trigger a page fault) for a 1 TiB heap with 4 MiB
	// arenas which is completely discontiguous. ~160µs is still a lot, but in
	// practice most platforms have 64 MiB arenas (which cuts this by a factor
	// of 16) and Go heaps are usually mostly contiguous, so the chance that
	// an addrRanges even grows to that size is extremely low.

	// Because we assume r is not currently represented in a,
	// findSucc gives us our insertion index.
	i := a.findSucc(r.start())
	coalescesDown := i > 0 && a.ranges[i-1].limit == r.base-1
	coalescesUp := i < len(a.ranges) && r.limit == a.ranges[i].base-1
	if coalescesUp && coalescesDown {
		// We have neighbors and they both border us.
		// Merge a.ranges[i-1], r, and a.ranges[i] together into a.ranges[i-1].
		a.ranges[i-1].limit = a.ranges[i].limit

		// Delete a.ranges[i].
		copy(a.ranges[i:], a.ranges[i+1:])
		a.ranges = a.ranges[:len(a.ranges)-1]
	} else if coalescesDown {
		// We have a neighbor at a lower address only and it borders us.
		// Merge the new space into a.ranges[i-1].
		a.ranges[i-1].limit = r.limit
	} else if coalescesUp {
		// We have a neighbor at a higher address only and it borders us.
		// Merge the new space into a.ranges[i].
		a.ranges[i].base = r.base
	} else {
		// We may or may not have neighbors which don't border us.
		// Add the new range.
		if len(a.ranges)+1 > cap(a.ranges) {
			// Grow the array. Note that this leaks the old array, but since
			// we're doubling we have at most 2x waste. For a 1 TiB heap and
			// 4 MiB arenas which are all discontiguous (both very conservative
			// assumptions), this would waste at most 4 MiB of memory.
			oldRanges := a.ranges
			ranges := (*notInHeapSlice)(unsafe.Pointer(&a.ranges))
			ranges.len = len(oldRanges) + 1
			ranges.cap = cap(oldRanges) * 2
			ranges.array = (*notInHeap)(persistentalloc(unsafe.Sizeof(addrRange{})*uintptr(ranges.cap), sys.PtrSize, a.sysStat))

			// Copy in the old array, but make space for the new range.
			copy(a.ranges[:i], oldRanges[:i])
			copy(a.ranges[i+1:], oldRanges[i:])
		} else {
			a.ranges = a.ranges[:len(a.ranges)+1]
			copy(a.ranges[i+1:], a.ranges[i:])
		}
		a.ranges[i] = r
	}
	a.totalBytes += r.size()
}

// removeLast removes and returns the highest-addressed contiguous range
// of a, or the last nbytes of that range, whichever is smaller. If a is
// empty, it returns an empty range.
func (a *addrRanges) removeLast(nbytes uintptr) addrRange {
	if len(a.ranges) == 0 {
		return dummyAddrRange
	}
	r := a.ranges[len(a.ranges)-1]
	size := r.size()
	if size > nbytes {
		newLimit := r.limit.sub(nbytes)
		a.ranges[len(a.ranges)-1].limit = newLimit
		a.totalBytes -= nbytes
		// Add 1 to newLimit so that it's aligned to a
		// page instead of 1 less. Adding one here is safe
		// from overflow because newLimit is strictly smaller
		// than r.limit.
		return addrRange{newLimit + 1, r.limit}
	}
	a.ranges = a.ranges[:len(a.ranges)-1]
	a.totalBytes -= size
	return r
}

// removeAbove removes the ranges of a which are above addr, and additionally
// splits any range containing addr.
func (a *addrRanges) removeAbove(addr uintptr) {
	pivot := a.findSucc(addr)
	if pivot == 0 {
		// addr is before all ranges in a.
		a.totalBytes = 0
		a.ranges = a.ranges[:0]
		return
	}
	removed := uintptr(0)
	for _, r := range a.ranges[pivot:] {
		removed += r.size()
	}
	if r := a.ranges[pivot-1]; r.contains(addr) {
		removed += r.size()
		r = r.subtract(makeAddrRange(addr, maxSearchAddr))
		removed -= r.size()
		a.ranges[pivot-1] = r
	}
	a.ranges = a.ranges[:pivot]
	a.totalBytes -= removed
}

// cloneInto makes a deep clone of a's state into b, re-using
// b's ranges if able.
func (a *addrRanges) cloneInto(b *addrRanges) {
	if len(a.ranges) > cap(b.ranges) {
		// Grow the array.
		ranges := (*notInHeapSlice)(unsafe.Pointer(&b.ranges))
		ranges.len = 0
		ranges.cap = cap(a.ranges)
		ranges.array = (*notInHeap)(persistentalloc(unsafe.Sizeof(addrRange{})*uintptr(ranges.cap), sys.PtrSize, b.sysStat))
	}
	b.ranges = b.ranges[:len(a.ranges)]
	b.totalBytes = a.totalBytes
	copy(b.ranges, a.ranges)
}
