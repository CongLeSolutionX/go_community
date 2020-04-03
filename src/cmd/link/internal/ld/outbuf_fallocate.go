// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build dragonfly freebsd linux openbsd

package ld

import "syscall"

func (out *OutBuf) FAllocate(length uint64) error {
	return syscall.Fallocate(out.f.Fd(), 0755, 0, length)
}
