// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !without_amd64

package main

import "cmd/link/internal/amd64"

func init() {
	archMain["amd64"] = amd64.Main
	archMain["amd64p32"] = amd64.Main
}
