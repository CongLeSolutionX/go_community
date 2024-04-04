// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unique

import (
	"internal/abi"
	"runtime"
	"testing"
)

// Set up special types. Because the internal maps are sharded by type,
// this will ensure that we're not overlapping with other tests.
type testString string
type testIntArr [4]int
type testEface any

func TestHandle(t *testing.T) {
	foo0 := Make[testString]("foo")
	bar0 := Make[testString]("bar")
	empty0 := Make[testString]("")

	foo1 := Make[testString]("foo")
	bar1 := Make[testString]("bar")
	empty1 := Make[testString]("")

	i0 := Make[testIntArr]([4]int{7, 77, 777, 7777})
	nilEface0 := Make[testEface](nil)

	i1 := Make[testIntArr]([4]int{7, 77, 777, 7777})
	nilEface1 := Make[testEface](nil)

	if foo0.Value() != foo1.Value() {
		t.Error("foo0.Value != foo1.Value")
	}
	if foo0.Value() != "foo" {
		t.Error("foo0.Value not foo")
	}
	if foo0 != foo1 {
		t.Error("foo0 != foo1")
	}

	if bar0.Value() != bar1.Value() {
		t.Error("bar0.Value != bar1.Value")
	}
	if bar0.Value() != "bar" {
		t.Error("bar0.Value not bar")
	}
	if bar0 != bar1 {
		t.Error("bar0 != bar1")
	}

	if i0.Value() != i1.Value() {
		t.Error("i0.Value != i1.Value")
	}
	if i0.Value() != [4]int{7, 77, 777, 7777} {
		t.Error("i1.Value not [4]int{7, 77, 777, 7777}")
	}
	if i0 != i1 {
		t.Error("i0 != i1")
	}

	if empty0.Value() != empty1.Value() {
		t.Error("empty0.Value != empty1.Value")
	}
	if empty0.Value() != "" {
		t.Error("empty0.Value not empty")
	}
	if empty0 != empty1 {
		t.Error("empty0 != empty1")
	}

	if nilEface0.Value() != nilEface1.Value() {
		t.Error("nilEface0.Value != nilEface1.Value")
	}
	if nilEface0.Value() != nil {
		t.Error("nilEface0.Value not nil")
	}
	if nilEface0 != nilEface1 {
		t.Error("nilEface0 != nilEface1")
	}

	drainMaps(t)
	checkMapsFor(t, testString("foo"))
	checkMapsFor(t, testString("bar"))
	checkMapsFor(t, testString(""))
	checkMapsFor(t, testIntArr([4]int{7, 77, 777, 7777}))
	checkMapsFor(t, testEface(nil))
}

// drainMaps ensures that the internal maps are drained.
func drainMaps(t *testing.T) {
	t.Helper()

	wait := make(chan struct{}, 1)

	// Set up a one-time notification for the next time the cleanup runs.
	// Note: this will only run if there's no other active cleanup, so
	// we can be sure that the next time cleanup runs, it'll see the new
	// notification.
	cleanupMu.Lock()
	cleanupNotify = append(cleanupNotify, func() {
		select {
		case wait <- struct{}{}:
		default:
		}
	})

	// Two GC cycles are necessary to clean up unique pointers.
	runtime.GC()
	runtime.GC()
	cleanupMu.Unlock()

	// Wait until cleanup runs.
	<-wait
}

func checkMapsFor[T comparable](t *testing.T, value T) {
	// Manually load the value out of the map.
	typ := abi.TypeOf(value)
	a, ok := uniqueMaps.Load(typ)
	if !ok {
		return
	}
	m := a.(*uniqueMap[T])
	wp, ok := m.Load(value)
	if !ok {
		return
	}
	if !wp.Nil() {
		t.Errorf("value %v still referenced a handle (or tiny block?) ", value)
		return
	}
	t.Errorf("failed to drain internal maps of %v", value)
}
