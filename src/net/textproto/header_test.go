// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package textproto

import (
	"reflect"
	"testing"
)

func TestMIMEHeaderSet(t *testing.T) {
	type toSet struct {
		key, val string
	}
	tests := []struct {
		base MIMEHeader
		sets []toSet
		exp  MIMEHeader
	}{
		{ // set nothing; nothing changes
			base: MIMEHeader{
				"Foo":  []string{"bar"},
				"Fizz": []string{"buzz", "baz"},
			},
			exp: MIMEHeader{
				"Foo":  []string{"bar"},
				"Fizz": []string{"buzz", "baz"},
			},
		},
		{ // overrides existing value and adds new canonicalized value
			base: MIMEHeader{
				"Foo":  []string{"bar"},
				"Fizz": []string{"buzz", "baz"},
			},
			sets: []toSet{
				{"Fizz", "Bot"},
				{"Fizz", "Bat"},
				{"text", "proto"},
			},
			exp: MIMEHeader{
				"Foo":  []string{"bar"},
				"Fizz": []string{"Bat"},
				"Text": []string{"proto"},
			},
		},
	}

	for i, test := range tests {
		for _, set := range test.sets {
			test.base.Set(set.key, set.val)
		}
		if !reflect.DeepEqual(test.base, test.exp) {
			t.Errorf("#%d: got %v != want %v", i, test.base, test.exp)
		}
	}
}

func BenchmarkMIMEHeaderSetReuse(b *testing.B) {
	h := make(MIMEHeader)
	for i := 0; i < b.N; i++ {
		h.Set("Foo", "bar")
	}
}

func BenchmarkMIMEHeaderSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		h := make(MIMEHeader)
		h.Set("Foo", "bar")
	}
}
