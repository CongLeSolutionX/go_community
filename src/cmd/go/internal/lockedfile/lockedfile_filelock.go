// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package lockedfile

import (
	"os"

	"cmd/go/internal/lockedfile/filelock"
)

func openFile(name string, flag int, perm os.FileMode) (_ *os.File, err error) {
	// On BSD systems, we could add the O_SHLOCK or O_EXLOCK flag to the OpenFile
	// call instead of locking separately, but we have to support separate locking
	// calls for Linux and Windows anyway, so it's simpler to use that approach
	// consistently.

	f, err := os.OpenFile(name, flag&^os.O_TRUNC, perm)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	if flag&os.O_WRONLY == os.O_WRONLY || flag&os.O_RDWR == os.O_RDWR {
		err = filelock.Lock(f)
	} else {
		err = filelock.RLock(f)
	}
	if err != nil {
		return nil, err
	}

	if flag&os.O_TRUNC == os.O_TRUNC {
		// The documentation for os.O_TRUNC says “if possible, truncate file when
		// opened”, so ignore errors from the Truncate call.
		f.Truncate(0)
	}
	return f, nil
}

func closeFile(f *os.File) error {
	// Since locking syscalls operate on file descriptors, we must unlock the file
	// while the descriptor is still valid — that is, before the file is closed —
	// and avoid unlocking files that are already closed.
	unlockErr := filelock.Unlock(f)

	if closeErr := f.Close(); closeErr != nil {
		return closeErr
	}
	return unlockErr
}
