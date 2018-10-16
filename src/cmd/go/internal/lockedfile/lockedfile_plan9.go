// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build plan9

package lockedfile

import (
	"math/rand"
	"os"
	"time"

	"cmd/go/internal/lockedfile/filelock"
)

func openFile(name string, flag int, perm os.FileMode) (_ *os.File, err error) {
	// Plan 9 uses a mode bit instead of explicit lock/unlock syscalls.
	//
	// Per http://man.cat-v.org/plan_9/5/stat: “Exclusive use files may be open
	// for I/O by only one fid at a time across all clients of the server. If a
	// second open is attempted, it draws an error.”
	//
	// So we can try to open a locked file, but if it fails we're on our own to
	// figure out when it becomes available. We'll use exponential backoff with
	// some jitter and an arbitrary limit of 500ms.

	nextSleep := 1 * time.Millisecond
	const maxSleep = 500 * time.Millisecond
	for {
		f, err := os.OpenFile(name, flag, perm|os.ModeExclusive)
		if err == nil {
			return f, nil
		}

		if !filelock.IsLocked(err) {
			return nil, err
		}

		time.Sleep(nextSleep)

		nextSleep += nextSleep
		if nextSleep > maxSleep {
			nextSleep = maxSleep
		}
		// Apply 10% jitter to avoid synchronizing collisions.
		nextSleep += time.Duration((0.1*rand.Float64() - 0.05) * float64(nextSleep))
	}
}

func closeFile(f *os.File) error {
	return f.Close()
}
