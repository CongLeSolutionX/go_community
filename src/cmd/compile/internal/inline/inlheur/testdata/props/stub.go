// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT (use 'go test -v -update-expected' instead.)
// See cmd/compile/internal/inline/inlheur/testdata/props/README.txt
// for more information on the format of this file.
// =^=^=

package stub

var M = map[string]int{}

// stub.go T_stub 18
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_stub() {
}

func ThisFunctionShouldBeIgnored(x int) {
	println(x)
}

// stub.go init.0 29
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func init() {
	ThisFunctionShouldBeIgnored(1)
}
