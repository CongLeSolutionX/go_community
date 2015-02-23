// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

var bloc uintptr
var memlock mutex

type memHdr struct {
	next *memHdr
	size uintptr
}

var memBase memHdr
var memFreelist *memHdr // sorted in ascending order

func memAlloc(n uintptr) unsafe.Pointer {
	if n < unsafe.Sizeof(memHdr{}) {
		n = unsafe.Sizeof(memHdr{})
	}
	prevp := memFreelist
	if prevp == nil {
		memBase.next = &memBase
		memFreelist = &memBase
		prevp = &memBase
		memBase.size = 0
	}
	p := prevp.next
	for {
		if p.size >= n {
			if p.size == n {
				prevp.next = p.next
			} else {
				p.size -= n
				p = (*memHdr)(unsafe.Pointer((uintptr(unsafe.Pointer(p))) + p.size))
				p.size = n
			}
			memFreelist = prevp
			memclr(unsafe.Pointer(p), unsafe.Sizeof(memHdr{}))
			return unsafe.Pointer(p)
		}
		if p == memFreelist {
			p = morecore(n)
			if p == nil {
				return nil
			}
		}
		prevp = p
		p = p.next
	}
}

func morecore(n uintptr) *memHdr {
	cp := sbrk(n)
	if cp == nil {
		return nil
	}
	memFree(unsafe.Pointer(cp), n)
	return memFreelist
}

func memFree(ap unsafe.Pointer, n uintptr) {
	bp := (*memHdr)(ap)
	bp.size = n
	bpn := uintptr(ap)
	p := memFreelist
	pn := uintptr(unsafe.Pointer(p))
	for {
		if bpn > pn && bpn < uintptr(unsafe.Pointer(p.next)) {
			break
		}
		if pn >= uintptr(unsafe.Pointer(p.next)) && (bpn > pn || bpn < uintptr(unsafe.Pointer(p.next))) {
			break
		}
		p = p.next
		pn = uintptr(unsafe.Pointer(p))
	}
	if bpn+bp.size == uintptr(unsafe.Pointer(p.next)) {
		bp.size += p.next.size
		bp.next = p.next.next
	} else {
		bp.next = p.next
	}
	if pn+p.size == bpn {
		p.size += bp.size
		p.next = bp.next
	} else {
		p.next = bp
	}
	memFreelist = p
}

func memRound(p uintptr) uintptr {
	return (p + _PAGESIZE - 1) &^ (_PAGESIZE - 1)
}

func initBloc() {
	bloc = memRound(uintptr(unsafe.Pointer(&end)))
}

func sbrk(n uintptr) unsafe.Pointer {
	// Plan 9 sbrk from /sys/src/libc/9sys/sbrk.c
	bl := bloc
	n = memRound(n)
	if brk_(unsafe.Pointer(bl+n)) < 0 {
		return nil
	}
	bloc += n
	return unsafe.Pointer(bl)
}

func sysAlloc(n uintptr, stat *uint64) unsafe.Pointer {
	lock(&memlock)
	p := memAlloc(n)
	unlock(&memlock)
	if p != nil {
		xadd64(stat, int64(n))
	}
	return p
}

func sysFree(v unsafe.Pointer, n uintptr, stat *uint64) {
	xadd64(stat, -int64(n))
	lock(&memlock)
	memclr(v, n)
	memFree(v, n)
	unlock(&memlock)
}

func sysUnused(v unsafe.Pointer, n uintptr) {
}

func sysUsed(v unsafe.Pointer, n uintptr) {
}

func sysMap(v unsafe.Pointer, n uintptr, reserved bool, stat *uint64) {
	// sysReserve has already allocated all heap memory,
	// but has not adjusted stats.
	xadd64(stat, int64(n))
}

func sysFault(v unsafe.Pointer, n uintptr) {
}

func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	*reserved = true
	lock(&memlock)
	p := memAlloc(n)
	unlock(&memlock)
	return p
}
