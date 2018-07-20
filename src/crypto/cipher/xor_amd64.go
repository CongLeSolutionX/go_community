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

	if cpu.X86.HasSSE2 {
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
func xorSSE2(dst, a, b []byte, n int)
