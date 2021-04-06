// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package amd64_test

import (
	"internal/testenv"
	"runtime"
	"testing"
)

type S struct {
	x [2]uint64
	p *uint64
	y [2]uint64
	q uint64
}

type M struct {
	x [8]uint64
	p *uint64
	y [8]uint64
	q uint64
}

type L struct {
	x [4096]uint64
	p *uint64
	y [4096]uint64
	q uint64
}

//go:noinline
func triggerZerorangeLarge(f, g, h uint64) (rv0 uint64) {
	ll := L{p: &f}
	da := f
	rv0 = f + g + h
	defer func(dl L, i uint64) {
		rv0 += dl.q + i
	}(ll, da)
	return rv0
}

//go:noinline
func triggerZerorangeMedium(f, g, h uint64) (rv0 uint64) {
	ll := M{p: &f}
	rv0 = f + g + h
	defer func(dm M, i uint64) {
		rv0 += dm.q + i
	}(ll, f)
	return rv0
}

//go:noinline
func triggerZerorangeSmall(f, g, h uint64) (rv0 uint64) {
	ll := S{p: &f}
	rv0 = f + g + h
	defer func(ds S, i uint64) {
		rv0 += ds.q + i
	}(ll, f)
	return rv0
}

// This test was created as a follow up to issue #45372, to insure
// that we have adequate coverage of the the compiler's
// architecture-specific "zerorange" function, which is invoked to
// zero out ambiguously live portions of the stack frame in certain
// specific circumstances.
//
// In the current compiler implementation, for zerorange to be
// invoked, we need to have an ambiguously live variable that needs
// zeroing. One way to trigger this is to have a function with an
// open-coded defer, where the opendefer function has an argument that
// contains a pointer (this is what's used below).
//
// At the moment this test doesn't do any specific checking for
// code sequence, or verification that things were properly set to zero,
// this seems as though it would be too tricky and would result
// in a "brittle" test.
//
// The small/medium/large scenarios below are inspired by the amd64
// implementation of zerorange, which generates different code
// depending on the size of the thing that needs to be zeroed out
// (I've verified at the time of the writing of this test that it
// exercises the various cases).
//
func TestZerorange45372(t *testing.T) {
	testenv.MustHaveGoBuild(t)

	// This test is intended to exercise the amd64-specific copy of zerorange.
	if runtime.GOARCH != "amd64" {
		t.Skip("amd64-only test")
	}

	if r := triggerZerorangeLarge(101, 303, 505); r != 1010 {
		t.Errorf("large: wanted %d got %d", 1010, r)
	}
	if r := triggerZerorangeMedium(101, 303, 505); r != 1010 {
		t.Errorf("medium: wanted %d got %d", 1010, r)
	}
	if r := triggerZerorangeSmall(101, 303, 505); r != 1010 {
		t.Errorf("small: wanted %d got %d", 1010, r)
	}

}
