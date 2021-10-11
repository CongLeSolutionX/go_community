// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Support for test coverage with redesigned coverage implementation.

package testing

import (
	"fmt"
	"internal/goexperiment"
	"os"
)

// NOTE: This type is internal to the testing infrastructure and may change.
// It is not covered (yet) by the Go 1 compatibility guidelines.
type Cover2 struct {
	Mode     string
	TearDown func(coverprofile string, gocoverdir string) (string, error)
}

var cover2 Cover2

// RegisterCover2 passes coverage setup information in 'c' to the testing
// package during "go test -cover" execution.
// NOTE: This function is internal to the testing infrastructure and may change.
// It is not covered (yet) by the Go 1 compatibility guidelines.
func RegisterCover2(c Cover2) {
	cover2 = c
}

// coverReport2 invokes a callback in _testmain.go that will
// emit coverage data at the point where test execution is complete,
// for "go test -cover" runs.
func coverReport2() {
	if !goexperiment.CoverageRedesign {
		panic("unexpected")
	}
	if errmsg, err := cover2.TearDown(*coverProfile, *gocoverdir); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", errmsg, err)
		os.Exit(2)
	}
}

// coverMode returns the selected coverage mode, if
// "-cover" is in effect for the current test.
func coverMode() string {
	if goexperiment.CoverageRedesign {
		return cover2.Mode
	}
	return cover.Mode
}
