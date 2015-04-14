// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build freebsd linux netbsd openbsd

package os

import (
	isyscall "internal/syscall"
	stdsyscall "syscall"
)

func (f *File) readdir(n int) (fi []FileInfo, err error) {
	dirname := f.name
	if dirname == "" {
		dirname = "."
	}
	names, err := f.Readdirnames(n)
	fi = make([]FileInfo, 0, len(names))
	for _, filename := range names {
		var st stdsyscall.Stat_t
		lerr := isyscall.Fstatat(f.fd, filename, &st, isyscall.AT_SYMLINK_NOFOLLOW)
		if IsNotExist(lerr) {
			// File disappeared between readdir + stat.
			// Just treat it as if it didn't exist.
			continue
		}
		if lerr != nil {
			return fi, lerr
		}
		fip := fileInfoFromStat(&st, dirname+"/"+filename)
		fi = append(fi, fip)
	}
	return fi, err
}
