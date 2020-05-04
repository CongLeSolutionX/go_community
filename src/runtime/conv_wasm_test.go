// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime_test

import (
	"testing"
)

var res int64
var ures uint64

func TestFloatTruncation(t *testing.T) {
	testdata := []struct {
		input      float64
		convInt64  int64
		convUInt64 uint64
		overflow   bool
	}{
		// max +- 1
		{
			input:      9223372036854775807,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		{
			input:      9223372036854775808,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		{
			input:      9223372036854775806,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		// trunc point +- 1
		{
			input:      9223372036854775295,
			convInt64:  9223372036854774784,
			convUInt64: 9223372036854774784,
		},
		{
			input:      9223372036854775296,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		{
			input:      9223372036854775294,
			convInt64:  9223372036854774784,
			convUInt64: 9223372036854774784,
		},
		// umax +- 1
		{
			input:      18446744073709551615,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		{
			input:      18446744073709551616,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		{
			input:    18446744073709551614,
			overflow: true,
		},
		// umax trunc +- 1
		{
			input:      18446744073709550591,
			convInt64:  -9223372036854775808,
			convUInt64: 18446744073709549568,
		},
		{
			input:      18446744073709550592,
			convInt64:  -9223372036854775808,
			convUInt64: 9223372036854775808,
		},
		{
			input:    18446744073709550590,
			overflow: true,
		},
	}
	for _, item := range testdata {
		// We just attempt a conversion without verifying the output
		// for those items which cannot be stored in a constant without overflowing.
		if item.overflow {
			res = int64(item.input)
			ures = uint64(item.input)
			continue
		}
		if got, want := int64(item.input), item.convInt64; got != want {
			t.Errorf("int64(%f): got %d, want %d", item.input, got, want)
		}
		if got, want := uint64(item.input), item.convUInt64; got != want {
			t.Errorf("uint64(%f): got %d, want %d", item.input, got, want)
		}
	}
}
