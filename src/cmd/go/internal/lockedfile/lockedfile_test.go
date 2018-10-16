// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// js and nacl do not support inter-process file locking.
// +build !js,!nacl

package lockedfile_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cmd/go/internal/lockedfile"
)

func mustTempDir(t *testing.T) (dir string, remove func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", filepath.Base(t.Name()))
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

const (
	quiescent            = 10 * time.Millisecond
	probablyStillBlocked = 10 * time.Second
)

func mustBlock(t *testing.T, desc string, f func()) (wait func(*testing.T)) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		f()
		close(done)
	}()

	select {
	case <-done:
		t.Fatalf("%s unexpectedly did not block", desc)
		return nil

	case <-time.After(quiescent):
		return func(t *testing.T) {
			t.Helper()
			select {
			case <-time.After(probablyStillBlocked):
				t.Fatalf("%s is unexpectedly still blocked after %v", desc, probablyStillBlocked)
			case <-done:
			}
		}
	}
}

func TestNonEmptySentinelReturnsIsExist(t *testing.T) {
	t.Parallel()

	dir, remove := mustTempDir(t)
	defer remove()

	path := filepath.Join(dir, "done")

	f, err := lockedfile.NewSentinel(path)
	if err != nil {
		t.Fatalf("NewSentinel: %v", err)
	}
	t.Logf("f, <nil> := NewSentinel(_)")
	defer f.Close()

	wait := mustBlock(t, "NewSentinel", func() {
		other, err := lockedfile.NewSentinel(path)
		if os.IsExist(err) {
			t.Logf("NewSentinel: os.IsExist(err) (as expected)")
		} else {
			if err == nil {
				other.Close()
			}
			t.Errorf("NewSentinel: err = %v; want os.IsExist(err)", err)
		}
	})

	_, err = f.WriteString("done")
	t.Logf(`f.WriteString("done"): err = %v`, err)
	if err != nil {
		t.Fail()
	}
	err = f.Close()
	t.Logf("f.Close() = %v", err)

	wait(t)
}

func TestSentinelResumesIfEmpty(t *testing.T) {
	t.Parallel()

	dir, remove := mustTempDir(t)
	defer remove()

	path := filepath.Join(dir, "done")
	f, err := lockedfile.NewSentinel(path)
	if err != nil {
		t.Fatalf("NewSentinel: %v", err)
	}
	t.Logf("f, <nil> := NewSentinel(_)")
	defer f.Close()

	wait := mustBlock(t, "NewSentinel", func() {
		other, err := lockedfile.NewSentinel(path)
		if err == nil {
			other.Close()
			t.Logf("other, <nil> := NewSentinel(_)")
		} else {
			t.Errorf("NewSentinel; err = %v; want <nil>", err)
		}
	})

	err = f.Close() // Without writing!
	t.Logf("f.Close() = %v", err)

	wait(t)
}

func TestMutexExcludes(t *testing.T) {
	t.Parallel()

	dir, remove := mustTempDir(t)
	defer remove()

	path := filepath.Join(dir, "lock")

	mu := lockedfile.MutexAt(path)
	t.Logf("mu := MutexAt(_)")

	unlock, err := mu.Lock()
	if err != nil {
		t.Fatalf("mu.Lock: %v", err)
	}
	t.Logf("unlock, _  := mu.Lock()")

	mu2 := lockedfile.MutexAt(mu.Path)
	t.Logf("mu2 := MutexAt(mu.Path)")

	wait := mustBlock(t, "mu2.Lock()", func() {
		unlock2, err := mu2.Lock()
		if err != nil {
			t.Errorf("mu2.Lock: %v", err)
			return
		}
		t.Logf("unlock2, _ := mu2.Lock()")
		t.Logf("unlock2()")
		unlock2()
	})

	t.Logf("unlock()")
	unlock()
	wait(t)
}

func TestReadWaitsForLock(t *testing.T) {
	t.Parallel()

	dir, remove := mustTempDir(t)
	defer remove()

	path := filepath.Join(dir, "timestamp.txt")

	f, err := lockedfile.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer f.Close()

	const (
		part1 = "part 1\n"
		part2 = "part 2\n"
	)
	_, err = f.WriteString(part1)
	if err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	t.Logf("WriteString(%q) = <nil>", part1)

	wait := mustBlock(t, "Read", func() {
		b, err := lockedfile.Read(path)
		if err != nil {
			t.Errorf("Read: %v", err)
			return
		}

		const want = part1 + part2
		got := string(b)
		if got == want {
			t.Logf("Read(_) = %q", got)
		} else {
			t.Errorf("Read(_) = %q, _; want %q", got, want)
		}
	})

	_, err = f.WriteString(part2)
	if err != nil {
		t.Errorf("WriteString: %v", err)
	} else {
		t.Logf("WriteString(%q) = <nil>", part2)
	}
	f.Close()

	wait(t)
}
