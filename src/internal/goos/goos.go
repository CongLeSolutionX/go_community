// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package goos contains GOOS-specific constants.
package goos

// The next line makes 'go generate' write the zgoos*.go files with
// per-OS information, including constants named Is$GOOS for every
// known GOOS. The constant is 1 on the current system, 0 otherwise;
// multiplying by them is useful for defining GOOS-specific constants.
//go:generate go run gengoos.go

// IsUnix reports whether goos is a unix system. This is the
// implementation of the "unix" build tag.
//
// If you add anything here, see matchtag in cmd/dist/build.go.
func IsUnix(goos string) bool {
	switch goos {
	case "aix", "android", "darwin", "dragonfly", "freebsd", "hurd",
		"illumos", "ios", "linux", "netbsd", "openbsd", "solaris":

		return true
	default:
		return false
	}
}
