// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"errors"
	"runtime"
	"sync"
	"syscall"
)

var getwdCache struct {
	sync.Mutex
	dir string
}

// Getwd returns an absolute path name corresponding to the
// current directory. If the current directory can be
// reached via multiple paths (due to symbolic links),
// Getwd may return any one of them.
//
// On Unix platforms, if the environment variable PWD
// provides an absolute name, and it is a name of the
// current directory, it is returned.
func Getwd() (dir string, err error) {
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		dir, err = syscall.Getwd()
		return dir, NewSyscallError("getwd", err)
	}

	// Clumsy but widespread kludge:
	// if $PWD is set and matches ".", use it.
	dir = Getenv("PWD")
	if len(dir) > 0 && dir[0] == '/' {
		dot, err := statNolog(".")
		if err != nil {
			return "", err
		}
		d, err := statNolog(dir)
		if err == nil && SameFile(dot, d) {
			return dir, nil
		}
	}

	// All platforms provide syscall.Getwd, so this should never happen.
	if !syscall.ImplementsGetwd {
		return "", errors.New(runtime.GOOS + "/" + runtime.GOARCH + " does not implement Getwd")
	}
	for {
		dir, err = syscall.Getwd()
		if err != syscall.EINTR {
			break
		}
	}
	return dir, NewSyscallError("getwd", err)
}
