// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Report a single error if we fail to convert a complex
// number to uint* or int*.

package p

func _() {
	_ = uint(-4 + 2i) // ERROR "cannot convert \-4 \+ 2i .type untyped complex. to type uint"
	_ = uint(-4 + 0i) // ERROR "constant -4 overflows uint"
	_ = int(-4 + 2i)  // ERROR "cannot convert \-4 \+ 2i .type untyped complex. to type int"
	_ = int(-4 + 0i)

	_ = float64(-4 + 2i) // ERROR "constant .\-4\+2i. truncated to real"
	_ = float64(-4 + 0i)

	s := []string{"a", "b", "c"}
	_ = s[-3.3] // ERROR "-3.3 truncated to integer" "invalid slice index \-3.3 .index must be non-negative."

	_ = 10 << -3.3 // ERROR "invalid negative shift count: \-3" "constant \-3.3 truncated to integer"
	_ = 10 << 3.0
	_ = 10 << 3.3       // ERROR "constant 3.3 truncated to integer"
	_ = 10 << (-4 + 2i) // ERROR "cannot use \-4 \+ 2i .type untyped complex. as type uint"

}
