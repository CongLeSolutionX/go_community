// +build !without_mips64

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "cmd/link/internal/mips64"

func init() {
	archMain["mips64"] = mips64.Main
	archMain["mips64le"] = mips64.Main
}
