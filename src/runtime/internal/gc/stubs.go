// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import _ "unsafe"

//go:linkname time_now time.now
func time_now() (sec int64, nsec int32)
