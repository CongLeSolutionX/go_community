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
	var fileSize int64

	lr, ok := r.(*io.LimitedReader)
	if ok {
		fileSize, r = lr.N, lr.R
		if fileSize <= 0 {
			return 0, nil, true
		}
	}

	f, ok := r.(*os.File)
	if !ok {
		return 0, nil, false
	}

	// As per https://docs.microsoft.com/en-us/windows/win32/api/mswsock/nf-mswsock-transmitfile
	// TransmitFile can be invoked in one call with at most
	// 2,147,483,646 bytes: the maximum value for a 32-bit integer minus 1.
	const _2GiB = int64(0x7fffffff - 1)

	switch {
	case fileSize <= _2GiB:
		// The fileSize is within sendfile's limits.
		return doSendFile(fd, lr, f, fileSize)

	default:
		// Now invoke doSendFile on the file in chunks upto 2GiB per chunk.
		for lr.N > 0 { // lr.N is decremented in every successful invocation of doSendFile.
			chunkSize := _2GiB
			if chunkSize > lr.N {
				chunkSize = lr.N
			}
			var nw int64
			nw, err, handled = doSendFile(fd, lr, f, chunkSize)
			if !handled || err != nil {
				// TODO: (@odeke-em, @alexbrainman) what should we do if !handled
				// in the middle of sending chunks?
				return
			}
			written += nw
		}
		return
	}
}

func doSendFile(fd *netFD, lr *io.LimitedReader, f *os.File, remain int64) (written int64, err error, handled bool) {
	done, err := poll.SendFile(&fd.pfd, syscall.Handle(f.Fd()), remain)

	if err != nil {
		return 0, wrapSyscallError("transmitfile", err), false
	}
	if lr != nil {
		lr.N -= int64(done)
	}
	return int64(done), nil, true
}
