// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This file is inserted into the build, overriding the 'ignore' constraint
// above, by cmd/go when running 'go test'. This means that the testing package
// flags are automatically initialized in test contexts while allowing non-test
// code to import testing without registering those flags.
//
// In Go 1.12 and earlier, the testing flags were unconditionally registered at
// initialization. Code in the wild grew to depend on the behavior: some tests
// call flag.Parse at initialization; worse, some libraries that may be imported
// by tests do the same thing. This workaround avoids breaking such code.
//
// See golang.org/issue/21051 and golang.org/issue/31859 for details.

package testing

func init() { Init() }
