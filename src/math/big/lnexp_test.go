// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package big

import (
	"testing"
)

func TestLn(t *testing.T) {
	// TODO(gri) fill this in
}

func TestExponential(t *testing.T) {
	for _, test := range []struct {
		x, want string
		prec    uint
	}{
		// TODO(gri) verify accuracy
		{"-Inf", "0", 64},
		{"-1", "0.36787944117144232158", 64},
		{"0", "1", 64},
		{"1", "2.7182818284590452354", 64},
		{"+Inf", "+Inf", 64},
	} {
		f, _, err := ParseFloat(test.x, 0, test.prec, ToNearestEven)
		if err != nil {
			t.Errorf("%v: %s", test, err)
			continue
		}

		var z Float
		got := z.exponential(f).Text('g', -1)
		if got != test.want {
			t.Errorf("%v: got %s; want %s", test, got, test.want)
		}
	}
}
