// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import _ "unsafe"

// initHook is invoked from the main package "init" routine in
// programs built with "-cover". This function is intended to be
// called only by the compiler.
func initHook() {
	runOnNonZeroExit := true
	runtime_addExitHook(emitCounterData, runOnNonZeroExit)
	emitMetaData()
}

//go:linkname runtime_addExitHook runtime.addExitHook
func runtime_addExitHook(f func(), runOnNonZeroExit bool)
