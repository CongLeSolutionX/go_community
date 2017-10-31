// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package failfast

import "testing"

func TestA(t *testing.T) { t.Log("LOG: TestA"); t.Parallel() }
func TestB(t *testing.T) { t.Log("LOG: TestB"); t.Parallel() }
func TestC(t *testing.T) { t.Log("LOG: TestC"); t.Parallel() }

func TestFailingA(t *testing.T) {
	t.Parallel()
	t.Log("LOG: TestFailingA - FAIL")
	t.Fail()
}

func TestFailingNotParallelB(t *testing.T) {
	// Edge-case testing, mixing unparallel tests too
	t.Log("LOG: TestFailingNotParallelB - FAIL")
	t.Fail()
}

func TestNotParallelB(t *testing.T) {
	// Edge-case testing, mixing unparallel tests too
	t.Log("LOG: TestNotParallelB")
}

func TestZ(t *testing.T) { t.Log("LOG: TestZ"); t.Parallel() }

func TestFailingB(t *testing.T) {
	t.Parallel()
	t.Log("LOG: TestFailingB - FAIL")
	t.Fail()
}

func TestFailingSubtestsA(t *testing.T) {
	t.Parallel()
	// t.Log("LOG: TestFailingSubtestsA - FAIL") // TODO
	t.Run("TestFailingSubtestsA1", func(t *testing.T) {
		t.Log("LOG: TestFailingSubtestsA1 - FAIL")
		t.Fail()
	})
	t.Run("TestFailingSubtestsA2", func(t *testing.T) {
		t.Log("LOG: TestFailingSubtestsA2 - FAIL")
		t.Fail()
	})
}
