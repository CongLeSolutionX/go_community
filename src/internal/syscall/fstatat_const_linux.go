// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !386,!arm,!arm64

package syscall

import stdsyscall "syscall"

const fstatatNum = stdsyscall.SYS_NEWFSTATAT

// AT_SYMLINK_NOFOLLOW forbid following symlinks in Fstatat
const AT_SYMLINK_NOFOLLOW = 0x100
