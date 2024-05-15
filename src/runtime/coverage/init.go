// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"internal/coverage/cfile"
	_ "unsafe"
)

// initHook is invoked from main.init in programs built with -cover.
// The call is emitted by the compiler.
func initHook(istest bool) {
	cfile.InitHook(istest)
}
