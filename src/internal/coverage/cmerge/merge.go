// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmerge

import (
	"fmt"
	"internal/coverage"
	"math"
)

// Merger provides state and methods to help manage the process of
// merging together coverage counter data for a given function, for
// tools that need to implicitly merge counter as they read multiple
// coverage counter data files.
type Merger struct {
	cmode    coverage.CounterMode
	cgran    coverage.CounterGranularity
	overflow bool
}

func (m *Merger) MergeCounters(dst, src []uint32) (error, bool) {
	if len(src) != len(dst) {
		return fmt.Errorf("merging counters: len(dst)=%d len(src)=%d", len(dst), len(src)), false
	}
	for i := 0; i < len(src); i++ {
		if m.cmode == coverage.CtrModeSet {
			if src[i] != 0 {
				dst[i] = 1
			}
		} else {
			dst[i] = m.SaturatingAdd(dst[i], src[i])
		}
	}
	defer func() {
		m.overflow = false
	}()
	return nil, m.overflow
}

func (m *Merger) SaturatingAdd(dst, src uint32) uint32 {
	d, s := uint64(dst), uint64(src)
	sum := d + s
	if uint64(uint32(sum)) != sum {
		if !m.overflow {
			m.overflow = true
		}
		return math.MaxUint32
	}
	return uint32(sum)
}

// SetModeAndGranularity records the counter mode and granularity for
// the current merge. In the specific case of merging across coverage
// data files from different binaries, where we're combining data from
// more than one meta-data file, we need to check for mode/granularity
// clashes.
func (cm *Merger) SetModeAndGranularity(mdf string, cmode coverage.CounterMode, cgran coverage.CounterGranularity) error {
	// Collect counter mode and granularity so as to detect clashes.
	if cm.cmode != coverage.CtrModeInvalid {
		if cm.cmode != cmode {
			return fmt.Errorf("counter mode clash while reading meta-data file %s: previous file had %s, new file has %s", mdf, cm.cmode.String(), cmode.String())
		}
		if cm.cgran != cgran {
			return fmt.Errorf("counter granularity clash while reading meta-data file %s: previous file had %s, new file has %s", mdf, cm.cgran.String(), cgran.String())
		}
	}
	cm.cmode = cmode
	cm.cgran = cgran
	return nil
}

func (cm *Merger) ResetModeAndGranularity() {
	cm.cmode = coverage.CtrModeInvalid
	cm.cgran = coverage.CtrGranularityInvalid
	cm.overflow = false
}

func (cm *Merger) Mode() coverage.CounterMode {
	return cm.cmode
}

func (cm *Merger) Granularity() coverage.CounterGranularity {
	return cm.cgran
}
