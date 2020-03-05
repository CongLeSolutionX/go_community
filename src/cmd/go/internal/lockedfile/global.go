// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockedfile

import (
	"fmt"
	"os"
	"sync"

	"cmd/go/internal/lockedfile/internal/filelock"
)

// SetGlobalParent configures a specific file that will be locked implicitly
// prior to locking any other file, and unlocked only after all other files
// have been unlocked.
//
// SetGlobalParent must be called at most once, before any file is locked.
//
// SetGlobalParent is only needed on platforms that may produce spurious EDEADLK
// errors (see the NeedsGlobalParent function), and only if multiple other files
// may be locked and unlocked in independent goroutines.
//
// If the named file does not yet exist, the lockedfile package will create it
// (but not its parent directory) the first time it is needed.
func SetGlobalParent(name string) {
	if name == "" {
		panic(fmt.Sprintf("lockedfile.SetGlobalParent called with an empty filename"))
	}

	ok := false
	globalParent.once.Do(func() {
		globalParent.name = name
		globalParent.state = make(chan globalParentState, 1)
		globalParent.state <- globalParentState{}
		ok = true
	})

	if !ok {
		if globalParent.name == "" {
			panic(fmt.Sprintf("lockedfile.SetGlobalParent called after a file has been locked"))
		} else {
			panic(fmt.Sprintf("lockedfile.SetGlobalParent called twice"))
		}
	}
}

// NeedsGlobalParent reports whether the current platform may return spurious
// EDEADLK errors if a global parent lockfile is not configured via
// SetGlobalParent.
//
// Spurious EDEADLK errors arise on platforms that compute deadlock graphs at
// the process, rather than thread, level. Consider processes P and Q, with
// threads P.1, P.2, and Q.3. The following trace is NOT a deadlock, but will be
// reported as a deadlock on systems that consider only process granularity:
//
// 	P.1 locks file A.
// 	Q.3 locks file B.
// 	Q.3 blocks on file A.
// 	P.2 blocks on file B. (This is erroneously reported as a deadlock.)
// 	P.1 unlocks file A.
// 	Q.3 unblocks and locks file A.
// 	Q.3 unlocks files A and B.
// 	P.2 unblocks and locks file B.
// 	P.2 unlocks file B.
//
// A global parent lock prevents the spurious EDEADLK by preventing process Q
// from locking any files at all until process P has released all of its locks.
func NeedsGlobalParent() bool {
	return filelock.PlatformHasEDEADLKBug()
}

type globalParentState struct {
	active int      // the number of open locked files guarded by the global parent
	file   *os.File // the lockfile itself; nil when unlocked
}

var globalParent struct {
	once  sync.Once
	name  string
	state chan globalParentState // 1-buffered; non-nil if a global parent is configured
}

func incGlobalCount() error {
	globalParent.once.Do(func() {}) // Disallow further configuration.
	if globalParent.state == nil {
		return nil
	}

	st := <-globalParent.state
	defer func() { globalParent.state <- st }()

	if st.active == 0 {
		f, err := openFile(globalParent.name, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		st.file = f
	}
	st.active++
	return nil
}

func decGlobalCount() error {
	globalParent.once.Do(func() {}) // Disallow further configuration.
	if globalParent.state == nil {
		return nil
	}

	st := <-globalParent.state
	defer func() { globalParent.state <- st }()

	if st.active <= 0 {
		panic("lockedfile: inconsistent active lock count")
	}

	st.active--
	if st.active != 0 {
		return nil
	}

	err := closeFile(st.file)
	st.file = nil
	return err
}
