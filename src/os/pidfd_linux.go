// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package os

import (
	"internal/syscall/unix"
	"sync"
	"syscall"
)

func pidfdOpen(pid int) (uintptr, error) {
	h, err := unix.PidFDOpen(pid, 0)
	switch err {
	case nil:
		return h, nil
	case syscall.ESRCH:
		return unsetHandle, ErrProcessDone
	}
	return unsetHandle, err
}

func pidfdSendSignal(handle uintptr, s syscall.Signal) (_ error, done bool) {
	if !canUsePidfdSendSignal() || handle == unsetHandle {
		return nil, false
	}
	e := unix.PidFDSendSignal(handle, s)
	switch e {
	case nil:
		return nil, true
	case syscall.ESRCH:
		return ErrProcessDone, true
	case syscall.ENOSYS:
		return e, false
	}
	return e, true
}

var (
	pidfdSendSignalOnce  sync.Once
	pidfdSendSignalWorks bool
)

func canUsePidfdSendSignal() bool {
	pidfdSendSignalOnce.Do(func() {
		if fd, err := syscall.Open("/proc/self", syscall.O_RDONLY, 0); err == nil {
			e := unix.PidFDSendSignal(uintptr(fd), 0)
			syscall.Close(fd)
			pidfdSendSignalWorks = (e == nil)
		}
	})

	return pidfdSendSignalWorks
}
