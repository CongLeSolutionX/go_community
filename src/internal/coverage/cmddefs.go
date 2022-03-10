// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

// These shared constants are used by both the cmd/cover tool (in
// hybrid instrumentation mode) and cmd/compile (in coverage "fixup" mode).
const MetaVarTag = "metavar"
const MetaHashTag = "metahash"
const MetaLenTag = "metalen"
const PkgIdVarTag = "pkgidvar"
const CounterModeTag = "countermode"
const CounterVarTag = "countervar"
const CounterGranularityTag = "countergranularity"
const CounterPrefixTag = "counterprefix"

// CoverPkgConfig is a bundle of information passed from the Go
// command to the cover command during "go build -cover" runs. The
// Go command creates and fills in a struct as below, then passes
// file containing the encoded JSON for the struct to the "cover"
// tool when instrumenting the source files in a Go package.
type CoverPkgConfig struct {
	// File into which cmd/cover should emit summary info
	// when instrumentation is complete.
	OutConfig string

	// Import path for the package being instrumented.
	PkgPath string

	// Package name.
	PkgName string

	// Package classification: one of
	//    "stdlib"  package is part of the Go standard library
	//    "none"    no go.mod in use
	//    "mainmod" Go cmd considers package to be in the main module
	//    "depmod"  go cmd considers this package to be a go.mod dependency
	PkgClassification string

	// Instrumentation granularity: one of "perfunc" or "perblock" (default)
	Granularity string

	// Module path for this package (empty if no go.mod in use)
	ModulePath string

	// Skip instrumentation for this package and emit only registration code.
	RegOnly bool
}
