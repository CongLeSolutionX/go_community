// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gen

type Uint8x64 struct {
	valAny
}

var kindUint8x64 = &kind{typ: "Uint8x64", reg: regClassZ}

func ConstUint8x64(c [64]uint8, name string) (y Uint8x64) {
	y.initOp(&op{op: "const", kind: y.kind(), c: c, name: name})
	return y
}

func (Uint8x64) kind() *kind {
	return kindUint8x64
}

func (Uint8x64) wrap(x *op) Uint8x64 {
	var y Uint8x64
	y.initOp(x)
	return y
}

func (x Uint8x64) Shuffle(shuf Uint8x64) (y Uint8x64) {
	if shuf.op.op == "const" {
		// TODO: There are often patterns we can take advantage of here. Sometimes
		// we can do a broadcast. Sometimes we can at least do a quadword
		// permutation instead of a full byte permutation.

		// Range check the shuffle
		for i, inp := range shuf.op.c.([64]uint8) {
			// 0xff is a special "don't care" value
			if !(inp == 0xff || inp < 64) {
				fatalf("shuffle[%d] = %d out of range [0, %d) or 0xff", i, inp, 64)
			}
		}
	}

	args := []*op{x.op, shuf.op}
	y.initOp(&op{op: "VPERMB", kind: y.kind(), args: args})
	return y
}

// TODO: The two-argument shuffle is a little weird. You almost want the
// receiver to be the shuffle and the two arguments to be the two inputs, but
// that's almost certainly *not* what you want for the single input shuffle.

func (x Uint8x64) Shuffle2(y Uint8x64, shuf Uint8x64) (z Uint8x64) {
	// Confusingly, the inputs are in the opposite order from what you'd expect.
	args := []*op{y.op, x.op, shuf.op}
	z.initOp(&op{op: "VPERMI2B", kind: z.kind(), args: args})
	return z
}

type Uint64x8 struct {
	valAny
}

var kindUint64x8 = &kind{typ: "Uint64x8", reg: regClassZ}

func ConstUint64x8(c [8]uint64, name string) (y Uint64x8) {
	y.initOp(&op{op: "const", kind: y.kind(), c: c, name: name})
	return y
}

func (Uint64x8) kind() *kind {
	return kindUint64x8
}

func (Uint64x8) wrap(x *op) Uint64x8 {
	var y Uint64x8
	y.initOp(x)
	return y
}

func (x Uint64x8) GF2P8Affine(y Uint8x64) (z Uint8x64) {
	z.initOp(&op{op: "VGF2P8AFFINEQB", kind: z.kind(), args: []*op{x.op, y.op}})
	return z
}
