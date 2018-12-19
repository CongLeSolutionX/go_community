package hash_test

import (
	"hash"
	"io"
	"testing"
)

func TestRuntimeHash(t *testing.T) {
	h := hash.NewRuntime()
	io.WriteString(h, "abc")
	io.WriteString(h, "def")
	h1 := h.Sum(nil)

	h.Reset()
	io.WriteString(h, "abc")
	io.WriteString(h, "def")
	h2 := h.Sum(nil)

	if string(h1) != string(h2) {
		t.Errorf("hashes differ: %08x, %08x", h1, h2)
	}

	// Writing the same message in different chunks
	// changes the hash ("with high probability").
	h.Reset()
	io.WriteString(h, "ab")
	io.WriteString(h, "cdef")
	h3 := h.Sum(nil)

	if string(h1) == string(h3) {
		t.Errorf("hashes same: %08x, %08x", h1, h2)
	}
}

func BenchmarkRuntimeHashNoAllocs(b *testing.B) {
	b.ReportAllocs()

	var sum [8]byte
	data := []byte("abc")

	// This loop should not allocate memory.
	// This relies on h.Write being a direct call;
	// an indirect call causes data to be heap-allocated.
	for i := 0; i < b.N; i++ {
		h := hash.NewRuntime()
		h.Write(data)
		h.WriteString("abc")
		_ = h.Sum(sum[:0])
	}
}
