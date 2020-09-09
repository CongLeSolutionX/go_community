// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !goexperiment.regabi

package cpu

const (
	// Number of argument and return registers.
	// As of the current ABI, there are zero of each,
	// so once we start using registers in our ABI,
	// these constants should be specified for their
	// respective architectures.
	//
	// Return registers across all architectures
	// are equivalent to argument registers,
	// so it's always safe to allocate the same
	// space for both by just using the *ParamRegisters
	// constants.
	IntParamRegisters   = 0
	FloatParamRegisters = 0
	MaxFloatBytes       = 0
)
