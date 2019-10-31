// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package modload

import (
	"os"
	"syscall"
)

func hasWritePermSys(path string, fi os.FileInfo) bool {
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}

	uid := os.Getuid()
	if fi.Mode()&0200 != 0 && uint32(uid) == sys.Uid {
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

	// The ordinary file permissions don't give this user permission to write, but
	// it may be still be allowed by ACL. If the user is not root, we can try
	// opening the file to find out.
	//
	// (If the user *is* root, opening the file would always succeed and we would
	// learn nothing, but ACLs probably don't mention the root user anyway.)
	if uid != 0 {
		if f, err := os.OpenFile(path, os.O_WRONLY, 0); err == nil {
			f.Close()
			return true
		}
	}

	return false
}
