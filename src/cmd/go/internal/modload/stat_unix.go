// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package modload

import (
	"os"
	"syscall"
)

func hasWritePermSys(fi os.FileInfo) bool {
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}

	if fi.Mode()&0200 != 0 && uint32(os.Getuid()) == sys.Uid {
		return true
	}

	if fi.Mode()&0020 != 0 {
		groups, _ := os.Getgroups()
		for _, gid := range groups {
			if uint32(gid) == sys.Gid {
				return true
			}
		}
	}

	return false
}
