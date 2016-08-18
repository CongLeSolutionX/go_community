// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package testenv provides information about what functionality
// is available in different testing environments run by the Go team.
//
// It is an internal package because these details are specific
// to the Go team's test setup (on build.golang.org) and not
// fundamental to tests in general.

// +build !windows

package testenv

import (
	"runtime"
)

func hasSymlink() bool {
	switch runtime.GOOS {
	case "android", "nacl", "plan9":
		return false
	}

	return true
}
