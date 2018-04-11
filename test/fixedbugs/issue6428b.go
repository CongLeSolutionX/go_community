// compile

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

import . "testing"

type S map[interface{}]int

func main() {
	_ = S{(*T).Run: 0} // using T here is a use of testing.T
}
