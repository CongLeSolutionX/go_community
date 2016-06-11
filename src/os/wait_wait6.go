// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"runtime"
	"syscall"
	"unsafe"
)

const _P_PID = 0

// blockUntilWaitable attempts to block until a call to p.Wait will
// succeed immediately, and returns whether it has done so.
// It does not actually call p.Wait.
func (p *Process) blockUntilWaitable() (bool, error) {
	// waitid expects a pointer to a siginfo_t, which is 128 bytes
	// on all systems. We don't care about the values it returns.
	var siginfo [128]byte
	psig := &siginfo[0]
	var status int32
	_, _, errno := syscall.Syscall6(syscall.SYS_WAIT6, _P_PID, uintptr(p.Pid), uintptr(unsafe.Pointer(&status)), syscall.WEXITED|syscall.WNOWAIT, 0, uintptr(unsafe.Pointer(psig)))
	runtime.KeepAlive(psig)
	if errno != 0 {
		if errno == syscall.ENOSYS {
			return false, nil
		}
		return false, NewSyscallError("wait6", errno)
	}
	return true, nil
}
