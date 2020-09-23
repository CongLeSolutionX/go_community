// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type T1 struct {
	f2 T2
}

type T2 struct { // ERROR "invalid recursive type: T2 refers to T1"
	f1 T1
}
