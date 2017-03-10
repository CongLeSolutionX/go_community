// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

const (
	// condition codes
	s390xFlagEQ = 1 << 3
	s390xFlagLT = 1 << 2
	s390xFlagGT = 1 << 1
	s390xFlagO  = 1 << 0
)
