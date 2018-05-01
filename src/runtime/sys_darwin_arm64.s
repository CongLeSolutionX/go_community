// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// System calls and other sys.stuff for ARM64, Darwin
// See http://fxr.watson.org/fxr/source/bsd/kern/syscalls.c?v=xnu-1228
// or /usr/include/sys/syscall.h (on a Mac) for system call numbers.

#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"

// Copied from /usr/include/sys/syscall.h
#define	SYS_gettimeofday   116
#define	SYS_kill           37
#define	SYS_getpid         20
#define	SYS_pthread_sigmask 329
#define	SYS_setitimer      83
#define	SYS___sysctl       202
#define	SYS_sigaction      46
#define	SYS_sigreturn      184
#define	SYS_kqueue         362
#define	SYS_kevent         363
#define	SYS_fcntl          92

TEXT notok<>(SB),NOSPLIT,$0
	MOVD	$0, R8
	MOVD	R8, (R8)
	B	0(PC)

TEXT runtime·open(SB),NOSPLIT,$0-20
	MOVD	name+0(FP), R0
	MOVW	mode+8(FP), R1
	MOVW	perm+12(FP), R2
	BL libc_open(SB)
	MOVW	R0, ret+16(FP)
	RET

TEXT runtime·closefd(SB),NOSPLIT,$0-12
	MOVW	fd+0(FP), R0
	BL	libc_close(SB)
	MOVW	R0, ret+8(FP)
	RET

TEXT runtime·write(SB),NOSPLIT,$0-28
	MOVW	fd+0(FP), R0
	MOVD	p+8(FP), R1
	MOVW	n+16(FP), R2
	BL	libc_write(SB)
	MOVW	R0, ret+24(FP)
	RET

TEXT runtime·read(SB),NOSPLIT,$0-28
	MOVW	fd+0(FP), R0
	MOVD	p+8(FP), R1
	MOVW	n+16(FP), R2
	BL libc_read(SB)
	MOVW	R0, ret+24(FP)
	RET

TEXT runtime·exit(SB),NOSPLIT|NOFRAME,$0
	MOVW	code+0(FP), R0
	BL	libc_exit(SB)
	MOVD	$1234, R0
	MOVD	$1002, R1
	MOVD	R0, (R1)	// fail hard

// Not used on Darwin.
TEXT runtime·exitThread(SB),NOSPLIT,$0-8
	MOVD	$1234, R0
	MOVD	$1003, R1
	MOVD	R0, (R1)	// fail hard

TEXT runtime·raiseproc(SB),NOSPLIT,$0
	MOVW	$SYS_getpid, R16
	SVC	$0x80
	// arg 1 pid already in R0 from getpid
	MOVW	sig+0(FP), R1	// arg 2 - signal
	MOVW	$1, R2	// arg 3 - posix
	MOVW	$SYS_kill, R16
	SVC	$0x80
	RET

TEXT runtime·mmap(SB),NOSPLIT,$0-48
	MOVD	addr+0(FP), R0
	MOVD	n+8(FP), R1
	MOVW	prot+16(FP), R2
	MOVW	flags+20(FP), R3
	MOVW	fd+24(FP), R4
	MOVW	off+28(FP), R5
	BL	libc_mmap(SB)
	MOVD	$0, R1
	MOVD	$-1, R2
	CMP	R0, R2
	BNE	ok
	BL libc_error(SB)
	MOVD	(R0), R1
	MOVD	$0, R0
ok:
	MOVD	R0, p+32(FP)
	MOVD	R1, err+40(FP)
	RET

TEXT runtime·munmap(SB),NOSPLIT,$0-16
	MOVD	addr+0(FP), R0
	MOVD	n+8(FP), R1
	BL	libc_munmap(SB)
	CMP $0, R0
	BEQ 2(PC)
	BL	notok<>(SB)
	RET

TEXT runtime·madvise(SB),NOSPLIT,$0-20
	MOVD	addr+0(FP), R0
	MOVD	n+8(FP), R1
	MOVW	flags+16(FP), R2
	BL	libc_madvise(SB)
	RET

TEXT runtime·setitimer(SB),NOSPLIT,$0
	MOVW	mode+0(FP), R0
	MOVD	new+8(FP), R1
	MOVD	old+16(FP), R2
	MOVW	$SYS_setitimer, R16
	SVC	$0x80
	RET

TEXT runtime·walltime(SB),NOSPLIT,$40-12
	MOVD	RSP, R0	// timeval
	MOVD	R0, R9	// this is how dyld calls gettimeofday
	MOVW	$0, R1	// zone
	MOVD	$0, R2	// see issue 16570
	MOVW	$SYS_gettimeofday, R16
	SVC	$0x80	// Note: x0 is tv_sec, w1 is tv_usec
	CMP	$0, R0
	BNE	inreg
	MOVD	0(RSP), R0
	MOVW	8(RSP), R1
inreg:
	MOVD	R0, sec+0(FP)
	MOVW	$1000, R3
	MUL	R3, R1
	MOVW	R1, nsec+8(FP)
	RET

TEXT runtime·nanotime(SB),NOSPLIT,$40
	MOVD	RSP, R0	// timeval
	MOVD	R0, R9	// this is how dyld calls gettimeofday
	MOVW	$0, R1	// zone
	MOVD	$0, R2	// see issue 16570
	MOVW	$SYS_gettimeofday, R16
	SVC	$0x80	// Note: x0 is tv_sec, w1 is tv_usec
	CMP	$0, R0
	BNE	inreg
	MOVD	0(RSP), R0
	MOVW	8(RSP), R1
inreg:
	MOVW	$1000000000, R3
	MUL	R3, R0
	MOVW	$1000, R3
	MUL	R3, R1
	ADD	R1, R0

	MOVD	R0, ret+0(FP)
	RET

TEXT runtime·sigfwd(SB),NOSPLIT,$0-32
	MOVW	sig+8(FP), R0
	MOVD	info+16(FP), R1
	MOVD	ctx+24(FP), R2
	MOVD	fn+0(FP), R11
	BL	(R11)
	RET

// Sigtramp's job is to call the actual signal handler.
// It is called with the following arguments on the stack:
//	LR	"return address" - ignored
//	R0	actual handler
//	R1	siginfo style - ignored
//	R2	signal number
//	R3	siginfo
//	R4	context
TEXT runtime·sigtramp(SB),NOSPLIT,$0
	// this might be called in external code context,
	// where g is not set.
	// first save R0, because runtime·load_g will clobber it
	MOVD.W	R0, -16(RSP)	// note: stack must be 16-byte aligned
	MOVB	runtime·iscgo(SB), R0
	CMP	$0, R0
	BEQ	2(PC)
	BL	runtime·load_g(SB)

	CMP	$0, g
	BNE	cont
	// fake function call stack frame for badsignal
	// we only need to pass R2 (signal number), but
	// badsignal will expect R2 at 8(RSP), so we also
	// push R1 onto stack. turns out we do need R1
	// to do sigreturn.
	MOVD.W	R1, -16(RSP)
	MOVD	R2, 8(RSP)
	MOVD	R4, 24(RSP)	// save ucontext, badsignal might clobber R4
	MOVD	$runtime·badsignal(SB), R26
	BL	(R26)
	MOVD	0(RSP), R1	// saved infostype
	MOVD	24(RSP), R0	// the ucontext
	ADD	$(16+16), RSP
	B	ret

cont:
	// Restore R0
	MOVD.P	16(RSP), R0

	// NOTE: some Darwin/ARM kernels always use the main stack to run the
	// signal handler. We need to switch to gsignal ourselves.
	MOVD	g_m(g), R11
	MOVD	m_gsignal(R11), R5
	MOVD	(g_stack+stack_hi)(R5), R6
	SUB	$64, R6

	// copy arguments for call to sighandler
	MOVD	R2, 8(R6)	// signal num
	MOVD	R3, 16(R6)	// signal info
	MOVD	R4, 24(R6)	// context
	MOVD	g, 32(R6)	// old_g

	// Backup ucontext and infostyle
	MOVD	R4, 40(R6)
	MOVD	R1, 48(R6)

	// switch stack and g
	MOVD	R6, RSP	// sigtramp is not re-entrant, so no need to back up RSP.
	MOVD	R5, g

	BL	(R0)

	// call sigreturn
	MOVD	40(RSP), R0	// saved ucontext
	MOVD	48(RSP), R1	// saved infostyle
ret:
	MOVW	$SYS_sigreturn, R16 // sigreturn(ucontext, infostyle)
	SVC	$0x80

	// if sigreturn fails, we can do nothing but exit
	B	runtime·exit(SB)

TEXT runtime·sigprocmask(SB),NOSPLIT,$0
	MOVW	how+0(FP), R0
	MOVD	new+8(FP), R1
	MOVD	old+16(FP), R2
	MOVW	$SYS_pthread_sigmask, R16
	SVC	$0x80
	BCC	2(PC)
	BL	notok<>(SB)
	RET

TEXT runtime·sigaction(SB),NOSPLIT,$0
	MOVW	mode+0(FP), R0
	MOVD	new+8(FP), R1
	MOVD	old+16(FP), R2
	MOVW	$SYS_sigaction, R16
	SVC	$0x80
	BCC	2(PC)
	BL	notok<>(SB)
	RET

TEXT runtime·usleep(SB),NOSPLIT,$0-4
	MOVW	usec+0(FP), R0
	BL	libc_usleep(SB)
	RET

TEXT runtime·sysctl(SB),NOSPLIT,$0
	MOVD	mib+0(FP), R0
	MOVW	miblen+8(FP), R1
	MOVD	out+16(FP), R2
	MOVD	size+24(FP), R3
	MOVD	dst+32(FP), R4
	MOVD	ndst+40(FP), R5
	MOVW	$SYS___sysctl, R16
	SVC	$0x80
	BCC	ok
	NEG	R0, R0
	MOVW	R0, ret+48(FP)
	RET
ok:
	MOVW	$0, R0
	MOVW	R0, ret+48(FP)
	RET

// uint32 mach_msg_trap(void*, uint32, uint32, uint32, uint32, uint32, uint32)
TEXT runtime·mach_msg_trap(SB),NOSPLIT,$0
	MOVD	h+0(FP), R0
	MOVW	op+8(FP), R1
	MOVW	send_size+12(FP), R2
	MOVW	rcv_size+16(FP), R3
	MOVW	rcv_name+20(FP), R4
	MOVW	timeout+24(FP), R5
	MOVW	notify+28(FP), R6
	MOVN	$30, R16
	SVC	$0x80
	MOVW	R0, ret+32(FP)
	RET

TEXT runtime·mach_task_self(SB),NOSPLIT,$0
	MOVN	$27, R16 // task_self_trap
	SVC	$0x80
	MOVW	R0, ret+0(FP)
	RET

TEXT runtime·mach_thread_self(SB),NOSPLIT,$0
	MOVN	$26, R16 // thread_self_trap
	SVC	$0x80
	MOVW	R0, ret+0(FP)
	RET

TEXT runtime·mach_reply_port(SB),NOSPLIT,$0
	MOVN	$25, R16	// mach_reply_port
	SVC	$0x80
	MOVW	R0, ret+0(FP)
	RET

// Mach provides trap versions of the semaphore ops,
// instead of requiring the use of RPC.

// uint32 mach_semaphore_wait(uint32)
TEXT runtime·mach_semaphore_wait(SB),NOSPLIT,$0
	MOVW	sema+0(FP), R0
	MOVN	$35, R16	// semaphore_wait_trap
	SVC	$0x80
	MOVW	R0, ret+8(FP)
	RET

// uint32 mach_semaphore_timedwait(uint32, uint32, uint32)
TEXT runtime·mach_semaphore_timedwait(SB),NOSPLIT,$0
	MOVW	sema+0(FP), R0
	MOVW	sec+4(FP), R1
	MOVW	nsec+8(FP), R2
	MOVN	$37, R16	// semaphore_timedwait_trap
	SVC	$0x80
	MOVW	R0, ret+16(FP)
	RET

// uint32 mach_semaphore_signal(uint32)
TEXT runtime·mach_semaphore_signal(SB),NOSPLIT,$0
	MOVW	sema+0(FP), R0
	MOVN	$32, R16	// semaphore_signal_trap
	SVC	$0x80
	MOVW	R0, ret+8(FP)
	RET

// uint32 mach_semaphore_signal_all(uint32)
TEXT runtime·mach_semaphore_signal_all(SB),NOSPLIT,$0
	MOVW	sema+0(FP), R0
	MOVN	$33, R16	// semaphore_signal_all_trap
	SVC	$0x80
	MOVW	R0, ret+8(FP)
	RET

// int32 runtime·kqueue(void)
TEXT runtime·kqueue(SB),NOSPLIT,$0
	MOVW	$SYS_kqueue, R16
	SVC	$0x80
	BCC	2(PC)
	NEG	R0, R0
	MOVW	R0, ret+0(FP)
	RET

// int32 runtime·kevent(int kq, Kevent *ch, int nch, Kevent *ev, int nev, Timespec *ts)
TEXT runtime·kevent(SB),NOSPLIT,$0
	MOVW	kq+0(FP), R0
	MOVD	ch+8(FP), R1
	MOVW	nch+16(FP), R2
	MOVD	ev+24(FP), R3
	MOVW	nev+32(FP), R4
	MOVD	ts+40(FP), R5
	MOVW	$SYS_kevent, R16
	SVC	$0x80
	BCC	2(PC)
	NEG	R0, R0
	MOVW	R0, ret+48(FP)
	RET

// int32 runtime·closeonexec(int32 fd)
TEXT runtime·closeonexec(SB),NOSPLIT,$0
	MOVW	fd+0(FP), R0
	MOVW	$2, R1	// F_SETFD
	MOVW	$1, R2	// FD_CLOEXEC
	MOVW	$SYS_fcntl, R16
	SVC	$0x80
	RET

// sigaltstack on some darwin/arm version is buggy and will always
// run the signal handler on the main stack, so our sigtramp has
// to do the stack switch ourselves.
TEXT runtime·sigaltstack(SB),NOSPLIT,$0
	RET

// Thread related functions
// Note: On darwin/arm64, the runtime always use runtime/cgo to
// create threads, so all thread related functions will just exit with a
// unique status.

TEXT runtime·mstart_stub(SB),NOSPLIT,$0
	MOVW	$44, R0
	BL	libc_exit(SB)
	RET

TEXT runtime·pthread_attr_init_trampoline(SB),NOSPLIT,$0-12
	MOVW	$45, R0
	BL	libc_exit(SB)
	MOVW	$0, ret+8(FP)
	RET

TEXT runtime·pthread_attr_setstacksize_trampoline(SB),NOSPLIT,$0-20
	MOVW	$46, R0
	BL	libc_exit(SB)
	MOVW	$0, ret+16(FP)
	RET

TEXT runtime·pthread_attr_setdetachstate_trampoline(SB),NOSPLIT,$0-20
	MOVW	$47, R0
	BL	libc_exit(SB)
	MOVW	$0, ret+16(FP)
	RET

TEXT runtime·pthread_create_trampoline(SB),NOSPLIT,$0-36
	MOVW	$48, R0
	BL	libc_exit(SB)
	MOVW	$0, ret+32(FP)
	RET

TEXT runtime·pthread_self_trampoline(SB),NOSPLIT,$0-8
	BL	libc_pthread_self(SB)
	MOVD	R0, ret+0(FP)
	RET

TEXT runtime·pthread_kill_trampoline(SB),NOSPLIT,$0-20
	MOVD	thread+0(FP), R0
	MOVD	sig+8(FP), R1
	BL	libc_pthread_kill(SB)
	MOVW	R0, ret+16(FP)
	RET
