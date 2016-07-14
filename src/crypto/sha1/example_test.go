// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sha1_test

import (
	"crypto/sha1"
	"fmt"
	"io"
)

func ExampleNew() {
	h := sha1.New()
	io.WriteString(h, "His money is twice tainted:")
	io.WriteString(h, " 'taint yours and 'taint mine.")
	fmt.Printf("%x", h.Sum(nil))
	// Output: 597f6a540010f94c15d71806a99a2c8710e747bd
}

func ExampleSum() {
	data := []byte("This page intentionally left blank.")
	fmt.Printf("%x", sha1.Sum(data))
	// Output: af064923bbf2301596aac4c273ba32178ebc4a96
}
