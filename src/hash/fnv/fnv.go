// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fnv implements FNV-1 and FNV-1a, non-cryptographic hash functions
// created by Glenn Fowler, Landon Curt Noll, and Phong Vo.
// See
// https://en.wikipedia.org/wiki/Fowler-Noll-Vo_hash_function.
package fnv

import (
	"errors"
	"hash"
)

type (
	sum32   uint32
	sum32a  uint32
	sum64   uint64
	sum64a  uint64
	sum128  [2]uint64
	sum128a [2]uint64
)

const (
	offset32        = 2166136261
	offset64        = 14695981039346656037
	offset128Lower  = 0x62b821756295c58d
	offset128Higher = 0x6c62272e07bb0142
	prime32         = 16777619
	prime64         = 1099511628211
	prime128Lower   = 0x13b
	prime128Shift   = 24
)

// New32 returns a new 32-bit FNV-1 hash.Hash.
// Its Sum method will lay the value out in big-endian byte order.
func New32() hash.Hash32 {
	var s sum32 = offset32
	return &s
}

// New32a returns a new 32-bit FNV-1a hash.Hash.
// Its Sum method will lay the value out in big-endian byte order.
func New32a() hash.Hash32 {
	var s sum32a = offset32
	return &s
}

// New64 returns a new 64-bit FNV-1 hash.Hash.
// Its Sum method will lay the value out in big-endian byte order.
func New64() hash.Hash64 {
	var s sum64 = offset64
	return &s
}

// New64a returns a new 64-bit FNV-1a hash.Hash.
// Its Sum method will lay the value out in big-endian byte order.
func New64a() hash.Hash64 {
	var s sum64a = offset64
	return &s
}

// New128 returns a new 128-bit FNV-1 hash.Hash.
// Its Sum method will lay the value out in big-endian byte order.
func New128() hash.Hash {
	var s sum128
	s[0] = offset128Higher
	s[1] = offset128Lower
	return &s
}

// New128a returns a new 128-bit FNV-1a hash.Hash.
// Its Sum method will lay the value out in big-endian byte order.
func New128a() hash.Hash {
	var s sum128a
	s[0] = offset128Higher
	s[1] = offset128Lower
	return &s
}

func (s *sum32) Reset()   { *s = offset32 }
func (s *sum32a) Reset()  { *s = offset32 }
func (s *sum64) Reset()   { *s = offset64 }
func (s *sum64a) Reset()  { *s = offset64 }
func (s *sum128) Reset()  { s[0] = offset128Higher; s[1] = offset128Lower }
func (s *sum128a) Reset() { s[0] = offset128Higher; s[1] = offset128Lower }

func (s *sum32) Sum32() uint32  { return uint32(*s) }
func (s *sum32a) Sum32() uint32 { return uint32(*s) }
func (s *sum64) Sum64() uint64  { return uint64(*s) }
func (s *sum64a) Sum64() uint64 { return uint64(*s) }

func (s *sum32) Write(data []byte) (int, error) {
	hash := *s
	for _, c := range data {
		hash *= prime32
		hash ^= sum32(c)
	}
	*s = hash
	return len(data), nil
}

func (s *sum32a) Write(data []byte) (int, error) {
	hash := *s
	for _, c := range data {
		hash ^= sum32a(c)
		hash *= prime32
	}
	*s = hash
	return len(data), nil
}

func (s *sum64) Write(data []byte) (int, error) {
	hash := *s
	for _, c := range data {
		hash *= prime64
		hash ^= sum64(c)
	}
	*s = hash
	return len(data), nil
}

func (s *sum64a) Write(data []byte) (int, error) {
	hash := *s
	for _, c := range data {
		hash ^= sum64a(c)
		hash *= prime64
	}
	*s = hash
	return len(data), nil
}

func (s *sum128) Write(data []byte) (int, error) {
	for _, c := range data {
		// Compute the multiplication in 4 parts to simplify carrying
		s1l := (s[1] & 0xffffffff) * prime128Lower
		s1h := (s[1] >> 32) * prime128Lower
		s0l := (s[0]&0xffffffff)*prime128Lower + (s[1]&0xffffffff)<<prime128Shift
		s0h := (s[0]>>32)*prime128Lower + (s[1]>>32)<<prime128Shift
		// Carries
		s1h += s1l >> 32
		s0l += s1h >> 32
		s0h += s0l >> 32
		// Update the values
		s[1] = (s1l & 0xffffffff) + (s1h << 32)
		s[0] = (s0l & 0xffffffff) + (s0h << 32)
		s[1] ^= uint64(c)
	}
	return len(data), nil
}

func (s *sum128a) Write(data []byte) (int, error) {
	for _, c := range data {
		s[1] ^= uint64(c)
		// Compute the multiplication in 4 parts to simplify carrying
		s1l := (s[1] & 0xffffffff) * prime128Lower
		s1h := (s[1] >> 32) * prime128Lower
		s0l := (s[0]&0xffffffff)*prime128Lower + (s[1]&0xffffffff)<<prime128Shift
		s0h := (s[0]>>32)*prime128Lower + (s[1]>>32)<<prime128Shift
		// Carries
		s1h += s1l >> 32
		s0l += s1h >> 32
		s0h += s0l >> 32
		// Update the values
		s[1] = (s1l & 0xffffffff) + (s1h << 32)
		s[0] = (s0l & 0xffffffff) + (s0h << 32)
	}
	return len(data), nil
}

func (s *sum32) Size() int   { return 4 }
func (s *sum32a) Size() int  { return 4 }
func (s *sum64) Size() int   { return 8 }
func (s *sum64a) Size() int  { return 8 }
func (s *sum128) Size() int  { return 16 }
func (s *sum128a) Size() int { return 16 }

func (s *sum32) BlockSize() int   { return 1 }
func (s *sum32a) BlockSize() int  { return 1 }
func (s *sum64) BlockSize() int   { return 1 }
func (s *sum64a) BlockSize() int  { return 1 }
func (s *sum128) BlockSize() int  { return 1 }
func (s *sum128a) BlockSize() int { return 1 }

func (s *sum32) Sum(in []byte) []byte {
	v := uint32(*s)
	return append(in, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (s *sum32a) Sum(in []byte) []byte {
	v := uint32(*s)
	return append(in, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (s *sum64) Sum(in []byte) []byte {
	v := uint64(*s)
	return append(in, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (s *sum64a) Sum(in []byte) []byte {
	v := uint64(*s)
	return append(in, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32), byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (s *sum128) Sum(in []byte) []byte {
	return append(in,
		byte(s[0]>>56), byte(s[0]>>48), byte(s[0]>>40), byte(s[0]>>32), byte(s[0]>>24), byte(s[0]>>16), byte(s[0]>>8), byte(s[0]),
		byte(s[1]>>56), byte(s[1]>>48), byte(s[1]>>40), byte(s[1]>>32), byte(s[1]>>24), byte(s[1]>>16), byte(s[1]>>8), byte(s[1]),
	)
}

func (s *sum128a) Sum(in []byte) []byte {
	return append(in,
		byte(s[0]>>56), byte(s[0]>>48), byte(s[0]>>40), byte(s[0]>>32), byte(s[0]>>24), byte(s[0]>>16), byte(s[0]>>8), byte(s[0]),
		byte(s[1]>>56), byte(s[1]>>48), byte(s[1]>>40), byte(s[1]>>32), byte(s[1]>>24), byte(s[1]>>16), byte(s[1]>>8), byte(s[1]),
	)
}

func (s *sum32) MarshalBinary() ([]byte, error) {
	return []byte{
		'f', 'n', 'v',
		0x01,
		byte(*s >> 24),
		byte(*s >> 16),
		byte(*s >> 8),
		byte(*s),
	}, nil
}

func (s *sum32a) MarshalBinary() ([]byte, error) {
	return []byte{
		'f', 'n', 'v',
		0x02,
		byte(*s >> 24),
		byte(*s >> 16),
		byte(*s >> 8),
		byte(*s),
	}, nil
}

func (s *sum64) MarshalBinary() ([]byte, error) {
	return []byte{
		'f', 'n', 'v',
		0x03,
		byte(*s >> 56),
		byte(*s >> 48),
		byte(*s >> 40),
		byte(*s >> 32),
		byte(*s >> 24),
		byte(*s >> 16),
		byte(*s >> 8),
		byte(*s),
	}, nil
}

func (s *sum64a) MarshalBinary() ([]byte, error) {
	return []byte{
		'f', 'n', 'v',
		0x04,
		byte(*s >> 56),
		byte(*s >> 48),
		byte(*s >> 40),
		byte(*s >> 32),
		byte(*s >> 24),
		byte(*s >> 16),
		byte(*s >> 8),
		byte(*s),
	}, nil
}

func (s *sum128) MarshalBinary() ([]byte, error) {
	return []byte{
		'f', 'n', 'v',
		0x05,
		byte(s[0] >> 56),
		byte(s[0] >> 48),
		byte(s[0] >> 40),
		byte(s[0] >> 32),
		byte(s[0] >> 24),
		byte(s[0] >> 16),
		byte(s[0] >> 8),
		byte(s[0]),
		byte(s[1] >> 56),
		byte(s[1] >> 48),
		byte(s[1] >> 40),
		byte(s[1] >> 32),
		byte(s[1] >> 24),
		byte(s[1] >> 16),
		byte(s[1] >> 8),
		byte(s[1]),
	}, nil
}

func (s *sum128a) MarshalBinary() ([]byte, error) {
	return []byte{
		'f', 'n', 'v',
		0x06,
		byte(s[0] >> 56),
		byte(s[0] >> 48),
		byte(s[0] >> 40),
		byte(s[0] >> 32),
		byte(s[0] >> 24),
		byte(s[0] >> 16),
		byte(s[0] >> 8),
		byte(s[0]),
		byte(s[1] >> 56),
		byte(s[1] >> 48),
		byte(s[1] >> 40),
		byte(s[1] >> 32),
		byte(s[1] >> 24),
		byte(s[1] >> 16),
		byte(s[1] >> 8),
		byte(s[1]),
	}, nil
}

func (s *sum32) UnmarshalBinary(data []byte) error {
	if len(data) != 8 || data[0] != 'f' || data[1] != 'n' || data[2] != 'v' || data[3] != 0x01 {
		return errors.New("hash/fnv: invalid state")
	}
	*s = sum32(data[4])<<24 | sum32(data[5])<<16 | sum32(data[6])<<8 | sum32(data[7])
	return nil
}

func (s *sum32a) UnmarshalBinary(data []byte) error {
	if len(data) != 8 || data[0] != 'f' || data[1] != 'n' || data[2] != 'v' || data[3] != 0x02 {
		return errors.New("hash/fnv: invalid state")
	}
	*s = sum32a(data[4])<<24 | sum32a(data[5])<<16 | sum32a(data[6])<<8 | sum32a(data[7])
	return nil
}

func (s *sum64) UnmarshalBinary(data []byte) error {
	if len(data) != 12 || data[0] != 'f' || data[1] != 'n' || data[2] != 'v' || data[3] != 0x03 {
		return errors.New("hash/fnv: invalid state")
	}
	*s = sum64(data[4])<<56 | sum64(data[5])<<48 | sum64(data[6])<<40 | sum64(data[7])<<32 |
		sum64(data[8])<<24 | sum64(data[9])<<16 | sum64(data[10])<<8 | sum64(data[11])
	return nil
}

func (s *sum64a) UnmarshalBinary(data []byte) error {
	if len(data) != 12 || data[0] != 'f' || data[1] != 'n' || data[2] != 'v' || data[3] != 0x04 {
		return errors.New("hash/fnv: invalid state")
	}
	*s = sum64a(data[4])<<56 | sum64a(data[5])<<48 | sum64a(data[6])<<40 | sum64a(data[7])<<32 |
		sum64a(data[8])<<24 | sum64a(data[9])<<16 | sum64a(data[10])<<8 | sum64a(data[11])
	return nil
}

func (s *sum128) UnmarshalBinary(data []byte) error {
	if len(data) != 20 || data[0] != 'f' || data[1] != 'n' || data[2] != 'v' || data[3] != 0x05 {
		return errors.New("hash/fnv: invalid state")
	}
	s[0] = uint64(data[4])<<56 | uint64(data[5])<<48 | uint64(data[6])<<40 | uint64(data[7])<<32 |
		uint64(data[8])<<24 | uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11])
	s[1] = uint64(data[12])<<56 | uint64(data[13])<<48 | uint64(data[14])<<40 | uint64(data[15])<<32 |
		uint64(data[16])<<24 | uint64(data[17])<<16 | uint64(data[18])<<8 | uint64(data[19])
	return nil
}

func (s *sum128a) UnmarshalBinary(data []byte) error {
	if len(data) != 20 || data[0] != 'f' || data[1] != 'n' || data[2] != 'v' || data[3] != 0x06 {
		return errors.New("hash/fnv: invalid state")
	}
	s[0] = uint64(data[4])<<56 | uint64(data[5])<<48 | uint64(data[6])<<40 | uint64(data[7])<<32 |
		uint64(data[8])<<24 | uint64(data[9])<<16 | uint64(data[10])<<8 | uint64(data[11])
	s[1] = uint64(data[12])<<56 | uint64(data[13])<<48 | uint64(data[14])<<40 | uint64(data[15])<<32 |
		uint64(data[16])<<24 | uint64(data[17])<<16 | uint64(data[18])<<8 | uint64(data[19])
	return nil
}
