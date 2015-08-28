// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iface

import (
	_base "runtime/internal/base"
)

//go:nosplit
func Gomcache() *_base.Mcache {
	return _base.Getg().M.Mcache
}
