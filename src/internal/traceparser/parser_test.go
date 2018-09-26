// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traceparser

import (
	"io/ioutil"
	"os"
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

func TestStats(t *testing.T) {
	fname := "./testdata/http_1_9_good"
	p, err := New(fname)
	if err != nil {
		t.Fatal(err)
	}
	stat := p.OSStats()
	if stat.Bytes == 0 || stat.Seeks == 0 || stat.Reads == 0 {
		t.Errorf("OSStats impossible %v", stat)
	}
	fd, err := os.Open(fname)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := ParseBuffer(fd)
	if err != nil {
		t.Fatal(err)
	}
	stat = pb.OSStats()
	if stat.Seeks != 0 || stat.Reads != 0 {
		t.Errorf("unexpected positive results %v", stat)
	}
}
