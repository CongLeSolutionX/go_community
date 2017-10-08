// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains tests for the useless-assignment checker.

package testdata

import "math/rand"

type ST struct {
	x int
	l []int
}

func (s *ST) SetX(x int) {
	// Accidental self-assignment; it should be "s.x = x"
	x = x // ERROR "self-assignment of x to x"
	// Another mistake
	s.x = s.x // ERROR "self-assignment of s.x to s.x"

	// Bail on any calls to avoid false positivse
	s.l[0] = s.l[0] // ERROR "self-assignment of s.l\[0\] to s.l\[0\]"
	s.l[num()] = s.l[num()]
	rng := rand.New(rand.NewSource(0))
	s.l[rng.Intn(len(s.l))] = s.l[rng.Intn(len(s.l))]
}

func num() int { return 2 }
