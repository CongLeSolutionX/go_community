// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hash

import (
	_core "runtime/internal/core"
)

const (
	HashSize = 1009
)

var (
	Hash [HashSize]*_core.Itab
)
