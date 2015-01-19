// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lock

import (
	_core "runtime/internal/core"
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

const usesLR = GOARCH != "amd64" && GOARCH != "amd64p32" && GOARCH != "386"

var (
	// initialized in tracebackinit
	GoexitPC    uintptr
	JmpdeferPC  uintptr
	McallPC     uintptr
	MorestackPC uintptr
	MstartPC    uintptr
	Rt0_goPC    uintptr
	SigpanicPC  uintptr

	externalthreadhandlerp uintptr // initialized elsewhere
)

// Generic traceback.  Handles runtime stack prints (pcbuf == nil),
// the runtime.Callers function (pcbuf != nil), as well as the garbage
// collector (callback != nil).  A little clunky to merge these, but avoids
// duplicating the code and all its subtlety.
func Gentraceback(pc0 uintptr, sp0 uintptr, lr0 uintptr, gp *_core.G, skip int, pcbuf *uintptr, max int, callback func(*Stkframe, unsafe.Pointer) bool, v unsafe.Pointer, flags uint) int {
	if GoexitPC == 0 {
		Throw("gentraceback before goexitPC initialization")
	}
	g := _core.Getg()
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
	gotraceback := Gotraceback(nil)
	if pc0 == ^uintptr(0) && sp0 == ^uintptr(0) { // Signal to fetch saved values from gp.
		if gp.Syscallsp != 0 {
			pc0 = gp.Syscallpc
			sp0 = gp.Syscallsp
			if usesLR {
				lr0 = 0
			}
		} else {
			pc0 = gp.Sched.Pc
			sp0 = gp.Sched.Sp
			if usesLR {
				lr0 = gp.Sched.Lr
			}
		}
	}

	nprint := 0
	var frame Stkframe
	frame.Pc = pc0
	frame.Sp = sp0
	if usesLR {
		frame.lr = lr0
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
		if usesLR {
			frame.Pc = *(*uintptr)(unsafe.Pointer(frame.Sp))
			frame.lr = 0
		} else {
			frame.Pc = uintptr(*(*_core.Uintreg)(unsafe.Pointer(frame.Sp)))
			frame.Sp += RegSize
		}
	}

	f := Findfunc(frame.Pc)
	if f == nil {
		if callback != nil {
			print("runtime: unknown pc ", _core.Hex(frame.Pc), "\n")
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
			frame.Fp = frame.Sp + uintptr(funcspdelta(f, frame.Pc))
			if !usesLR {
				// On x86, call instruction pushes return PC before entering new function.
				frame.Fp += RegSize
			}
		}
		var flr *Func
		if topofstack(f) {
			frame.lr = 0
			flr = nil
		} else if usesLR && f.Entry == JmpdeferPC {
			// jmpdefer modifies SP/LR/PC non-atomically.
			// If a profiling interrupt arrives during jmpdefer,
			// the stack unwind may see a mismatched register set
			// and get confused. Stop if we see PC within jmpdefer
			// to avoid that confusion.
			// See golang.org/issue/8153.
			if callback != nil {
				Throw("traceback_arm: found jmpdefer when tracing with callback")
			}
			frame.lr = 0
		} else {
			if usesLR {
				if n == 0 && frame.Sp < frame.Fp || frame.lr == 0 {
					frame.lr = *(*uintptr)(unsafe.Pointer(frame.Sp))
				}
			} else {
				if frame.lr == 0 {
					frame.lr = uintptr(*(*_core.Uintreg)(unsafe.Pointer(frame.Fp - RegSize)))
				}
			}
			flr = Findfunc(frame.lr)
			if flr == nil {
				// This happens if you get a profiling interrupt at just the wrong time.
				// In that context it is okay to stop early.
				// But if callback is set, we're doing a garbage collection and must
				// get everything, so crash loudly.
				if callback != nil {
					print("runtime: unexpected return pc for ", Gofuncname(f), " called from ", _core.Hex(frame.lr), "\n")
					Throw("unknown caller pc")
				}
			}
		}

		frame.Varp = frame.Fp
		if !usesLR {
			// On x86, call instruction pushes return PC before entering new function.
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
			if usesLR {
				frame.Argp += _core.PtrSize
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
			if !callback((*Stkframe)(_core.Noescape(unsafe.Pointer(&frame))), v) {
				return n
			}
		}
		if printing {
			if (flags&TraceRuntimeFrames) != 0 || showframe(f, gp) {
				// Print during crash.
				//	main(0x1, 0x2, 0x3)
				//		/home/rsc/go/src/runtime/x.go:23 +0xf
				//
				tracepc := frame.Pc // back up to CALL instruction for funcline.
				if (n > 0 || flags&TraceTrap == 0) && frame.Pc > f.Entry && !waspanic {
					tracepc--
				}
				print(Gofuncname(f), "(")
				argp := (*[100]uintptr)(unsafe.Pointer(frame.Argp))
				for i := uintptr(0); i < frame.Arglen/_core.PtrSize; i++ {
					if i >= 10 {
						print(", ...")
						break
					}
					if i != 0 {
						print(", ")
					}
					print(_core.Hex(argp[i]))
				}
				print(")\n")
				file, line := Funcline(f, tracepc)
				print("\t", file, ":", line)
				if frame.Pc > f.Entry {
					print(" +", _core.Hex(frame.Pc-f.Entry))
				}
				if g.M.Throwing > 0 && gp == g.M.Curg || gotraceback >= 2 {
					print(" fp=", _core.Hex(frame.Fp), " sp=", _core.Hex(frame.Sp))
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
		frame.Pc = frame.lr
		frame.lr = 0
		frame.Sp = frame.Fp
		frame.Fp = 0
		frame.Argmap = nil

		// On link register architectures, sighandler saves the LR on stack
		// before faking a call to sigpanic.
		if usesLR && waspanic {
			x := *(*uintptr)(unsafe.Pointer(frame.Sp))
			frame.Sp += _core.PtrSize
			f = Findfunc(frame.Pc)
			frame.Fn = f
			if f == nil {
				frame.Pc = x
			} else if f.frame == 0 {
				frame.lr = x
			}
		}
	}

	if pcbuf == nil && callback == nil {
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
			print("runtime: g", gp.Goid, ": leftover defer sp=", _core.Hex(_defer.Sp), " pc=", _core.Hex(_defer.Pc), "\n")
		}
		for _defer = gp.Defer; _defer != nil; _defer = _defer.Link {
			print("\tdefer ", _defer, " sp=", _core.Hex(_defer.Sp), " pc=", _core.Hex(_defer.Pc), "\n")
		}
		Throw("traceback has leftover defers")
	}

	return n
}

func SetArgInfo(frame *Stkframe, f *Func, needArgMap bool) {
	frame.Arglen = uintptr(f.args)
	if needArgMap && f.args == ArgsSizeUnknown {
		// Extract argument bitmaps for reflect stubs from the calls they made to reflect.
		switch Gofuncname(f) {
		case "reflect.makeFuncStub", "reflect.methodValueCall":
			arg0 := frame.Sp
			if usesLR {
				arg0 += _core.PtrSize
			}
			fn := *(**[2]uintptr)(unsafe.Pointer(arg0))
			if fn[0] != f.Entry {
				print("runtime: confused by ", Gofuncname(f), "\n")
				Throw("reflect mismatch")
			}
			bv := (*Bitvector)(unsafe.Pointer(fn[1]))
			frame.Arglen = uintptr(bv.N / 2 * _core.PtrSize)
			frame.Argmap = bv
		}
	}
}

func printcreatedby(gp *_core.G) {
	// Show what created goroutine, except main goroutine (goid 1).
	pc := gp.Gopc
	f := Findfunc(pc)
	if f != nil && showframe(f, gp) && gp.Goid != 1 {
		print("created by ", Gofuncname(f), "\n")
		tracepc := pc // back up to CALL instruction for funcline.
		if pc > f.Entry {
			tracepc -= PCQuantum
		}
		file, line := Funcline(f, tracepc)
		print("\t", file, ":", line)
		if pc > f.Entry {
			print(" +", _core.Hex(pc-f.Entry))
		}
		print("\n")
	}
}

func Traceback(pc uintptr, sp uintptr, lr uintptr, gp *_core.G) {
	Traceback1(pc, sp, lr, gp, 0)
}

func Traceback1(pc uintptr, sp uintptr, lr uintptr, gp *_core.G, flags uint) {
	var n int
	if Readgstatus(gp)&^Gscan == Gsyscall {
		// Override registers if blocked in system call.
		pc = gp.Syscallpc
		sp = gp.Syscallsp
		flags &^= TraceTrap
	}
	// Print traceback. By default, omits runtime frames.
	// If that means we print nothing at all, repeat forcing all frames printed.
	n = Gentraceback(pc, sp, lr, gp, 0, nil, _TracebackMaxFrames, nil, nil, flags)
	if n == 0 && (flags&TraceRuntimeFrames) == 0 {
		n = Gentraceback(pc, sp, lr, gp, 0, nil, _TracebackMaxFrames, nil, nil, flags|TraceRuntimeFrames)
	}
	if n == _TracebackMaxFrames {
		print("...additional frames elided...\n")
	}
	printcreatedby(gp)
}

func showframe(f *Func, gp *_core.G) bool {
	g := _core.Getg()
	if g.M.Throwing > 0 && gp != nil && (gp == g.M.Curg || gp == g.M.Caughtsig) {
		return true
	}
	traceback := Gotraceback(nil)
	name := Gostringnocopy(Funcname(f))

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

func Goroutineheader(gp *_core.G) {
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

func Tracebackothers(me *_core.G) {
	level := Gotraceback(nil)

	// Show the current goroutine first, if we haven't already.
	g := _core.Getg()
	gp := g.M.Curg
	if gp != nil && gp != me {
		print("\n")
		Goroutineheader(gp)
		Traceback(^uintptr(0), ^uintptr(0), 0, gp)
	}

	Lock(&Allglock)
	for _, gp := range Allgs {
		if gp == me || gp == g.M.Curg || Readgstatus(gp) == Gdead || gp.Issystem && level < 2 {
			continue
		}
		print("\n")
		Goroutineheader(gp)
		if Readgstatus(gp)&^Gscan == Grunning {
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
