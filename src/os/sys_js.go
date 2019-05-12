// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build js,wasm

package os

import "syscall"

// supportsCloseOnExec reports whether the platform supports the
// O_CLOEXEC flag.
const supportsCloseOnExec = false

func sysProcAttrHasNonDefaultEnvironment(sys *syscall.SysProcAttr) bool {
	return false
}
