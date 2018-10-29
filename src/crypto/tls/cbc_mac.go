// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import "crypto/subtle"

// rotateSecretOffset rotates b by n bytes, such that b[i] is set to what was
// previously b[(i+n)%len(b)]. The contents of b are undefined if n < 0 or
// n >= len(b)
func rotateSecretOffset(b []byte, n int) {
	scratch := make([]byte, len(b))
	// Rotate by powers of 2, conditionally including each, to add to n.
	// This gives a constant memory access pattern and an O(N log(N))
	// secret rotation.
	for i := uint(0); (1<<i) != 0 && (1<<i) < len(b); i++ {
		// Rotate b by 1<<i into scratch.
		copy(scratch, b[1<<i:])
		copy(scratch[len(b)-1<<i:], b)
		// mask is all zeros if this rotation should be kept and all
		// ones if it should be discarded.
		mask := byte(((n >> i) & 1) - 1)
		for j := range b {
			b[j] = (b[j] & mask) | (scratch[j] & ^mask)
		}
	}
}

// copySecretSlice returns a copy of in[start:start+length] where start is
// considered secret. It returns an undefined slice of the correct length if
// start is out of bounds or if start and length exceed 2**31-1.
func copySecretSlice(in []byte, start, length int) []byte {
	end := start + length
	ret := make([]byte, length)
	rotate := 0
	// Map in[i] to ret[i%length], discarding all values but those in
	// range. This copies the desired bytes, but rotated out of order.
	for i, v := range in {
		inRange := subtle.ConstantTimeLessOrEq(start, i)
		inRange &= ^subtle.ConstantTimeLessOrEq(end, i)
		ret[i%length] |= v & byte(^(inRange - 1))

		// This is equivalent to setting rotate to start % length, but
		// avoids a variable-time division.
		isStart := subtle.ConstantTimeEq(int32(i), int32(start))
		rotate |= (i % length) & ^(isStart - 1)
	}

	// Fix the rotation.
	rotateSecretOffset(ret, rotate)
	return ret
}
