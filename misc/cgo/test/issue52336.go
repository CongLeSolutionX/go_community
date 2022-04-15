// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !ppc64le

package cgotest

import "testing"

// Issue 52336: generate ELFv2 ABI save/restore functions from go linker.
//              These are calls are made when compiling with -Os on ppc64le.

func test52336(t *testing.T) {
	t.Skipf("Test not needed on non-ppc64le target")
}
