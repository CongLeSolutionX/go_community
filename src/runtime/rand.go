// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"internal/goarch"
	"runtime/internal/math"
	"unsafe"
)

type chacha8 struct {
	seed    [32]byte
	buf     [32]uint64
	counter uint64
	n       int
}

//go:nosplit
func fastrand() uint32 {
	mp := getg().m
	// Implement wyrand: https://github.com/wangyi-fudan/wyhash
	// Only the platform that math.Mul64 can be lowered
	// by the compiler should be in this list.
	if goarch.IsAmd64|goarch.IsArm64|goarch.IsPpc64|
		goarch.IsPpc64le|goarch.IsMips64|goarch.IsMips64le|
		goarch.IsS390x|goarch.IsRiscv64|goarch.IsLoong64 == 1 {
		mp.fastrand += 0xa0761d6478bd642f
		hi, lo := math.Mul64(mp.fastrand, mp.fastrand^0xe7037ed1a0b428db)
		return uint32(hi ^ lo)
	}

	// Implement xorshift64+: 2 32-bit xorshift sequences added together.
	// Shift triplet [17,7,16] was calculated as indicated in Marsaglia's
	// Xorshift paper: https://www.jstatsoft.org/article/view/v008i14/xorshift.pdf
	// This generator passes the SmallCrush suite, part of TestU01 framework:
	// http://simul.iro.umontreal.ca/testu01/tu01.html
	t := (*[2]uint32)(unsafe.Pointer(&mp.fastrand))
	s1, s0 := t[0], t[1]
	s1 ^= s1 << 17
	s1 = s1 ^ s0 ^ s1>>7 ^ s0>>16
	t[0], t[1] = s0, s1
	return s0 + s1
}

//go:nosplit
func fastrandn(n uint32) uint32 {
	// This is similar to fastrand() % n, but faster.
	// See https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32(uint64(fastrand()) * uint64(n) >> 32)
}

func chacha8block(counter uint64, seed *[32]byte, blocks *[32]uint64)

//go:nosplit
func fastrand64() uint64 {
	mp := getg().m
	c := &mp.chacha8
	if c.n == 0 {
		mp.locks++
		chacha8block(c.counter, &c.seed, &c.buf)
		c.counter += 4
		c.n = len(c.buf)
		mp.locks--
	}
	c.n--
	return c.buf[^c.n&31]
}

func fastrandu() uint {
	if goarch.PtrSize == 4 {
		return uint(fastrand())
	}
	return uint(fastrand64())
}

//go:linkname rand_fastrand64 math/rand.fastrand64
func rand_fastrand64() uint64 { return fastrand64() }

//go:linkname randv2_fastrand64 math/rand/v2.fastrand64
func randv2_fastrand64() uint64 { return fastrand64() }

//go:linkname sync_fastrandn sync.fastrandn
func sync_fastrandn(n uint32) uint32 { return fastrandn(n) }

//go:linkname net_fastrandu net.fastrandu
func net_fastrandu() uint { return fastrandu() }

//go:linkname os_fastrand os.fastrand
func os_fastrand() uint32 { return fastrand() }
