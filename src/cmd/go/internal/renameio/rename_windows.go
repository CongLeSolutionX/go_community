// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renameio

import (
	"internal/syscall/windows"
	"os"
	"syscall"
	"time"
	_ "unsafe"
)

//go:linkname os_fixLongPath os.fixLongPath
func os_fixLongPath(path string) string

// rename is like os.Rename, but uses ReplaceFile instead of MoveFileEx.
//
// TODO(bcmills): For Go 1.14, try changing os.Rename itself to do this.
func rename(oldpath, newpath string) error {
	from, err := syscall.UTF16PtrFromString(os_fixLongPath(oldpath))
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(os_fixLongPath(newpath))
	if err != nil {
		return err
	}

	var start time.Time
	for {
		err = windows.MoveFileEx(from, to, windows.MOVEFILE_WRITE_THROUGH)
		if err == nil {
			break
		}

		if os.IsExist(err) {
			replaceErr := windows.ReplaceFile(to, from, nil, windows.REPLACEFILE_IGNORE_MERGE_ERRORS|windows.REPLACEFILE_IGNORE_ACL_ERRORS, 0, 0)
			// If ReplaceFile returns a “cannot file the file specified” error, the
			// error message from MoveFileEx is likely to be more helpful.
			if !os.IsNotExist(replaceErr) {
				err = replaceErr
			}
		}

		if !isAccessDeniedError(err) {
			break
		}

		// Windows seems to occasionally trigger spurious "Access is denied" errors
		// here (see golang.org/issue/31247). We're not sure why. It's probably
		// worth a little extra latency to avoid propagating the spurious errors.
		if start.IsZero() {
			start = time.Now()
		} else if time.Since(start) >= 500*time.Millisecond {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if err != nil {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: err}
	}
	return nil
}
