// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package par

import "testing"

func TestCache(t *testing.T) {
	var cache Cache[int, int]

	n := 1
	v := cache.Do(1, func() int { n++; return n })
	if v != 2 {
		t.Fatalf("cache.Do(1) did not run f")
	}
	v = cache.Do(1, func() int { n++; return n })
	if v != 2 {
		t.Fatalf("cache.Do(1) ran f again!")
	}
	v = cache.Do(2, func() int { n++; return n })
	if v != 3 {
		t.Fatalf("cache.Do(2) did not run f")
	}
	v = cache.Do(1, func() int { n++; return n })
	if v != 2 {
		t.Fatalf("cache.Do(1) did not returned saved value from original cache.Do(1)")
	}
}
