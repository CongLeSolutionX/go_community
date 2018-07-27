// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu_test

import (
	. "internal/cpu"
	"runtime"
	"testing"
)

func TestARMminimalFeatures(t *testing.T) {
	switch runtime.GOOS {
	case "linux", "freebsd", "android":
	default:
		t.Skipf("%s/arm is not supported", runtime.GOOS)
	}
	if !ARM.HasSWP {
		t.Fatalf("HasSWP expected true, got false")
	}
	if !ARM.HasHALF {
		t.Fatalf("HasHALF expected true, got false")
	}
}
