package cipher

import "internal/cpu"

func xor(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return 0
	}

	if cpu.X86.HasAVX512 {
		if n < 16 {
			xorBytesNoSIMD(dst, a, b)
		} else if n < 32 {
			xorSSE2(dst, a, b, n)
		} else if n < 64 {
			xorAVX2(dst, a, b, n)
		} else {
			xorAVX512(dst, a, b, n)
		}
	} else if cpu.X86.HasAVX2 {
		if n < 16 {
			xorBytesNoSIMD(dst, a, b)
		} else if n < 32 {
			xorSSE2(dst, a, b, n)
		} else {
			xorAVX2(dst, a, b, n)
		}
	} else if cpu.X86.HasSSE2 {
		if n < 16 {
			xorBytesNoSIMD(dst, a, b)
		} else {
			xorSSE2(dst, a, b, n)
		}
	} else {
		xorBytesNoSIMD(dst, a, b)
	}

	return n
}

//go:noescape
func xorAVX512(dst, a, b []byte, n int)

//go:noescape
func xorAVX2(dst, a, b []byte, n int)

//go:noescape
func xorSSE2(dst, a, b []byte, n int)
