// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I1 = interface {
	I2
}

// BAD: "interface {}" in error message should be replaced by "I1"
type I2 interface { // ERROR "invalid recursive type: I2" "I2 refers to$" "interface {} refers to$" "I2$"
	I1
}
