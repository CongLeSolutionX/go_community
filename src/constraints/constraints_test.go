// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package constraints

import (
	"bytes"
	"fmt"
	"internal/testenv"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type testSigned[T Signed] struct{ f T }
type testUnsigned[T Unsigned] struct{ f T }
type testInteger[T Integer] struct{ f T }
type testFloat[T Float] struct{ f T }
type testComplex[T Complex] struct{ f T }
type testOrdered[T Ordered] struct{ f T }
type testSlice[T Slice[E], E any] struct{ f T }
type testMap[T Map[K, V], K comparable, V any] struct{ f T }
type testChan[T Chan[E], E any] struct{ f T }

// TestTypes passes if it compiles.
type TestTypes struct {
	fs1 testSigned[int]
	fs2 testSigned[int64]
	fu1 testUnsigned[uint]
	fu2 testUnsigned[uintptr]
	fi1 testInteger[int8]
	fi2 testInteger[uint8]
	fi3 testInteger[uintptr]
	ff1 testFloat[float32]
	fc1 testComplex[complex64]
	fo1 testOrdered[int]
	fo2 testOrdered[float64]
	fo3 testOrdered[string]
	fl1 testSlice[[]int, int]
	fm1 testMap[map[int]bool, int, bool]
	fh1 testChan[chan int, int]
}

func infer1[S Slice[E], E any](s S, v E) S                     { return s }
func infer2[M Map[K, V], K comparable, V any](m M, k K, v V) M { return m }
func infer3[C Chan[E], E any](c C, v E) C                      { return c }

func TestInference(t *testing.T) {
	var empty interface{}

	type S []int
	empty = infer1(S{}, 0)
	if _, ok := empty.(S); !ok {
		t.Errorf("infer1(S) returned %T, expected S", empty)
	}

	type M map[int]bool
	empty = infer2(M{}, 0, false)
	if _, ok := empty.(M); !ok {
		t.Errorf("infer2(M) returned %T, expected M", empty)
	}

	type C chan bool
	empty = infer3(make(C), true)
	if _, ok := empty.(C); !ok {
		t.Errorf("infer3(C) returned %T, expected C", empty)
	}
}

var prolog = []byte(`
package constrainttest

import "constraints"

type testSigned[T constraints.Signed] struct{ f T }
type testUnsigned[T constraints.Unsigned] struct{ f T }
type testInteger[T constraints.Integer] struct{ f T }
type testFloat[T constraints.Float] struct{ f T }
type testComplex[T constraints.Complex] struct{ f T }
type testOrdered[T constraints.Ordered] struct{ f T }
type testSlice[T constraints.Slice[E], E any] struct{ f T }
type testMap[T constraints.Map[K, V], K comparable, V any] struct{ f T }
type testChan[T constraints.Chan[E], E any] struct{ f T }
`)

func TestFailure(t *testing.T) {
	testenv.MustHaveGoBuild(t)
	gocmd := testenv.GoToolPath(t)
	tmpdir := t.TempDir()

	if err := os.WriteFile(filepath.Join(tmpdir, "go.mod"), []byte("module constraintest"), 0666); err != nil {
		t.Fatal(err)
	}

	for i, test := range []struct {
		constraint, typ string
	}{
		{"testSigned", "uint"},
		{"testUnsigned", "int"},
		{"testInteger", "float32"},
		{"testFloat", "int8"},
		{"testComplex", "float64"},
		{"testOrdered", "bool"},
		{"testSlice", "int, int"},
		{"testMap", "string, string, string"},
		{"testChan", "[]int, int"},
	} {
		i := i
		test := test
		t.Run(fmt.Sprintf("%s %d", test.constraint, i), func(t *testing.T) {
			t.Parallel()
			name := fmt.Sprintf("go%d.go", i)
			f, err := os.Create(filepath.Join(tmpdir, name))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write(prolog); err != nil {
				t.Fatal(err)
			}
			if _, err := fmt.Fprintf(f, "var V %s[%s]\n", test.constraint, test.typ); err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(gocmd, "build", name)
			cmd.Dir = tmpdir
			if out, err := cmd.CombinedOutput(); err == nil {
				t.Error("build succeeded, but expected to fail")
			} else if len(out) > 0 {
				t.Logf("%s", out)
				const want = "does not satisfy"
				if !bytes.Contains(out, []byte(want)) {
					t.Errorf("output does not include %q", want)
				}
			} else {
				t.Error("no error output, expected something")
			}
		})
	}
}
