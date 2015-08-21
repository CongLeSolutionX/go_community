// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package base

// Keep a cached value to make gotraceback fast,
// since we call it on every call to gentraceback.
// The cached value is a uint32 in which the low bit
// is the "crash" setting and the top 31 bits are the
// gotraceback value.
var Traceback_cache uint32 = 2 << 1

// The GOTRACEBACK environment variable controls the
// behavior of a Go program that is crashing and exiting.
//	GOTRACEBACK=0   suppress all tracebacks
//	GOTRACEBACK=1   default behavior - show tracebacks but exclude runtime frames
//	GOTRACEBACK=2   show tracebacks including runtime frames
//	GOTRACEBACK=crash   show tracebacks including runtime frames, then crash (core dump etc)
//go:nosplit
func gotraceback(crash *bool) int32 {
	_g_ := Getg()
	if crash != nil {
		*crash = false
	}
	if _g_.M.Traceback != 0 {
		return int32(_g_.M.Traceback)
	}
	if crash != nil {
		*crash = Traceback_cache&1 != 0
	}
	return int32(Traceback_cache >> 1)
}

// Holds variables parsed from GODEBUG env var,
// except for "memprofilerate" since there is an
// existing int var for that value, which may
// already have an initial value.
var Debug struct {
	Allocfreetrace    int32
	Efence            int32
	Gccheckmark       int32
	Gcpacertrace      int32
	Gcshrinkstackoff  int32
	Gcstackbarrieroff int32
	Gcstoptheworld    int32
	Gctrace           int32
	Invalidptr        int32
	Sbrk              int32
	Scavenge          int32
	Scheddetail       int32
	Schedtrace        int32
	Wbshadow          int32
}

// Poor mans 64-bit division.
// This is a very special function, do not use it if you are not sure what you are doing.
// int64 division is lowered into _divv() call on 386, which does not fit into nosplit functions.
// Handles overflow in a time-specific manner.
//go:nosplit
func Timediv(v int64, div int32, rem *int32) int32 {
	res := int32(0)
	for bit := 30; bit >= 0; bit-- {
		if v >= int64(div)<<uint(bit) {
			v = v - (int64(div) << uint(bit))
			res += 1 << uint(bit)
		}
	}
	if v >= int64(div) {
		if rem != nil {
			*rem = 0
		}
		return 0x7fffffff
	}
	if rem != nil {
		*rem = int32(v)
	}
	return res
}

// Helpers for Go. Must be NOSPLIT, must only call NOSPLIT functions, and must not block.

//go:nosplit
func Acquirem() *M {
	_g_ := Getg()
	_g_.M.Locks++
	return _g_.M
}

//go:nosplit
func Releasem(mp *M) {
	_g_ := Getg()
	mp.Locks--
	if mp.Locks == 0 && _g_.Preempt {
		// restore the preemption request in case we've cleared it in newstack
		_g_.Stackguard0 = StackPreempt
	}
}
