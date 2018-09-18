// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package invariant reports invariant violations robustly.
package invariant

import (
	"fmt"
	"sync"
)

func defaultAction(desc string) {
	if len(desc) > 0 {
		panic("invariant violated: " + desc)
	}
	panic("invariant violated")
}

var (
	mu        sync.Mutex
	reporters []func(string)
)

// ReportViolationsTo registers f as a callback to be invoked if an invariant check fails,
// instead of the default action and after any previous ReportTo callbacks.
//
// If any such callback panics, no further callbacks will be invoked for the
// violation. If all callbacks return normally, calls to Check or Checkf for
// violated invariants will return false.
func ReportViolationsTo(f func(desc string)) {
	Check(f != nil, "Argument to invariant.ReportViolationsTo must be non-nil.")
	mu.Lock()
	reporters = append(reporters, f)
	mu.Unlock()
}

// Check verifies that ok is true.
// If it is, Check returns true immediately.
//
// Otherwise, Check reports that the invariant, optionally described by
// fmt.Sprint(args...), was violated.
//
// If a hook has been registered by a call to ReportViolationsTo, that hook is
// invoked directly from Check.
// Otherwise, Check takes an unspecified action (normally resulting in a panic,
// test failure, or program termination) to report the violation.
// If the hook function returns, Check returns false.
func Check(ok bool, v ...interface{}) bool {
	if ok {
		return true
	}

	desc := fmt.Sprint(v...)

	mu.Lock()
	rs := reporters
	mu.Unlock()

	if len(rs) == 0 {
		defaultAction(desc)
	} else {
		for _, f := range rs {
			f(desc)
		}
	}
	return false
}

// Checkf is like Check, but formats the arguments using fmt.Sprintf instead of
// fmt.Sprint.
func Checkf(ok bool, format string, args ...interface{}) bool {
	if ok {
		return true
	}

	desc := fmt.Sprintf(format, args...)

	mu.Lock()
	rs := reporters
	mu.Unlock()

	if len(rs) == 0 {
		defaultAction(desc)
	} else {
		for _, f := range rs {
			f(desc)
		}
	}
	return false
}
