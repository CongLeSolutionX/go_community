// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"errors"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"
)

func (p *Process) wait() (ps *ProcessState, err error) {
	handle := atomic.LoadUintptr(&p.handle)
	s, e := syscall.WaitForSingleObject(syscall.Handle(handle), syscall.INFINITE)
	switch s {
	case syscall.WAIT_OBJECT_0:
		break
	case syscall.WAIT_FAILED:
		return nil, NewSyscallError("WaitForSingleObject", e)
	default:
		return nil, errors.New("os: unexpected result from WaitForSingleObject")
	}
	var ec uint32
	e = syscall.GetExitCodeProcess(syscall.Handle(handle), &ec)
	if e != nil {
		return nil, NewSyscallError("GetExitCodeProcess", e)
	}
	var u syscall.Rusage
	e = syscall.GetProcessTimes(syscall.Handle(handle), &u.CreationTime, &u.ExitTime, &u.KernelTime, &u.UserTime)
	if e != nil {
		return nil, NewSyscallError("GetProcessTimes", e)
	}
	p.setDone()
	// NOTE(brainman): It seems that sometimes process is not dead
	// when WaitForSingleObject returns. But we do not know any
	// other way to wait for it. Sleeping for a while seems to do
	// the trick sometimes. So we will sleep and smell the roses.
	defer time.Sleep(5 * time.Millisecond)
	defer p.Release()
	return &ProcessState{p.Pid, syscall.WaitStatus{ExitCode: ec}, &u}, nil
}

func terminateProcess(pid, exitcode int) error {
	h, e := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if e != nil {
		return NewSyscallError("OpenProcess", e)
	}
	defer syscall.CloseHandle(h)
	e = syscall.TerminateProcess(h, uint32(exitcode))
	return NewSyscallError("TerminateProcess", e)
}

func (p *Process) signal(sig Signal) error {
	handle := atomic.LoadUintptr(&p.handle)
	if handle == uintptr(syscall.InvalidHandle) {
		return syscall.EINVAL
	}
	if p.done() {
		return errors.New("os: process already finished")
	}
	if sig == Kill {
		err := terminateProcess(p.Pid, 1)
		runtime.KeepAlive(p)
		return err
	}
	// TODO(rsc): Handle Interrupt too?
	return syscall.Errno(syscall.EWINDOWS)
}

func (p *Process) release() error {
	handle := atomic.LoadUintptr(&p.handle)
	if handle == uintptr(syscall.InvalidHandle) {
		return syscall.EINVAL
	}
	e := syscall.CloseHandle(syscall.Handle(handle))
	if e != nil {
		return NewSyscallError("CloseHandle", e)
	}
	atomic.StoreUintptr(&p.handle, uintptr(syscall.InvalidHandle))
	// no need for a finalizer anymore
	runtime.SetFinalizer(p, nil)
	return nil
}

func findProcess(pid int) (p *Process, err error) {
	const da = syscall.STANDARD_RIGHTS_READ |
		syscall.PROCESS_QUERY_INFORMATION | syscall.SYNCHRONIZE
	h, e := syscall.OpenProcess(da, false, uint32(pid))
	if e != nil {
		return nil, NewSyscallError("OpenProcess", e)
	}
	return newProcess(pid, uintptr(h)), nil
}

func isCmdSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

// commandLineToArgv splits a command line into individual argument strings, following the
// Windows conventions documented at https://msdn.microsoft.com/en-us/library/17w5ykft.aspx.
func commandLineToArgv(cmd string) []string {
	var argv []string
	i := 0
	for {
		for i < len(cmd) && isCmdSpace(cmd[i]) {
			i++
		}
		if i == len(cmd) {
			break
		}

		var arg []byte
		inquote := false
		for {
			nslash := 0
			for i < len(cmd) && cmd[i] == '\\' {
				i++
				nslash++
			}

			if i < len(cmd) && cmd[i] == '"' {
				for s := 0; s < nslash/2; s++ {
					arg = append(arg, '\\')
				}

				if nslash%2 == 0 {
					inquote = !inquote
					// Special case: if the next character is also a quote,
					// then this quote gets included.
					if !inquote && i+1 < len(cmd) && cmd[i+1] == '"' {
						arg = append(arg, '"')
						i++
					}
				} else {
					arg = append(arg, '"')
				}
			} else {
				for nslash > 0 {
					arg = append(arg, '\\')
					nslash--
				}

				if i == len(cmd) || (!inquote && isCmdSpace(cmd[i])) {
					break
				}

				arg = append(arg, cmd[i])
			}

			i++
		}

		argv = append(argv, string(arg))
	}
	return argv
}

func init() {
	cmd := syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(syscall.GetCommandLine()))[:])

	if len(cmd) != 0 {
		Args = commandLineToArgv(cmd)
	} else {
		// No command line was provided, so get argv[0] from the module name.
		dll := syscall.MustLoadDLL("kernel32.dll")
		defer dll.Release()
		fn := dll.MustFindProc("GetModuleFileNameW")

		p := make([]uint16, syscall.MAX_PATH)
		r, _, err := fn.Call(0, uintptr(unsafe.Pointer(&p[0])), uintptr(len(p)))
		n := uint32(r)
		if n == 0 || n >= uint32(len(p)) {
			panic(err)
		}
		Args = []string{string(utf16.Decode(p[:n]))}
	}
}

func ftToDuration(ft *syscall.Filetime) time.Duration {
	n := int64(ft.HighDateTime)<<32 + int64(ft.LowDateTime) // in 100-nanosecond intervals
	return time.Duration(n*100) * time.Nanosecond
}

func (p *ProcessState) userTime() time.Duration {
	return ftToDuration(&p.rusage.UserTime)
}

func (p *ProcessState) systemTime() time.Duration {
	return ftToDuration(&p.rusage.KernelTime)
}
