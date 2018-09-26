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

var (
	// testfiles from the old trace parser
	otherDir  = "../trace/testdata/"
	othergood = []string{"http_1_9_good", "http_1_10_good", "http_1_11_good",
		"stress_1_9_good", "stress_1_10_good", "stress_1_11_good",
		"stress_start_stop_1_9_good", "stress_start_stop_1_10_good", "stress_start_stop_1_11_good",
		"user_task_span_1_11_good",
	}
	otherbad = []string{"http_1_5_good", "http_1_7_good",
		"stress_1_5_good", "stress_1_5_unordered", "stress_1_7_good",
		"stress_start_stop_1_5_good", "stress_start_stop_1_7_good",
	}
)

func TestRemoteFiles(t *testing.T) {
	type test struct {
		fname string
		ok    bool
	}
	var table []test
	for _, f := range othergood {
		table = append(table, test{filepath.Join(otherDir, f), true})
	}
	for _, f := range otherbad {
		table = append(table, test{filepath.Join(otherDir, f), false})
	}
	for _, x := range table {
		p, err := New(x.fname)
		if err == nil {
			err = p.Parse(0, 1<<62, nil)
		}
		if err == nil != x.ok {
			t.Errorf("%s: got %v, expected %v", x.fname, err == nil, x.ok)
		}
	}
}

func TestLocalFiles(t *testing.T) {

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
	fname := "./testdata/2ccf452e473ded814ea880c602488637fc27e549.good"
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
