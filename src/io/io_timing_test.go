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
	b.RunParallel(func(pb *testing.PB) {
		var (
			src bytesReader

			// Use zeroWriter instead of ioutil.Discard here,
			// since io.Discard may have unexpected side effects
			// for this benchmark.
			dst zeroWriter
		)
		for pb.Next() {
			src.Reset(size)
			n, err := io.Copy(&dst, &src)
			if err != nil {
				b.Fatalf("unexpected error returned from io.Copy: %s", err)
			}
			if n != int64(size) {
				b.Fatalf("unexpected number of bytes copied: %d. Expecting %d", n, size)
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

type zeroWriter struct{}

func (zw *zeroWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
