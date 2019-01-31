// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This program used to give an extra error that was both unnecessary and
// confusing:
//
//     cannot refer to unexported name types._

package main

import "go/types"

func _() {
        var _ types.const // ERROR "syntax error: unexpected const, expecting name"
}
