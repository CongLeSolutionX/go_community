// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package linknamed provides some package-private method,
// function, and variable, for testing go:linkname.
package linknamed

type universe int64

func (u universe) answer() int64 {
	return int64(u) + 42
}

var singleton universe

func increment() {
	singleton += 1
}
