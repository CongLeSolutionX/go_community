// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !without_s390x

package main

import "cmd/link/internal/s390x"

func init() {
	archMain["s390x"] = s390x.Main
}
