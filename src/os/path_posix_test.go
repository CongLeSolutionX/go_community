// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !plan9

package os_test

import (
	"syscall"
)

// Declare a local EROFS so that the symbol is always
// present and can be referenced by tests. See
// path_plan9_test.go for details.
const EROFS = syscall.EROFS
