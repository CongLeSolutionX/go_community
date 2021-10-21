// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"internal/coverage"
	"math"
)

type counterMerge struct {
	cmode    coverage.CounterMode
	overflow bool
}

func (cm *counterMerge) mergeCounters(dst, src []uint32) {
	if len(src) != len(dst) {
		panic("bad")
	}
	for i := 0; i < len(src); i++ {
		if cm.cmode == coverage.CtrModeSet {
			if src[i] != 0 {
				dst[i] = 1
			}
		} else {
			dst[i] = cm.saturatingAdd(dst[i], src[i])
		}
	}
}

func (cm *counterMerge) saturatingAdd(dst, src uint32) uint32 {
	d, s := uint64(dst), uint64(src)
	sum := d + s
	if uint64(uint32(sum)) != sum {
		if !cm.overflow {
			cm.overflow = true
			warn("uint32 overflow during counter merge")
		}
		return math.MaxUint32
	}
	return uint32(sum)
}
