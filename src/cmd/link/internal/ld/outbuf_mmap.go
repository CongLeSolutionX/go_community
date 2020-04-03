// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux openbsd

package ld

import (
	"syscall"
)

func (out *OutBuf) Mmap(filesize uint64) error {
	err := out.f.Truncate(int64(filesize))
	if err != nil {
		Exitf("resize output file failed: %v", err)
	}
	out.length = filesize
	out.buf, err = syscall.Mmap(int(out.f.Fd()), 0, int(filesize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_FILE)
	return err
}

func (out *OutBuf) Munmap() {
	if out.buf == nil {
		return
	}
	syscall.Munmap(out.buf)
	out.buf = nil
	if _, err := out.f.Seek(out.off, 0); err != nil {
		Exitf("seek output file failed: %v", err)
	}
}
