// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

import (
	"unsafe"
)

//extern SigTabTT runtime·sigtab[];

var Sigset_all = ^uint32(0)

func unimplemented(name string) {
	println(name, "not implemented")
	*(*int)(unsafe.Pointer(uintptr(1231))) = 1231
}

//go:nosplit
func semawakeup(mp *M) {
	mach_semrelease(uint32(mp.Waitsema))
}

//go:nosplit
func Semacreate() uintptr {
	var x uintptr
	Systemstack(func() {
		x = uintptr(mach_semcreate())
	})
	return x
}

// May run with m.p==nil, so write barriers are not allowed.
//go:nowritebarrier
func newosproc(mp *M, stk unsafe.Pointer) {
	mp.tls[0] = uintptr(mp.Id) // so 386 asm can find it
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.G0, " id=", mp.Id, "/", int(mp.tls[0]), " ostk=", &mp, "\n")
	}

	var oset uint32
	Sigprocmask(SIG_SETMASK, &Sigset_all, &oset)
	errno := Bsdthread_create(stk, unsafe.Pointer(mp), FuncPC(Mstart))
	Sigprocmask(SIG_SETMASK, &oset, nil)

	if errno < 0 {
		print("runtime: failed to create new OS thread (have ", mcount(), " already; errno=", -errno, ")\n")
		Throw("runtime.newosproc")
	}
}

// Called to initialize a new m (including the bootstrap m).
// Called on the parent thread (main thread in case of bootstrap), can allocate memory.
func mpreinit(mp *M) {
	mp.Gsignal = Malg(32 * 1024) // OS X wants >= 8K
	mp.Gsignal.M = mp
}

func Msigsave(mp *M) {
	smask := (*uint32)(unsafe.Pointer(&mp.Sigmask))
	if unsafe.Sizeof(*smask) > unsafe.Sizeof(mp.Sigmask) {
		Throw("insufficient storage for signal mask")
	}
	Sigprocmask(SIG_SETMASK, nil, smask)
}

// Called to initialize a new m (including the bootstrap m).
// Called on the new thread, can not allocate memory.
func Minit() {
	// Initialize signal handling.
	_g_ := Getg()
	Signalstack(&_g_.M.Gsignal.Stack)

	// restore signal mask from m.sigmask and unblock essential signals
	nmask := *(*uint32)(unsafe.Pointer(&_g_.M.Sigmask))
	for i := range Sigtable {
		if Sigtable[i].Flags&SigUnblock != 0 {
			nmask &^= 1 << (uint32(i) - 1)
		}
	}
	Sigprocmask(SIG_SETMASK, &nmask, nil)
}

// Mach IPC, to get at semaphores
// Definitions are in /usr/include/mach on a Mac.

func Macherror(r int32, fn string) {
	print("mach error ", fn, ": ", r, "\n")
	Throw("mach error")
}

const _DebugMach = false

var zerondr machndr

func mach_msgh_bits(a, b uint32) uint32 {
	return a | b<<8
}

func mach_msg(h *Machheader, op int32, send_size, rcv_size, rcv_name, timeout, notify uint32) int32 {
	// TODO: Loop on interrupt.
	return mach_msg_trap(unsafe.Pointer(h), op, send_size, rcv_size, rcv_name, timeout, notify)
}

// Mach RPC (MIG)
const (
	_MinMachMsg = 48
	_MachReply  = 100
)

type codemsg struct {
	h    Machheader
	ndr  machndr
	code int32
}

func Machcall(h *Machheader, maxsize int32, rxsize int32) int32 {
	_g_ := Getg()
	port := _g_.M.machport
	if port == 0 {
		port = mach_reply_port()
		_g_.M.machport = port
	}

	h.Msgh_bits |= mach_msgh_bits(MACH_MSG_TYPE_COPY_SEND, MACH_MSG_TYPE_MAKE_SEND_ONCE)
	h.msgh_local_port = port
	h.msgh_reserved = 0
	id := h.Msgh_id

	if _DebugMach {
		p := (*[10000]unsafe.Pointer)(unsafe.Pointer(h))
		print("send:\t")
		var i uint32
		for i = 0; i < h.Msgh_size/uint32(unsafe.Sizeof(p[0])); i++ {
			print(" ", p[i])
			if i%8 == 7 {
				print("\n\t")
			}
		}
		if i%8 != 0 {
			print("\n")
		}
	}
	ret := mach_msg(h, MACH_SEND_MSG|MACH_RCV_MSG, h.Msgh_size, uint32(maxsize), port, 0, 0)
	if ret != 0 {
		if _DebugMach {
			print("mach_msg error ", ret, "\n")
		}
		return ret
	}
	if _DebugMach {
		p := (*[10000]unsafe.Pointer)(unsafe.Pointer(h))
		var i uint32
		for i = 0; i < h.Msgh_size/uint32(unsafe.Sizeof(p[0])); i++ {
			print(" ", p[i])
			if i%8 == 7 {
				print("\n\t")
			}
		}
		if i%8 != 0 {
			print("\n")
		}
	}
	if h.Msgh_id != id+_MachReply {
		if _DebugMach {
			print("mach_msg _MachReply id mismatch ", h.Msgh_id, " != ", id+_MachReply, "\n")
		}
		return -303 // MIG_REPLY_MISMATCH
	}
	// Look for a response giving the return value.
	// Any call can send this back with an error,
	// and some calls only have return values so they
	// send it back on success too.  I don't quite see how
	// you know it's one of these and not the full response
	// format, so just look if the message is right.
	c := (*codemsg)(unsafe.Pointer(h))
	if uintptr(h.Msgh_size) == unsafe.Sizeof(*c) && h.Msgh_bits&MACH_MSGH_BITS_COMPLEX == 0 {
		if _DebugMach {
			print("mig result ", c.code, "\n")
		}
		return c.code
	}
	if h.Msgh_size != uint32(rxsize) {
		if _DebugMach {
			print("mach_msg _MachReply size mismatch ", h.Msgh_size, " != ", rxsize, "\n")
		}
		return -307 // MIG_ARRAY_TOO_LARGE
	}
	return 0
}

// Semaphores!

const (
	Tmach_semcreate = 3418
	Rmach_semcreate = Tmach_semcreate + _MachReply

	Tmach_semdestroy = 3419
	Rmach_semdestroy = Tmach_semdestroy + _MachReply

	KERN_ABORTED             = 14
	KERN_OPERATION_TIMED_OUT = 49
)

type tmach_semcreatemsg struct {
	h      Machheader
	ndr    machndr
	policy int32
	value  int32
}

type rmach_semcreatemsg struct {
	h         Machheader
	body      Machbody
	semaphore Machport
}

func mach_semcreate() uint32 {
	var m [256]uint8
	tx := (*tmach_semcreatemsg)(unsafe.Pointer(&m))
	rx := (*rmach_semcreatemsg)(unsafe.Pointer(&m))

	tx.h.Msgh_bits = 0
	tx.h.Msgh_size = uint32(unsafe.Sizeof(*tx))
	tx.h.Msgh_remote_port = Mach_task_self()
	tx.h.Msgh_id = Tmach_semcreate
	tx.ndr = zerondr

	tx.policy = 0 // 0 = SYNC_POLICY_FIFO
	tx.value = 0

	for {
		r := Machcall(&tx.h, int32(unsafe.Sizeof(m)), int32(unsafe.Sizeof(*rx)))
		if r == 0 {
			break
		}
		if r == KERN_ABORTED { // interrupted
			continue
		}
		Macherror(r, "semaphore_create")
	}
	if rx.body.Msgh_descriptor_count != 1 {
		unimplemented("mach_semcreate desc count")
	}
	return rx.semaphore.Name
}

// The other calls have simple system call traps in sys_darwin_{amd64,386}.s

func mach_semaphore_wait(sema uint32) int32
func mach_semaphore_timedwait(sema, sec, nsec uint32) int32
func mach_semaphore_signal(sema uint32) int32

func semasleep1(ns int64) int32 {
	_g_ := Getg()

	if ns >= 0 {
		var nsecs int32
		secs := Timediv(ns, 1000000000, &nsecs)
		r := mach_semaphore_timedwait(uint32(_g_.M.Waitsema), uint32(secs), uint32(nsecs))
		if r == KERN_ABORTED || r == KERN_OPERATION_TIMED_OUT {
			return -1
		}
		if r != 0 {
			Macherror(r, "semaphore_wait")
		}
		return 0
	}

	for {
		r := mach_semaphore_wait(uint32(_g_.M.Waitsema))
		if r == 0 {
			break
		}
		if r == KERN_ABORTED { // interrupted
			continue
		}
		Macherror(r, "semaphore_wait")
	}
	return 0
}

//go:nosplit
func semasleep(ns int64) int32 {
	var r int32
	Systemstack(func() {
		r = semasleep1(ns)
	})
	return r
}

//go:nosplit
func mach_semrelease(sem uint32) {
	for {
		r := mach_semaphore_signal(sem)
		if r == 0 {
			break
		}
		if r == KERN_ABORTED { // interrupted
			continue
		}

		// mach_semrelease must be completely nosplit,
		// because it is called from Go code.
		// If we're going to die, start that process on the system stack
		// to avoid a Go stack split.
		Systemstack(func() { Macherror(r, "semaphore_signal") })
	}
}

//go:nosplit
func Osyield() {
	Usleep(1)
}

func Setsig(i int32, fn uintptr, restart bool) {
	var sa sigactiont
	Memclr(unsafe.Pointer(&sa), unsafe.Sizeof(sa))
	sa.sa_flags = SA_SIGINFO | SA_ONSTACK
	if restart {
		sa.sa_flags |= SA_RESTART
	}
	sa.sa_mask = ^uint32(0)
	sa.sa_tramp = unsafe.Pointer(FuncPC(sigtramp)) // runtime·sigtramp's job is to call into real handler
	*(*uintptr)(unsafe.Pointer(&sa.__sigaction_u)) = fn
	sigaction(uint32(i), &sa, nil)
}

func setsigstack(i int32) {
	Throw("setsigstack")
}

func Getsig(i int32) uintptr {
	var sa sigactiont
	Memclr(unsafe.Pointer(&sa), unsafe.Sizeof(sa))
	sigaction(uint32(i), nil, &sa)
	return *(*uintptr)(unsafe.Pointer(&sa.__sigaction_u))
}

func Signalstack(s *Stack) {
	var st stackt
	if s == nil {
		st.ss_flags = SS_DISABLE
	} else {
		st.ss_sp = (*byte)(unsafe.Pointer(s.Lo))
		st.ss_size = s.Hi - s.Lo
		st.ss_flags = 0
	}
	sigaltstack(&st, nil)
}

func Updatesigmask(m Sigmask) {
	Sigprocmask(SIG_SETMASK, &m[0], nil)
}
