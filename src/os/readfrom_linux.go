// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/poll"
	"io"
	"syscall"
)

var pollCopyFileRange = poll.CopyFileRange

func isStreamSock(fd int) bool {
	sa, err := syscall.Getsockname(fd)
	if err != nil {
		return false
	}

	switch sa.(type) {
	case *syscall.SockaddrInet4, *syscall.SockaddrInet6:
		sotype, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)
		if err != nil {
			return false
		}
		return sotype == syscall.SOCK_STREAM
	default:
		return false
	}
}

func (f *File) readFrom(r io.Reader) (written int64, handled bool, err error) {
	written, handled, err = f.copyFileRange(r)
	if handled {
		return
	}

	remain := int64(1 << 62)

	lr, ok := r.(*io.LimitedReader)
	if ok {
		remain, r = lr.N, lr.R
		if remain <= 0 {
			return 0, true, nil
		}
	}

	sc, ok := r.(syscall.Conn)
	if !ok {
		return
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return
	}

	var srcFD int

	if err = rc.Control(func(fd uintptr) {
		srcFD = int(fd)
	}); err != nil {
		return
	}

	// TODO(panjf2000): avoid the system calls and take a preconceived guess that r is a TCP socket?
	// Even if it's not, we can still tell by checking if the returned error from poll.Splice is EINVAL.
	isStream := isStreamSock(srcFD)
	if !isStream {
		return
	}

	src := &poll.FD{
		Sysfd:         srcFD,
		IsStream:      isStream,
		ZeroReadIsEOF: isStream,
	}

	var (
		sent       int64
		errSplice  error
		errSyscall string
	)
	err = rc.Read(func(fd uintptr) bool {
		sent, handled, errSyscall, errSplice = poll.Splice(&f.pfd, src, remain)
		written += sent
		if errSplice != nil && errSplice.Error() == "waiting for unsupported file type" {
			return false
		}

		return true
	})
	if err == nil {
		err = errSplice
	}

	if lr != nil {
		lr.N = remain - written
	}

	return written, handled, NewSyscallError(errSyscall, err)
}

func (f *File) copyFileRange(r io.Reader) (written int64, handled bool, err error) {
	// copy_file_range(2) does not support destinations opened with
	// O_APPEND, so don't even try.
	if f.appendMode {
		return 0, false, nil
	}

	remain := int64(1 << 62)

	lr, ok := r.(*io.LimitedReader)
	if ok {
		remain, r = lr.N, lr.R
		if remain <= 0 {
			return 0, true, nil
		}
	}

	src, ok := r.(*File)
	if !ok {
		return 0, false, nil
	}
	if src.checkValid("ReadFrom") != nil {
		// Avoid returning the error as we report handled as false,
		// leave further error handling as the responsibility of the caller.
		return 0, false, nil
	}

	written, handled, err = pollCopyFileRange(&f.pfd, &src.pfd, remain)
	if lr != nil {
		lr.N -= written
	}
	return written, handled, NewSyscallError("copy_file_range", err)
}
