// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import _ "unsafe"

// onExitHook is registered with the runtime as an exit hook
// for programs that are built with "-coverage". This function
// is not intended to be user-visible or user-callable.
func onExitHook() {
	emitCounterData()
}

// initHook is invoked from the main package "init" routine in
// programs built with "-cover". This function is intended to be
// called only by the compiler.
func initHook() {
	runOnNonZeroExit := true
	runtime_addExitHook(onExitHook, runOnNonZeroExit)
	emitMetaData()
}

//go:linkname runtime_addExitHook runtime.addExitHook
func runtime_addExitHook(f func(), runOnNonZeroExit bool)
