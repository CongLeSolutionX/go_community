// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file shall contain regression reproductions to lock-in
// and prevent future slip-ups.
package testing_test

import "testing"

// Reported in https://go.dev/issue/64402
func TestRegressionRefactorDeadlockIssue64402(t *testing.T) {
	t.Run("outer", func(*testing.T) {
		t.Run("inner", func(t *testing.T) {
			t.Log("Hello World!")
		})
	})
}
