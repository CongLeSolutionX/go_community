// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmtsort_test

import (
	"fmt"
	"internal/fmtsort"
	"math"
	"reflect"
	"strings"
	"testing"
)

type sortTest struct {
	data  interface{} // Always a map.
	print string      // Printed result using our custom printer.
}

var sortTests = []sortTest{
	{
		map[int]string{7: "bar", -3: "foo"},
		"-3:foo 7:bar",
	},
	{
		map[uint8]string{7: "bar", 3: "foo"},
		"3:foo 7:bar",
	},
	{
		map[string]string{"7": "bar", "3": "foo"},
		"3:foo 7:bar",
	},
	{
		map[float64]string{7: "bar", -3: "foo", math.NaN(): "nan", math.Inf(0): "inf"},
		"NaN:nan -3:foo 7:bar +Inf:inf",
	},
	{
		map[complex128]string{7 + 2i: "bar2", 7 + 1i: "bar", -3: "foo", complex(math.NaN(), 0i): "nan", complex(math.Inf(0), 0i): "inf"},
		"(NaN+0i):nan (-3+0i):foo (7+1i):bar (7+2i):bar2 (+Inf+0i):inf",
	},
	{
		map[bool]string{true: "true", false: "false"},
		"false:false true:true",
	},
	{
		chanMap(),
		"CHAN0:0 CHAN1:1 CHAN2:2",
	},
	{
		pointerMap(),
		"PTR0:0 PTR1:1 PTR2:2",
	},
	{
		map[toy]string{toy{7, 2}: "72", toy{7, 1}: "71", toy{3, 4}: "34"},
		"{3 4}:34 {7 1}:71 {7 2}:72",
	},
	{
		map[[2]int]string{{7, 2}: "72", {7, 1}: "71", {3, 4}: "34"},
		"[3 4]:34 [7 1]:71 [7 2]:72",
	},
	{
		map[interface{}]string{7: "7", 4: "4", 3: "3", nil: "nil"},
		"<nil>:nil 3:3 4:4 7:7",
	},
}

func sprint(data interface{}) string {
	om := fmtsort.Sort(reflect.ValueOf(data))
	if om == nil {
		return "nil"
	}
	b := new(strings.Builder)
	for i, key := range om.Key {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteString(sprintKey(key))
		b.WriteRune(':')
		b.WriteString(fmt.Sprint(om.Value[i]))
	}
	return b.String()
}

// sprintKey formats a reflect.Value but gives reproducible values for some
// problematic types such as pointers. Note that it only does special handling
// for the troublesome types used in the test cases; it is not a general
// printer.
func sprintKey(key reflect.Value) string {
	switch str := key.Type().String(); str {
	case "*int":
		ptr := key.Interface().(*int)
		for i := range ints {
			if ptr == &ints[i] {
				return fmt.Sprintf("PTR%d", i)
			}
		}
		return "PTR???"
	case "chan int":
		c := key.Interface().(chan int)
		for i := range chans {
			if c == chans[i] {
				return fmt.Sprintf("CHAN%d", i)
			}
		}
		return "CHAN???"
	default:
		return fmt.Sprint(key)
	}
}

var (
	ints  [3]int
	chans = [3]chan int{make(chan int), make(chan int), make(chan int)}
)

func pointerMap() map[*int]string {
	m := make(map[*int]string)
	for i := 2; i >= 0; i-- {
		m[&ints[i]] = fmt.Sprint(i)
	}
	return m
}

func chanMap() map[chan int]string {
	m := make(map[chan int]string)
	for i := 2; i >= 0; i-- {
		m[chans[i]] = fmt.Sprint(i)
	}
	return m
}

type toy struct {
	A int // Exported.
	b int // Unexported.
}

func TestOrder(t *testing.T) {
	for _, test := range sortTests {
		got := sprint(test.data)
		if got != test.print {
			t.Errorf("%s: got %q, want %q", reflect.TypeOf(test.data), got, test.print)
		}
	}
}
