// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Malloc small size classes.
//
// See malloc.h for overview.
//
// The size classes are chosen so that rounding an allocation
// request up to the next size class wastes at most 12.5% (1.125x).
//
// Each size class has its own page count that gets allocated
// and chopped up when new objects of the size class are needed.
// That page count is chosen so that chopping up the run of
// pages into objects of the given size wastes at most 12.5% (1.125x)
// of the memory.  It is not necessary that the cutoff here be
// the same as above.
//
// The two sources of waste multiply, so the worst possible case
// for the above constraints would be that allocations of some
// size might have a 26.6% (1.266x) overhead.
// In practice, only one of the wastes comes into play for a
// given size (sizes < 512 waste mainly on the round-up,
// sizes > 512 waste mainly on the page chopping).
//
// TODO(rsc): Compute max waste for any given size.

package runtime

import (
	_core "runtime/internal/core"
	_gc "runtime/internal/gc"
	_maps "runtime/internal/maps"
	_sched "runtime/internal/sched"
)

// Returns size of the memory block that mallocgc will allocate if you ask for the size.
func roundupsize(size uintptr) uintptr {
	if size < _core.MaxSmallSize {
		if size <= 1024-8 {
			return uintptr(_gc.Class_to_size[_maps.Size_to_class8[(size+7)>>3]])
		} else {
			return uintptr(_gc.Class_to_size[_maps.Size_to_class128[(size-1024+127)>>7]])
		}
	}
	if size+_core.PageSize < size {
		return size
	}
	return _sched.Round(size, _core.PageSize)
}
