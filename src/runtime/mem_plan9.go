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

var base memHdr
var freep *memHdr

func memAlloc(n uintptr) unsafe.Pointer {
	var p *memHdr

	nunits := (n+unsafe.Sizeof(*p)-1)/unsafe.Sizeof(*p) + 1
	prevp := freep
	if prevp == nil {
		base.next = &base
		freep = &base
		prevp = &base
		base.size = 0
	}
	p = prevp.next
	for {
		if p.size >= nunits {
			if p.size == nunits {
				prevp.next = p.next
			} else {
				p.size -= nunits
				p = (*memHdr)(unsafe.Pointer((uintptr(unsafe.Pointer(p))) + p.size))
				p.size = nunits
			}
			freep = prevp
			return unsafe.Pointer((uintptr(unsafe.Pointer(p))) + unsafe.Sizeof(*p))
		}
		if p == freep {
			p = morecore(nunits)
			if p == nil {
				return nil
			}
		}
		prevp = p
		p = p.next
	}
}

func morecore(nu uintptr) *memHdr {
	var up *memHdr

	cp := sbrk(nu * unsafe.Sizeof(*up))
	if cp == nil {
		return nil
	}
	up = (*memHdr)(cp)
	up.size = nu
	memFree(unsafe.Pointer((uintptr(unsafe.Pointer(up))) + unsafe.Sizeof(*up)))
	return freep
}

func memFree(ap unsafe.Pointer) {
	var p *memHdr

	bp := (*memHdr)(unsafe.Pointer((uintptr(unsafe.Pointer(ap))) - unsafe.Sizeof(*p)))
	bpn := uintptr(unsafe.Pointer(bp))
	pn := uintptr(unsafe.Pointer(p))
	for p = freep; !(bpn > pn && bpn < uintptr(unsafe.Pointer(p.next))); p = p.next {
		bpn = uintptr(unsafe.Pointer(bp))
		pn = uintptr(unsafe.Pointer(p))
		if pn >= uintptr(unsafe.Pointer(p.next)) && (bpn > pn || bpn < uintptr(unsafe.Pointer(p.next))) {
			break
		}
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
	freep = p
	memclr(ap, p.size*unsafe.Sizeof(*p))
}

func memRound(p uintptr) uintptr {
	return (p + _PAGESIZE - 1) &^ (_PAGESIZE - 1)
}

func initBloc() {
	bloc = memRound(uintptr(unsafe.Pointer(&end)))
}

func sbrk(n uintptr) unsafe.Pointer {
	lock(&memlock)
	// Plan 9 sbrk from /sys/src/libc/9sys/sbrk.c
	bl := bloc
	n = memRound(n)
	if brk_(unsafe.Pointer(bl+n)) < 0 {
		unlock(&memlock)
		return nil
	}
	bloc += n
	unlock(&memlock)
	return unsafe.Pointer(bl)
}

func sysAlloc(n uintptr, stat *uint64) unsafe.Pointer {
	p := memAlloc(n)
	if p != nil {
		xadd64(stat, int64(n))
	}
	return p
}

func sysFree(v unsafe.Pointer, n uintptr, stat *uint64) {
	xadd64(stat, -int64(n))
	lock(&memlock)
	memFree(v)
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
	return memAlloc(n)
}
