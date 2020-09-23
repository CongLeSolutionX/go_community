// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I1 = interface {
	I2
}

type I2 interface { // ERROR "(?s)invalid recursive type: I2.\tI2 refers to.\tI2$"
	I1
}
