// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"internal/poll"
	"io"
	"os"
	_ "unsafe"
)

// sendFile copies the contents of r to c using the sendfile
// system call to minimize copies.
//
// if handled == true, sendFile returns the number (potentially zero) of bytes
// copied and any non-EOF error.
//
// if handled == false, sendFile performed no work.
func sendFile(c *netFD, r io.Reader) (written int64, err error, handled bool) {
	var remain int64 = 1<<63 - 1 // by default, copy until EOF
	var pos int64 = -1           // pos == -1 means use current position
	var f *os.File

	// TODO: We could potentially also support nested Limited/SectionReaders.
	switch tr := r.(type) {
	case *io.LimitedReader:
		if wf, ok := tr.R.(*os.File); ok {
			remain, f = tr.N, wf
		}
	case *io.SectionReader:
		ra, base, n := tr.Outer()
		if wf, ok := ra.(*os.File); ok {
			off, err := tr.Seek(0, io.SeekCurrent)
			if err != nil {
				// SectionReader.Seek(0, SeekCurrent) never returns error
				// for a valid SectionReader. If we get an error here
				// something is wrong with the SectionReader, so bail out.
				return 0, err, true
			}
			pos, remain, f = base+off, n-off, wf
		}
	case *os.File:
		f = tr
	}
	if f == nil {
		return 0, nil, false
	} else if remain <= 0 {
		return 0, nil, true
	}

	sc, err := f.SyscallConn()
	if err != nil {
		return 0, nil, false
	}

	var werr error
	err = sc.Read(func(fd uintptr) bool {
		written, werr, handled = poll.SendFile(&c.pfd, int(fd), pos, remain)
		return true
	})
	if err == nil {
		err = werr
	}

	if sendFileTestHook != nil {
		sendFileTestHook(&c.pfd, f, pos, remain, written, err, handled)
	}

	switch tr := r.(type) {
	case *io.LimitedReader:
		tr.N = remain - written
	case *io.SectionReader:
		_, serr := tr.Seek(written, io.SeekCurrent)
		if err == nil && serr != nil {
			return written, serr, true
		}
	}

	return written, wrapSyscallError("sendfile", err), handled
}

var sendFileTestHook func(pfd *poll.FD, f *os.File, pos, remain, written int64, err error, handled bool)
