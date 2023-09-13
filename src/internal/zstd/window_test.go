// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zstd

import (
	"fmt"
	"testing"
)

func TestWindowSave(t *testing.T) {
	w := window{size: 5}

	for _, tc := range []struct {
		save []string
		data string
		off  int
	}{
		{
			save: []string{""},
			data: "",
			off:  0,
		},
		{
			save: []string{"foo"},
			data: "foo",
			off:  0,
		},
		{
			save: []string{"hello"},
			data: "hello",
			off:  0,
		},
		{
			save: []string{"hello!"},
			data: "ello!",
			off:  0,
		},
		{
			save: []string{"hello", "!"},
			data: "!ello",
			off:  1,
		},
		{
			save: []string{"foo", "b"},
			data: "foob",
			off:  0,
		},
		{
			save: []string{"foo", "ba"},
			data: "fooba",
			off:  0,
		},
		{
			save: []string{"foo", "bar"},
			data: "rooba",
			off:  1,
		},
		{
			save: []string{"foo", "bar", "baz"},
			data: "rbaza",
			off:  4,
		},
		{
			save: []string{"foo", "bar", "baz", "qux"},
			data: "uxazq",
			off:  2,
		},
	} {
		t.Run(fmt.Sprintf("%v", tc.save), func(t *testing.T) {
			w.reset()

			for _, s := range tc.save {
				w.save([]byte(s))
			}

			if w.len() != uint32(len(tc.data)) {
				t.Errorf("wrong data length: got: %d, want: %d", w.len(), len(tc.data))
			}
			if string(w.data) != tc.data {
				t.Errorf("wrong data: got: %s, want: %s", string(w.data), tc.data)
			}
			if w.off != tc.off {
				t.Errorf("wrong offset: got: %d, want: %d", w.off, tc.off)
			}
		})
	}
}

func TestWindowAppendTo(t *testing.T) {
	w := window{size: 5}

	for _, tc := range []struct {
		save     []string
		from, to uint32
		expected string
	}{
		{},
		{
			save:     []string{"foo"},
			from:     0,
			to:       2,
			expected: "fo",
		},
		{
			save:     []string{"hello"},
			from:     1,
			to:       4,
			expected: "ell",
		},
		{
			save:     []string{"hello"},
			from:     0,
			to:       5,
			expected: "hello",
		},
		{
			save:     []string{"hello", "!"},
			from:     0,
			to:       5,
			expected: "ello!",
		},
		{
			save:     []string{"hello!"},
			from:     0,
			to:       5,
			expected: "ello!",
		},
		{
			save:     []string{"foo", "bar"},
			from:     0,
			to:       0,
			expected: "",
		},
		{
			save:     []string{"foo", "bar"},
			from:     0,
			to:       5,
			expected: "oobar",
		},
		{
			save:     []string{"foo", "bar"},
			from:     1,
			to:       4,
			expected: "oba",
		},
		{
			save:     []string{"foo", "bar"},
			from:     3,
			to:       5,
			expected: "ar",
		},
		{
			save:     []string{"foo", "bar"},
			from:     4,
			to:       5,
			expected: "r",
		},
	} {
		t.Run(fmt.Sprintf("%v %d %d", tc.save, tc.from, tc.to), func(t *testing.T) {
			w.reset()

			for _, s := range tc.save {
				w.save([]byte(s))
			}

			got := string(w.appendTo(nil, tc.from, tc.to))
			if got != tc.expected {
				t.Errorf("got: %q, want: %q", got, tc.expected)
			}
		})
	}
}
