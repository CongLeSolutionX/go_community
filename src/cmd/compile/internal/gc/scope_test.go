// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gc

import (
	"cmd/internal/goobj"
	"internal/testenv"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type pcrange struct {
	low, high int
}

type testscope struct {
	ranges []pcrange
	vars   []string
	scopes []testscope
}

func (tgt *testscope) match(out testscope) bool {
	if len(tgt.vars) != len(out.vars) {
		return false
	}

	for i := range tgt.vars {
		if tgt.vars[i] != out.vars[i] {
			return false
		}
	}

	if len(tgt.scopes) != len(out.scopes) {
		return false
	}

	for i := range tgt.scopes {
		if !tgt.scopes[i].match(out.scopes[i]) {
			return false
		}
	}

	return true
}

// recursively tests if two siblings in the scope tree have overlapping PC
func (s *testscope) siblingOverlap() bool {
	for i := range s.scopes {
		for j := i + 1; j < len(s.scopes); j++ {
			if s.scopes[i].overlapAny(&s.scopes[j]) {
				return true
			}
		}
	}
	return false
}

func (s1 *testscope) overlapAny(s2 *testscope) bool {
	for i := range s1.ranges {
		for j := range s2.ranges {
			if (s1.ranges[i].low >= s2.ranges[j].low && s1.ranges[i].low < s2.ranges[j].high) || (s1.ranges[i].high > s2.ranges[j].low && s1.ranges[i].high <= s2.ranges[j].high) {
				return true
			}
		}
	}
	return false
}

// recursively tests that all children are contained within the range of their parent
func (s *testscope) parentCovers(root bool) bool {
	for i := range s.scopes {
		if !root {
			if !s.covers(&s.scopes[i]) {
				return false
			}
		}
		if !s.scopes[i].parentCovers(false) {
			return false
		}
	}
	return true
}

func (s1 *testscope) covers(s2 *testscope) bool {
	for j := range s2.ranges {
		found := false
		for i := range s1.ranges {
			if s1.ranges[i].low <= s2.ranges[j].low && s1.ranges[i].high >= s2.ranges[j].high {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// returns true if s or one of its childs has no pc ranges
func (s *testscope) emptyScope(root bool) bool {
	if !root {
		if len(s.ranges) == 0 {
			return true
		}
	}
	for i := range s.scopes {
		if s.scopes[i].emptyScope(false) {
			return true
		}
	}
	return false
}

type testcase struct {
	name   string
	symbol string
	file   string
	target testscope
}

func scopeTest(t *testing.T, test testcase) {
	testenv.MustHaveGoBuild(t)
	dir, err := ioutil.TempDir("", test.name)
	if err != nil {
		t.Fatalf("could not create directory: %v", err)
	}
	defer os.RemoveAll(dir)

	compileToObj(t, dir, test.file)
}

func TestNestedForScope(t *testing.T) {
	scopeTest(t, testcase{
		name:   "TestNestedForScope",
		symbol: "main.main",
		file: `
			package main

			func main() {
				for i := 0; i < 5; i++ {
					for i := 0; i < 5; i++ {
						//...
					}
					//...
				}
			}`,
		target: testscope{
			scopes: []testscope{
				testscope{
					vars: []string{"var i"},
					scopes: []testscope{
						testscope{vars: []string{"var i"}}}}}},
	})
}

func compileToObj(t *testing.T, dir, file string) (*goobj.Package, *os.File) {
	src := filepath.Join(dir, "test.go")
	dst := filepath.Join(dir, "out.o")

	f, err := os.Create(src)
	if err != nil {
		panic(err)
	}
	f.Write([]byte(file))
	f.Close()

	cmd := exec.Command(testenv.GoToolPath(t), "tool", "compile", "-N", "-l", "-o", dst, src)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Logf("compile: %s\n", string(b))
		panic(err)
	}

	f, err = os.Open(dst)
	if err != nil {
		panic(err)
	}
	pkg, err := goobj.Parse(f, "main")
	if err != nil {
		panic(err)
	}

	return pkg, f
}
