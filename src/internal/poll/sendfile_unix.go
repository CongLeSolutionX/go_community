// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || linux || solaris

package poll

import (
	"runtime"
	"syscall"
)

// SendFile wraps the sendfile system call.
func SendFile(dstFD *FD, src int, size int64) (n int64, err error, handled bool) {
	if runtime.GOOS == "linux" {
		// Linux's sendfile doesn't require any setup:
		// It sends from the current position of the source file,
		// updates the position of the source after sending,
		// and sends everything when the size is 0.
		return sendFile(dstFD, src, nil, size)
	}

	// Darwin/FreeBSD/DragonFly/Solaris's sendfile implementation
	// doesn't use the current position of the file --
	// if you pass it offset 0, it starts from offset 0.
	// There's no way to tell it "start from current position",
	// so we have to manage that explicitly.
	const (
		seekStart   = 0
		seekCurrent = 1
		seekEnd     = 2
	)
	start, err := ignoringEINTR2(func() (int64, error) {
		return syscall.Seek(src, 0, seekCurrent)
	})
	if err != nil {
		return 0, err, false
	}

	// Solaris requires us to pass a length to send,
	// rather than accepting 0 as "send everything".
	//
	// Seek to the end of the source file to find its length.
	mustReposition := false
	if runtime.GOOS == "solaris" && size == 0 {
		end, err := ignoringEINTR2(func() (int64, error) {
			return syscall.Seek(src, 0, seekEnd)
		})
		if err != nil {
			return 0, err, false
		}
		size = end - start
		mustReposition = true
	}

	pos := start
	n, err, handled = sendFile(dstFD, src, &pos, size)
	if n > 0 || mustReposition {
		ignoringEINTR2(func() (int64, error) {
			return syscall.Seek(src, start+n, seekStart)
		})
	}
	return n, err, handled
}

// sendFile wraps the sendfile system call.
func sendFile(dstFD *FD, src int, offset *int64, size int64) (written int64, err error, handled bool) {
	defer func() {
		TestHookDidSendFile(dstFD, src, written, err, handled)
	}()
	if err := dstFD.writeLock(); err != nil {
		return 0, err, false
	}
	defer dstFD.writeUnlock()

	if err := dstFD.pd.prepareWrite(dstFD.isFile); err != nil {
		return 0, err, false
	}

	dst := dstFD.Sysfd
	for {
		chunk := 0
		if size > 0 {
			chunk = int(size - written)
		}
		var n int
		n, err = sendFileChunk(dst, src, offset, chunk)
		if n > 0 {
			written += int64(n)
		}
		if runtime.GOOS == "solaris" {
			// A quirk on Solaris: sendfile() claims to support out_fd
			// as a regular file but returns EINVAL when the out_fd
			// is not a socket of SOCK_STREAM, while it actually sends
			// out data anyway and updates the file offset.
			if err == syscall.EINVAL && written > 0 {
				err = nil
			}
		}
		if err == nil {
			if n == 0 {
				break
			}
		} else if err == syscall.EAGAIN {
			err = nil
		} else if err == syscall.EINTR {
			continue
		} else {
			// This includes syscall.ENOSYS (no kernel
			// support) and syscall.EINVAL (fd types which
			// don't implement sendfile), and other errors.
			// We should end the loop when there is no error
			// returned from sendfile(2) or it is not a retryable error.
			break
		}
		if size > 0 && written >= size {
			break
		}
		if err = dstFD.pd.waitWrite(dstFD.isFile); err != nil {
			break
		}
	}
	handled = written != 0 || (err != syscall.ENOSYS && err != syscall.EINVAL)
	return
}

func sendFileChunk(dst, src int, offset *int64, size int) (n int, err error) {
	start := *offset
	n, err = syscall.Sendfile(dst, src, offset, size)
	switch runtime.GOOS {
	case "linux":
		// We always pass a nil offset on Linux at this time.
	case "solaris":
		// Trust the offset, not the return value from sendfile.
		n = int(*offset - start)
	default:
		if n > 0 {
			// The BSD implementations of syscall.Sendfile don't
			// update the offset parameter (despite it being a *int64).
			//
			// Trust the return value from sendfile, not the offset.
			*offset = start + int64(n)
		}
	}
	return
}
