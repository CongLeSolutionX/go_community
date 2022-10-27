// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || linux
// +build darwin linux

package ld

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestFallocate(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "a.out")
	out := NewOutBuf(nil)
	err := out.Open(filename)
	if err != nil {
		t.Fatalf("Open file failed: %v", err)
	}
	defer out.Close()

	// Try fallocate first.
	for {
		err = out.fallocate(1 << 10)
		if err == syscall.EOPNOTSUPP { // The underlying file system may not support fallocate
			t.Skip("fallocate is not supported")
		}
		if err == syscall.EINTR {
			continue // try again
		}
		if err != nil {
			t.Fatalf("fallocate failed: %v", err)
		}
		break
	}

	// Mmap 1 MiB initially, and grow to 2 and 3 MiB.
	// Check if the file size and disk usage is expected.
	for _, sz := range []int64{1 << 20, 2 << 20, 3 << 20} {
		err = out.Mmap(uint64(sz))
		if err != nil {
			t.Fatalf("Mmap failed: %v", err)
		}
		stat, err := os.Stat(filename)
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}
		if got := stat.Size(); got != sz {
			t.Errorf("unexpected file size: got %d, want %d", got, sz)
		}
		// The number of blocks must be enough for the requested size.
		// We used to require an exact match, but it appears that
		// some file systems allocate a few extra blocks in some cases.
		// See issue #41127.
		if got, want := stat.Sys().(*syscall.Stat_t).Blocks, (sz+511)/512; got < want {
			t.Errorf("unexpected disk usage: got %d blocks, want at least %d", got, want)
		}
		out.munmap()
	}
}

func BenchmarkFallocate(b *testing.B) {
	dir := b.TempDir()
	filename := filepath.Join(dir, "a.out")
	out := NewOutBuf(nil)

	err := out.Open(filename)
	if err != nil {
		b.Fatalf("Open file failed: %v", err)
	}
	defer out.Close()

	for {
		err = out.fallocate(100 << 20)
		if err == syscall.EOPNOTSUPP {
			b.Skip("fallocate is not supported")
		}
		if err == syscall.EINTR {
			continue
		}
		if err != nil {
			b.Fatalf("fallocate failed: %v", err)
		}
		out.Close()
		os.Remove(filename)
		break
	}

	var fsize = []int64{
		100, 200, 500, 800, 1000, 1234, 2030, 4567, 5600, 5710, 6620,
		6643, 8000, 9012, 10010, 10240, 12010, 12345, 20000, 28000,
		35000, 53000, 54321, 65535, 81020, 88050, 100200, 110010, 111000,
		111100, 118100, 168200, 1682200, 10682200, 10882200, 20682200,
		20692200, 31698200, 40691200, 48691200, 50730000, 83886080,
	}

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		err := out.Open(filename)
		if err != nil {
			b.Fatalf("Open file failed: %v", err)
		}
		defer out.Close()

		for _, sz := range fsize {
			err = out.Mmap(uint64(sz))
			if err != nil {
				b.Fatalf("fallocate failed: %v", err)
			}
			stat, err := os.Stat(filename)
			if err != nil {
				b.Fatalf("Stat failed: %v", err)
			}
			if got, want := stat.Sys().(*syscall.Stat_t).Blocks, (sz+511)/512; got < want {
				b.Errorf("unexpected disk usage: got %d blocks, want at least %d\n", got, want)
			}
			out.munmap()
		}
		out.Close()
		os.Remove(filename)
	}
}
