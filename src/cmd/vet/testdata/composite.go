// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the test for untagged struct literals.

package testdata

import (
	"flag"
	"go/scanner"
	"image"
	"unicode"

	"unknownpkg"
)

var Okay1 = []string{
	"Name",
	"Usage",
	"DefValue",
}

var Okay2 = map[string]bool{
	"Name":     true,
	"Usage":    true,
	"DefValue": true,
}

var Okay3 = struct {
	X string
	Y string
	Z string
}{
	"Name",
	"Usage",
	"DefValue",
}

type MyStruct struct {
	X string
	Y string
	Z string
}

var Okay4 = MyStruct{
	"Name",
	"Usage",
	"DefValue",
}

// Testing is awkward because we need to reference things from a separate package
// to trigger the warnings.

var BadStructLiteralUsedInTests = flag.Flag{ // ERROR "unkeyed fields"
	"Name",
	"Usage",
	nil, // Value
	"DefValue",
}

// SpecialCase is an (aptly named) slice of CaseRange to test issue 9171.
var GoodNamedSliceLiteralUsedInTests = unicode.SpecialCase{
	{Lo: 1, Hi: 2},
	unicode.CaseRange{Lo: 1, Hi: 2},
}
var badNamedSliceLiteralUsedInTests = unicode.SpecialCase{
	{1, 2},                  // ERROR "unkeyed fields"
	unicode.CaseRange{1, 2}, // ERROR "unkeyed fields"
}

// ErrorList is a slice, so no warnings should be emitted.
var scannerErrorListTest = scanner.ErrorList{nil, nil}

// Check whitelisted structs: if vet is run with --compositewhitelist=false,
// this line triggers an error.
var whitelistedP = image.Point{1, 2}

// Do not check type from unknown package.
// See issue 15408.
var x = unknownpkg.Foobar{"foo", "bar"}
