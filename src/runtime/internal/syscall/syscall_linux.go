// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package syscall provides the syscall primitives required for the runtime.
package syscall

// TODO(https://go.dev/issue/51087): This package is incomplete and currently
// only contains very minimal support for Linux.

func Syscall(num, a1, a2, a3, a4, a5, a6 uintptr) (r1, r2, errno uintptr)
