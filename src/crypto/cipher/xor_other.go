// +build !amd64 noasm

package cipher

func xor(dst, a, b []int) int {
	return xorBytesNoSIMD(dst, a, b)
}
