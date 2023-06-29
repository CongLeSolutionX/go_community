// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// DO NOT EDIT COMMENTS (use 'go test -v -update-expected' instead)

package stub

var M = map[string]int{}

// T_stub
// =====
// {"Flags":0,"RecvrParamFlags":null,"ReturnFlags":null}
// =-=-=
func T_stub() {
}

func ThisFunctionShouldBeIgnored() {
	panic("no")
}
