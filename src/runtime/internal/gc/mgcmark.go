// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: marking and scanning

package gc

import (
	_base "runtime/internal/base"
	"unsafe"
)

// Scan all of the stacks, greying (or graying if in America) the referents
// but not blackening them since the mark write barrier isn't installed.
//go:nowritebarrier
func gcscan_m() {
	_g_ := _base.Getg()

	// Grab the g that called us and potentially allow rescheduling.
	// This allows it to be scanned like other goroutines.
	mastergp := _g_.M.Curg
	_base.Casgstatus(mastergp, _base.Grunning, _base.Gwaiting)
	mastergp.Waitreason = "garbage collection scan"

	// Span sweeping has been done by finishsweep_m.
	// Long term we will want to make this goroutine runnable
	// by placing it onto a scanenqueue state and then calling
	// runtime·restartg(mastergp) to make it Grunnable.
	// At the bottom we will want to return this p back to the scheduler.

	// Prepare flag indicating that the scan has not been completed.
	local_allglen := gcResetGState()

	_base.Work.Ndone = 0
	useOneP := uint32(1) // For now do not do this in parallel.
	//	ackgcphase is not needed since we are not scanning running goroutines.
	parforsetup(_base.Work.Markfor, useOneP, uint32(_base.RootCount+local_allglen), false, markroot)
	_base.Parfordo(_base.Work.Markfor)

	_base.Lock(&_base.Allglock)
	// Check that gc work is done.
	for i := 0; i < local_allglen; i++ {
		gp := _base.Allgs[i]
		if !gp.Gcscandone {
			_base.Throw("scan missed a g")
		}
	}
	_base.Unlock(&_base.Allglock)

	_base.Casgstatus(mastergp, _base.Gwaiting, _base.Grunning)
	// Let the g that called us continue to run.
}

// ptrmask for an allocation containing a single pointer.
var oneptrmask = [...]uint8{1}

//go:nowritebarrier
func markroot(desc *_base.Parfor, i uint32) {
	// TODO: Consider using getg().m.p.ptr().gcw.
	var gcw _base.GcWork

	// Note: if you add a case here, please also update heapdump.go:dumproots.
	switch i {
	case _base.RootData:
		for datap := &_base.Firstmoduledata; datap != nil; datap = datap.Next {
			scanblock(datap.Data, datap.Edata-datap.Data, datap.Gcdatamask.Bytedata, &gcw)
		}

	case _base.RootBss:
		for datap := &_base.Firstmoduledata; datap != nil; datap = datap.Next {
			scanblock(datap.Bss, datap.Ebss-datap.Bss, datap.Gcbssmask.Bytedata, &gcw)
		}

	case _base.RootFinalizers:
		for fb := Allfin; fb != nil; fb = fb.Alllink {
			scanblock(uintptr(unsafe.Pointer(&fb.Fin[0])), uintptr(fb.Cnt)*unsafe.Sizeof(fb.Fin[0]), &finptrmask[0], &gcw)
		}

	case _base.RootSpans:
		// mark MSpan.specials
		sg := _base.Mheap_.Sweepgen
		for spanidx := uint32(0); spanidx < uint32(len(_base.Work.Spans)); spanidx++ {
			s := _base.Work.Spans[spanidx]
			if s.State != _base.MSpanInUse {
				continue
			}
			if !_base.UseCheckmark && s.Sweepgen != sg {
				// sweepgen was updated (+2) during non-checkmark GC pass
				print("sweep ", s.Sweepgen, " ", sg, "\n")
				_base.Throw("gc: unswept span")
			}
			for sp := s.Specials; sp != nil; sp = sp.Next {
				if sp.Kind != KindSpecialFinalizer {
					continue
				}
				// don't mark finalized object, but scan it so we
				// retain everything it points to.
				spf := (*Specialfinalizer)(unsafe.Pointer(sp))
				// A finalizer can be set for an inner byte of an object, find object beginning.
				p := uintptr(s.Start<<_base.XPageShift) + uintptr(spf.Special.Offset)/s.Elemsize*s.Elemsize
				if _base.Gcphase != _base.GCscan {
					_base.Scanobject(p, &gcw) // scanned during mark termination
				}
				scanblock(uintptr(unsafe.Pointer(&spf.Fn)), _base.PtrSize, &oneptrmask[0], &gcw)
			}
		}

	case _base.RootFlushCaches:
		if _base.Gcphase != _base.GCscan { // Do not flush mcaches during GCscan phase.
			Flushallmcaches()
		}

	default:
		// the rest is scanning goroutine stacks
		if uintptr(i-_base.RootCount) >= _base.Allglen {
			_base.Throw("markroot: bad index")
		}
		gp := _base.Allgs[i-_base.RootCount]

		// remember when we've first observed the G blocked
		// needed only to output in traceback
		status := _base.Readgstatus(gp) // We are not in a scan state
		if (status == _base.Gwaiting || status == _base.Gsyscall) && gp.Waitsince == 0 {
			gp.Waitsince = _base.Work.Tstart
		}

		// Shrink a stack if not much of it is being used but not in the scan phase.
		if _base.Gcphase == _base.GCmarktermination {
			// Shrink during STW GCmarktermination phase thus avoiding
			// complications introduced by shrinking during
			// non-STW phases.
			shrinkstack(gp)
		}

		scang(gp)
	}

	gcw.Dispose()
}

//go:nowritebarrier
func Scanstack(gp *_base.G) {
	if gp.Gcscanvalid {
		if _base.Gcphase == _base.GCmarktermination {
			gcRemoveStackBarriers(gp)
		}
		return
	}

	if _base.Readgstatus(gp)&_base.Gscan == 0 {
		print("runtime:scanstack: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _base.Hex(_base.Readgstatus(gp)), "\n")
		_base.Throw("scanstack - bad status")
	}

	switch _base.Readgstatus(gp) &^ _base.Gscan {
	default:
		print("runtime: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _base.Readgstatus(gp), "\n")
		_base.Throw("mark - bad status")
	case _base.Gdead:
		return
	case _base.Grunning:
		print("runtime: gp=", gp, ", goid=", gp.Goid, ", gp->atomicstatus=", _base.Readgstatus(gp), "\n")
		_base.Throw("scanstack: goroutine not stopped")
	case _base.Grunnable, _base.Gsyscall, _base.Gwaiting:
		// ok
	}

	if gp == _base.Getg() {
		_base.Throw("can't scan our own stack")
	}
	mp := gp.M
	if mp != nil && mp.Helpgc != 0 {
		_base.Throw("can't scan gchelper stack")
	}

	var sp, barrierOffset, nextBarrier uintptr
	if gp.Syscallsp != 0 {
		sp = gp.Syscallsp
	} else {
		sp = gp.Sched.Sp
	}
	switch _base.Gcphase {
	case _base.GCscan:
		// Install stack barriers during stack scan.
		barrierOffset = _base.FirstStackBarrierOffset
		nextBarrier = sp + barrierOffset

		if _base.Debug.Gcstackbarrieroff > 0 {
			nextBarrier = ^uintptr(0)
		}

		if gp.StkbarPos != 0 || len(gp.Stkbar) != 0 {
			// If this happens, it's probably because we
			// scanned a stack twice in the same phase.
			print("stkbarPos=", gp.StkbarPos, " len(stkbar)=", len(gp.Stkbar), " goid=", gp.Goid, " gcphase=", _base.Gcphase, "\n")
			_base.Throw("g already has stack barriers")
		}

	case _base.GCmarktermination:
		if int(gp.StkbarPos) == len(gp.Stkbar) {
			// gp hit all of the stack barriers (or there
			// were none). Re-scan the whole stack.
			nextBarrier = ^uintptr(0)
		} else {
			// Only re-scan up to the lowest un-hit
			// barrier. Any frames above this have not
			// executed since the _GCscan scan of gp and
			// any writes through up-pointers to above
			// this barrier had write barriers.
			nextBarrier = gp.Stkbar[gp.StkbarPos].SavedLRPtr
			if _base.DebugStackBarrier {
				print("rescan below ", _base.Hex(nextBarrier), " in [", _base.Hex(sp), ",", _base.Hex(gp.Stack.Hi), ") goid=", gp.Goid, "\n")
			}
		}

		gcRemoveStackBarriers(gp)

	default:
		_base.Throw("scanstack in wrong phase")
	}

	gcw := &_base.Getg().M.P.Ptr().Gcw
	n := 0
	scanframe := func(frame *_base.Stkframe, unused unsafe.Pointer) bool {
		scanframeworker(frame, unused, gcw)

		if frame.Fp > nextBarrier {
			// We skip installing a barrier on bottom-most
			// frame because on LR machines this LR is not
			// on the stack.
			if _base.Gcphase == _base.GCscan && n != 0 {
				gcInstallStackBarrier(gp, frame)
				barrierOffset *= 2
				nextBarrier = sp + barrierOffset
			} else if _base.Gcphase == _base.GCmarktermination {
				// We just scanned a frame containing
				// a return to a stack barrier. Since
				// this frame never returned, we can
				// stop scanning.
				return false
			}
		}
		n++

		return true
	}
	_base.Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, 0x7fffffff, scanframe, nil, 0)
	tracebackdefers(gp, scanframe, nil)
	if _base.Gcphase == _base.GCmarktermination {
		gcw.Dispose()
	}
	gp.Gcscanvalid = true
}

// Scan a stack frame: local variables and function arguments/results.
//go:nowritebarrier
func scanframeworker(frame *_base.Stkframe, unused unsafe.Pointer, gcw *_base.GcWork) {

	f := frame.Fn
	targetpc := frame.Continpc
	if targetpc == 0 {
		// Frame is dead.
		return
	}
	if _base.DebugGC > 1 {
		print("scanframe ", _base.Funcname(f), "\n")
	}
	if targetpc != f.Entry {
		targetpc--
	}
	pcdata := Pcdatavalue(f, _base.PCDATA_StackMapIndex, targetpc)
	if pcdata == -1 {
		// We do not have a valid pcdata value but there might be a
		// stackmap for this function.  It is likely that we are looking
		// at the function prologue, assume so and hope for the best.
		pcdata = 0
	}

	// Scan local variables if stack frame has been allocated.
	size := frame.Varp - frame.Sp
	var minsize uintptr
	switch _base.Thechar {
	case '6', '8':
		minsize = 0
	case '7':
		minsize = SpAlign
	default:
		minsize = _base.PtrSize
	}
	if size > minsize {
		stkmap := (*Stackmap)(Funcdata(f, _base.FUNCDATA_LocalsPointerMaps))
		if stkmap == nil || stkmap.N <= 0 {
			print("runtime: frame ", _base.Funcname(f), " untyped locals ", _base.Hex(frame.Varp-size), "+", _base.Hex(size), "\n")
			_base.Throw("missing stackmap")
		}

		// Locals bitmap information, scan just the pointers in locals.
		if pcdata < 0 || pcdata >= stkmap.N {
			// don't know where we are
			print("runtime: pcdata is ", pcdata, " and ", stkmap.N, " locals stack map entries for ", _base.Funcname(f), " (targetpc=", targetpc, ")\n")
			_base.Throw("scanframe: bad symbol table")
		}
		bv := Stackmapdata(stkmap, pcdata)
		size = uintptr(bv.N) * _base.PtrSize
		scanblock(frame.Varp-size, size, bv.Bytedata, gcw)
	}

	// Scan arguments.
	if frame.Arglen > 0 {
		var bv _base.Bitvector
		if frame.Argmap != nil {
			bv = *frame.Argmap
		} else {
			stkmap := (*Stackmap)(Funcdata(f, _base.FUNCDATA_ArgsPointerMaps))
			if stkmap == nil || stkmap.N <= 0 {
				print("runtime: frame ", _base.Funcname(f), " untyped args ", _base.Hex(frame.Argp), "+", _base.Hex(frame.Arglen), "\n")
				_base.Throw("missing stackmap")
			}
			if pcdata < 0 || pcdata >= stkmap.N {
				// don't know where we are
				print("runtime: pcdata is ", pcdata, " and ", stkmap.N, " args stack map entries for ", _base.Funcname(f), " (targetpc=", targetpc, ")\n")
				_base.Throw("scanframe: bad symbol table")
			}
			bv = Stackmapdata(stkmap, pcdata)
		}
		scanblock(frame.Argp, uintptr(bv.N)*_base.PtrSize, bv.Bytedata, gcw)
	}
}

// gcInstallStackBarrier installs a stack barrier over the return PC of frame.
//go:nowritebarrier
func gcInstallStackBarrier(gp *_base.G, frame *_base.Stkframe) {
	if frame.Lr == 0 {
		if _base.DebugStackBarrier {
			print("not installing stack barrier with no LR, goid=", gp.Goid, "\n")
		}
		return
	}

	// Save the return PC and overwrite it with stackBarrier.
	var lrUintptr uintptr
	if _base.UsesLR {
		lrUintptr = frame.Sp
	} else {
		lrUintptr = frame.Fp - _base.RegSize
	}
	lrPtr := (*_base.Uintreg)(unsafe.Pointer(lrUintptr))
	if _base.DebugStackBarrier {
		print("install stack barrier at ", _base.Hex(lrUintptr), " over ", _base.Hex(*lrPtr), ", goid=", gp.Goid, "\n")
		if uintptr(*lrPtr) != frame.Lr {
			print("frame.lr=", _base.Hex(frame.Lr))
			_base.Throw("frame.lr differs from stack LR")
		}
	}

	gp.Stkbar = gp.Stkbar[:len(gp.Stkbar)+1]
	stkbar := &gp.Stkbar[len(gp.Stkbar)-1]
	stkbar.SavedLRPtr = lrUintptr
	stkbar.SavedLRVal = uintptr(*lrPtr)
	*lrPtr = _base.Uintreg(_base.StackBarrierPC)
}

// gcRemoveStackBarriers removes all stack barriers installed in gp's stack.
//go:nowritebarrier
func gcRemoveStackBarriers(gp *_base.G) {
	if _base.DebugStackBarrier && gp.StkbarPos != 0 {
		print("hit ", gp.StkbarPos, " stack barriers, goid=", gp.Goid, "\n")
	}

	// Remove stack barriers that we didn't hit.
	for _, stkbar := range gp.Stkbar[gp.StkbarPos:] {
		GcRemoveStackBarrier(gp, stkbar)
	}

	// Clear recorded stack barriers so copystack doesn't try to
	// adjust them.
	gp.StkbarPos = 0
	gp.Stkbar = gp.Stkbar[:0]
}

// gcRemoveStackBarrier removes a single stack barrier. It is the
// inverse operation of gcInstallStackBarrier.
//
// This is nosplit to ensure gp's stack does not move.
//
//go:nowritebarrier
//go:nosplit
func GcRemoveStackBarrier(gp *_base.G, stkbar _base.Stkbar) {
	if _base.DebugStackBarrier {
		print("remove stack barrier at ", _base.Hex(stkbar.SavedLRPtr), " with ", _base.Hex(stkbar.SavedLRVal), ", goid=", gp.Goid, "\n")
	}
	lrPtr := (*_base.Uintreg)(unsafe.Pointer(stkbar.SavedLRPtr))
	if val := *lrPtr; val != _base.Uintreg(_base.StackBarrierPC) {
		_base.Printlock()
		print("at *", _base.Hex(stkbar.SavedLRPtr), " expected stack barrier PC ", _base.Hex(_base.StackBarrierPC), ", found ", _base.Hex(val), ", goid=", gp.Goid, "\n")
		print("gp.stkbar=")
		_base.GcPrintStkbars(gp.Stkbar)
		print(", gp.stkbarPos=", gp.StkbarPos, ", gp.stack=[", _base.Hex(gp.Stack.Lo), ",", _base.Hex(gp.Stack.Hi), ")\n")
		_base.Throw("stack barrier lost")
	}
	*lrPtr = _base.Uintreg(stkbar.SavedLRVal)
}

// gcDrainUntilPreempt blackens grey objects until g.preempt is set.
// This is best-effort, so it will return as soon as it is unable to
// get work, even though there may be more work in the system.
//go:nowritebarrier
func gcDrainUntilPreempt(gcw *_base.GcWork, flushScanCredit int64) {
	if !_base.WriteBarrierEnabled {
		println("gcphase =", _base.Gcphase)
		_base.Throw("gcDrainUntilPreempt phase incorrect")
	}

	var lastScanFlush, nextScanFlush int64
	if flushScanCredit != -1 {
		lastScanFlush = gcw.ScanWork
		nextScanFlush = lastScanFlush + flushScanCredit
	} else {
		nextScanFlush = int64(^uint64(0) >> 1)
	}

	gp := _base.Getg()
	for !gp.Preempt {
		// If the work queue is empty, balance. During
		// concurrent mark we don't really know if anyone else
		// can make use of this work, but even if we're the
		// only worker, the total cost of this per cycle is
		// only O(_WorkbufSize) pointer copies.
		if _base.Work.Full == 0 && _base.Work.Partial == 0 {
			gcw.Balance()
		}

		b := gcw.TryGet()
		if b == 0 {
			// No more work
			break
		}
		_base.Scanobject(b, gcw)

		// Flush background scan work credit to the global
		// account if we've accumulated enough locally so
		// mutator assists can draw on it.
		if gcw.ScanWork >= nextScanFlush {
			credit := gcw.ScanWork - lastScanFlush
			_base.Xaddint64(&_base.GcController.BgScanCredit, credit)
			lastScanFlush = gcw.ScanWork
			nextScanFlush = lastScanFlush + flushScanCredit
		}
	}
	if flushScanCredit != -1 {
		credit := gcw.ScanWork - lastScanFlush
		_base.Xaddint64(&_base.GcController.BgScanCredit, credit)
	}
}

// scanblock scans b as scanobject would, but using an explicit
// pointer bitmap instead of the heap bitmap.
//
// This is used to scan non-heap roots, so it does not update
// gcw.bytesMarked or gcw.scanWork.
//
//go:nowritebarrier
func scanblock(b0, n0 uintptr, ptrmask *uint8, gcw *_base.GcWork) {
	// Use local copies of original parameters, so that a stack trace
	// due to one of the throws below shows the original block
	// base and extent.
	b := b0
	n := n0

	arena_start := _base.Mheap_.Arena_start
	arena_used := _base.Mheap_.Arena_used

	for i := uintptr(0); i < n; {
		// Find bits for the next word.
		bits := uint32(*Addb(ptrmask, i/(_base.PtrSize*8)))
		if bits == 0 {
			i += _base.PtrSize * 8
			continue
		}
		for j := 0; j < 8 && i < n; j++ {
			if bits&1 != 0 {
				// Same work as in scanobject; see comments there.
				obj := *(*uintptr)(unsafe.Pointer(b + i))
				if obj != 0 && arena_start <= obj && obj < arena_used {
					if obj, hbits, span := _base.HeapBitsForObject(obj); obj != 0 {
						_base.Greyobject(obj, b, i, hbits, span, gcw)
					}
				}
			}
			bits >>= 1
			i += _base.PtrSize
		}
	}
}

//go:nowritebarrier
func initCheckmarks() {
	_base.UseCheckmark = true
	for _, s := range _base.Work.Spans {
		if s.State == _base.XMSpanInUse {
			HeapBitsForSpan(s.Base()).InitCheckmarkSpan(s.Layout())
		}
	}
}

func clearCheckmarks() {
	_base.UseCheckmark = false
	for _, s := range _base.Work.Spans {
		if s.State == _base.XMSpanInUse {
			HeapBitsForSpan(s.Base()).ClearCheckmarkSpan(s.Layout())
		}
	}
}
