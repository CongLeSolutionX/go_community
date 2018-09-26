// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traceparser

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestFiles(t *testing.T) {

	files, err := ioutil.ReadDir("./testdata")
	if err != nil {
		t.Fatalf("failed to read ./testdata: %v", err)
	}
	for _, f := range files {
		fname := filepath.Join("./testdata", f.Name())
		p, err := New(fname)
		if err == nil {
			err = p.Parse(0, 1<<62, nil)
		}
		//t.Logf("%v %s", err, f.Name())
		switch {
		case strings.Contains(f.Name(), "good"),
			strings.Contains(f.Name(), "weird"):
			if err != nil {
				t.Errorf("unexpected failure %v %s", err, f.Name())
			}
		case strings.Contains(f.Name(), "bad"):
			if err == nil {
				t.Errorf("bad file did not fail %s", f.Name())
			}
		default:
			t.Errorf("untyped file %v %s", err, f.Name())
		}
	}
}
