// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package issue16153

// original test case
const (
	x uint8 = 255
	Y       = 256
)

// variations
const (
	x uint8 = 255
	Y
)

const (
	X int64 = iota
	Y       = 1
)

const (
	X int64 = iota
	Y
)
