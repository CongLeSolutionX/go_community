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

	src := &poll.FD{
		Sysfd:         srcFD,
		IsStream:      true,
		ZeroReadIsEOF: true,
	}

	var (
		sent       int64
		errSplice  error
		errSyscall string
	)
	err = rc.Read(func(fd uintptr) bool {
		sent, handled, errSyscall, errSplice = poll.Splice(&f.pfd, src, remain)
		written += sent

		// Since src is a mocked poll.FD, it has no src.pd.runtimeCtx, so poll.Splice may
		// return an error with "waiting for unsupported file type" when getting EAGAIN error and then calling
		// poll.pollDesc.waitRead inside, thus we need to wait until r is readable again and retry.
		if errSplice != nil && errSplice.Error() == "waiting for unsupported file type" {
			return false
		}

		return true
	})
	// EINVAL indicates r is not a valid target to use splice, ignore it.
	if errSplice == syscall.EINVAL {
		return 0, false, nil
	}
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
