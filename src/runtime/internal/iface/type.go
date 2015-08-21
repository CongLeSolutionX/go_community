// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime type representation.

package iface

import (
	_base "runtime/internal/base"
)

type imethod struct {
	name    *string
	pkgpath *string
	_type   *_base.Type
}

type Interfacetype struct {
	typ  _base.Type
	Mhdr []imethod
}
