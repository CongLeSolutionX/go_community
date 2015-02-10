// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implementation of runtime/debug.WriteHeapDump.  Writes all
// objects in the heap plus additional info (roots, threads,
// finalizers, etc.) to a file.

// The format of the dumped file is described at
// http://golang.org/s/go14heapdump.

package runtime

import (
	_channels "runtime/internal/channels"
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_heapdump "runtime/internal/heapdump"
	_lock "runtime/internal/lock"
	_sched "runtime/internal/sched"
	_schedinit "runtime/internal/schedinit"
	_sem "runtime/internal/sem"
	"unsafe"
)

var tmpbuf []byte

func dwritebyte(b byte) {
	_heapdump.Dwrite(unsafe.Pointer(&b), 1)
}

func flush() {
	_core.Write(_heapdump.Dumpfd, (unsafe.Pointer)(&_heapdump.Buf), int32(_heapdump.Nbuf))
	_heapdump.Nbuf = 0
}

// Cache of types that have been serialized already.
// We use a type's hash field to pick a bucket.
// Inside a bucket, we keep a list of types that
// have been serialized so far, most recently used first.
// Note: when a bucket overflows we may end up
// serializing a type more than once.  That's ok.
const (
	typeCacheBuckets = 256
	typeCacheAssoc   = 4
)

type typeCacheBucket struct {
	t [typeCacheAssoc]*_core.Type
}

var typecache [typeCacheBuckets]typeCacheBucket

func dumpbool(b bool) {
	if b {
		_heapdump.Dumpint(1)
	} else {
		_heapdump.Dumpint(0)
	}
}

func dumpslice(b []byte) {
	_heapdump.Dumpint(uint64(len(b)))
	if len(b) > 0 {
		_heapdump.Dwrite(unsafe.Pointer(&b[0]), uintptr(len(b)))
	}
}

// dump information for a type
func dumptype(t *_core.Type) {
	if t == nil {
		return
	}

	// If we've definitely serialized the type before,
	// no need to do it again.
	b := &typecache[t.Hash&(typeCacheBuckets-1)]
	if t == b.t[0] {
		return
	}
	for i := 1; i < typeCacheAssoc; i++ {
		if t == b.t[i] {
			// Move-to-front
			for j := i; j > 0; j-- {
				b.t[j] = b.t[j-1]
			}
			b.t[0] = t
			return
		}
	}

	// Might not have been dumped yet.  Dump it and
	// remember we did so.
	for j := typeCacheAssoc - 1; j > 0; j-- {
		b.t[j] = b.t[j-1]
	}
	b.t[0] = t

	// dump the type
	_heapdump.Dumpint(_heapdump.TagType)
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(t))))
	_heapdump.Dumpint(uint64(t.Size))
	if t.X == nil || t.X.Pkgpath == nil || t.X.Name == nil {
		_heapdump.Dumpstr(*t.String)
	} else {
		pkgpath := (*_lock.StringStruct)(unsafe.Pointer(&t.X.Pkgpath))
		name := (*_lock.StringStruct)(unsafe.Pointer(&t.X.Name))
		_heapdump.Dumpint(uint64(uintptr(pkgpath.Len) + 1 + uintptr(name.Len)))
		_heapdump.Dwrite(pkgpath.Str, uintptr(pkgpath.Len))
		dwritebyte('.')
		_heapdump.Dwrite(name.Str, uintptr(name.Len))
	}
	dumpbool(t.Kind&_channels.KindDirectIface == 0 || t.Kind&_channels.KindNoPointers == 0)
}

// dump an object
func dumpobj(obj unsafe.Pointer, size uintptr, bv _lock.Bitvector) {
	dumpbvtypes(&bv, obj)
	_heapdump.Dumpint(_heapdump.TagObject)
	_heapdump.Dumpint(uint64(uintptr(obj)))
	_heapdump.Dumpmemrange(obj, size)
	dumpfields(bv)
}

func dumpfinalizer(obj unsafe.Pointer, fn *_core.Funcval, fint *_core.Type, ot *_gc.Ptrtype) {
	_heapdump.Dumpint(_heapdump.TagFinalizer)
	_heapdump.Dumpint(uint64(uintptr(obj)))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(fn))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(fn.Fn))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(fint))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(ot))))
}

type childInfo struct {
	// Information passed up from the callee frame about
	// the layout of the outargs region.
	argoff uintptr         // where the arguments start in the frame
	arglen uintptr         // size of args region
	args   _lock.Bitvector // if args.n >= 0, pointer map of args region
	sp     *uint8          // callee sp
	depth  uintptr         // depth in call stack (0 == most recent)
}

// dump kinds & offsets of interesting fields in bv
func dumpbv(cbv *_lock.Bitvector, offset uintptr) {
	bv := _gc.Gobv(*cbv)
	for i := uintptr(0); i < uintptr(bv.N); i += _sched.TypeBitsWidth {
		switch bv.Bytedata[i/8] >> (i % 8) & _sched.TypeMask {
		default:
			_lock.Throw("unexpected pointer bits")
		case _sched.TypeDead:
			// typeDead has already been processed in makeheapobjbv.
			// We should only see it in stack maps, in which case we should continue processing.
		case _sched.TypeScalar:
			// ok
		case _sched.TypePointer:
			_heapdump.Dumpint(_heapdump.FieldKindPtr)
			_heapdump.Dumpint(uint64(offset + i/_sched.TypeBitsWidth*_core.PtrSize))
		}
	}
}

func dumpframe(s *_lock.Stkframe, arg unsafe.Pointer) bool {
	child := (*childInfo)(arg)
	f := s.Fn

	// Figure out what we can about our stack map
	pc := s.Pc
	if pc != f.Entry {
		pc--
	}
	pcdata := _gc.Pcdatavalue(f, _lock.PCDATA_StackMapIndex, pc)
	if pcdata == -1 {
		// We do not have a valid pcdata value but there might be a
		// stackmap for this function.  It is likely that we are looking
		// at the function prologue, assume so and hope for the best.
		pcdata = 0
	}
	stkmap := (*_gc.Stackmap)(_gc.Funcdata(f, _lock.FUNCDATA_LocalsPointerMaps))

	// Dump any types we will need to resolve Efaces.
	if child.args.N >= 0 {
		dumpbvtypes(&child.args, unsafe.Pointer(s.Sp+child.argoff))
	}
	var bv _lock.Bitvector
	if stkmap != nil && stkmap.N > 0 {
		bv = _gc.Stackmapdata(stkmap, pcdata)
		dumpbvtypes(&bv, unsafe.Pointer(s.Varp-uintptr(bv.N/_sched.TypeBitsWidth*_core.PtrSize)))
	} else {
		bv.N = -1
	}

	// Dump main body of stack frame.
	_heapdump.Dumpint(_heapdump.TagStackFrame)
	_heapdump.Dumpint(uint64(s.Sp))                              // lowest address in frame
	_heapdump.Dumpint(uint64(child.depth))                       // # of frames deep on the stack
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(child.sp)))) // sp of child, or 0 if bottom of stack
	_heapdump.Dumpmemrange(unsafe.Pointer(s.Sp), s.Fp-s.Sp)      // frame contents
	_heapdump.Dumpint(uint64(f.Entry))
	_heapdump.Dumpint(uint64(s.Pc))
	_heapdump.Dumpint(uint64(s.Continpc))
	name := _lock.Funcname(f)
	if name == "" {
		name = "unknown function"
	}
	_heapdump.Dumpstr(name)

	// Dump fields in the outargs section
	if child.args.N >= 0 {
		dumpbv(&child.args, child.argoff)
	} else {
		// conservative - everything might be a pointer
		for off := child.argoff; off < child.argoff+child.arglen; off += _core.PtrSize {
			_heapdump.Dumpint(_heapdump.FieldKindPtr)
			_heapdump.Dumpint(uint64(off))
		}
	}

	// Dump fields in the local vars section
	if stkmap == nil {
		// No locals information, dump everything.
		for off := child.arglen; off < s.Varp-s.Sp; off += _core.PtrSize {
			_heapdump.Dumpint(_heapdump.FieldKindPtr)
			_heapdump.Dumpint(uint64(off))
		}
	} else if stkmap.N < 0 {
		// Locals size information, dump just the locals.
		size := uintptr(-stkmap.N)
		for off := s.Varp - size - s.Sp; off < s.Varp-s.Sp; off += _core.PtrSize {
			_heapdump.Dumpint(_heapdump.FieldKindPtr)
			_heapdump.Dumpint(uint64(off))
		}
	} else if stkmap.N > 0 {
		// Locals bitmap information, scan just the pointers in
		// locals.
		dumpbv(&bv, s.Varp-uintptr(bv.N)/_sched.TypeBitsWidth*_core.PtrSize-s.Sp)
	}
	_heapdump.Dumpint(_heapdump.FieldKindEol)

	// Record arg info for parent.
	child.argoff = s.Argp - s.Fp
	child.arglen = s.Arglen
	child.sp = (*uint8)(unsafe.Pointer(s.Sp))
	child.depth++
	stkmap = (*_gc.Stackmap)(_gc.Funcdata(f, _lock.FUNCDATA_ArgsPointerMaps))
	if stkmap != nil {
		child.args = _gc.Stackmapdata(stkmap, pcdata)
	} else {
		child.args.N = -1
	}
	return true
}

func dumpgoroutine(gp *_core.G) {
	var sp, pc, lr uintptr
	if gp.Syscallsp != 0 {
		sp = gp.Syscallsp
		pc = gp.Syscallpc
		lr = 0
	} else {
		sp = gp.Sched.Sp
		pc = gp.Sched.Pc
		lr = gp.Sched.Lr
	}

	_heapdump.Dumpint(_heapdump.TagGoroutine)
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(gp))))
	_heapdump.Dumpint(uint64(sp))
	_heapdump.Dumpint(uint64(gp.Goid))
	_heapdump.Dumpint(uint64(gp.Gopc))
	_heapdump.Dumpint(uint64(_lock.Readgstatus(gp)))
	dumpbool(gp.Issystem)
	dumpbool(false) // isbackground
	_heapdump.Dumpint(uint64(gp.Waitsince))
	_heapdump.Dumpstr(gp.Waitreason)
	_heapdump.Dumpint(uint64(uintptr(gp.Sched.Ctxt)))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(gp.M))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(gp.Defer))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(gp.Panic))))

	// dump stack
	var child childInfo
	child.args.N = -1
	child.arglen = 0
	child.sp = nil
	child.depth = 0
	_lock.Gentraceback(pc, sp, lr, gp, 0, nil, 0x7fffffff, dumpframe, _core.Noescape(unsafe.Pointer(&child)), 0)

	// dump defer & panic records
	for d := gp.Defer; d != nil; d = d.Link {
		_heapdump.Dumpint(_heapdump.TagDefer)
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(d))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(gp))))
		_heapdump.Dumpint(uint64(d.Sp))
		_heapdump.Dumpint(uint64(d.Pc))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(d.Fn))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(d.Fn.Fn))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(d.Link))))
	}
	for p := gp.Panic; p != nil; p = p.Link {
		_heapdump.Dumpint(_heapdump.TagPanic)
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(p))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(gp))))
		eface := (*_core.Eface)(unsafe.Pointer(&p.Arg))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(eface.Type))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(eface.Data))))
		_heapdump.Dumpint(0) // was p->defer, no longer recorded
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(p.Link))))
	}
}

func dumpgs() {
	// goroutines & stacks
	for i := 0; uintptr(i) < _gc.Allglen; i++ {
		gp := _lock.Allgs[i]
		status := _lock.Readgstatus(gp) // The world is stopped so gp will not be in a scan state.
		switch status {
		default:
			print("runtime: unexpected G.status ", _core.Hex(status), "\n")
			_lock.Throw("dumpgs in STW - bad status")
		case _lock.Gdead:
			// ok
		case _lock.Grunnable,
			_lock.Gsyscall,
			_lock.Gwaiting:
			dumpgoroutine(gp)
		}
	}
}

func finq_callback(fn *_core.Funcval, obj unsafe.Pointer, nret uintptr, fint *_core.Type, ot *_gc.Ptrtype) {
	_heapdump.Dumpint(_heapdump.TagQueuedFinalizer)
	_heapdump.Dumpint(uint64(uintptr(obj)))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(fn))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(fn.Fn))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(fint))))
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(ot))))
}

func dumproots() {
	// data segment
	dumpbvtypes(&_gc.Gcdatamask, unsafe.Pointer(&_gc.Data))
	_heapdump.Dumpint(_heapdump.TagData)
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(&_gc.Data))))
	_heapdump.Dumpmemrange(unsafe.Pointer(&_gc.Data), uintptr(unsafe.Pointer(&_gc.Edata))-uintptr(unsafe.Pointer(&_gc.Data)))
	dumpfields(_gc.Gcdatamask)

	// bss segment
	dumpbvtypes(&_gc.Gcbssmask, unsafe.Pointer(&_gc.Bss))
	_heapdump.Dumpint(_heapdump.TagBSS)
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(&_gc.Bss))))
	_heapdump.Dumpmemrange(unsafe.Pointer(&_gc.Bss), uintptr(unsafe.Pointer(&_gc.Ebss))-uintptr(unsafe.Pointer(&_gc.Bss)))
	dumpfields(_gc.Gcbssmask)

	// MSpan.types
	allspans := _gc.H_allspans
	for spanidx := uint32(0); spanidx < _lock.Mheap_.Nspan; spanidx++ {
		s := allspans[spanidx]
		if s.State == _sched.MSpanInUse {
			// Finalizers
			for sp := s.Specials; sp != nil; sp = sp.Next {
				if sp.Kind != _gc.KindSpecialFinalizer {
					continue
				}
				spf := (*_gc.Specialfinalizer)(unsafe.Pointer(sp))
				p := unsafe.Pointer((uintptr(s.Start) << _core.PageShift) + uintptr(spf.Special.Offset))
				dumpfinalizer(p, spf.Fn, spf.Fint, spf.Ot)
			}
		}
	}

	// Finalizer queue
	iterate_finq(finq_callback)
}

// Bit vector of free marks.
// Needs to be as big as the largest number of objects per span.
var freemark [_core.PageSize / 8]bool

func dumpobjs() {
	for i := uintptr(0); i < uintptr(_lock.Mheap_.Nspan); i++ {
		s := _gc.H_allspans[i]
		if s.State != _sched.MSpanInUse {
			continue
		}
		p := uintptr(s.Start << _core.PageShift)
		size := s.Elemsize
		n := (s.Npages << _core.PageShift) / size
		if n > uintptr(len(freemark)) {
			_lock.Throw("freemark array doesn't have enough entries")
		}
		for l := s.Freelist; l.Ptr() != nil; l = l.Ptr().Next {
			freemark[(uintptr(l)-p)/size] = true
		}
		for j := uintptr(0); j < n; j, p = j+1, p+size {
			if freemark[j] {
				freemark[j] = false
				continue
			}
			dumpobj(unsafe.Pointer(p), size, makeheapobjbv(p, size))
		}
	}
}

func dumpparams() {
	_heapdump.Dumpint(_heapdump.TagParams)
	x := uintptr(1)
	if *(*byte)(unsafe.Pointer(&x)) == 1 {
		dumpbool(false) // little-endian ptrs
	} else {
		dumpbool(true) // big-endian ptrs
	}
	_heapdump.Dumpint(_core.PtrSize)
	_heapdump.Dumpint(uint64(_lock.Mheap_.Arena_start))
	_heapdump.Dumpint(uint64(_lock.Mheap_.Arena_used))
	_heapdump.Dumpint(_lock.Thechar)
	_heapdump.Dumpstr(_schedinit.Goexperiment)
	_heapdump.Dumpint(uint64(_lock.Ncpu))
}

func itab_callback(tab *_core.Itab) {
	t := tab.Type
	// Dump a map from itab* to the type of its data field.
	// We want this map so we can deduce types of interface referents.
	if t.Kind&_channels.KindDirectIface == 0 {
		// indirect - data slot is a pointer to t.
		dumptype(t.Ptrto)
		_heapdump.Dumpint(_heapdump.TagItab)
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(tab))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(t.Ptrto))))
	} else if t.Kind&_channels.KindNoPointers == 0 {
		// t is pointer-like - data slot is a t.
		dumptype(t)
		_heapdump.Dumpint(_heapdump.TagItab)
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(tab))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(t))))
	} else {
		// Data slot is a scalar.  Dump type just for fun.
		// With pointer-only interfaces, this shouldn't happen.
		dumptype(t)
		_heapdump.Dumpint(_heapdump.TagItab)
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(tab))))
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(t))))
	}
}

func dumpitabs() {
	iterate_itabs(itab_callback)
}

func dumpms() {
	for mp := _lock.Allm; mp != nil; mp = mp.Alllink {
		_heapdump.Dumpint(_heapdump.TagOSThread)
		_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(mp))))
		_heapdump.Dumpint(uint64(mp.Id))
		_heapdump.Dumpint(mp.Procid)
	}
}

func dumpmemstats() {
	_heapdump.Dumpint(_heapdump.TagMemStats)
	_heapdump.Dumpint(_lock.Memstats.Alloc)
	_heapdump.Dumpint(_lock.Memstats.Total_alloc)
	_heapdump.Dumpint(_lock.Memstats.Sys)
	_heapdump.Dumpint(_lock.Memstats.Nlookup)
	_heapdump.Dumpint(_lock.Memstats.Nmalloc)
	_heapdump.Dumpint(_lock.Memstats.Nfree)
	_heapdump.Dumpint(_lock.Memstats.Heap_alloc)
	_heapdump.Dumpint(_lock.Memstats.Heap_sys)
	_heapdump.Dumpint(_lock.Memstats.Heap_idle)
	_heapdump.Dumpint(_lock.Memstats.Heap_inuse)
	_heapdump.Dumpint(_lock.Memstats.Heap_released)
	_heapdump.Dumpint(_lock.Memstats.Heap_objects)
	_heapdump.Dumpint(_lock.Memstats.Stacks_inuse)
	_heapdump.Dumpint(_lock.Memstats.Stacks_sys)
	_heapdump.Dumpint(_lock.Memstats.Mspan_inuse)
	_heapdump.Dumpint(_lock.Memstats.Mspan_sys)
	_heapdump.Dumpint(_lock.Memstats.Mcache_inuse)
	_heapdump.Dumpint(_lock.Memstats.Mcache_sys)
	_heapdump.Dumpint(_lock.Memstats.Buckhash_sys)
	_heapdump.Dumpint(_lock.Memstats.Gc_sys)
	_heapdump.Dumpint(_lock.Memstats.Other_sys)
	_heapdump.Dumpint(_lock.Memstats.Next_gc)
	_heapdump.Dumpint(_lock.Memstats.Last_gc)
	_heapdump.Dumpint(_lock.Memstats.Pause_total_ns)
	for i := 0; i < 256; i++ {
		_heapdump.Dumpint(_lock.Memstats.Pause_ns[i])
	}
	_heapdump.Dumpint(uint64(_lock.Memstats.Numgc))
}

func dumpmemprof_callback(b *_sem.Bucket, nstk uintptr, pstk *uintptr, size, allocs, frees uintptr) {
	stk := (*[100000]uintptr)(unsafe.Pointer(pstk))
	_heapdump.Dumpint(_heapdump.TagMemProf)
	_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(b))))
	_heapdump.Dumpint(uint64(size))
	_heapdump.Dumpint(uint64(nstk))
	for i := uintptr(0); i < nstk; i++ {
		pc := stk[i]
		f := _lock.Findfunc(pc)
		if f == nil {
			var buf [64]byte
			n := len(buf)
			n--
			buf[n] = ')'
			if pc == 0 {
				n--
				buf[n] = '0'
			} else {
				for pc > 0 {
					n--
					buf[n] = "0123456789abcdef"[pc&15]
					pc >>= 4
				}
			}
			n--
			buf[n] = 'x'
			n--
			buf[n] = '0'
			n--
			buf[n] = '('
			dumpslice(buf[n:])
			_heapdump.Dumpstr("?")
			_heapdump.Dumpint(0)
		} else {
			_heapdump.Dumpstr(_lock.Funcname(f))
			if i > 0 && pc > f.Entry {
				pc--
			}
			file, line := _lock.Funcline(f, pc)
			_heapdump.Dumpstr(file)
			_heapdump.Dumpint(uint64(line))
		}
	}
	_heapdump.Dumpint(uint64(allocs))
	_heapdump.Dumpint(uint64(frees))
}

func dumpmemprof() {
	iterate_memprof(dumpmemprof_callback)
	allspans := _gc.H_allspans
	for spanidx := uint32(0); spanidx < _lock.Mheap_.Nspan; spanidx++ {
		s := allspans[spanidx]
		if s.State != _sched.MSpanInUse {
			continue
		}
		for sp := s.Specials; sp != nil; sp = sp.Next {
			if sp.Kind != _gc.KindSpecialProfile {
				continue
			}
			spp := (*_gc.Specialprofile)(unsafe.Pointer(sp))
			p := uintptr(s.Start<<_core.PageShift) + uintptr(spp.Special.Offset)
			_heapdump.Dumpint(_heapdump.TagAllocSample)
			_heapdump.Dumpint(uint64(p))
			_heapdump.Dumpint(uint64(uintptr(unsafe.Pointer(spp.B))))
		}
	}
}

var dumphdr = []byte("go1.4 heap dump\n")

func mdump() {
	// make sure we're done sweeping
	for i := uintptr(0); i < uintptr(_lock.Mheap_.Nspan); i++ {
		s := _gc.H_allspans[i]
		if s.State == _sched.MSpanInUse {
			_gc.MSpan_EnsureSwept(s)
		}
	}
	_core.Memclr(unsafe.Pointer(&typecache), unsafe.Sizeof(typecache))
	_heapdump.Dwrite(unsafe.Pointer(&dumphdr[0]), uintptr(len(dumphdr)))
	dumpparams()
	dumpitabs()
	dumpobjs()
	dumpgs()
	dumpms()
	dumproots()
	dumpmemstats()
	dumpmemprof()
	_heapdump.Dumpint(_heapdump.TagEOF)
	flush()
}

func writeheapdump_m(fd uintptr) {
	_g_ := _core.Getg()
	_sched.Casgstatus(_g_.M.Curg, _lock.Grunning, _lock.Gwaiting)
	_g_.Waitreason = "dumping heap"

	// Update stats so we can dump them.
	// As a side effect, flushes all the MCaches so the MSpan.freelist
	// lists contain all the free objects.
	_gc.Updatememstats(nil)

	// Set dump file.
	_heapdump.Dumpfd = fd

	// Call dump routine.
	mdump()

	// Reset dump file.
	_heapdump.Dumpfd = 0
	if tmpbuf != nil {
		_sched.SysFree(unsafe.Pointer(&tmpbuf[0]), uintptr(len(tmpbuf)), &_lock.Memstats.Other_sys)
		tmpbuf = nil
	}

	_sched.Casgstatus(_g_.M.Curg, _lock.Gwaiting, _lock.Grunning)
}

// dumpint() the kind & offset of each field in an object.
func dumpfields(bv _lock.Bitvector) {
	dumpbv(&bv, 0)
	_heapdump.Dumpint(_heapdump.FieldKindEol)
}

// The heap dump reader needs to be able to disambiguate
// Eface entries.  So it needs to know every type that might
// appear in such an entry.  The following routine accomplishes that.
// TODO(rsc, khr): Delete - no longer possible.

// Dump all the types that appear in the type field of
// any Eface described by this bit vector.
func dumpbvtypes(bv *_lock.Bitvector, base unsafe.Pointer) {
}

func makeheapobjbv(p uintptr, size uintptr) _lock.Bitvector {
	// Extend the temp buffer if necessary.
	nptr := size / _core.PtrSize
	if uintptr(len(tmpbuf)) < nptr*_sched.TypeBitsWidth/8+1 {
		if tmpbuf != nil {
			_sched.SysFree(unsafe.Pointer(&tmpbuf[0]), uintptr(len(tmpbuf)), &_lock.Memstats.Other_sys)
		}
		n := nptr*_sched.TypeBitsWidth/8 + 1
		p := _lock.SysAlloc(n, &_lock.Memstats.Other_sys)
		if p == nil {
			_lock.Throw("heapdump: out of memory")
		}
		tmpbuf = (*[1 << 30]byte)(p)[:n]
	}
	// Convert heap bitmap to type bitmap.
	i := uintptr(0)
	hbits := _sched.HeapBitsForAddr(p)
	for ; i < nptr; i++ {
		bits := hbits.TypeBits()
		if bits == _sched.TypeDead {
			break // end of object
		}
		hbits = hbits.Next()
		tmpbuf[i*_sched.TypeBitsWidth/8] &^= (_sched.TypeMask << ((i * _sched.TypeBitsWidth) % 8))
		tmpbuf[i*_sched.TypeBitsWidth/8] |= bits << ((i * _sched.TypeBitsWidth) % 8)
	}
	return _lock.Bitvector{int32(i * _sched.TypeBitsWidth), &tmpbuf[0]}
}
