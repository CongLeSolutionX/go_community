// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package par

import "testing"

func TestCache(t *testing.T) {
	var cache Cache[int, int]

	n := 1
	check := func(key, want int, wantRan bool) {
		oldN := n
		v, err := cache.Do(key, func() (int, error) { n++; return n, nil })
		if err != nil {
			t.Fatal(err)
		}
		ran := oldN != n
		if ran && !wantRan {
			t.Fatalf("cache.Do(%d): ran f again", key)
		} else if !ran && wantRan {
			t.Fatalf("cache.Do(%d): did not run f", key)
		}
		if v != want {
			t.Fatalf("cache.Do(%d): got %d, want %d", key, v, want)
		}
	}

	check(1, 2, true)
	check(1, 2, false)
	check(2, 3, true)
	check(1, 2, false)
}
