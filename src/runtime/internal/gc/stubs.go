// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import _ "unsafe"

import (
	_base "runtime/internal/base"
)

const SpAlign = 1*(1-_base.Goarch_arm64) + 16*_base.Goarch_arm64 // SP alignment: 1 normally, 16 for ARM64

//go:linkname time_now time.now
func time_now() (sec int64, nsec int32)

func Unixnanotime() int64 {
	sec, nsec := time_now()
	return sec*1e9 + int64(nsec)
}
