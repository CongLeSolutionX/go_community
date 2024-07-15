// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtle

import (
	"internal/cpu"
	"runtime"
	"testing"
)

func TestWithDataIndependentTiming(t *testing.T) {
	if !cpu.ARM64.HasDIT {
		t.Skipf("DIT is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	WithDataIndependentTiming(func() {
		if !ditEnabled() {
			t.Fatal("dit not enabled within WithDataIndependentTiming closure")
		}

		WithDataIndependentTiming(func() {
			if !ditEnabled() {
				t.Fatal("dit not enabled within nested WithDataIndependentTiming closure")
			}
		})

		if !ditEnabled() {
			t.Fatal("dit not enabled after return from nested WithDataIndependentTiming closure")
		}
	})

	if ditEnabled() {
		t.Fatal("dit not unset after returning from WithDataIndependentTiming closure")
	}
}
