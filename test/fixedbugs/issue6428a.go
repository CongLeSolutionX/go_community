// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import . "testing" // ERROR "imported and not used"

type S struct {
	T int
}

func main() {
	_ = S{T: 0} // using T here is not a use of testing.T
}
