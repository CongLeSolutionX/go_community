// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !race

// Dummy race detection API, used when not built with -race.

package iface

import (
	_base "runtime/internal/base"
	"unsafe"
)

func Raceacquire(addr unsafe.Pointer) { _base.Throw("race") }
