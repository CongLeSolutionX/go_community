package cipher

//go:noescape
func xorBytesSSE2(dst, a, b []byte, n int)
