package x

import _ "unsafe"

func indexByte(xs []byte, b byte) int { // ERROR "indexByte xs does not escape"
	for i, x := range xs {
		if x == b {
			return i
		}
	}
	return -1
}
