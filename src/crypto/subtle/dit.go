// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtle

import "runtime"

// WithDataIndependentTiming enables architecture specific features which ensure
// that the timing of specific instructions is independent of their inputs
// before executing f. On f returning it disables these features.
//
// WithDataIndependentTiming should only be used when f is written to make use
// of constant-time operations. WithDataIndependentTiming does not make
// variable-time code constant-time.
//
// WithDataIndependentTiming locks the current goroutine to the OS thread for
// the duration of f. Calls to WithDataIndependentTiming may be nested.
//
// On Arm64 processors with FEAT_DIT, WithDataIndependentTiming enables
// PSTATE.DIT. See https://developer.arm.com/documentation/ka005181/1-0/?lang=en.
//
// On all other architectures WithDataIndependentTiming executes f immediately
// with no other side-effects.
func WithDataIndependentTiming(f func()) {
	if !ditSupported {
		f()
		return
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	alreadyEnabled := enableDIT()

	// disableDIT is called in a deferred function so that if f panics we will
	// still disable DIT, in case the panic is recovered further up the stack.
	defer func() {
		if !alreadyEnabled {
			disableDIT()
		}
	}()

	f()
}
