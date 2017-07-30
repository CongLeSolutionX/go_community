// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
)

// This is exactly what a test would do without a TestMain.
// It's here only so that there is at least one package in the
// standard library with a TestMain, so that code is executed.

func TestMain(m *M) {
	os.Exit(m.Run())
}

func TestListTests(t *T) {
	var matchString = func(pat, str string) (bool, error) {
		if ok, err := regexp.MatchString(pat, str); err != nil || !ok {
			return false, err
		}
		return true, nil
	}

	testInternal := struct {
		tests      []InternalTest
		benchmarks []InternalBenchmark
		examples   []InternalExample
	}{
		tests: []InternalTest{
			InternalTest{
				Name: "TestSimple",
				F: func(t *T) {
					_ = fmt.Sprint("Test simple")
				},
			},
		},
		benchmarks: []InternalBenchmark{
			InternalBenchmark{
				Name: "BenchmarkSimple",
				F: func(b *B) {
					b.StopTimer()
					b.StartTimer()
					for i := 0; i < b.N; i++ {
						_ = fmt.Sprint("Test for bench")
					}
				},
			},
		},
		examples: []InternalExample{
			InternalExample{
				Name: "ExampleSimple",
				F: func() {
					fmt.Println("Test with Output.")

					// Output: Test with Output.
				},
				Output:    "Test with Output.\n",
				Unordered: false,
			},
			InternalExample{
				Name: "ExampleWithEmptyOutput",
				F: func() {
					fmt.Println("")

					// Output:
				},
				Output:    "",
				Unordered: false,
			},
		},
	}

	testCases := []struct {
		listFlag string
		output   []string
	}{
		{"Test", []string{"TestSimple"}},
		{"Benchmark", []string{"BenchmarkSimple"}},
		{"Example", []string{"ExampleSimple", "ExampleWithEmptyOutput"}},
	}

	for _, tc := range testCases {
		flag.Set("test.list", tc.listFlag)
		flag.Parse()
		results, err := listTests(matchString, testInternal.tests, testInternal.benchmarks, testInternal.examples)
		if !reflect.DeepEqual(results, tc.output) || err != nil {
			t.Errorf("Test case for list &s failed.", tc.listFlag)
		}
	}
}
