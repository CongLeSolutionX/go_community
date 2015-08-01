// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Garbage collector: marking and scanning

package iface

import (
	_base "runtime/internal/base"
	_gc "runtime/internal/gc"
)

// gcAssistAlloc records and allocation of size bytes and, if
// allowAssist is true, may assist GC scanning in proportion to the
// allocations performed by this mutator since the last assist.
//
// It should only be called if gcAssistAlloc != 0.
//
// This must be called with preemption disabled.
//go:nowritebarrier
func gcAssistAlloc(size uintptr, allowAssist bool) {
	// Find the G responsible for this assist.
	gp := _base.Getg()
	if gp.M.Curg != nil {
		gp = gp.M.Curg
	}

	// Record allocation.
	gp.Gcalloc += size

	if !allowAssist {
		return
	}

	// Compute the amount of assist scan work we need to do.
	scanWork := int64(_base.GcController.AssistRatio*float64(gp.Gcalloc)) - gp.Gcscanwork
	// scanWork can be negative if the last assist scanned a large
	// object and we're still ahead of our assist goal.
	if scanWork <= 0 {
		return
	}

	// Steal as much credit as we can from the background GC's
	// scan credit. This is racy and may drop the background
	// credit below 0 if two mutators steal at the same time. This
	// will just cause steals to fail until credit is accumulated
	// again, so in the long run it doesn't really matter, but we
	// do have to handle the negative credit case.
	bgScanCredit := atomicloadint64(&_base.GcController.BgScanCredit)
	stolen := int64(0)
	if bgScanCredit > 0 {
		if bgScanCredit < scanWork {
			stolen = bgScanCredit
		} else {
			stolen = scanWork
		}
		_base.Xaddint64(&_base.GcController.BgScanCredit, -scanWork)

		scanWork -= stolen
		gp.Gcscanwork += stolen

		if scanWork == 0 {
			return
		}
	}

	// Perform assist work
	_base.Systemstack(func() {
		if _base.Atomicload(&_base.GcBlackenEnabled) == 0 {
			// The gcBlackenEnabled check in malloc races with the
			// store that clears it but an atomic check in every malloc
			// would be a performance hit.
			// Instead we recheck it here on the non-preemptable system
			// stack to determine if we should preform an assist.
			return
		}
		// Track time spent in this assist. Since we're on the
		// system stack, this is non-preemptible, so we can
		// just measure start and end time.
		startTime := _base.Nanotime()

		decnwait := _base.Xadd(&_base.Work.Nwait, -1)
		if decnwait == _base.Work.Nproc {
			println("runtime: work.nwait =", decnwait, "work.nproc=", _base.Work.Nproc)
			_base.Throw("nwait > work.nprocs")
		}

		// drain own cached work first in the hopes that it
		// will be more cache friendly.
		gcw := &_base.Getg().M.P.Ptr().Gcw
		startScanWork := gcw.ScanWork
		gcDrainN(gcw, scanWork)
		// Record that we did this much scan work.
		gp.Gcscanwork += gcw.ScanWork - startScanWork
		// If we are near the end of the mark phase
		// dispose of the gcw.
		if _base.GcBlackenPromptly {
			gcw.Dispose()
		}
		// If this is the last worker and we ran out of work,
		// signal a completion point.
		incnwait := _base.Xadd(&_base.Work.Nwait, +1)
		if incnwait > _base.Work.Nproc {
			println("runtime: work.nwait=", incnwait,
				"work.nproc=", _base.Work.Nproc,
				"gcBlackenPromptly=", _base.GcBlackenPromptly)
			_base.Throw("work.nwait > work.nproc")
		}

		if incnwait == _base.Work.Nproc && _base.Work.Full == 0 && _base.Work.Partial == 0 {
			// This has reached a background completion
			// point.
			if _base.GcBlackenPromptly {
				if _base.Work.BgMark1.Done == 0 {
					_base.Throw("completing mark 2, but bgMark1.done == 0")
				}
				_base.Work.BgMark2.Complete()
			} else {
				_base.Work.BgMark1.Complete()
			}
		}
		duration := _base.Nanotime() - startTime
		_p_ := gp.M.P.Ptr()
		_p_.GcAssistTime += duration
		if _p_.GcAssistTime > gcAssistTimeSlack {
			_base.Xaddint64(&_base.GcController.AssistTime, _p_.GcAssistTime)
			_p_.GcAssistTime = 0
		}
	})
}

// gcUnwindBarriers marks all stack barriers up the frame containing
// sp as hit and removes them. This is used during stack unwinding for
// panic/recover and by heapBitsBulkBarrier to force stack re-scanning
// when its destination is on the stack.
//
// This is nosplit to ensure gp's stack does not move.
//
//go:nosplit
func GcUnwindBarriers(gp *_base.G, sp uintptr) {
	// On LR machines, if there is a stack barrier on the return
	// from the frame containing sp, this will mark it as hit even
	// though it isn't, but it's okay to be conservative.
	before := gp.StkbarPos
	for int(gp.StkbarPos) < len(gp.Stkbar) && gp.Stkbar[gp.StkbarPos].SavedLRPtr < sp {
		_gc.GcRemoveStackBarrier(gp, gp.Stkbar[gp.StkbarPos])
		gp.StkbarPos++
	}
	if _base.DebugStackBarrier && gp.StkbarPos != before {
		print("skip barriers below ", _base.Hex(sp), " in goid=", gp.Goid, ": ")
		_base.GcPrintStkbars(gp.Stkbar[before:gp.StkbarPos])
		print("\n")
	}
}

// gcDrainN blackens grey objects until it has performed roughly
// scanWork units of scan work. This is best-effort, so it may perform
// less work if it fails to get a work buffer. Otherwise, it will
// perform at least n units of work, but may perform more because
// scanning is always done in whole object increments.
//go:nowritebarrier
func gcDrainN(gcw *_base.GcWork, scanWork int64) {
	targetScanWork := gcw.ScanWork + scanWork
	for gcw.ScanWork < targetScanWork {
		// This might be a good place to add prefetch code...
		// if(wbuf.nobj > 4) {
		//         PREFETCH(wbuf->obj[wbuf.nobj - 3];
		//  }
		b := gcw.TryGet()
		if b == 0 {
			return
		}
		_base.Scanobject(b, gcw)
	}
}

// If gcBlackenPromptly is true we are in the second mark phase phase so we allocate black.
//go:nowritebarrier
func gcmarknewobject_m(obj, size uintptr) {
	if _base.UseCheckmark && !_base.GcBlackenPromptly { // The world should be stopped so this should not happen.
		_base.Throw("gcmarknewobject called while doing checkmark")
	}
	_base.HeapBitsForAddr(obj).SetMarked()
	_base.Xadd64(&_base.Work.BytesMarked, int64(size))
}
