// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renameio

import (
	"os"
	"time"
)

// rename is like os.Rename, but retries in case of ERROR_ACCESS_DENIED.
func rename(oldpath, newpath string) error {
	var (
		start time.Time
		err   error
	)
	for {
		err = os.Rename(oldpath, newpath)
		if err == nil {
			break
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
