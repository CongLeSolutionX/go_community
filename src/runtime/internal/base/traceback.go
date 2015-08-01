// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

// The code in this file implements stack trace walking for all architectures.
// The most important fact about a given architecture is whether it uses a link register.
// On systems with link registers, the prologue for a non-leaf function stores the
// incoming value of LR at the bottom of the newly allocated stack frame.
// On systems without link registers, the architecture pushes a return PC during
// the call instruction, so the return PC ends up above the stack frame.
// In this file, the return PC is always called LR, no matter how it was found.
//
// To date, the opposite of a link register architecture is an x86 architecture.
// This code may need to change if some other kind of non-link-register
// architecture comes along.
//
// The other important fact is the size of a pointer: on 32-bit systems the LR
// takes up only 4 bytes on the stack, while on 64-bit systems it takes up 8 bytes.
// Typically this is ptrSize.
//
// As an exception, amd64p32 has ptrSize == 4 but the CALL instruction still
// stores an 8-byte return PC onto the stack. To accommodate this, we use regSize
// as the size of the architecture-pushed return PC.
//
// usesLR is defined below. ptrSize and regSize are defined in stubs.go.

const UsesLR = GOARCH != "amd64" && GOARCH != "amd64p32" && GOARCH != "386"

var (
	// initialized in tracebackinit
	GoexitPC         uintptr
	JmpdeferPC       uintptr
	McallPC          uintptr
	MorestackPC      uintptr
	MstartPC         uintptr
	Rt0_goPC         uintptr
	SigpanicPC       uintptr
	RunfinqPC        uintptr
	BackgroundgcPC   uintptr
	BgsweepPC        uintptr
	ForcegchelperPC  uintptr
	TimerprocPC      uintptr
	GcBgMarkWorkerPC uintptr
	SystemstackPC    uintptr
	StackBarrierPC   uintptr

	GogoPC uintptr

	externalthreadhandlerp uintptr // initialized elsewhere
)

// Generic traceback.  Handles runtime stack prints (pcbuf == nil),
// the runtime.Callers function (pcbuf != nil), as well as the garbage
// collector (callback != nil).  A little clunky to merge these, but avoids
// duplicating the code and all its subtlety.
func Gentraceback(pc0, sp0, lr0 uintptr, gp *G, skip int, pcbuf *uintptr, max int, callback func(*Stkframe, unsafe.Pointer) bool, v unsafe.Pointer, flags uint) int {
	if GoexitPC == 0 {
		Throw("gentraceback before goexitPC initialization")
	}
	g := Getg()
	if g == gp && g == g.M.Curg {
		// The starting sp has been passed in as a uintptr, and the caller may
		// have other uintptr-typed stack references as well.
		// If during one of the calls that got us here or during one of the
		// callbacks below the stack must be grown, all these uintptr references
		// to the stack will not be updated, and gentraceback will continue
		// to inspect the old stack memory, which may no longer be valid.
		// Even if all the variables were updated correctly, it is not clear that
		// we want to expose a traceback that begins on one stack and ends
		// on another stack. That could confuse callers quite a bit.
		// Instead, we require that gentraceback and any other function that
		// accepts an sp for the current goroutine (typically obtained by
		// calling getcallersp) must not run on that goroutine's stack but
		// instead on the g0 stack.
		Throw("gentraceback cannot trace user goroutine on its own stack")
	}
	gotraceback := gotraceback(nil)

	// Fix up returns to the stack barrier by fetching the
	// original return PC from gp.stkbar.
	stkbar := gp.Stkbar[gp.StkbarPos:]

	if pc0 == ^uintptr(0) && sp0 == ^uintptr(0) { // Signal to fetch saved values from gp.
		if gp.Syscallsp != 0 {
			pc0 = gp.Syscallpc
			sp0 = gp.Syscallsp
			if UsesLR {
				lr0 = 0
			}
		} else {
			pc0 = gp.Sched.Pc
			sp0 = gp.Sched.Sp
			if UsesLR {
				lr0 = gp.Sched.Lr
			}
		}
	}

	nprint := 0
	var frame Stkframe
	frame.Pc = pc0
	frame.Sp = sp0
	if UsesLR {
		frame.Lr = lr0
	}
	waspanic := false
	printing := pcbuf == nil && callback == nil
	_defer := gp.Defer

	for _defer != nil && uintptr(_defer.Sp) == _NoArgs {
		_defer = _defer.Link
	}

	// If the PC is zero, it's likely a nil function call.
	// Start in the caller's frame.
	if frame.Pc == 0 {
		if UsesLR {
			frame.Pc = *(*uintptr)(unsafe.Pointer(frame.Sp))
			frame.Lr = 0
		} else {
			frame.Pc = uintptr(*(*Uintreg)(unsafe.Pointer(frame.Sp)))
			frame.Sp += RegSize
		}
	}

	f := Findfunc(frame.Pc)
	if f == nil {
		if callback != nil {
			print("runtime: unknown pc ", Hex(frame.Pc), "\n")
			Throw("unknown pc")
		}
		return 0
	}
	frame.Fn = f

	n := 0
	for n < max {
		// Typically:
		//	pc is the PC of the running function.
		//	sp is the stack pointer at that program counter.
		//	fp is the frame pointer (caller's stack pointer) at that program counter, or nil if unknown.
		//	stk is the stack containing sp.
		//	The caller's program counter is lr, unless lr is zero, in which case it is *(uintptr*)sp.
		f = frame.Fn

		// Found an actual function.
		// Derive frame pointer and link register.
		if frame.Fp == 0 {
			// We want to jump over the systemstack switch. If we're running on the
			// g0, this systemstack is at the top of the stack.
			// if we're not on g0 or there's a no curg, then this is a regular call.
			sp := frame.Sp
			if flags&_TraceJumpStack != 0 && f.Entry == SystemstackPC && gp == g.M.G0 && gp.M.Curg != nil {
				sp = gp.M.Curg.Sched.Sp
				stkbar = gp.M.Curg.Stkbar[gp.M.Curg.StkbarPos:]
			}
			frame.Fp = sp + uintptr(funcspdelta(f, frame.Pc))
			if !UsesLR {
				// On x86, call instruction pushes return PC before entering new function.
				frame.Fp += RegSize
			}
		}
		var flr *Func
		if topofstack(f) {
			frame.Lr = 0
			flr = nil
		} else if UsesLR && f.Entry == JmpdeferPC {
			// jmpdefer modifies SP/LR/PC non-atomically.
			// If a profiling interrupt arrives during jmpdefer,
			// the stack unwind may see a mismatched register set
			// and get confused. Stop if we see PC within jmpdefer
			// to avoid that confusion.
			// See golang.org/issue/8153.
			if callback != nil {
				Throw("traceback_arm: found jmpdefer when tracing with callback")
			}
			frame.Lr = 0
		} else {
			var lrPtr uintptr
			if UsesLR {
				if n == 0 && frame.Sp < frame.Fp || frame.Lr == 0 {
					lrPtr = frame.Sp
					frame.Lr = *(*uintptr)(unsafe.Pointer(lrPtr))
				}
			} else {
				if frame.Lr == 0 {
					lrPtr = frame.Fp - RegSize
					frame.Lr = uintptr(*(*Uintreg)(unsafe.Pointer(lrPtr)))
				}
			}
			if frame.Lr == StackBarrierPC {
				// Recover original PC.
				if stkbar[0].SavedLRPtr != lrPtr {
					print("found next stack barrier at ", Hex(lrPtr), "; expected ")
					GcPrintStkbars(stkbar)
					print("\n")
					Throw("missed stack barrier")
				}
				frame.Lr = stkbar[0].SavedLRVal
				stkbar = stkbar[1:]
			}
			flr = Findfunc(frame.Lr)
			if flr == nil {
				// This happens if you get a profiling interrupt at just the wrong time.
				// In that context it is okay to stop early.
				// But if callback is set, we're doing a garbage collection and must
				// get everything, so crash loudly.
				if callback != nil {
					print("runtime: unexpected return pc for ", Funcname(f), " called from ", Hex(frame.Lr), "\n")
					Throw("unknown caller pc")
				}
			}
		}

		frame.Varp = frame.Fp
		if !UsesLR {
			// On x86, call instruction pushes return PC before entering new function.
			frame.Varp -= RegSize
		}

		// If framepointer_enabled and there's a frame, then
		// there's a saved bp here.
		if Framepointer_enabled && GOARCH == "amd64" && frame.Varp > frame.Sp {
			frame.Varp -= RegSize
		}

		// Derive size of arguments.
		// Most functions have a fixed-size argument block,
		// so we can use metadata about the function f.
		// Not all, though: there are some variadic functions
		// in package runtime and reflect, and for those we use call-specific
		// metadata recorded by f's caller.
		if callback != nil || printing {
			frame.Argp = frame.Fp
			if UsesLR {
				frame.Argp += PtrSize
			}
			SetArgInfo(&frame, f, callback != nil)
		}

		// Determine frame's 'continuation PC', where it can continue.
		// Normally this is the return address on the stack, but if sigpanic
		// is immediately below this function on the stack, then the frame
		// stopped executing due to a trap, and frame.pc is probably not
		// a safe point for looking up liveness information. In this panicking case,
		// the function either doesn't return at all (if it has no defers or if the
		// defers do not recover) or it returns from one of the calls to
		// deferproc a second time (if the corresponding deferred func recovers).
		// It suffices to assume that the most recent deferproc is the one that
		// returns; everything live at earlier deferprocs is still live at that one.
		frame.Continpc = frame.Pc
		if waspanic {
			if _defer != nil && _defer.Sp == frame.Sp {
				frame.Continpc = _defer.Pc
			} else {
				frame.Continpc = 0
			}
		}

		// Unwind our local defer stack past this frame.
		for _defer != nil && (_defer.Sp == frame.Sp || _defer.Sp == _NoArgs) {
			_defer = _defer.Link
		}

		if skip > 0 {
			skip--
			goto skipped
		}

		if pcbuf != nil {
			(*[1 << 20]uintptr)(unsafe.Pointer(pcbuf))[n] = frame.Pc
		}
		if callback != nil {
			if !callback((*Stkframe)(Noescape(unsafe.Pointer(&frame))), v) {
				return n
			}
		}
		if printing {
			if (flags&_TraceRuntimeFrames) != 0 || showframe(f, gp) {
				// Print during crash.
				//	main(0x1, 0x2, 0x3)
				//		/home/rsc/go/src/runtime/x.go:23 +0xf
				//
				tracepc := frame.Pc // back up to CALL instruction for funcline.
				if (n > 0 || flags&_TraceTrap == 0) && frame.Pc > f.Entry && !waspanic {
					tracepc--
				}
				print(Funcname(f), "(")
				argp := (*[100]uintptr)(unsafe.Pointer(frame.Argp))
				for i := uintptr(0); i < frame.Arglen/PtrSize; i++ {
					if i >= 10 {
						print(", ...")
						break
					}
					if i != 0 {
						print(", ")
					}
					print(Hex(argp[i]))
				}
				print(")\n")
				file, line := Funcline(f, tracepc)
				print("\t", file, ":", line)
				if frame.Pc > f.Entry {
					print(" +", Hex(frame.Pc-f.Entry))
				}
				if g.M.Throwing > 0 && gp == g.M.Curg || gotraceback >= 2 {
					print(" fp=", Hex(frame.Fp), " sp=", Hex(frame.Sp))
				}
				print("\n")
				nprint++
			}
		}
		n++

	skipped:
		waspanic = f.Entry == SigpanicPC

		// Do not unwind past the bottom of the stack.
		if flr == nil {
			break
		}

		// Unwind to next frame.
		frame.Fn = flr
		frame.Pc = frame.Lr
		frame.Lr = 0
		frame.Sp = frame.Fp
		frame.Fp = 0
		frame.Argmap = nil

		// On link register architectures, sighandler saves the LR on stack
		// before faking a call to sigpanic.
		if UsesLR && waspanic {
			x := *(*uintptr)(unsafe.Pointer(frame.Sp))
			frame.Sp += PtrSize
			if GOARCH == "arm64" {
				// arm64 needs 16-byte aligned SP, always
				frame.Sp += PtrSize
			}
			f = Findfunc(frame.Pc)
			frame.Fn = f
			if f == nil {
				frame.Pc = x
			} else if funcspdelta(f, frame.Pc) == 0 {
				frame.Lr = x
			}
		}
	}

	if printing {
		n = nprint
	}

	// If callback != nil, we're being called to gather stack information during
	// garbage collection or stack growth. In that context, require that we used
	// up the entire defer stack. If not, then there is a bug somewhere and the
	// garbage collection or stack growth may not have seen the correct picture
	// of the stack. Crash now instead of silently executing the garbage collection
	// or stack copy incorrectly and setting up for a mysterious crash later.
	//
	// Note that panic != nil is okay here: there can be leftover panics,
	// because the defers on the panic stack do not nest in frame order as
	// they do on the defer stack. If you have:
	//
	//	frame 1 defers d1
	//	frame 2 defers d2
	//	frame 3 defers d3
	//	frame 4 panics
	//	frame 4's panic starts running defers
	//	frame 5, running d3, defers d4
	//	frame 5 panics
	//	frame 5's panic starts running defers
	//	frame 6, running d4, garbage collects
	//	frame 6, running d2, garbage collects
	//
	// During the execution of d4, the panic stack is d4 -> d3, which
	// is nested properly, and we'll treat frame 3 as resumable, because we
	// can find d3. (And in fact frame 3 is resumable. If d4 recovers
	// and frame 5 continues running, d3, d3 can recover and we'll
	// resume execution in (returning from) frame 3.)
	//
	// During the execution of d2, however, the panic stack is d2 -> d3,
	// which is inverted. The scan will match d2 to frame 2 but having
	// d2 on the stack until then means it will not match d3 to frame 3.
	// This is okay: if we're running d2, then all the defers after d2 have
	// completed and their corresponding frames are dead. Not finding d3
	// for frame 3 means we'll set frame 3's continpc == 0, which is correct
	// (frame 3 is dead). At the end of the walk the panic stack can thus
	// contain defers (d3 in this case) for dead frames. The inversion here
	// always indicates a dead frame, and the effect of the inversion on the
	// scan is to hide those dead frames, so the scan is still okay:
	// what's left on the panic stack are exactly (and only) the dead frames.
	//
	// We require callback != nil here because only when callback != nil
	// do we know that gentraceback is being called in a "must be correct"
	// context as opposed to a "best effort" context. The tracebacks with
	// callbacks only happen when everything is stopped nicely.
	// At other times, such as when gathering a stack for a profiling signal
	// or when printing a traceback during a crash, everything may not be
	// stopped nicely, and the stack walk may not be able to complete.
	// It's okay in those situations not to use up the entire defer stack:
	// incomplete information then is still better than nothing.
	if callback != nil && n < max && _defer != nil {
		if _defer != nil {
			print("runtime: g", gp.Goid, ": leftover defer sp=", Hex(_defer.Sp), " pc=", Hex(_defer.Pc), "\n")
		}
		for _defer = gp.Defer; _defer != nil; _defer = _defer.Link {
			print("\tdefer ", _defer, " sp=", Hex(_defer.Sp), " pc=", Hex(_defer.Pc), "\n")
		}
		Throw("traceback has leftover defers")
	}

	if callback != nil && n < max && len(stkbar) > 0 {
		print("runtime: g", gp.Goid, ": leftover stack barriers ")
		GcPrintStkbars(stkbar)
		print("\n")
		Throw("traceback has leftover stack barriers")
	}

	return n
}

func SetArgInfo(frame *Stkframe, f *Func, needArgMap bool) {
	frame.Arglen = uintptr(f.args)
	if needArgMap && f.args == ArgsSizeUnknown {
		// Extract argument bitmaps for reflect stubs from the calls they made to reflect.
		switch Funcname(f) {
		case "reflect.makeFuncStub", "reflect.methodValueCall":
			arg0 := frame.Sp
			if UsesLR {
				arg0 += PtrSize
			}
			fn := *(**[2]uintptr)(unsafe.Pointer(arg0))
			if fn[0] != f.Entry {
				print("runtime: confused by ", Funcname(f), "\n")
				Throw("reflect mismatch")
			}
			bv := (*Bitvector)(unsafe.Pointer(fn[1]))
			frame.Arglen = uintptr(bv.N * PtrSize)
			frame.Argmap = bv
		}
	}
}

func printcreatedby(gp *G) {
	// Show what created goroutine, except main goroutine (goid 1).
	pc := gp.Gopc
	f := Findfunc(pc)
	if f != nil && showframe(f, gp) && gp.Goid != 1 {
		print("created by ", Funcname(f), "\n")
		tracepc := pc // back up to CALL instruction for funcline.
		if pc > f.Entry {
			tracepc -= PCQuantum
		}
		file, line := Funcline(f, tracepc)
		print("\t", file, ":", line)
		if pc > f.Entry {
			print(" +", Hex(pc-f.Entry))
		}
		print("\n")
	}
}

func Traceback(pc, sp, lr uintptr, gp *G) {
	traceback1(pc, sp, lr, gp, 0)
}

// tracebacktrap is like traceback but expects that the PC and SP were obtained
// from a trap, not from gp->sched or gp->syscallpc/gp->syscallsp or getcallerpc/getcallersp.
// Because they are from a trap instead of from a saved pair,
// the initial PC must not be rewound to the previous instruction.
// (All the saved pairs record a PC that is a return address, so we
// rewind it into the CALL instruction.)
func tracebacktrap(pc, sp, lr uintptr, gp *G) {
	traceback1(pc, sp, lr, gp, _TraceTrap)
}

func traceback1(pc, sp, lr uintptr, gp *G, flags uint) {
	var n int
	if Readgstatus(gp)&^Gscan == Gsyscall {
		// Override registers if blocked in system call.
		pc = gp.Syscallpc
		sp = gp.Syscallsp
		flags &^= _TraceTrap
	}
	// Print traceback. By default, omits runtime frames.
	// If that means we print nothing at all, repeat forcing all frames printed.
	n = Gentraceback(pc, sp, lr, gp, 0, nil, _TracebackMaxFrames, nil, nil, flags)
	if n == 0 && (flags&_TraceRuntimeFrames) == 0 {
		n = Gentraceback(pc, sp, lr, gp, 0, nil, _TracebackMaxFrames, nil, nil, flags|_TraceRuntimeFrames)
	}
	if n == _TracebackMaxFrames {
		print("...additional frames elided...\n")
	}
	printcreatedby(gp)
}

func Callers(skip int, pcbuf []uintptr) int {
	sp := Getcallersp(unsafe.Pointer(&skip))
	pc := uintptr(Getcallerpc(unsafe.Pointer(&skip)))
	gp := Getg()
	var n int
	Systemstack(func() {
		n = Gentraceback(pc, sp, 0, gp, skip, &pcbuf[0], len(pcbuf), nil, nil, 0)
	})
	return n
}

func Gcallers(gp *G, skip int, pcbuf []uintptr) int {
	return Gentraceback(^uintptr(0), ^uintptr(0), 0, gp, skip, &pcbuf[0], len(pcbuf), nil, nil, 0)
}

func showframe(f *Func, gp *G) bool {
	g := Getg()
	if g.M.Throwing > 0 && gp != nil && (gp == g.M.Curg || gp == g.M.caughtsig.Ptr()) {
		return true
	}
	traceback := gotraceback(nil)
	name := Funcname(f)

	// Special case: always show runtime.panic frame, so that we can
	// see where a panic started in the middle of a stack trace.
	// See golang.org/issue/5832.
	if name == "runtime.panic" {
		return true
	}

	return traceback > 1 || f != nil && contains(name, ".") && (!hasprefix(name, "runtime.") || isExportedRuntime(name))
}

// isExportedRuntime reports whether name is an exported runtime function.
// It is only for runtime functions, so ASCII A-Z is fine.
func isExportedRuntime(name string) bool {
	const n = len("runtime.")
	return len(name) > n && name[:n] == "runtime." && 'A' <= name[n] && name[n] <= 'Z'
}

var gStatusStrings = [...]string{
	Gidle:      "idle",
	Grunnable:  "runnable",
	Grunning:   "running",
	Gsyscall:   "syscall",
	Gwaiting:   "waiting",
	Gdead:      "dead",
	Genqueue:   "enqueue",
	Gcopystack: "copystack",
}

func Goroutineheader(gp *G) {
	gpstatus := Readgstatus(gp)

	// Basic string status
	var status string
	if 0 <= gpstatus && gpstatus < uint32(len(gStatusStrings)) {
		status = gStatusStrings[gpstatus]
	} else if gpstatus&Gscan != 0 && 0 <= gpstatus&^Gscan && gpstatus&^Gscan < uint32(len(gStatusStrings)) {
		status = gStatusStrings[gpstatus&^Gscan]
	} else {
		status = "???"
	}

	// Override.
	if (gpstatus == Gwaiting || gpstatus == Gscanwaiting) && gp.Waitreason != "" {
		status = gp.Waitreason
	}

	// approx time the G is blocked, in minutes
	var waitfor int64
	gpstatus &^= Gscan // drop the scan bit
	if (gpstatus == Gwaiting || gpstatus == Gsyscall) && gp.Waitsince != 0 {
		waitfor = (Nanotime() - gp.Waitsince) / 60e9
	}
	print("goroutine ", gp.Goid, " [", status)
	if waitfor >= 1 {
		print(", ", waitfor, " minutes")
	}
	if gp.Lockedm != nil {
		print(", locked to thread")
	}
	print("]:\n")
}

func Tracebackothers(me *G) {
	level := gotraceback(nil)

	// Show the current goroutine first, if we haven't already.
	g := Getg()
	gp := g.M.Curg
	if gp != nil && gp != me {
		print("\n")
		Goroutineheader(gp)
		Traceback(^uintptr(0), ^uintptr(0), 0, gp)
	}

	Lock(&Allglock)
	for _, gp := range Allgs {
		if gp == me || gp == g.M.Curg || Readgstatus(gp) == Gdead || IsSystemGoroutine(gp) && level < 2 {
			continue
		}
		print("\n")
		Goroutineheader(gp)
		// Note: gp.m == g.m occurs when tracebackothers is
		// called from a signal handler initiated during a
		// systemstack call.  The original G is still in the
		// running state, and we want to print its stack.
		if gp.M != g.M && Readgstatus(gp)&^Gscan == Grunning {
			print("\tgoroutine running on other thread; stack unavailable\n")
			printcreatedby(gp)
		} else {
			Traceback(^uintptr(0), ^uintptr(0), 0, gp)
		}
	}
	Unlock(&Allglock)
}

// Does f mark the top of a goroutine stack?
func topofstack(f *Func) bool {
	pc := f.Entry
	return pc == GoexitPC ||
		pc == MstartPC ||
		pc == McallPC ||
		pc == MorestackPC ||
		pc == Rt0_goPC ||
		externalthreadhandlerp != 0 && pc == externalthreadhandlerp
}

// isSystemGoroutine reports whether the goroutine g must be omitted in
// stack dumps and deadlock detector.
func IsSystemGoroutine(gp *G) bool {
	pc := gp.Startpc
	return pc == RunfinqPC && !FingRunning ||
		pc == BackgroundgcPC ||
		pc == BgsweepPC ||
		pc == ForcegchelperPC ||
		pc == TimerprocPC ||
		pc == GcBgMarkWorkerPC
}
