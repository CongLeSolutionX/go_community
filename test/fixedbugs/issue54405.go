// errorcheck

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var x struct{ f *NotAType } // ERROR "undefined: NotAType"
var _ = x.f == nil

var y *NotAType // ERROR "undefined: NotAType"
var _ = y == nil
