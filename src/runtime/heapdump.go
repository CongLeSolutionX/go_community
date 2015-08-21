// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Implementation of runtime/debug.WriteHeapDump.  Writes all
// objects in the heap plus additional info (roots, threads,
// finalizers, etc.) to a file.

// The format of the dumped file is described at
// https://golang.org/s/go14heapdump.

package runtime

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
	_iface "runtime/internal/iface"
	_print "runtime/internal/print"
	"unsafe"
)

//go:linkname runtime_debug_WriteHeapDump runtime/debug.WriteHeapDump
func runtime_debug_WriteHeapDump(fd uintptr) {
	stopTheWorld("write heap dump")

	_base.Systemstack(func() {
		writeheapdump_m(fd)
	})

	startTheWorld()
}

const (
	fieldKindEol       = 0
	fieldKindPtr       = 1
	fieldKindIface     = 2
	fieldKindEface     = 3
	tagEOF             = 0
	tagObject          = 1
	tagOtherRoot       = 2
	tagType            = 3
	tagGoroutine       = 4
	tagStackFrame      = 5
	tagParams          = 6
	tagFinalizer       = 7
	tagItab            = 8
	tagOSThread        = 9
	tagMemStats        = 10
	tagQueuedFinalizer = 11
	tagData            = 12
	tagBSS             = 13
	tagDefer           = 14
	tagPanic           = 15
	tagMemProf         = 16
	tagAllocSample     = 17
)

var dumpfd uintptr // fd to write the dump to.
var tmpbuf []byte

// buffer of pending write data
const (
	bufSize = 4096
)

var buf [bufSize]byte
var nbuf uintptr

func dwrite(data unsafe.Pointer, len uintptr) {
	if len == 0 {
		return
	}
	if nbuf+len <= bufSize {
		copy(buf[nbuf:], (*[bufSize]byte)(data)[:len])
		nbuf += len
		return
	}

	_print.Write(dumpfd, (unsafe.Pointer)(&buf), int32(nbuf))
	if len >= bufSize {
		_print.Write(dumpfd, data, int32(len))
		nbuf = 0
	} else {
		copy(buf[:], (*[bufSize]byte)(data)[:len])
		nbuf = len
	}
}

func dwritebyte(b byte) {
	dwrite(unsafe.Pointer(&b), 1)
}

func flush() {
	_print.Write(dumpfd, (unsafe.Pointer)(&buf), int32(nbuf))
	nbuf = 0
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
	t [typeCacheAssoc]*_base.Type
}

var typecache [typeCacheBuckets]typeCacheBucket

// dump a uint64 in a varint format parseable by encoding/binary
func dumpint(v uint64) {
	var buf [10]byte
	var n int
	for v >= 0x80 {
		buf[n] = byte(v | 0x80)
		n++
		v >>= 7
	}
	buf[n] = byte(v)
	n++
	dwrite(unsafe.Pointer(&buf), uintptr(n))
}

func dumpbool(b bool) {
	if b {
		dumpint(1)
	} else {
		dumpint(0)
	}
}

// dump varint uint64 length followed by memory contents
func dumpmemrange(data unsafe.Pointer, len uintptr) {
	dumpint(uint64(len))
	dwrite(data, len)
}

func dumpslice(b []byte) {
	dumpint(uint64(len(b)))
	if len(b) > 0 {
		dwrite(unsafe.Pointer(&b[0]), uintptr(len(b)))
	}
}

func dumpstr(s string) {
	sp := (*_base.StringStruct)(unsafe.Pointer(&s))
	dumpmemrange(sp.Str, uintptr(sp.Len))
}

// dump information for a type
func dumptype(t *_base.Type) {
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
	dumpint(tagType)
	dumpint(uint64(uintptr(unsafe.Pointer(t))))
	dumpint(uint64(t.Size))
	if t.X == nil || t.X.Pkgpath == nil || t.X.Name == nil {
		dumpstr(*t.String)
	} else {
		pkgpath := (*_base.StringStruct)(unsafe.Pointer(&t.X.Pkgpath))
		name := (*_base.StringStruct)(unsafe.Pointer(&t.X.Name))
		dumpint(uint64(uintptr(pkgpath.Len) + 1 + uintptr(name.Len)))
		dwrite(pkgpath.Str, uintptr(pkgpath.Len))
		dwritebyte('.')
		dwrite(name.Str, uintptr(name.Len))
	}
	dumpbool(t.Kind&_iface.KindDirectIface == 0 || t.Kind&_iface.KindNoPointers == 0)
}

// dump an object
func dumpobj(obj unsafe.Pointer, size uintptr, bv _base.Bitvector) {
	dumpbvtypes(&bv, obj)
	dumpint(tagObject)
	dumpint(uint64(uintptr(obj)))
	dumpmemrange(obj, size)
	dumpfields(bv)
}

func dumpotherroot(description string, to unsafe.Pointer) {
	dumpint(tagOtherRoot)
	dumpstr(description)
	dumpint(uint64(uintptr(to)))
}

func dumpfinalizer(obj unsafe.Pointer, fn *_base.Funcval, fint *_base.Type, ot *_gc.Ptrtype) {
	dumpint(tagFinalizer)
	dumpint(uint64(uintptr(obj)))
	dumpint(uint64(uintptr(unsafe.Pointer(fn))))
	dumpint(uint64(uintptr(unsafe.Pointer(fn.Fn))))
	dumpint(uint64(uintptr(unsafe.Pointer(fint))))
	dumpint(uint64(uintptr(unsafe.Pointer(ot))))
}

type childInfo struct {
	// Information passed up from the callee frame about
	// the layout of the outargs region.
	argoff uintptr         // where the arguments start in the frame
	arglen uintptr         // size of args region
	args   _base.Bitvector // if args.n >= 0, pointer map of args region
	sp     *uint8          // callee sp
	depth  uintptr         // depth in call stack (0 == most recent)
}

// dump kinds & offsets of interesting fields in bv
func dumpbv(cbv *_base.Bitvector, offset uintptr) {
	bv := _gc.Gobv(*cbv)
	for i := uintptr(0); i < uintptr(bv.N); i++ {
		if bv.Bytedata[i/8]>>(i%8)&1 == 1 {
			dumpint(fieldKindPtr)
			dumpint(uint64(offset + i*_base.PtrSize))
		}
	}
}

func dumpframe(s *_base.Stkframe, arg unsafe.Pointer) bool {
	child := (*childInfo)(arg)
	f := s.Fn

	// Figure out what we can about our stack map
	pc := s.Pc
	if pc != f.Entry {
		pc--
	}
	pcdata := _gc.Pcdatavalue(f, _base.PCDATA_StackMapIndex, pc)
	if pcdata == -1 {
		// We do not have a valid pcdata value but there might be a
		// stackmap for this function.  It is likely that we are looking
		// at the function prologue, assume so and hope for the best.
		pcdata = 0
	}
	stkmap := (*_gc.Stackmap)(_gc.Funcdata(f, _base.FUNCDATA_LocalsPointerMaps))

	// Dump any types we will need to resolve Efaces.
	if child.args.N >= 0 {
		dumpbvtypes(&child.args, unsafe.Pointer(s.Sp+child.argoff))
	}
	var bv _base.Bitvector
	if stkmap != nil && stkmap.N > 0 {
		bv = _gc.Stackmapdata(stkmap, pcdata)
		dumpbvtypes(&bv, unsafe.Pointer(s.Varp-uintptr(bv.N*_base.PtrSize)))
	} else {
		bv.N = -1
	}

	// Dump main body of stack frame.
	dumpint(tagStackFrame)
	dumpint(uint64(s.Sp))                              // lowest address in frame
	dumpint(uint64(child.depth))                       // # of frames deep on the stack
	dumpint(uint64(uintptr(unsafe.Pointer(child.sp)))) // sp of child, or 0 if bottom of stack
	dumpmemrange(unsafe.Pointer(s.Sp), s.Fp-s.Sp)      // frame contents
	dumpint(uint64(f.Entry))
	dumpint(uint64(s.Pc))
	dumpint(uint64(s.Continpc))
	name := _base.Funcname(f)
	if name == "" {
		name = "unknown function"
	}
	dumpstr(name)

	// Dump fields in the outargs section
	if child.args.N >= 0 {
		dumpbv(&child.args, child.argoff)
	} else {
		// conservative - everything might be a pointer
		for off := child.argoff; off < child.argoff+child.arglen; off += _base.PtrSize {
			dumpint(fieldKindPtr)
			dumpint(uint64(off))
		}
	}

	// Dump fields in the local vars section
	if stkmap == nil {
		// No locals information, dump everything.
		for off := child.arglen; off < s.Varp-s.Sp; off += _base.PtrSize {
			dumpint(fieldKindPtr)
			dumpint(uint64(off))
		}
	} else if stkmap.N < 0 {
		// Locals size information, dump just the locals.
		size := uintptr(-stkmap.N)
		for off := s.Varp - size - s.Sp; off < s.Varp-s.Sp; off += _base.PtrSize {
			dumpint(fieldKindPtr)
			dumpint(uint64(off))
		}
	} else if stkmap.N > 0 {
		// Locals bitmap information, scan just the pointers in
		// locals.
		dumpbv(&bv, s.Varp-uintptr(bv.N)*_base.PtrSize-s.Sp)
	}
	dumpint(fieldKindEol)

	// Record arg info for parent.
	child.argoff = s.Argp - s.Fp
	child.arglen = s.Arglen
	child.sp = (*uint8)(unsafe.Pointer(s.Sp))
	child.depth++
	stkmap = (*_gc.Stackmap)(_gc.Funcdata(f, _base.FUNCDATA_ArgsPointerMaps))
	if stkmap != nil {
		child.args = _gc.Stackmapdata(stkmap, pcdata)
	} else {
		child.args.N = -1
	}
	return true
}

func dumpgoroutine(gp *_base.G) {
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

	dumpint(tagGoroutine)
	dumpint(uint64(uintptr(unsafe.Pointer(gp))))
	dumpint(uint64(sp))
	dumpint(uint64(gp.Goid))
	dumpint(uint64(gp.Gopc))
	dumpint(uint64(_base.Readgstatus(gp)))
	dumpbool(_base.IsSystemGoroutine(gp))
	dumpbool(false) // isbackground
	dumpint(uint64(gp.Waitsince))
	dumpstr(gp.Waitreason)
	dumpint(uint64(uintptr(gp.Sched.Ctxt)))
	dumpint(uint64(uintptr(unsafe.Pointer(gp.M))))
	dumpint(uint64(uintptr(unsafe.Pointer(gp.Defer))))
	dumpint(uint64(uintptr(unsafe.Pointer(gp.Panic))))

	// dump stack
	var child childInfo
	child.args.N = -1
	child.arglen = 0
	child.sp = nil
	child.depth = 0
	_base.Gentraceback(pc, sp, lr, gp, 0, nil, 0x7fffffff, dumpframe, _base.Noescape(unsafe.Pointer(&child)), 0)

	// dump defer & panic records
	for d := gp.Defer; d != nil; d = d.Link {
		dumpint(tagDefer)
		dumpint(uint64(uintptr(unsafe.Pointer(d))))
		dumpint(uint64(uintptr(unsafe.Pointer(gp))))
		dumpint(uint64(d.Sp))
		dumpint(uint64(d.Pc))
		dumpint(uint64(uintptr(unsafe.Pointer(d.Fn))))
		dumpint(uint64(uintptr(unsafe.Pointer(d.Fn.Fn))))
		dumpint(uint64(uintptr(unsafe.Pointer(d.Link))))
	}
	for p := gp.Panic; p != nil; p = p.Link {
		dumpint(tagPanic)
		dumpint(uint64(uintptr(unsafe.Pointer(p))))
		dumpint(uint64(uintptr(unsafe.Pointer(gp))))
		eface := (*_iface.Eface)(unsafe.Pointer(&p.Arg))
		dumpint(uint64(uintptr(unsafe.Pointer(eface.Type))))
		dumpint(uint64(uintptr(unsafe.Pointer(eface.Data))))
		dumpint(0) // was p->defer, no longer recorded
		dumpint(uint64(uintptr(unsafe.Pointer(p.Link))))
	}
}

func dumpgs() {
	// goroutines & stacks
	for i := 0; uintptr(i) < _base.Allglen; i++ {
		gp := _base.Allgs[i]
		status := _base.Readgstatus(gp) // The world is stopped so gp will not be in a scan state.
		switch status {
		default:
			print("runtime: unexpected G.status ", _base.Hex(status), "\n")
			_base.Throw("dumpgs in STW - bad status")
		case _base.Gdead:
			// ok
		case _base.Grunnable,
			_base.Gsyscall,
			_base.Gwaiting:
			dumpgoroutine(gp)
		}
	}
}

func finq_callback(fn *_base.Funcval, obj unsafe.Pointer, nret uintptr, fint *_base.Type, ot *_gc.Ptrtype) {
	dumpint(tagQueuedFinalizer)
	dumpint(uint64(uintptr(obj)))
	dumpint(uint64(uintptr(unsafe.Pointer(fn))))
	dumpint(uint64(uintptr(unsafe.Pointer(fn.Fn))))
	dumpint(uint64(uintptr(unsafe.Pointer(fint))))
	dumpint(uint64(uintptr(unsafe.Pointer(ot))))
}

func dumproots() {
	// TODO(mwhudson): dump datamask etc from all objects
	// data segment
	dumpbvtypes(&_base.Firstmoduledata.Gcdatamask, unsafe.Pointer(_base.Firstmoduledata.Data))
	dumpint(tagData)
	dumpint(uint64(_base.Firstmoduledata.Data))
	dumpmemrange(unsafe.Pointer(_base.Firstmoduledata.Data), _base.Firstmoduledata.Edata-_base.Firstmoduledata.Data)
	dumpfields(_base.Firstmoduledata.Gcdatamask)

	// bss segment
	dumpbvtypes(&_base.Firstmoduledata.Gcbssmask, unsafe.Pointer(_base.Firstmoduledata.Bss))
	dumpint(tagBSS)
	dumpint(uint64(_base.Firstmoduledata.Bss))
	dumpmemrange(unsafe.Pointer(_base.Firstmoduledata.Bss), _base.Firstmoduledata.Ebss-_base.Firstmoduledata.Bss)
	dumpfields(_base.Firstmoduledata.Gcbssmask)

	// MSpan.types
	allspans := _gc.H_allspans
	for spanidx := uint32(0); spanidx < _base.Mheap_.Nspan; spanidx++ {
		s := allspans[spanidx]
		if s.State == _base.XMSpanInUse {
			// Finalizers
			for sp := s.Specials; sp != nil; sp = sp.Next {
				if sp.Kind != _gc.KindSpecialFinalizer {
					continue
				}
				spf := (*_gc.Specialfinalizer)(unsafe.Pointer(sp))
				p := unsafe.Pointer((uintptr(s.Start) << _base.XPageShift) + uintptr(spf.Special.Offset))
				dumpfinalizer(p, spf.Fn, spf.Fint, spf.Ot)
			}
		}
	}

	// Finalizer queue
	iterate_finq(finq_callback)
}

// Bit vector of free marks.
// Needs to be as big as the largest number of objects per span.
var freemark [_base.XPageSize / 8]bool

func dumpobjs() {
	for i := uintptr(0); i < uintptr(_base.Mheap_.Nspan); i++ {
		s := _gc.H_allspans[i]
		if s.State != _base.XMSpanInUse {
			continue
		}
		p := uintptr(s.Start << _base.XPageShift)
		size := s.Elemsize
		n := (s.Npages << _base.XPageShift) / size
		if n > uintptr(len(freemark)) {
			_base.Throw("freemark array doesn't have enough entries")
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
	dumpint(tagParams)
	x := uintptr(1)
	if *(*byte)(unsafe.Pointer(&x)) == 1 {
		dumpbool(false) // little-endian ptrs
	} else {
		dumpbool(true) // big-endian ptrs
	}
	dumpint(_base.PtrSize)
	dumpint(uint64(_base.Mheap_.Arena_start))
	dumpint(uint64(_base.Mheap_.Arena_used))
	dumpint(_base.Thechar)
	dumpstr(goexperiment)
	dumpint(uint64(_base.Ncpu))
}

func itab_callback(tab *_iface.Itab) {
	t := tab.Type
	// Dump a map from itab* to the type of its data field.
	// We want this map so we can deduce types of interface referents.
	if t.Kind&_iface.KindDirectIface == 0 {
		// indirect - data slot is a pointer to t.
		dumptype(t.Ptrto)
		dumpint(tagItab)
		dumpint(uint64(uintptr(unsafe.Pointer(tab))))
		dumpint(uint64(uintptr(unsafe.Pointer(t.Ptrto))))
	} else if t.Kind&_iface.KindNoPointers == 0 {
		// t is pointer-like - data slot is a t.
		dumptype(t)
		dumpint(tagItab)
		dumpint(uint64(uintptr(unsafe.Pointer(tab))))
		dumpint(uint64(uintptr(unsafe.Pointer(t))))
	} else {
		// Data slot is a scalar.  Dump type just for fun.
		// With pointer-only interfaces, this shouldn't happen.
		dumptype(t)
		dumpint(tagItab)
		dumpint(uint64(uintptr(unsafe.Pointer(tab))))
		dumpint(uint64(uintptr(unsafe.Pointer(t))))
	}
}

func dumpitabs() {
	iterate_itabs(itab_callback)
}

func dumpms() {
	for mp := _base.Allm; mp != nil; mp = mp.Alllink {
		dumpint(tagOSThread)
		dumpint(uint64(uintptr(unsafe.Pointer(mp))))
		dumpint(uint64(mp.Id))
		dumpint(mp.Procid)
	}
}

func dumpmemstats() {
	dumpint(tagMemStats)
	dumpint(_base.Memstats.Alloc)
	dumpint(_base.Memstats.Total_alloc)
	dumpint(_base.Memstats.Sys)
	dumpint(_base.Memstats.Nlookup)
	dumpint(_base.Memstats.Nmalloc)
	dumpint(_base.Memstats.Nfree)
	dumpint(_base.Memstats.Heap_alloc)
	dumpint(_base.Memstats.Heap_sys)
	dumpint(_base.Memstats.Heap_idle)
	dumpint(_base.Memstats.Heap_inuse)
	dumpint(_base.Memstats.Heap_released)
	dumpint(_base.Memstats.Heap_objects)
	dumpint(_base.Memstats.Stacks_inuse)
	dumpint(_base.Memstats.Stacks_sys)
	dumpint(_base.Memstats.Mspan_inuse)
	dumpint(_base.Memstats.Mspan_sys)
	dumpint(_base.Memstats.Mcache_inuse)
	dumpint(_base.Memstats.Mcache_sys)
	dumpint(_base.Memstats.Buckhash_sys)
	dumpint(_base.Memstats.Gc_sys)
	dumpint(_base.Memstats.Other_sys)
	dumpint(_base.Memstats.Next_gc)
	dumpint(_base.Memstats.Last_gc)
	dumpint(_base.Memstats.Pause_total_ns)
	for i := 0; i < 256; i++ {
		dumpint(_base.Memstats.Pause_ns[i])
	}
	dumpint(uint64(_base.Memstats.Numgc))
}

func dumpmemprof_callback(b *_gc.Bucket, nstk uintptr, pstk *uintptr, size, allocs, frees uintptr) {
	stk := (*[100000]uintptr)(unsafe.Pointer(pstk))
	dumpint(tagMemProf)
	dumpint(uint64(uintptr(unsafe.Pointer(b))))
	dumpint(uint64(size))
	dumpint(uint64(nstk))
	for i := uintptr(0); i < nstk; i++ {
		pc := stk[i]
		f := _base.Findfunc(pc)
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
			dumpstr("?")
			dumpint(0)
		} else {
			dumpstr(_base.Funcname(f))
			if i > 0 && pc > f.Entry {
				pc--
			}
			file, line := _base.Funcline(f, pc)
			dumpstr(file)
			dumpint(uint64(line))
		}
	}
	dumpint(uint64(allocs))
	dumpint(uint64(frees))
}

func dumpmemprof() {
	iterate_memprof(dumpmemprof_callback)
	allspans := _gc.H_allspans
	for spanidx := uint32(0); spanidx < _base.Mheap_.Nspan; spanidx++ {
		s := allspans[spanidx]
		if s.State != _base.XMSpanInUse {
			continue
		}
		for sp := s.Specials; sp != nil; sp = sp.Next {
			if sp.Kind != _gc.KindSpecialProfile {
				continue
			}
			spp := (*_gc.Specialprofile)(unsafe.Pointer(sp))
			p := uintptr(s.Start<<_base.XPageShift) + uintptr(spp.Special.Offset)
			dumpint(tagAllocSample)
			dumpint(uint64(p))
			dumpint(uint64(uintptr(unsafe.Pointer(spp.B))))
		}
	}
}

var dumphdr = []byte("go1.5 heap dump\n")

func mdump() {
	// make sure we're done sweeping
	for i := uintptr(0); i < uintptr(_base.Mheap_.Nspan); i++ {
		s := _gc.H_allspans[i]
		if s.State == _base.XMSpanInUse {
			_gc.MSpan_EnsureSwept(s)
		}
	}
	_base.Memclr(unsafe.Pointer(&typecache), unsafe.Sizeof(typecache))
	dwrite(unsafe.Pointer(&dumphdr[0]), uintptr(len(dumphdr)))
	dumpparams()
	dumpitabs()
	dumpobjs()
	dumpgs()
	dumpms()
	dumproots()
	dumpmemstats()
	dumpmemprof()
	dumpint(tagEOF)
	flush()
}

func writeheapdump_m(fd uintptr) {
	_g_ := _base.Getg()
	_base.Casgstatus(_g_.M.Curg, _base.Grunning, _base.Gwaiting)
	_g_.Waitreason = "dumping heap"

	// Update stats so we can dump them.
	// As a side effect, flushes all the MCaches so the MSpan.freelist
	// lists contain all the free objects.
	updatememstats(nil)

	// Set dump file.
	dumpfd = fd

	// Call dump routine.
	mdump()

	// Reset dump file.
	dumpfd = 0
	if tmpbuf != nil {
		_base.SysFree(unsafe.Pointer(&tmpbuf[0]), uintptr(len(tmpbuf)), &_base.Memstats.Other_sys)
		tmpbuf = nil
	}

	_base.Casgstatus(_g_.M.Curg, _base.Gwaiting, _base.Grunning)
}

// dumpint() the kind & offset of each field in an object.
func dumpfields(bv _base.Bitvector) {
	dumpbv(&bv, 0)
	dumpint(fieldKindEol)
}

// The heap dump reader needs to be able to disambiguate
// Eface entries.  So it needs to know every type that might
// appear in such an entry.  The following routine accomplishes that.
// TODO(rsc, khr): Delete - no longer possible.

// Dump all the types that appear in the type field of
// any Eface described by this bit vector.
func dumpbvtypes(bv *_base.Bitvector, base unsafe.Pointer) {
}

func makeheapobjbv(p uintptr, size uintptr) _base.Bitvector {
	// Extend the temp buffer if necessary.
	nptr := size / _base.PtrSize
	if uintptr(len(tmpbuf)) < nptr/8+1 {
		if tmpbuf != nil {
			_base.SysFree(unsafe.Pointer(&tmpbuf[0]), uintptr(len(tmpbuf)), &_base.Memstats.Other_sys)
		}
		n := nptr/8 + 1
		p := _base.SysAlloc(n, &_base.Memstats.Other_sys)
		if p == nil {
			_base.Throw("heapdump: out of memory")
		}
		tmpbuf = (*[1 << 30]byte)(p)[:n]
	}
	// Convert heap bitmap to pointer bitmap.
	for i := uintptr(0); i < nptr/8+1; i++ {
		tmpbuf[i] = 0
	}
	i := uintptr(0)
	hbits := _base.HeapBitsForAddr(p)
	for ; i < nptr; i++ {
		if i >= 2 && !hbits.IsMarked() {
			break // end of object
		}
		if hbits.IsPointer() {
			tmpbuf[i/8] |= 1 << (i % 8)
		}
		hbits = hbits.Next()
	}
	return _base.Bitvector{int32(i), &tmpbuf[0]}
}
