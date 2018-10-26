// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ppc64 ppc64le

package cpu_test

import (
	. "internal/cpu"
	"testing"
)

func TestPPC64minimalFeatures(t *testing.T) {
	if !PPC64.IsPOWER8 {
		t.Fatalf("IsPOWER8 expected true, got false")
	}
}
