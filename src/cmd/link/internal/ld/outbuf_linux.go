// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ld

import "syscall"

func (out *OutBuf) fallocate(size uint64) error {
	// 1.Original implementation:
	//   return syscall.Fallocate(int(out.f.Fd()), 0, 0, int64(size))
	//
	// 2.The current implementation is to fix issues #53116 and #53804, if your
	//   kernel version is not between 5.16-rc4 and 6.1, you can use the original
	//   simpler version of the implementation.
	var st syscall.Stat_t

	err := syscall.Fstat(int(out.f.Fd()), &st)
	if err != nil {
		return err
	}

	cursize := uint64(st.Blocks) * 512
	if size <= cursize {
		return nil
	}

	return syscall.Fallocate(int(out.f.Fd()), 0, int64(cursize), int64(size-cursize))
}
