// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64,!s390x

package elliptic

import (
	"math/big"
)

var (
	p256 p256Curve
)

func initP256Arch() {
	// Use default pure golang implementation
	p256 = p256Curve{p256Params}
}
