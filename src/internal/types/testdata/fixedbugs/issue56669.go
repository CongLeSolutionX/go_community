// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f[P any](_ func(P)) {}

type myT struct{}

func _() {
	f(func(int) {})
	f(func(myT) {})
	// for a failing example, see the type-checker test TestIssue56669
}
