// +build !amd64 noasm

package cipher

func xor(dst, a, b []byte) int {
	return xorBytesNoSIMD(dst, a, b)
}
