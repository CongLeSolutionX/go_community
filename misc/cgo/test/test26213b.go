// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo

package cgotest

import (
	"testing"

	"./issue26213/test26213cpp"
)

func test26213b(t *testing.T) {
	test26213cpp.Test26213cpp(t)
}
