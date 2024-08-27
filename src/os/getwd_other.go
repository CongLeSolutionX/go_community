// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !plan9

package os

import (
	"sync"
	"syscall"
)

var getwdCache struct {
	sync.Mutex
	dir string
}

func getwd() (dir string, err error) {
	// Clumsy but widespread kludge:
	// if $PWD is set and matches ".", use it.
	var dot FileInfo
	dir = Getenv("PWD")
	if len(dir) > 0 && dir[0] == '/' {
		dot, err = statNolog(".")
		if err != nil {
			return "", err
		}
		d, err := statNolog(dir)
		if err == nil && SameFile(dot, d) {
			return dir, nil
		}
		// If err is ENAMETOOLONG here, the syscall.Getwd below will
		// fail with the same error, too, but let's give it a try
		// anyway as the fallback code is much slower.
	}

	// If the operating system provides a Getwd call, use it.
	// Otherwise, we're trying to find our way back to ".".
	if syscall.ImplementsGetwd {
		for {
			dir, err = syscall.Getwd()
			if err != syscall.EINTR {
				break
			}
		}
		if err == syscall.ENAMETOOLONG {
			goto Fallback
		}
		return dir, NewSyscallError("getwd", err)
	}

Fallback:
	if dot == nil {
		dot, err = statNolog(".")
		if err != nil {
			return "", err
		}
	}
	// Apply same kludge but to cached dir instead of $PWD.
	getwdCache.Lock()
	dir = getwdCache.dir
	getwdCache.Unlock()
	if len(dir) > 0 {
		d, err := statNolog(dir)
		if err == nil && SameFile(dot, d) {
			return dir, nil
		}
	}

	// Root is a special case because it has no parent
	// and ends in a slash.
	root, err := statNolog("/")
	if err != nil {
		// Can't stat root - no hope.
		return "", err
	}
	if SameFile(root, dot) {
		return "/", nil
	}

	// General algorithm: find name in parent
	// and then find name of parent. Each iteration
	// adds /name to the beginning of dir.
	dir = ""
	for parent := ".."; ; parent = "../" + parent {
		if len(parent) >= 1024 { // Sanity check
			return "", NewSyscallError("getwd", syscall.ENAMETOOLONG)
		}
		fd, err := openDirNolog(parent)
		if err != nil {
			return "", err
		}

		for {
			names, err := fd.Readdirnames(100)
			if err != nil {
				fd.Close()
				// Readdirnames can return io.EOF or other error.
				// In any case, we're here because syscall.Getwd
				// is not implemented or failed with ENAMETOOLONG,
				// so return the most sensible error.
				if syscall.ImplementsGetwd {
					return "", NewSyscallError("getwd", syscall.ENAMETOOLONG)
				}
				return "", NewSyscallError("getwd", syscall.ENOSYS)
			}
			for _, name := range names {
				d, _ := lstatNolog(parent + "/" + name)
				if SameFile(d, dot) {
					dir = "/" + name + dir
					goto Found
				}
			}
		}

	Found:
		pd, err := fd.Stat()
		fd.Close()
		if err != nil {
			return "", err
		}
		if SameFile(pd, root) {
			break
		}
		// Set up for next round.
		dot = pd
	}

	// Save answer as hint to avoid the expensive path next time.
	getwdCache.Lock()
	getwdCache.dir = dir
	getwdCache.Unlock()

	return dir, nil
}
