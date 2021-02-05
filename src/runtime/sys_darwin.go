// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

// The X versions of syscall expect the libc call to return a 64-bit result.
// Otherwise (the non-X version) expects a 32-bit result.
// This distinction is required because an error is indicated by returning -1,
// and we need to know whether to check 32 or 64 bits of the result.
// (Some libc functions that return 32 bits put junk in the upper 32 bits of AX.)

//go:linkname syscall_syscall syscall.syscall
//go:nosplit
//go:cgo_unsafe_args
func syscall_syscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	entersyscall()
	libcCall(unsafe.Pointer(funcPC(syscall)), unsafe.Pointer(&fn))
	exitsyscall()
	return
}
func syscall()

//go:linkname syscall_syscallX syscall.syscallX
//go:nosplit
//go:cgo_unsafe_args
func syscall_syscallX(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	entersyscallblock()
	libcCall(unsafe.Pointer(funcPC(syscallX)), unsafe.Pointer(&fn))
	exitsyscall()
	return
}
func syscallX()

//go:linkname syscall_syscall6 syscall.syscall6
//go:nosplit
//go:cgo_unsafe_args
func syscall_syscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	entersyscall()
	libcCall(unsafe.Pointer(funcPC(syscall6)), unsafe.Pointer(&fn))
	exitsyscall()
	return
}
func syscall6()

//go:linkname syscall_syscall6X syscall.syscall6X
//go:nosplit
//go:cgo_unsafe_args
func syscall_syscall6X(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	entersyscall()
	libcCall(unsafe.Pointer(funcPC(syscall6X)), unsafe.Pointer(&fn))
	exitsyscall()
	return
}
func syscall6X()

//go:linkname syscall_syscallPtr syscall.syscallPtr
//go:nosplit
//go:cgo_unsafe_args
func syscall_syscallPtr(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	entersyscall()
	libcCall(unsafe.Pointer(funcPC(syscallPtr)), unsafe.Pointer(&fn))
	exitsyscall()
	return
}
func syscallPtr()

//go:linkname syscall_rawSyscall syscall.rawSyscall
//go:nosplit
//go:cgo_unsafe_args
func syscall_rawSyscall(fn, a1, a2, a3 uintptr) (r1, r2, err uintptr) {
	libcCall(unsafe.Pointer(funcPC(syscall)), unsafe.Pointer(&fn))
	return
}

//go:linkname syscall_rawSyscall6 syscall.rawSyscall6
//go:nosplit
//go:cgo_unsafe_args
func syscall_rawSyscall6(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, err uintptr) {
	libcCall(unsafe.Pointer(funcPC(syscall6)), unsafe.Pointer(&fn))
	return
}

// syscallNoErr is used in crypto/x509 to call into Security.framework and CF.

//go:linkname crypto_x509_syscall crypto/x509/internal/macos.syscall
//go:nosplit
//go:cgo_unsafe_args
func crypto_x509_syscall(fn, a1, a2, a3, a4, a5, a6 uintptr) (r1 uintptr) {
	entersyscall()
	libcCall(unsafe.Pointer(funcPC(syscallNoErr)), unsafe.Pointer(&fn))
	exitsyscall()
	return
}
func syscallNoErr()

// The *_trampoline functions convert from the Go calling convention to the C calling convention
// and then call the underlying libc function.  They are defined in sys_darwin_$ARCH.s.

//go:nosplit
//go:cgo_unsafe_args
func pthread_attr_init(attr *pthreadattr) int32 {
	return libcCall(unsafe.Pointer(pthread_attr_init_trampoline_addr), unsafe.Pointer(&attr))
}

var pthread_attr_init_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_attr_getstacksize(attr *pthreadattr, size *uintptr) int32 {
	return libcCall(unsafe.Pointer(pthread_attr_getstacksize_trampoline_addr), unsafe.Pointer(&attr))
}

var pthread_attr_getstacksize_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_attr_setdetachstate(attr *pthreadattr, state int) int32 {
	return libcCall(unsafe.Pointer(pthread_attr_setdetachstate_trampoline_addr), unsafe.Pointer(&attr))
}

var pthread_attr_setdetachstate_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_create(attr *pthreadattr, start uintptr, arg unsafe.Pointer) int32 {
	return libcCall(unsafe.Pointer(pthread_create_trampoline_addr), unsafe.Pointer(&attr))
}

var pthread_create_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func raise(sig uint32) {
	libcCall(unsafe.Pointer(raise_trampoline_addr), unsafe.Pointer(&sig))
}

var raise_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_self() (t pthread) {
	libcCall(unsafe.Pointer(pthread_self_trampoline_addr), unsafe.Pointer(&t))
	return
}

var pthread_self_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_kill(t pthread, sig uint32) {
	libcCall(unsafe.Pointer(pthread_kill_trampoline_addr), unsafe.Pointer(&t))
	return
}

var pthread_kill_trampoline_addr uintptr

// mmap is used to do low-level memory allocation via mmap. Don't allow stack
// splits, since this function (used by sysAlloc) is called in a lot of low-level
// parts of the runtime and callers often assume it won't acquire any locks.
// go:nosplit
func mmap(addr unsafe.Pointer, n uintptr, prot, flags, fd int32, off uint32) (unsafe.Pointer, int) {
	args := struct {
		addr            unsafe.Pointer
		n               uintptr
		prot, flags, fd int32
		off             uint32
		ret1            unsafe.Pointer
		ret2            int
	}{addr, n, prot, flags, fd, off, nil, 0}
	libcCall(unsafe.Pointer(mmap_trampoline_addr), unsafe.Pointer(&args))
	return args.ret1, args.ret2
}

var mmap_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func munmap(addr unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(munmap_trampoline_addr), unsafe.Pointer(&addr))
}

var munmap_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func madvise(addr unsafe.Pointer, n uintptr, flags int32) {
	libcCall(unsafe.Pointer(madvise_trampoline_addr), unsafe.Pointer(&addr))
}

var madvise_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func mlock(addr unsafe.Pointer, n uintptr) {
	libcCall(unsafe.Pointer(mlock_trampoline_addr), unsafe.Pointer(&addr))
}

var mlock_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func read(fd int32, p unsafe.Pointer, n int32) int32 {
	return libcCall(unsafe.Pointer(read_trampoline_addr), unsafe.Pointer(&fd))
}

var read_trampoline_addr uintptr

func pipe() (r, w int32, errno int32) {
	var p [2]int32
	errno = libcCall(unsafe.Pointer(pipe_trampoline_addr), noescape(unsafe.Pointer(&p)))
	return p[0], p[1], errno
}

var pipe_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func closefd(fd int32) int32 {
	return libcCall(unsafe.Pointer(close_trampoline_addr), unsafe.Pointer(&fd))
}

var close_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
//
// This is exported via linkname to assembly in runtime/cgo.
//go:linkname exit
func exit(code int32) {
	libcCall(unsafe.Pointer(exit_trampoline_addr), unsafe.Pointer(&code))
}

var exit_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func usleep(usec uint32) {
	libcCall(unsafe.Pointer(usleep_trampoline_addr), unsafe.Pointer(&usec))
}

var usleep_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func write1(fd uintptr, p unsafe.Pointer, n int32) int32 {
	return libcCall(unsafe.Pointer(write_trampoline_addr), unsafe.Pointer(&fd))
}

var write_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func open(name *byte, mode, perm int32) (ret int32) {
	return libcCall(unsafe.Pointer(open_trampoline_addr), unsafe.Pointer(&name))
}

var open_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func nanotime1() int64 {
	var r struct {
		t            int64  // raw timer
		numer, denom uint32 // conversion factors. nanoseconds = t * numer / denom.
	}
	libcCall(unsafe.Pointer(nanotime_trampoline_addr), unsafe.Pointer(&r))
	// Note: Apple seems unconcerned about overflow here. See
	// https://developer.apple.com/library/content/qa/qa1398/_index.html
	// Note also, numer == denom == 1 is common.
	t := r.t
	if r.numer != 1 {
		t *= int64(r.numer)
	}
	if r.denom != 1 {
		t /= int64(r.denom)
	}
	return t
}

var nanotime_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func walltime1() (int64, int32) {
	var t timespec
	libcCall(unsafe.Pointer(walltime_trampoline_addr), unsafe.Pointer(&t))
	return t.tv_sec, int32(t.tv_nsec)
}

var walltime_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func sigaction(sig uint32, new *usigactiont, old *usigactiont) {
	libcCall(unsafe.Pointer(sigaction_trampoline_addr), unsafe.Pointer(&sig))
}

var sigaction_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func sigprocmask(how uint32, new *sigset, old *sigset) {
	libcCall(unsafe.Pointer(sigprocmask_trampoline_addr), unsafe.Pointer(&how))
}

var sigprocmask_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func sigaltstack(new *stackt, old *stackt) {
	if new != nil && new.ss_flags&_SS_DISABLE != 0 && new.ss_size == 0 {
		// Despite the fact that Darwin's sigaltstack man page says it ignores the size
		// when SS_DISABLE is set, it doesn't. sigaltstack returns ENOMEM
		// if we don't give it a reasonable size.
		// ref: http://lists.llvm.org/pipermail/llvm-commits/Week-of-Mon-20140421/214296.html
		new.ss_size = 32768
	}
	libcCall(unsafe.Pointer(sigaltstack_trampoline_addr), unsafe.Pointer(&new))
}

var sigaltstack_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func raiseproc(sig uint32) {
	libcCall(unsafe.Pointer(raiseproc_trampoline_addr), unsafe.Pointer(&sig))
}

var raiseproc_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func setitimer(mode int32, new, old *itimerval) {
	libcCall(unsafe.Pointer(setitimer_trampoline_addr), unsafe.Pointer(&mode))
}

var setitimer_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func sysctl(mib *uint32, miblen uint32, oldp *byte, oldlenp *uintptr, newp *byte, newlen uintptr) int32 {
	return libcCall(unsafe.Pointer(sysctl_trampoline_addr), unsafe.Pointer(&mib))
}

var sysctl_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func sysctlbyname(name *byte, oldp *byte, oldlenp *uintptr, newp *byte, newlen uintptr) int32 {
	return libcCall(unsafe.Pointer(sysctlbyname_trampoline_addr), unsafe.Pointer(&name))
}

var sysctlbyname_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func fcntl(fd, cmd, arg int32) int32 {
	return libcCall(unsafe.Pointer(fcntl_trampoline_addr), unsafe.Pointer(&fd))
}

var fcntl_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func kqueue() int32 {
	v := libcCall(unsafe.Pointer(kqueue_trampoline_addr), nil)
	return v
}

var kqueue_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func kevent(kq int32, ch *keventt, nch int32, ev *keventt, nev int32, ts *timespec) int32 {
	return libcCall(unsafe.Pointer(kevent_trampoline_addr), unsafe.Pointer(&kq))
}

var kevent_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_mutex_init(m *pthreadmutex, attr *pthreadmutexattr) int32 {
	return libcCall(unsafe.Pointer(pthread_mutex_init_trampoline_addr), unsafe.Pointer(&m))
}

var pthread_mutex_init_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_mutex_lock(m *pthreadmutex) int32 {
	return libcCall(unsafe.Pointer(pthread_mutex_lock_trampoline_addr), unsafe.Pointer(&m))
}

var pthread_mutex_lock_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_mutex_unlock(m *pthreadmutex) int32 {
	return libcCall(unsafe.Pointer(pthread_mutex_unlock_trampoline_addr), unsafe.Pointer(&m))
}

var pthread_mutex_unlock_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_cond_init(c *pthreadcond, attr *pthreadcondattr) int32 {
	return libcCall(unsafe.Pointer(pthread_cond_init_trampoline_addr), unsafe.Pointer(&c))
}

var pthread_cond_init_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_cond_wait(c *pthreadcond, m *pthreadmutex) int32 {
	return libcCall(unsafe.Pointer(pthread_cond_wait_trampoline_addr), unsafe.Pointer(&c))
}

var pthread_cond_wait_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_cond_timedwait_relative_np(c *pthreadcond, m *pthreadmutex, t *timespec) int32 {
	return libcCall(unsafe.Pointer(pthread_cond_timedwait_relative_np_trampoline_addr), unsafe.Pointer(&c))
}

var pthread_cond_timedwait_relative_np_trampoline_addr uintptr

//go:nosplit
//go:cgo_unsafe_args
func pthread_cond_signal(c *pthreadcond) int32 {
	return libcCall(unsafe.Pointer(pthread_cond_signal_trampoline_addr), unsafe.Pointer(&c))
}

var pthread_cond_signal_trampoline_addr uintptr

// Not used on Darwin, but must be defined.
func exitThread(wait *uint32) {
}

//go:nosplit
func closeonexec(fd int32) {
	fcntl(fd, _F_SETFD, _FD_CLOEXEC)
}

//go:nosplit
func setNonblock(fd int32) {
	flags := fcntl(fd, _F_GETFL, 0)
	fcntl(fd, _F_SETFL, flags|_O_NONBLOCK)
}

// Tell the linker that the libc_* functions are to be found
// in a system library, with the libc_ prefix missing.

//go:cgo_import_dynamic libc_pthread_attr_init pthread_attr_init "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_attr_getstacksize pthread_attr_getstacksize "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_attr_setdetachstate pthread_attr_setdetachstate "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_create pthread_create "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_self pthread_self "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_kill pthread_kill "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_exit _exit "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_raise raise "/usr/lib/libSystem.B.dylib"

//go:cgo_import_dynamic libc_open open "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_close close "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_read read "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_write write "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pipe pipe "/usr/lib/libSystem.B.dylib"

//go:cgo_import_dynamic libc_mmap mmap "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_munmap munmap "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_madvise madvise "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_mlock mlock "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_error __error "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_usleep usleep "/usr/lib/libSystem.B.dylib"

//go:cgo_import_dynamic libc_mach_timebase_info mach_timebase_info "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_mach_absolute_time mach_absolute_time "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_clock_gettime clock_gettime "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_sigaction sigaction "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_sigmask pthread_sigmask "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_sigaltstack sigaltstack "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_getpid getpid "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_kill kill "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_setitimer setitimer "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_sysctl sysctl "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_sysctlbyname sysctlbyname "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_fcntl fcntl "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_kqueue kqueue "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_kevent kevent "/usr/lib/libSystem.B.dylib"

//go:cgo_import_dynamic libc_pthread_mutex_init pthread_mutex_init "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_mutex_lock pthread_mutex_lock "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_mutex_unlock pthread_mutex_unlock "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_cond_init pthread_cond_init "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_cond_wait pthread_cond_wait "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_cond_timedwait_relative_np pthread_cond_timedwait_relative_np "/usr/lib/libSystem.B.dylib"
//go:cgo_import_dynamic libc_pthread_cond_signal pthread_cond_signal "/usr/lib/libSystem.B.dylib"
