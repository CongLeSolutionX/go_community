// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"runtime"
	"testing"
)

// BenchmarkDelay is so we can know what the delay actually is
func BenchmarkDelay(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runtime.Delay()
	}
}
