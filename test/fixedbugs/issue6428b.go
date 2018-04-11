// compile

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import . "math"

type S map[float64]int

func main() {
	_ = S{Pi: 0} // using Pi here is a use of package math
}
