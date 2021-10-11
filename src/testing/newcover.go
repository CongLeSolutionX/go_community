// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build goexperiment.coverageredesign
// +build goexperiment.coverageredesign

// Support for test coverage with redesigned coverage implementation.

package testing

// RegisterCover2 passes coverage setup information in 'c' to the testing
// package, and returns variable pointers to be used by the testmain
// package during "go test -cover" execution.
// NOTE: This function is internal to the testing infrastructure and may change.
// It is not covered (yet) by the Go 1 compatibility guidelines.
func RegisterCover2(c Cover) (**string, **string) {
	cover = c
	return &coverProfile, &gocoverdir
}
