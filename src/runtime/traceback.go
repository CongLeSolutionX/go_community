// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_lock "runtime/internal/lock"
)

var gScanStatusStrings = [...]string{
	0:               "scan",
	_lock.Grunnable: "scanrunnable",
	_lock.Grunning:  "scanrunning",
	_lock.Gsyscall:  "scansyscall",
	_lock.Gwaiting:  "scanwaiting",
	_lock.Gdead:     "scandead",
	_lock.Genqueue:  "scanenqueue",
}
