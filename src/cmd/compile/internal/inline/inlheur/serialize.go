// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package inlheur

func (fp *FuncProps) SerializeToString() string {
	if fp == nil {
		return ""
	}
	sl := make([]byte, 0, 256)
	sl = writeUleb128(sl, uint64(fp.Flags))
	sl = writeUleb128(sl, uint64(len(fp.RecvrParamFlags)))
	for _, pf := range fp.RecvrParamFlags {
		sl = writeUleb128(sl, uint64(pf))
	}
	sl = writeUleb128(sl, uint64(len(fp.ReturnFlags)))
	for _, rf := range fp.ReturnFlags {
		sl = writeUleb128(sl, uint64(rf))
	}
	return string(sl)
}

func DeserializeFromString(s string) *FuncProps {
	if len(s) == 0 {
		return nil
	}
	var fp FuncProps
	var v uint64
	sl := []byte(s)
	v, sl = readULEB128(sl)
	fp.Flags = FuncPropBits(v)
	v, sl = readULEB128(sl)
	fp.RecvrParamFlags = make([]ParamPropBits, v)
	for i := range fp.RecvrParamFlags {
		v, sl = readULEB128(sl)
		fp.RecvrParamFlags[i] = ParamPropBits(v)
	}
	v, sl = readULEB128(sl)
	fp.ReturnFlags = make([]ReturnPropBits, v)
	for i := range fp.ReturnFlags {
		v, sl = readULEB128(sl)
		fp.ReturnFlags[i] = ReturnPropBits(v)
	}
	return &fp
}

func readULEB128(sl []byte) (value uint64, rsl []byte) {
	var shift uint

	for {
		b := sl[0]
		sl = sl[1:]
		value |= (uint64(b&0x7F) << shift)
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return value, sl
}

func writeUleb128(sl []byte, v uint64) []byte {
	if v < 128 {
		sl = append(sl, uint8(v))
		return sl
	}
	more := true
	for more {
		c := uint8(v & 0x7f)
		v >>= 7
		more = v != 0
		if more {
			c |= 0x80
		}
		sl = append(sl, c)
	}
	return sl
}
