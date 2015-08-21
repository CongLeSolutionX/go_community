// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Memory allocator, based on tcmalloc.
// http://goog-perftools.sourceforge.net/doc/tcmalloc.html

// The main allocator works in runs of pages.
// Small allocation sizes (up to and including 32 kB) are
// rounded to one of about 100 size classes, each of which
// has its own free list of objects of exactly that size.
// Any free page of memory can be split into a set of objects
// of one size class, which are then managed using free list
// allocators.
//
// The allocator's data structures are:
//
//	FixAlloc: a free-list allocator for fixed-size objects,
//		used to manage storage used by the allocator.
//	MHeap: the malloc heap, managed at page (4096-byte) granularity.
//	MSpan: a run of pages managed by the MHeap.
//	MCentral: a shared free list for a given size class.
//	MCache: a per-thread (in Go, per-P) cache for small objects.
//	MStats: allocation statistics.
//
// Allocating a small object proceeds up a hierarchy of caches:
//
//	1. Round the size up to one of the small size classes
//	   and look in the corresponding MCache free list.
//	   If the list is not empty, allocate an object from it.
//	   This can all be done without acquiring a lock.
//
//	2. If the MCache free list is empty, replenish it by
//	   taking a bunch of objects from the MCentral free list.
//	   Moving a bunch amortizes the cost of acquiring the MCentral lock.
//
//	3. If the MCentral free list is empty, replenish it by
//	   allocating a run of pages from the MHeap and then
//	   chopping that memory into objects of the given size.
//	   Allocating many objects amortizes the cost of locking
//	   the heap.
//
//	4. If the MHeap is empty or has no page runs large enough,
//	   allocate a new group of pages (at least 1MB) from the
//	   operating system.  Allocating a large run of pages
//	   amortizes the cost of talking to the operating system.
//
// Freeing a small object proceeds up the same hierarchy:
//
//	1. Look up the size class for the object and add it to
//	   the MCache free list.
//
//	2. If the MCache free list is too long or the MCache has
//	   too much memory, return some to the MCentral free lists.
//
//	3. If all the objects in a given span have returned to
//	   the MCentral list, return that span to the page heap.
//
//	4. If the heap has too much memory, return some to the
//	   operating system.
//
//	TODO(rsc): Step 4 is not implemented.
//
// Allocating and freeing a large object uses the page heap
// directly, bypassing the MCache and MCentral free lists.
//
// The small objects on the MCache and MCentral free lists
// may or may not be zeroed.  They are zeroed if and only if
// the second word of the object is zero.  A span in the
// page heap is zeroed unless s->needzero is set. When a span
// is allocated to break into small objects, it is zeroed if needed
// and s->needzero is set. There are two main benefits to delaying the
// zeroing this way:
//
//	1. stack frames allocated from the small object lists
//	   or the page heap can avoid zeroing altogether.
//	2. the cost of zeroing when reusing a small object is
//	   charged to the mutator, not the garbage collector.
//
// This code was written with an eye toward translating to Go
// in the future.  Methods have the form Type_Method(Type *t, ...).

package runtime

import (
	_base "runtime/internal/base"
	_iface "runtime/internal/iface"
	"unsafe"
)

// OS-defined helpers:
//
// sysAlloc obtains a large chunk of zeroed memory from the
// operating system, typically on the order of a hundred kilobytes
// or a megabyte.
// NOTE: sysAlloc returns OS-aligned memory, but the heap allocator
// may use larger alignment, so the caller must be careful to realign the
// memory obtained by sysAlloc.
//
// SysUnused notifies the operating system that the contents
// of the memory region are no longer needed and can be reused
// for other purposes.
// SysUsed notifies the operating system that the contents
// of the memory region are needed again.
//
// SysFree returns it unconditionally; this is only used if
// an out-of-memory error has been detected midway through
// an allocation.  It is okay if SysFree is a no-op.
//
// SysReserve reserves address space without allocating memory.
// If the pointer passed to it is non-nil, the caller wants the
// reservation there, but SysReserve can still choose another
// location if that one is unavailable.  On some systems and in some
// cases SysReserve will simply check that the address space is
// available and not actually reserve it.  If SysReserve returns
// non-nil, it sets *reserved to true if the address space is
// reserved, false if it has merely been checked.
// NOTE: SysReserve returns OS-aligned memory, but the heap allocator
// may use larger alignment, so the caller must be careful to realign the
// memory obtained by sysAlloc.
//
// SysMap maps previously reserved address space for use.
// The reserved argument is true if the address space was really
// reserved, not merely checked.
//
// SysFault marks a (already sysAlloc'd) region to fault
// if accessed.  Used only for debugging the runtime.

func mallocinit() {
	initSizes()

	if _iface.Class_to_size[_base.XTinySizeClass] != _base.TinySize {
		_base.Throw("bad TinySizeClass")
	}

	var p, bitmapSize, spansSize, pSize, limit uintptr
	var reserved bool

	// limit = runtime.memlimit();
	// See https://golang.org/issue/5049
	// TODO(rsc): Fix after 1.1.
	limit = 0

	// Set up the allocation arena, a contiguous area of memory where
	// allocated data will be found.  The arena begins with a bitmap large
	// enough to hold 4 bits per allocated word.
	if _base.PtrSize == 8 && (limit == 0 || limit > 1<<30) {
		// On a 64-bit machine, allocate from a single contiguous reservation.
		// 512 GB (MaxMem) should be big enough for now.
		//
		// The code will work with the reservation at any address, but ask
		// SysReserve to use 0x0000XXc000000000 if possible (XX=00...7f).
		// Allocating a 512 GB region takes away 39 bits, and the amd64
		// doesn't let us choose the top 17 bits, so that leaves the 9 bits
		// in the middle of 0x00c0 for us to choose.  Choosing 0x00c0 means
		// that the valid memory addresses will begin 0x00c0, 0x00c1, ..., 0x00df.
		// In little-endian, that's c0 00, c1 00, ..., df 00. None of those are valid
		// UTF-8 sequences, and they are otherwise as far away from
		// ff (likely a common byte) as possible.  If that fails, we try other 0xXXc0
		// addresses.  An earlier attempt to use 0x11f8 caused out of memory errors
		// on OS X during thread allocations.  0x00c0 causes conflicts with
		// AddressSanitizer which reserves all memory up to 0x0100.
		// These choices are both for debuggability and to reduce the
		// odds of a conservative garbage collector (as is still used in gccgo)
		// not collecting memory because some non-pointer block of memory
		// had a bit pattern that matched a memory address.
		//
		// Actually we reserve 544 GB (because the bitmap ends up being 32 GB)
		// but it hardly matters: e0 00 is not valid UTF-8 either.
		//
		// If this fails we fall back to the 32 bit memory mechanism
		//
		// However, on arm64, we ignore all this advice above and slam the
		// allocation at 0x40 << 32 because when using 4k pages with 3-level
		// translation buffers, the user address space is limited to 39 bits
		// On darwin/arm64, the address space is even smaller.
		arenaSize := _base.Round(_base.MaxMem, _base.XPageSize)
		bitmapSize = arenaSize / (_base.PtrSize * 8 / 4)
		spansSize = arenaSize / _base.XPageSize * _base.PtrSize
		spansSize = _base.Round(spansSize, _base.XPageSize)
		for i := 0; i <= 0x7f; i++ {
			switch {
			case _base.GOARCH == "arm64" && _base.GOOS == "darwin":
				p = uintptr(i)<<40 | _base.UintptrMask&(0x0013<<28)
			case _base.GOARCH == "arm64":
				p = uintptr(i)<<40 | _base.UintptrMask&(0x0040<<32)
			default:
				p = uintptr(i)<<40 | _base.UintptrMask&(0x00c0<<32)
			}
			pSize = bitmapSize + spansSize + arenaSize + _base.XPageSize
			p = uintptr(_base.SysReserve(unsafe.Pointer(p), pSize, &reserved))
			if p != 0 {
				break
			}
		}
	}

	if p == 0 {
		// On a 32-bit machine, we can't typically get away
		// with a giant virtual address space reservation.
		// Instead we map the memory information bitmap
		// immediately after the data segment, large enough
		// to handle another 2GB of mappings (256 MB),
		// along with a reservation for an initial arena.
		// When that gets used up, we'll start asking the kernel
		// for any memory anywhere and hope it's in the 2GB
		// following the bitmap (presumably the executable begins
		// near the bottom of memory, so we'll have to use up
		// most of memory before the kernel resorts to giving out
		// memory before the beginning of the text segment).
		//
		// Alternatively we could reserve 512 MB bitmap, enough
		// for 4GB of mappings, and then accept any memory the
		// kernel threw at us, but normally that's a waste of 512 MB
		// of address space, which is probably too much in a 32-bit world.

		// If we fail to allocate, try again with a smaller arena.
		// This is necessary on Android L where we share a process
		// with ART, which reserves virtual memory aggressively.
		arenaSizes := []uintptr{
			512 << 20,
			256 << 20,
			128 << 20,
		}

		for _, arenaSize := range arenaSizes {
			bitmapSize = _base.MaxArena32 / (_base.PtrSize * 8 / 4)
			spansSize = _base.MaxArena32 / _base.XPageSize * _base.PtrSize
			if limit > 0 && arenaSize+bitmapSize+spansSize > limit {
				bitmapSize = (limit / 9) &^ ((1 << _base.XPageShift) - 1)
				arenaSize = bitmapSize * 8
				spansSize = arenaSize / _base.XPageSize * _base.PtrSize
			}
			spansSize = _base.Round(spansSize, _base.XPageSize)

			// SysReserve treats the address we ask for, end, as a hint,
			// not as an absolute requirement.  If we ask for the end
			// of the data segment but the operating system requires
			// a little more space before we can start allocating, it will
			// give out a slightly higher pointer.  Except QEMU, which
			// is buggy, as usual: it won't adjust the pointer upward.
			// So adjust it upward a little bit ourselves: 1/4 MB to get
			// away from the running binary image and then round up
			// to a MB boundary.
			p = _base.Round(_base.Firstmoduledata.End+(1<<18), 1<<20)
			pSize = bitmapSize + spansSize + arenaSize + _base.XPageSize
			p = uintptr(_base.SysReserve(unsafe.Pointer(p), pSize, &reserved))
			if p != 0 {
				break
			}
		}
		if p == 0 {
			_base.Throw("runtime: cannot reserve arena virtual address space")
		}
	}

	// PageSize can be larger than OS definition of page size,
	// so SysReserve can give us a PageSize-unaligned pointer.
	// To overcome this we ask for PageSize more and round up the pointer.
	p1 := _base.Round(p, _base.XPageSize)

	_base.Mheap_.Spans = (**_base.Mspan)(unsafe.Pointer(p1))
	_base.Mheap_.Bitmap = p1 + spansSize
	_base.Mheap_.Arena_start = p1 + (spansSize + bitmapSize)
	_base.Mheap_.Arena_used = _base.Mheap_.Arena_start
	_base.Mheap_.Arena_end = p + pSize
	_base.Mheap_.Arena_reserved = reserved

	if _base.Mheap_.Arena_start&(_base.XPageSize-1) != 0 {
		println("bad pagesize", _base.Hex(p), _base.Hex(p1), _base.Hex(spansSize), _base.Hex(bitmapSize), _base.Hex(_base.XPageSize), "start", _base.Hex(_base.Mheap_.Arena_start))
		_base.Throw("misrounded allocation in mallocinit")
	}

	// Initialize the rest of the allocator.
	mHeap_Init(&_base.Mheap_, spansSize)
	_g_ := _base.Getg()
	_g_.M.Mcache = _base.Allocmcache()
}

// sysReserveHigh reserves space somewhere high in the address space.
// sysReserve doesn't actually reserve the full amount requested on
// 64-bit systems, because of problems with ulimit. Instead it checks
// that it can get the first 64 kB and assumes it can grab the rest as
// needed. This doesn't work well with the "let the kernel pick an address"
// mode, so don't do that. Pick a high address instead.
func sysReserveHigh(n uintptr, reserved *bool) unsafe.Pointer {
	if _base.PtrSize == 4 {
		return _base.SysReserve(nil, n, reserved)
	}

	for i := 0; i <= 0x7f; i++ {
		p := uintptr(i)<<40 | _base.UintptrMask&(0x00c0<<32)
		*reserved = false
		p = uintptr(_base.SysReserve(unsafe.Pointer(p), n, reserved))
		if p != 0 {
			return unsafe.Pointer(p)
		}
	}

	return _base.SysReserve(nil, n, reserved)
}

//go:linkname reflect_unsafe_New reflect.unsafe_New
func reflect_unsafe_New(typ *_base.Type) unsafe.Pointer {
	return _iface.Newobject(typ)
}

// implementation of make builtin for slices
func newarray(typ *_base.Type, n uintptr) unsafe.Pointer {
	flags := uint32(0)
	if typ.Kind&_iface.KindNoPointers != 0 {
		flags |= _base.XFlagNoScan
	}
	if int(n) < 0 || (typ.Size > 0 && n > _base.MaxMem/uintptr(typ.Size)) {
		panic("runtime: allocation size out of range")
	}
	return _iface.Mallocgc(uintptr(typ.Size)*n, typ, flags)
}

//go:linkname reflect_unsafe_NewArray reflect.unsafe_NewArray
func reflect_unsafe_NewArray(typ *_base.Type, n uintptr) unsafe.Pointer {
	return newarray(typ, n)
}

// rawmem returns a chunk of pointerless memory.  It is
// not zeroed.
func rawmem(size uintptr) unsafe.Pointer {
	return _iface.Mallocgc(size, nil, _base.XFlagNoScan|_base.XFlagNoZero)
}
