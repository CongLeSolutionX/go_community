// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.swaplencap
// +build !goexperiment.swaplencap

package goarch

const minStackAlignRegs = 1 // the minimum alignment allowed in registers
