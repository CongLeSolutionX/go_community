// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package subtle

import (
	"internal/cpu"
	"testing"
)

func TestWithDataIndependentTiming(t *testing.T) {
	if !cpu.ARM64.HasDIT {
		t.Skip("CPU does not support DIT")
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

func TestDITPanic(t *testing.T) {
	if !cpu.ARM64.HasDIT {
		t.Skip("CPU does not support DIT")
	}

	defer func() {
		e := recover()
		if e == nil {
			t.Fatal("didn't panic")
		}
		if ditEnabled() {
			t.Error("DIT still enabled after panic inside of WithDataIndependentTiming closure")
		}
	}()

	WithDataIndependentTiming(func() {
		if !ditEnabled() {
			t.Fatal("dit not enabled within WithDataIndependentTiming closure")
		}

		panic("bad")
	})
}
