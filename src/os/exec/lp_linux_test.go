// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package exec

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestFindExecutableVsNoexec(t *testing.T) {
	tmp := t.TempDir()

	// Check if faccessat2 is available.
	if _, _, err := syscall.Syscall6(syscall.SYS_FACCESSAT2, 0, 0, 0, 0, 0, 0); err == syscall.ENOSYS || err == syscall.EPERM {
		t.Skip("requires faccessat2(2) syscall")
	}

	// Create a tmpfs mount.
	err := syscall.Mount("tmpfs", tmp, "tmpfs", 0, "")
	if err != nil {
		// Most probably root is required.
		t.Skipf("requires ability to mount tmpfs (%v)", err)
	}
	t.Cleanup(func() { _ = syscall.Unmount(tmp, 0) })

	// Create an executable.
	path := filepath.Join(tmp, "program")
	err = os.WriteFile(path, []byte("#!/bin/sh\necho 123\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	// Check that it works as expected.
	err = findExecutable(path)
	if err != nil {
		t.Fatalf("findExecutable: want nil, got %v", err)
	}

	if err := Command(path).Run(); err != nil {
		t.Fatalf("exec: want nil, got %v", err)
	}

	// Remount with noexec flag.
	err = syscall.Mount("", tmp, "", syscall.MS_REMOUNT|syscall.MS_NOEXEC, "")
	if err != nil {
		t.Fatalf("requires ability to bind mount (%v)", err)
	}

	// It's a test prerequisite that the executable fails to run
	// after we remount tmpfs with noexec.
	if err := Command(path).Run(); err == nil {
		t.Skip("exec: want error, got nil")
	}

	err = findExecutable(path)
	if err == nil {
		t.Fatalf("findExecutable: want error, got nil")
	}
}
