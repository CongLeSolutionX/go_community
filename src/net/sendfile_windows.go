// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"internal/poll"
	"io"
	"os"
	"syscall"
)

// sendFile copies the contents of r to c using the TransmitFile
// system call to minimize copies.
//
// if handled == true, sendFile returns the number of bytes copied and any
// non-EOF error.
//
// if handled == false, sendFile performed no work.
func sendFile(fd *netFD, r io.Reader) (written int64, err error, handled bool) {
	var n int64 = 0 // by default, copy until EOF.

	// afterEachSend and remaining maintain the state of the current
	// number of written and remaining bytes, per SendFile invocation.
	var afterEachSend func(written int64)
	var remaining func() int64

	lr, ok := r.(*io.LimitedReader)
	if ok {
		n, r = lr.N, lr.R
		if n <= 0 {
			return 0, nil, true
		}
		afterEachSend = func(written int64) {
			lr.N -= written
		}
		remaining = func() int64 { return lr.N }
	}

	f, ok := r.(*os.File)
	if !ok {
		return 0, nil, false
	}

	if n == 0 || afterEachSend == nil || remaining == nil { // We haven't yet infered the size.
		// Try to infer the file size from stat-ing instead.
		fi, err := f.Stat()
		if err != nil {
			return 0, err, false
		}
		n = fi.Size()
		afterEachSend = func(written int64) {
			if written > 0 {
				n -= written
			}
		}
		remaining = func() int64 { return n }
	}

	fh := syscall.Handle(f.Fd())

	// TransmitFile can be invoked in one call with at most
	// 2,147,483,646 bytes: the maximum value for a 32-bit integer minus 1.
	// See https://docs.microsoft.com/en-us/windows/win32/api/mswsock/nf-mswsock-transmitfile
	const maxChunkSizePerCall = int64(0x7fffffff - 1)

	switch {
	case n <= maxChunkSizePerCall:
		// The file is within sendfile's limits.
		written, err = doSendFile(fd, fh, n)
		afterEachSend(written)

	default:
		// Now invoke doSendFile on the file in chunks of upto 2GiB per chunk.
		for {
			nLeft := remaining()
			if nLeft <= 0 {
				break
			}
			chunkSize := maxChunkSizePerCall
			if chunkSize > nLeft {
				chunkSize = nLeft
			}
			var nw int64
			nw, err = doSendFile(fd, fh, chunkSize)
			if err != nil {
				break
			}
			written += nw
			afterEachSend(nw)
		}
	}

	// If any byte was copied, regardless of any error
	// encountered mid-way, handled must be set to true.
	return written, err, written > 0
}

// doSendFile is a helper to invoke poll.SendFile.
func doSendFile(fd *netFD, fh syscall.Handle, remain int64) (written int64, err error) {
	done, err := poll.SendFile(&fd.pfd, fh, remain)
	if err != nil {
		return 0, wrapSyscallError("transmitfile", err)
	}
	return int64(done), nil
}
