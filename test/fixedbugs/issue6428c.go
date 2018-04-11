// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type S struct {
	F int
}

func main() {
	var F int // ERROR "declared and not used"
	_ = S{F: 0} // using F here is not a use of variable F
}
