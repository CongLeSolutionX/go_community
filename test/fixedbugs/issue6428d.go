// compile

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type S map[int]int

func main() {
	var F int
	_ = S{F: 0} // using F here is a use of variable F
}
