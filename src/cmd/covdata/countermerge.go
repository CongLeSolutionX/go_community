// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"internal/coverage"
	"internal/coverage/decodemeta"
	"math"
)

// counterMerge provides state and methods to help manage the process
// of merging together counter data for functions. All of the various
// tools (including subtract, textfmt, etc) may have to perform
// implicit merging of counters if the selected input directories have
// more than one counter data file for a given program.
type counterMerge struct {
	cmode    coverage.CounterMode
	cgran    coverage.CounterGranularity
	overflow bool
}

func (cm *counterMerge) mergeCounters(dst, src []uint32) {
	if len(src) != len(dst) {
		panic(fmt.Sprintf("MergeCounters: len(dst)=%d len(src)=%d", len(dst), len(src)))
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

func (cm *counterMerge) resetModeAndGranularity() {
	cm.cmode = coverage.CtrModeInvalid
	cm.cgran = coverage.CtrGranularityInvalid
}

// setModeAndGranularity records the counter mode and granularity
// for the current merge. In the specific case of -pcombine merges,
// where we're combining data from more than one meta-data file,
// it also checks for mode/granularity clashes.
func (cm *counterMerge) setModeAndGranularity(mdf string, mfr *decodemeta.CoverageMetaFileReader) {
	// Collect counter mode and granularity so as to detect clashes.
	newgran := mfr.CounterGranularity()
	newmode := mfr.CounterMode()
	if cm.cmode != coverage.CtrModeInvalid {
		if cm.cmode != newmode {
			fatal("counter mode clash while reading meta-data file %s: previous file had %s, new file has %s", mdf, cm.cmode.String(), newmode.String())
		}
		if cm.cgran != newgran {
			fatal("counter granularity clash while reading meta-data file %s: previous file had %s, new file has %s", mdf, cm.cgran.String(), newgran.String())
		}
	}
	cm.cmode = newmode
	cm.cgran = newgran
}
