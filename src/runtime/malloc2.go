// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	_lock "runtime/internal/lock"
)

type mcachelist struct {
	list  *_lock.Mlink
	nlist uint32
}
