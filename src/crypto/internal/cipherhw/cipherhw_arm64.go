// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build arm64,!gccgo,!appengine

package cipherhw

import "internal/cpu"

func AESGCMSupport() bool {
	return cpu.ARM64.HasAES && cpu.ARM64.HasPMULL
}
