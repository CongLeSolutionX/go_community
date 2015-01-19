// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_schedinit "runtime/internal/schedinit"
	_ "unsafe"
)

var tls0 [8]uintptr // available storage for m0's TLS; not necessarily used; opaque to GC

func makeStringSlice(n int) []string {
	return make([]string, n)
}

//go:linkname syscall_runtime_envs syscall.runtime_envs
func syscall_runtime_envs() []string { return _schedinit.Envs }

//go:linkname os_runtime_args os.runtime_args
func os_runtime_args() []string { return _schedinit.Argslice }
