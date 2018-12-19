package hash

import _ "unsafe"

// Runtime is an non-crypographic hash function
// based on the efficient implementation used by the Go runtime.
//
// It is deterministic only within the lifetime of a single process,
// so hashes should not be stored outside the program.
type Runtime struct {
	x uintptr
}

func NewRuntime() *Runtime { return new(Runtime) } // TODO: redundant?

const wordsize = 4 << (^uintptr(0) >> 32 & 1) // 4 or 8

func (h *Runtime) BlockSize() int { return 1 }

func (h *Runtime) Reset() { h.x = 0 }

func (h *Runtime) Size() int { return wordsize }

func (h *Runtime) Sum(b []byte) []byte {
	n := len(b)
	res := append(b, "........"[:wordsize]...)

	b = res[n : n+wordsize]
	v := h.x
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	if wordsize == 8 {
		b[4] = byte(v >> 32)
		b[5] = byte(v >> 40)
		b[6] = byte(v >> 48)
		b[7] = byte(v >> 56)
	}

	return res
}

func (h *Runtime) Sum64() uint64 { return uint64(h.x) }

func (h *Runtime) Write(p []byte) (n int, err error) {
	h.x = runtimeBytesHash(p, h.x)
	return len(p), nil
}

func (h *Runtime) WriteString(s string) (n int, err error) {
	h.x = runtimeStringHash(s, h.x)
	return len(s), nil
}

//go:linkname runtimeStringHash runtime.stringHash
func runtimeStringHash(s string, seed uintptr) uintptr

//go:linkname runtimeBytesHash runtime.bytesHash
func runtimeBytesHash(b []byte, seed uintptr) uintptr
