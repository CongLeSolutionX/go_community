// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type T1 struct { // ERROR "(?s)invalid recursive type: T1.*T1 refers to.*T2 refers to.*T1$"
	f2 T2
}

type T2 struct {
	f1 T1
}
