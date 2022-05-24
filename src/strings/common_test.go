// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings_test

import (
	"fmt"
	"strings"
	"testing"
)

func TestCommonPrefix(t *testing.T) {
	x := strings.Repeat("-", 1000000)
	for _, k := range []int{3, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 21, 78, 365, 1071, 16320, len(x)} {
		t.Run(fmt.Sprint(k), func(t *testing.T) {
			// y is a near copy of x with a '+' in one position.
			y := x
			if k < len(y) {
				y = x[:k] + "+" + x[k+1:]
			}

			got := strings.CommonPrefixLen(x, y)
			want := naive(x, y)
			if got != want {
				t.Fatalf("k=%d: got %d, want %d", k, got, want)
			}
		})
	}
}

// naive is a reference implementation of CommonPrefixLen.
func naive(a, b string) (i int) {
	for i = 0; i < len(a) && i < len(b) && a[i] == b[i]; i++ {
	}
	return i
}

func BenchmarkCommonPrefix(b *testing.B) {
	for _, k := range []int{3, 7, 17, 21, 78, 365, 1071, 16320, 1000000} {
		b.Run(fmt.Sprint(k), func(b *testing.B) {
			common := strings.Repeat("-", k)
			x := common + "x"
			y := common + "y"
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if strings.CommonPrefixLen(x, y) != k {
					b.Fatal("oops")
				}
			}
		})
	}
}
