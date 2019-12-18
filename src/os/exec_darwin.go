// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

/*
#include <sys/proc.h>
*/
import "C"

import (
	"errors"
	"syscall"
	"unsafe"
)

type kProc C.struct_extern_proc

func findProcess(pid int) (p *Process, err error) {
	args := [4]int32{1 /* CTL_KERN */, 14 /* KERN_PROC */, 1 /* KERN_PROC_PID */, int32(pid)}
	args_len := uintptr(len(args))

	// Get the size first.
	size := uintptr(0)
	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&args[0])),
		args_len,
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return nil, errno
	}
	buf := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&args[0])),
		args_len,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0,
	)
	if errno != 0 {
		return nil, errno
	}
	kp := (*kProc)(unsafe.Pointer(&buf[0]))
	if kp.p_pid == 0 {
		return nil, ErrProcessNotExist
	}
	if int(kp.p_pid) != pid {
		// Impossible but should be reported nonetheless.
		return nil, errors.New("Pid mismatch")
	}
	return newProcess(pid, 0), nil
}
