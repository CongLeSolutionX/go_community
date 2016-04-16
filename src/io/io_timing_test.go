package io_test

import (
	"io"
	"testing"
)

func BenchmarkCopy16(b *testing.B) {
	benchmarkCopySize(b, 16)
}

func BenchmarkCopy64(b *testing.B) {
	benchmarkCopySize(b, 64)
}

func BenchmarkCopy256(b *testing.B) {
	benchmarkCopySize(b, 256)
}

func BenchmarkCopy1K(b *testing.B) {
	benchmarkCopySize(b, 1024)
}

func BenchmarkCopy4K(b *testing.B) {
	benchmarkCopySize(b, 4*1024)
}

func BenchmarkCopy16K(b *testing.B) {
	benchmarkCopySize(b, 16*1024)
}

func BenchmarkCopy64K(b *testing.B) {
	benchmarkCopySize(b, 64*1024)
}

func BenchmarkCopy256K(b *testing.B) {
	benchmarkCopySize(b, 256*1024)
}

func BenchmarkCopy1M(b *testing.B) {
	benchmarkCopySize(b, 1024*1024)
}

func benchmarkCopySize(b *testing.B, size int) {
	b.SetBytes(int64(size))
	b.RunParallel(func(pb *testing.PB) {
		var (
			src bytesReader
			dst bytesWriter
		)
		for pb.Next() {
			src.Reset(size)
			n, err := io.Copy(&dst, &src)
			if err != nil {
				b.Errorf("unexpected error returned from io.Copy: %s", err)
				break
			}
			if n != int64(size) {
				b.Errorf("unexpected number of bytes copied: %d. Expecting %d", n, size)
				break
			}
		}
	})
}

type bytesReader struct {
	size int
	n    int
}

func (br *bytesReader) Read(p []byte) (int, error) {
	n := br.size - br.n
	if n <= 0 {
		return 0, io.EOF
	}
	if n > len(p) {
		n = len(p)
	}
	// Emulate reading to p.
	for i := 0; i < n; i++ {
		p[i] = byte(br.n + i)
	}
	br.n += n
	return n, nil
}

func (br *bytesReader) Reset(size int) {
	br.size = size
	br.n = 0
}

type bytesWriter struct {
	s uint64
}

func (zw *bytesWriter) Write(p []byte) (int, error) {
	n := len(p)
	// Emulate writing to p.
	for i := 0; i < n; i++ {
		zw.s += uint64(p[i])
	}
	return n, nil
}
