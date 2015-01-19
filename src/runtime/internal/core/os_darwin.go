// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package core

//go:noescape
func Sigprocmask(sig uint32, new, old *uint32)

//go:noescape
func sigaltstack(new, old *Stackt)
