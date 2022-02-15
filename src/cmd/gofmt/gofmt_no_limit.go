// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(linux || darwin || freebsd || openbsd || netbsd || solaris || dragonfly || aix)

package main

import "runtime"

func openFilelimit() int {
	switch runtime.GOOS {
	case "windows":
		// NOTE: Windows support up to 1.6M
		n := (1 << 10) * runtime.GOMAXPROCS(0)
		if n > maxFileOpen {
			return maxFileOpen
		}
		return n
	default:
		// For now, this is arbitrarily set to 200, based on the observation that many
		// platforms default to a kernel limit of 256.
		// File descriptors opened from outside of this package are not tracked,
		// so this limit may be approximate.
		return 200
	}
}
