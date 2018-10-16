// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build plan9

package filelock

import (
	"os"
	"strings"
	"syscall"
)

type lockType int8

const (
	readLock lockType = iota
	writeLock
)

func lock(f File, lt lockType) error {
	return &os.PathError{
		Op:   lt.String(),
		Path: f.Name(),
		Err:  ErrNotSupported,
	}
}

func unlock(f File) error {
	return &os.PathError{
		Op:   "Unlock",
		Path: f.Name(),
		Err:  ErrNotSupported,
	}
}

func isNotSupported(err error) bool {
	return err == ErrNotSupported
}

// Even though plan9 doesn't support the Lock/RLock/Unlock functions to
// manipulate already-open files, IsLocked is still meaningful: os.OpenFile
// itself may return errors that indicate that a file with the ModeExclusive bit
// set is already open.
func isLocked(err error) bool {
	errStr, ok := underlyingError(err).(syscall.ErrorString)
	if !ok {
		return false
	}
	s := string(errStr)

	// Per http://man.cat-v.org/plan_9/5/stat: “Exclusive use files may be open
	// for I/O by only one fid at a time across all clients of the server. If a
	// second open is attempted, it draws an error.”
	//
	// Too bad nobody documented which error!
	//
	// One of the programs in the plan9 distribution seems to check for the
	// strings "file is locked" and "exclusive lock", so we'll at least check for
	// those. We've also observed the string "exclusive use file already open"
	// during testing.
	for _, frag := range []string{
		"file is locked",
		"exclusive lock",
		"exclusive use file already open",
	} {
		if strings.Contains(s, frag) {
			return true
		}
	}

	return false
}
