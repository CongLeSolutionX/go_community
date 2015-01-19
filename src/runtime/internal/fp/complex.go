// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fp

func Isnan(f float64) bool { return f != f }

func Nan() float64 {
	var f float64 = 0
	return f / f
}

func Posinf() float64 {
	var f float64 = MaxFloat64
	return f * f
}
