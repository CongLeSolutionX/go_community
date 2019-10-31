// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package modload

import (
	"os"
	"os/user"
	"syscall"
)

func hasWritePermSys(fi os.FileInfo) bool {
	sys, ok := fi.Sys().(*syscall.Dir)
	if !ok {
		return false
	}

	u, err := user.Current()
	if err != nil {
		return false
	}

	if fi.Mode()&0200 != 0 && u.Uid == sys.Uid {
		return true
	}

	if fi.Mode()&0020 != 0 {
		groups, _ := u.GroupIds()
		for _, gid := range groups {
			if gid == sys.Gid {
				return true
			}
		}
	}

	return false
}
