// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cov_test

import (
	"cmd/internal/cov"
	"crypto/md5"
	"fmt"
	"internal/coverage"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPodCollection(t *testing.T) {
	//testenv.MustHaveGoBuild(t)

	mkdir := func(d string, perm os.FileMode) string {
		dp := filepath.Join(t.TempDir(), d)
		if err := os.Mkdir(dp, perm); err != nil {
			t.Fatal(err)
		}
		return dp
	}

	mkfile := func(d string, fn string) string {
		fp := filepath.Join(d, fn)
		if err := ioutil.WriteFile(fp, []byte("foo"), 0666); err != nil {
			t.Fatal(err)
		}
		return fp
	}

	mkmeta := func(dir string, tag string) string {
		hash := md5.Sum([]byte(tag))
		fn := fmt.Sprintf("%s.%x", coverage.MetaFilePref, hash)
		return mkfile(dir, fn)
	}

	mkcounter := func(dir string, tag string, nt int) string {
		hash := md5.Sum([]byte(tag))
		fn := fmt.Sprintf("%s.%x.%d", coverage.CounterFilePref, hash, nt)
		return mkfile(dir, fn)
	}

	trim := func(path string) string {
		b := filepath.Base(path)
		d := filepath.Dir(path)
		db := filepath.Base(d)
		return filepath.Join(db, b)
	}

	podToString := func(p cov.Pod) string {
		rv := trim(p.MetaFile) + " [ "
		for _, df := range p.CounterDataFiles {
			rv += trim(df) + " "
		}
		return rv + "]"
	}

	// Create a couple of directories.
	o1 := mkdir("o1", 0777)
	o2 := mkdir("o2", 0777)

	// Add some random files (not coverage related)
	mkfile(o1, "blah.txt")
	mkfile(o1, "something.exe")

	// Add a meta-data file with two counter files to first dir.
	mkmeta(o1, "m1")
	mkcounter(o1, "m1", 1)
	mkcounter(o1, "m1", 2)
	mkcounter(o1, "m1", 2)

	// Add a counter file with no associated meta file.
	mkcounter(o1, "orphan", 9)

	// Add a meta-data file with three counter files to second dir.
	mkmeta(o2, "m2")
	mkcounter(o2, "m2", 1)
	mkcounter(o2, "m2", 2)
	mkcounter(o2, "m2", 3)

	// Add a third counter file to the second dir. This is kind of
	// an odd case (don't expect to see this in practice) since the
	// counter file is in one dir and the meta file in another, but
	// we might as well support it.
	mkcounter(o2, "m1", 11)

	// Collect pods.
	pods, err := cov.CollectPods([]string{o1, o2}, true)
	if err != nil {
		t.Fatal(err)
	}

	// Verify pods
	if len(pods) != 2 {
		t.Fatalf("expected 2 pods got %d pods", len(pods))
	}

	expected := []string{
		"o1/covmeta.ae7be26cdaa742ca148068d5ac90eaca [ o1/covcounters.ae7be26cdaa742ca148068d5ac90eaca.1 o1/covcounters.ae7be26cdaa742ca148068d5ac90eaca.2 o2/covcounters.ae7be26cdaa742ca148068d5ac90eaca.11 ]",
		"o2/covmeta.aaf2f89992379705dac844c0a2a1d45f [ o2/covcounters.aaf2f89992379705dac844c0a2a1d45f.1 o2/covcounters.aaf2f89992379705dac844c0a2a1d45f.2 o2/covcounters.aaf2f89992379705dac844c0a2a1d45f.3 ]",
	}
	for k, exp := range expected {
		got := podToString(pods[k])
		if exp != got {
			t.Errorf("pod %d: expected %q got %q", k, exp, got)
		}
	}

	// Check handling of unreadable dir.
	dbad := mkdir("bad", 0)
	_, err = cov.CollectPods([]string{dbad}, true)
	if err == nil {
		t.Errorf("exected error due to unreadable dir")
	}
}
