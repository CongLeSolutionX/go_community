// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maps

import (
	_core "runtime/internal/core"
)

var Size_to_class8 [1024/8 + 1]int8
var Size_to_class128 [(_core.MaxSmallSize-1024)/128 + 1]int8
