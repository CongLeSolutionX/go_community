// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

type hook struct {
	f                func()
	runOnNonZeroExit bool
}

var hooks []hook
var runningHooks bool

func addExitHook(f func(), runOnNonZeroExit bool) {
	hooks = append(hooks, hook{f: f, runOnNonZeroExit: runOnNonZeroExit})
}

func runHook(f func()) (caughtPanic bool) {
	defer func() {
		if x := recover(); x != nil {
			caughtPanic = true
		}
	}()
	f()
	return
}

func runExitHooks(exitCode int) {
	if runningHooks {
		throw("internal error: exit hook invoked exit")
	}
	runningHooks = true
	// Hooks are run in reverse order of registration: first hook added
	// is the last one run.
	for i := range hooks {
		h := hooks[len(hooks)-i-1]
		if exitCode != 0 && !h.runOnNonZeroExit {
			continue
		}
		if caughtPanic := runHook(h.f); caughtPanic {
			throw("internal error: exit hook invoked panic")
		}
	}
	hooks = nil
}
