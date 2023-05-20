// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gover

import (
	"runtime"
	"strings"
)

// TestVersion is initialized in the go command test binary
// to be $TESTGO_VERSION, to allow tests to override the
// go command's idea of its own version as returned by Local.
var TestVersion string

// Local returns the local Go version, the one implemented by this go command.
func Local() string {
	v := runtime.Version()
	if TestVersion != "" {
		v = TestVersion
	}
	if strings.HasPrefix(v, "go") {
		return strings.TrimPrefix(v, "go")
	}
	if strings.HasPrefix(v, "devel go1") {
		v = strings.TrimPrefix(v, "devel go")
		v, _, _ = strings.Cut(v, "-")
		return v
	}
	return "unknown"
}
