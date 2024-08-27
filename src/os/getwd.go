// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"sync"
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
//
// On Unix, this function might be able to return paths
// longer than supported by the underlying OS, but it is
// more expensive, and a resulting long path could not be
// used for any filesystem-related operations. If this is
// not desirable, use [syscall.Getwd] instead.
func Getwd() (dir string, err error) {
	return getwd()
}
