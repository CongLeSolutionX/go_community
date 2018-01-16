// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtest_fatal

import "testing"

func TestSubtestfatal(t *testing.T) {
	t.Run("sub", func(tt *testing.T) {
		t.Fatalf("Fatal from subtest")
	})
}
