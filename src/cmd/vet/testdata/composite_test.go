// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains the test for untagged struct literals.

package testdata_test

type MyTestPkgStruct struct {
	X string
	Y string
}

var OkayNonTestTypeFromTest = []MyStruct{
	{"a", "b", "c"},
	{"0", "1", "2"},
}

var OkayTestTypeFromTest = []MyTestStruct{
	{"a", "b"},
	{"0", "1"},
}
