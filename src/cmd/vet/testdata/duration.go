// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This file contains tests for the duration checker.

package testdata

import (
	"flag"
	"time"
)

var (
	_ = time.Duration(5)
	_ = time.Sleep(0)
	_ = time.Sleep(1 * 10)
	_ = time.Sleep(30) // ERROR "time.Duration without unit: 30"
	_ = time.Sleep(30 * time.Nanosecond)
	_ = flag.Duration("period", 10, "`period` in seconds") // ERROR "time.Duration without unit: 10"
	_ = flag.Duration("delay", 1000, "delay in nanoseconds")
)
