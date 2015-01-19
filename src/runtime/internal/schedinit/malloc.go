// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schedinit

import (
	_gc "runtime/internal/gc"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
)

// round size up to next size class
func Goroundupsize(size uintptr) uintptr {
	if size < _sched.MaxSmallSize {
		if size <= 1024-8 {
			return uintptr(_gc.Class_to_size[_maps.Size_to_class8[(size+7)>>3]])
		}
		return uintptr(_gc.Class_to_size[_maps.Size_to_class128[(size-1024+127)>>7]])
	}
	if size+_sched.PageSize < size {
		return size
	}
	return (size + _sched.PageSize - 1) &^ _sched.PageMask
}
