// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// See malloc.h for overview.
//
// TODO(rsc): double-check stats.

package schedinit

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	"unsafe"
)

func mallocinit() {
	initSizes()

	if _gc.Class_to_size[_core.TinySizeClass] != _core.TinySize {
		_lock.Throw("bad TinySizeClass")
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
	if _core.PtrSize == 8 && (limit == 0 || limit > 1<<30) {
		// On a 64-bit machine, allocate from a single contiguous reservation.
		// 128 GB (MaxMem) should be big enough for now.
		//
		// The code will work with the reservation at any address, but ask
		// SysReserve to use 0x0000XXc000000000 if possible (XX=00...7f).
		// Allocating a 128 GB region takes away 37 bits, and the amd64
		// doesn't let us choose the top 17 bits, so that leaves the 11 bits
		// in the middle of 0x00c0 for us to choose.  Choosing 0x00c0 means
		// that the valid memory addresses will begin 0x00c0, 0x00c1, ..., 0x00df.
		// In little-endian, that's c0 00, c1 00, ..., df 00. None of those are valid
		// UTF-8 sequences, and they are otherwise as far away from
		// ff (likely a common byte) as possible.  If that fails, we try other 0xXXc0
		// addresses.  An earlier attempt to use 0x11f8 caused out of memory errors
		// on OS X during thread allocations.  0x00c0 causes conflicts with
		// AddressSanitizer which reserves all memory up to 0x0100.
		// These choices are both for debuggability and to reduce the
		// odds of the conservative garbage collector not collecting memory
		// because some non-pointer block of memory had a bit pattern
		// that matched a memory address.
		//
		// Actually we reserve 136 GB (because the bitmap ends up being 8 GB)
		// but it hardly matters: e0 00 is not valid UTF-8 either.
		//
		// If this fails we fall back to the 32 bit memory mechanism
		arenaSize := _lock.Round(_core.MaxMem, _core.PageSize)
		bitmapSize = arenaSize / (_core.PtrSize * 8 / 4)
		spansSize = arenaSize / _core.PageSize * _core.PtrSize
		spansSize = _lock.Round(spansSize, _core.PageSize)
		for i := 0; i <= 0x7f; i++ {
			p = uintptr(i)<<40 | _lock.UintptrMask&(0x00c0<<32)
			pSize = bitmapSize + spansSize + arenaSize + _core.PageSize
			p = uintptr(_sched.SysReserve(unsafe.Pointer(p), pSize, &reserved))
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
		}

		for _, arenaSize := range arenaSizes {
			bitmapSize = _sched.MaxArena32 / (_core.PtrSize * 8 / 4)
			spansSize = _sched.MaxArena32 / _core.PageSize * _core.PtrSize
			if limit > 0 && arenaSize+bitmapSize+spansSize > limit {
				bitmapSize = (limit / 9) &^ ((1 << _core.PageShift) - 1)
				arenaSize = bitmapSize * 8
				spansSize = arenaSize / _core.PageSize * _core.PtrSize
			}
			spansSize = _lock.Round(spansSize, _core.PageSize)

			// SysReserve treats the address we ask for, end, as a hint,
			// not as an absolute requirement.  If we ask for the end
			// of the data segment but the operating system requires
			// a little more space before we can start allocating, it will
			// give out a slightly higher pointer.  Except QEMU, which
			// is buggy, as usual: it won't adjust the pointer upward.
			// So adjust it upward a little bit ourselves: 1/4 MB to get
			// away from the running binary image and then round up
			// to a MB boundary.
			p = _lock.Round(uintptr(unsafe.Pointer(&end))+(1<<18), 1<<20)
			pSize = bitmapSize + spansSize + arenaSize + _core.PageSize
			p = uintptr(_sched.SysReserve(unsafe.Pointer(p), pSize, &reserved))
			if p != 0 {
				break
			}
		}
		if p == 0 {
			_lock.Throw("runtime: cannot reserve arena virtual address space")
		}
	}

	// PageSize can be larger than OS definition of page size,
	// so SysReserve can give us a PageSize-unaligned pointer.
	// To overcome this we ask for PageSize more and round up the pointer.
	p1 := _lock.Round(p, _core.PageSize)

	_lock.Mheap_.Spans = (**_core.Mspan)(unsafe.Pointer(p1))
	_lock.Mheap_.Bitmap = p1 + spansSize
	_lock.Mheap_.Arena_start = p1 + (spansSize + bitmapSize)
	_lock.Mheap_.Arena_used = _lock.Mheap_.Arena_start
	_lock.Mheap_.Arena_end = p + pSize
	_lock.Mheap_.Arena_reserved = reserved

	if _lock.Mheap_.Arena_start&(_core.PageSize-1) != 0 {
		println("bad pagesize", _core.Hex(p), _core.Hex(p1), _core.Hex(spansSize), _core.Hex(bitmapSize), _core.Hex(_core.PageSize), "start", _core.Hex(_lock.Mheap_.Arena_start))
		_lock.Throw("misrounded allocation in mallocinit")
	}

	// Initialize the rest of the allocator.
	mHeap_Init(&_lock.Mheap_, spansSize)
	_g_ := _core.Getg()
	_g_.M.Mcache = _lock.Allocmcache()
}

// sysReserveHigh reserves space somewhere high in the address space.
// sysReserve doesn't actually reserve the full amount requested on
// 64-bit systems, because of problems with ulimit. Instead it checks
// that it can get the first 64 kB and assumes it can grab the rest as
// needed. This doesn't work well with the "let the kernel pick an address"
// mode, so don't do that. Pick a high address instead.
func sysReserveHigh(n uintptr, reserved *bool) unsafe.Pointer {
	if _core.PtrSize == 4 {
		return _sched.SysReserve(nil, n, reserved)
	}

	for i := 0; i <= 0x7f; i++ {
		p := uintptr(i)<<40 | _lock.UintptrMask&(0x00c0<<32)
		*reserved = false
		p = uintptr(_sched.SysReserve(unsafe.Pointer(p), n, reserved))
		if p != 0 {
			return unsafe.Pointer(p)
		}
	}

	return _sched.SysReserve(nil, n, reserved)
}

var end struct{}
