// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux

package ld

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestFallocate(t *testing.T) {
	dir, err := ioutil.TempDir("", "TestFallocate")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	filename := filepath.Join(dir, "a.out")
	out := NewOutBuf(nil)
	err = out.Open(filename)
	if err != nil {
		t.Fatalf("Open file failed: %v", err)
	}
	defer out.Close()

	// Skip testing filesystems that don't support fallocate.
	stat := syscall.Statfs_t{}
	err = syscall.Statfs(dir, &stat)
	if err != nil {
		t.Fatalf("Statfs failed: %v", err)
	}

	// Common filesystem type list: https://man7.org/linux/man-pages/man2/statfs.2.html,
	// but ZFS is not included.
	// Type 0x2fc12fc1 for ZFS was got from live Linux system, official document is not available,
	// also be mentioned in minio project, https://github.com/minio/minio/blob/master/pkg/disk/type_linux.go
	if stat.Type == 0x2fc12fc1 {
		// This test is not supported on openZFS: https://github.com/openzfs/zfs/issues/11073
		t.Skip("fallocate is not supported on ZFS")
	}

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
