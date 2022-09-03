// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/poll"
	"internal/reflectlite"
	"io"
	"syscall"
)

var (
	pollCopyFileRange = poll.CopyFileRange
	pollSplice        = poll.Splice
)

func (f *File) readFrom(r io.Reader) (written int64, handled bool, err error) {
	written, handled, err = f.copyFileRange(r)
	if handled {
		return
	}
	return f.spliceToFile(r)
}

// checkIfStream checks if the fd is a streaming descriptor.
func checkIfStream(fd int) bool {
	sa, err := syscall.Getsockname(fd)
	if err != nil {
		return false
	}
	switch sa.(type) {
	case *syscall.SockaddrUnix, *syscall.SockaddrInet4, *syscall.SockaddrInet6:
		sotype, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)
		if err != nil {
			return false
		}
		return sotype == syscall.SOCK_STREAM
	default:
		return false
	}
}

func (f *File) spliceToFile(r io.Reader) (written int64, handled bool, err error) {
	var (
		remain int64
		lr     *io.LimitedReader
	)
	if lr, r, remain = tryLimitedReader(r); remain <= 0 {
		return 0, true, nil
	}

	sc, ok := r.(syscall.Conn)
	if !ok {
		return
	}
	rc, err := sc.SyscallConn()
	if err != nil {
		return
	}

	var rfd int
	if err = rc.Control(func(fd uintptr) {
		rfd = int(fd)
	}); err != nil {
		return
	}

	// TODO(panjf2000): avoid the system calls and take a preconceived guess that r is a streaming descriptor?
	// Even if it's not, we can still tell by checking if the returned error from poll.Splice is EINVAL.
	isStream := checkIfStream(rfd)
	if !isStream {
		return
	}

	pfd := getPollFD(r)
	if pfd == nil || pfd.Sysfd != rfd {
		return
	}
	var syscallName string
	written, handled, syscallName, err = pollSplice(&f.pfd, pfd, remain)

	if lr != nil {
		lr.N = remain - written
	}

	return written, handled, NewSyscallError(syscallName, err)
}

// getPollFD pulls the poll.FD out of the net.Conn(net.TCPConn/net.UnixConn) via reflection.
// Note that the test for validating this function is TestGetPollFDFromReader, located in readfrom_linux_test.go
func getPollFD(r io.Reader) *poll.FD {
	v := reflectlite.ValueOf(r).Elem()
	if !v.IsValid() {
		return nil
	}
	vfd := v.Field(0).Field(0)
	if !vfd.IsValid() || vfd.IsNil() {
		return nil
	}
	return (*poll.FD)(vfd.InternalUnsafePointer())
}

func (f *File) copyFileRange(r io.Reader) (written int64, handled bool, err error) {
	// copy_file_range(2) does not support destinations opened with
	// O_APPEND, so don't even try.
	if f.appendMode {
		return 0, false, nil
	}

	var (
		remain int64
		lr     *io.LimitedReader
	)
	if lr, r, remain = tryLimitedReader(r); remain <= 0 {
		return 0, true, nil
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

// tryLimitedReader tries to assert the io.Reader to io.LimitedReader, it returns the io.LimitedReader,
// the underlying io.Reader and the remaining amount of bytes if the assertion succeeds,
// otherwise it just returns the original io.Reader and the theoretical unlimited remaining amount of bytes.
func tryLimitedReader(r io.Reader) (*io.LimitedReader, io.Reader, int64) {
	remain := int64(1 << 62)

	lr, ok := r.(*io.LimitedReader)
	if !ok {
		return nil, r, remain
	}

	remain = lr.N
	return lr, lr.R, remain
}
