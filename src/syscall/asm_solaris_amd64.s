// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "textflag.h"

//
// System calls for solaris/amd64 are implemented in ../runtime/syscall_solaris.go
//

TEXT syscall·sysvicall6(SB),NOSPLIT,$0
	JMP	runtime·syscall_sysvicall6(SB)

TEXT syscall·rawSysvicall6(SB),NOSPLIT,$0
	JMP	runtime·syscall_rawsysvicall6(SB)

TEXT syscall·chdir(SB),NOSPLIT,$0
	JMP	runtime·syscall_chdir(SB)

TEXT syscall·chroot1(SB),NOSPLIT,$0
	JMP	runtime·syscall_chroot(SB)

TEXT syscall·close(SB),NOSPLIT,$0
	JMP	runtime·syscall_close(SB)

TEXT syscall·execve(SB),NOSPLIT,$0
	JMP	runtime·syscall_execve(SB)

TEXT runtime∕internal∕base·Exit(SB),NOSPLIT,$0
	JMP	runtime·syscall_exit(SB)

TEXT syscall·fcntl1(SB),NOSPLIT,$0
	JMP	runtime·syscall_fcntl(SB)

TEXT syscall·forkx(SB),NOSPLIT,$0
	JMP	runtime·syscall_forkx(SB)

TEXT syscall·gethostname(SB),NOSPLIT,$0
	JMP	runtime·syscall_gethostname(SB)

TEXT syscall·getpid(SB),NOSPLIT,$0
	JMP	runtime·syscall_getpid(SB)

TEXT syscall·ioctl(SB),NOSPLIT,$0
	JMP	runtime·syscall_ioctl(SB)

TEXT syscall·pipe(SB),NOSPLIT,$0
	JMP	runtime·syscall_pipe(SB)

TEXT syscall·RawSyscall(SB),NOSPLIT,$0
	JMP	runtime·syscall_rawsyscall(SB)

TEXT syscall·setgid(SB),NOSPLIT,$0
	JMP	runtime·syscall_setgid(SB)

TEXT syscall·setgroups1(SB),NOSPLIT,$0
	JMP	runtime·syscall_setgroups(SB)

TEXT syscall·setsid(SB),NOSPLIT,$0
	JMP	runtime·syscall_setsid(SB)

TEXT syscall·setuid(SB),NOSPLIT,$0
	JMP	runtime·syscall_setuid(SB)

TEXT syscall·setpgid(SB),NOSPLIT,$0
	JMP	runtime·syscall_setpgid(SB)

TEXT ·Syscall(SB),NOSPLIT,$0
	JMP	runtime·syscall_syscall(SB)

TEXT syscall·wait4(SB),NOSPLIT,$0
	JMP	runtime·syscall_wait4(SB)

TEXT syscall·write1(SB),NOSPLIT,$0
	JMP	runtime·syscall_write(SB)
