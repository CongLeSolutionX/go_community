// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package protopprof

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"internal/pprof/profile"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strconv"
	"testing"
	"time"
)

// Profile collects a CPU utilization profile and
// writes it to w as a compressed profile.proto. It's used by
// TestProfileParse.
func Profile(w http.ResponseWriter, r *http.Request) {
	sec, _ := strconv.ParseInt(r.FormValue("seconds"), 10, 64)
	if sec == 0 {
		sec = 30
	}
	var buf bytes.Buffer
	// Collect the CPU profile in legacy format in buf.
	startTime := time.Now()
	if err := pprof.StartCPUProfile(&buf); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not enable CPU profiling: %s\n", err)
		return
	}
	time.Sleep(time.Duration(sec) * time.Second)
	pprof.StopCPUProfile()

	const untagged = false
	p, err := TranslateCPUProfile(buf.Bytes(), startTime)
	writeResponse(w, p, err)
}

func writeResponse(w http.ResponseWriter, p *profile.Profile, err error) {
	if err == nil && runtime.GOOS == "linux" {
		err = addMappings(p)
	}
	if err == nil {
		symbolize(p)
		w.Header().Set("Content-Type", "application/octet-stream")
		err = p.Write(w)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

// Helper function to initialize empty cpu profile with sampling period provided.
func createEmptyProfileWithPeriod(t *testing.T, periodMs uint64,
	byteOrder binary.ByteOrder, wordSize int) bytes.Buffer {
	// Mock the sample header produced by cpu profiler. Write a sample
	// period of 2000 microseconds, followed by no samples.
	buf := new(bytes.Buffer)
	// Profile header is as follows:
	// The first, third and fifth words are 0. The second word is 3.
	// The fourth word is the period.
	// EOD marker:
	// The sixth word -- count is initialized to 0 above.
	// The code below sets the seventh word -- nstk to 1
	// The eighth word -- addr is initialized to 0 above.
	words := []uint64{0, 3, 0, periodMs, 0, 0, 1, 0}
	for _, n := range words {
		var err error
		switch wordSize {
		case 8:
			err = binary.Write(buf, byteOrder, uint64(n))
		case 4:
			err = binary.Write(buf, byteOrder, uint32(n))
		}
		if err != nil {
			t.Fatalf("createEmptyProfileWithPeriod failed: %v", err)
		}
	}
	return *buf
}

// Helper function to initialize cpu profile with two sample values.
func createProfileWithTwoSamples(t *testing.T, periodMs uintptr, count1 uintptr, count2 uintptr,
	address1 uintptr, address2 uintptr, byteOrder binary.ByteOrder, wordSize int) bytes.Buffer {
	// Mock the sample header produced by cpu profiler. Write a sample
	// period of 2000 microseconds, followed by no samples.
	buf := new(bytes.Buffer)
	words := []uint64{0, 3, 0, uint64(periodMs), 0, uint64(count1), 2,
		uint64(address1), uint64(address1 + 2),
		uint64(count2), 2, uint64(address2), uint64(address2 + 2),
		0, uint64(1), 0}
	for _, n := range words {
		var err error
		switch wordSize {
		case 8:
			err = binary.Write(buf, byteOrder, uint64(n))
		case 4:
			err = binary.Write(buf, byteOrder, uint32(n))
		}
		if err != nil {
			t.Fatalf("createProfileWithTwoSamples failed: %v", err)
		}
	}
	return *buf
}

// Tests that server creates a cpu profile handler that outputs a parsable Profile profile.
func TestCPUProfileParse(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/protoprofilez" {
				Profile(w, r)
			}
		}))
	defer srv.Close()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	resp, err := http.Get(srv.URL + "/protoprofilez")
	runtime.ReadMemStats(&after)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	_, err = profile.Parse(resp.Body)
	if err != nil {
		t.Fatalf("Could not parse Profile profile: %v", err)
	}
}

// Tests ProfileWriter.Flush() parses correct sampling period in an otherwise empty cpu profile.
func TestFlushSamplingPeriod(t *testing.T) {
	for _, byteOrder := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		for _, wordSize := range []int{4, 8} {
			// A test server with mock cpu profile data.
			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					startTime := time.Now()
					b := createEmptyProfileWithPeriod(t, 2000, byteOrder, wordSize)
					p, err := TranslateCPUProfile(b.Bytes(), startTime)
					if err != nil {
						t.Fatalf("translate failed: %v", err)
					}
					if err := p.Write(w); err != nil {
						t.Fatalf("write failed: %v", err)
					}
				}))
			defer srv.Close()
			resp, err := http.Get(srv.URL)
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			defer resp.Body.Close()
			p, err := profile.Parse(resp.Body)
			if err != nil {
				t.Fatalf("Could not parse Profile profile: %v", err)
			}
			// Expected PeriodType and SampleType.
			expectedPeriodType := &profile.ValueType{Type: "cpu", Unit: "nanoseconds"}
			expectedSampleType := []*profile.ValueType{
				{Type: "samples", Unit: "count"},
				{Type: "cpu", Unit: "nanoseconds"},
			}
			if p.Period != 2000*1000 || !reflect.DeepEqual(p.PeriodType, expectedPeriodType) ||
				!reflect.DeepEqual(p.SampleType, expectedSampleType) || p.Sample != nil {
				t.Fatalf("Unexpected Profile fields with byteOrder:%v, word size:%v", byteOrder, wordSize)
			}
		}
	}
}
func getSampleAsString(sample []*profile.Sample) string {
	var str string
	for _, x := range sample {
		for _, y := range x.Location {
			if y.Mapping != nil {
				str += fmt.Sprintf("Mapping:%v\n", *y.Mapping)
			}
			str += fmt.Sprintf("Location:%v\n", y)
		}
		str += fmt.Sprintf("Sample:%v\n", *x)
	}
	return str
}

// Tests ProfileWriter.Flush() parses a cpu profile with sample values present.
func TestFlushWithSamples(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("test requires a system with /proc/self/maps")
	}
	// Figure out two addresses from /proc/self/maps.
	mmap, err := ioutil.ReadFile("/proc/self/maps")
	if err != nil {
		t.Fatal("Cannot read /proc/self/maps")
	}
	rd := bytes.NewReader(mmap)
	mprof := &profile.Profile{}
	if err = mprof.ParseMemoryMap(rd); err != nil {
		t.Fatalf("Cannot parse /proc/self/maps")
	}
	if len(mprof.Mapping) < 2 {
		t.Fatalf("Less than two mappings")
	}
	address1 := mprof.Mapping[0].Start
	address2 := mprof.Mapping[1].Start
	// A test server with mock cpu profile data.
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			b := createProfileWithTwoSamples(t, 2000, 20, 40, uintptr(address1),
				uintptr(address2), binary.LittleEndian, 8)
			p, err := TranslateCPUProfile(b.Bytes(), startTime)
			writeResponse(w, p, err)
		}))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	p, err := profile.Parse(resp.Body)
	if err != nil {
		t.Fatalf("Could not parse Profile profile: %v", err)
	}
	// Expected PeriodType, SampleType and Sample.
	expectedPeriodType := &profile.ValueType{Type: "cpu", Unit: "nanoseconds"}
	expectedSampleType := []*profile.ValueType{
		{Type: "samples", Unit: "count"},
		{Type: "cpu", Unit: "nanoseconds"},
	}
	expectedSample := []*profile.Sample{
		{Value: []int64{20, 20 * 2000 * 1000}, Location: []*profile.Location{
			{ID: 1, Mapping: mprof.Mapping[0], Address: address1},
			{ID: 2, Mapping: mprof.Mapping[0], Address: address1 + 1},
		}},
		{Value: []int64{40, 40 * 2000 * 1000}, Location: []*profile.Location{
			{ID: 3, Mapping: mprof.Mapping[1], Address: address2},
			{ID: 4, Mapping: mprof.Mapping[1], Address: address2 + 1},
		}},
	}
	if p.Period != 2000*1000 {
		t.Fatalf("Sampling periods do not match")
	}
	if !reflect.DeepEqual(p.PeriodType, expectedPeriodType) {
		t.Fatalf("Period types do not match")
	}
	if !reflect.DeepEqual(p.SampleType, expectedSampleType) {
		t.Fatalf("Sample types do not match")
	}
	if !reflect.DeepEqual(p.Sample, expectedSample) {
		t.Fatalf("Samples do not match: Expected: %v, Got:%v", getSampleAsString(expectedSample),
			getSampleAsString(p.Sample))
	}
}

type fakeFunc struct {
	name   string
	file   string
	lineno int
}

func (f *fakeFunc) Name() string {
	return f.name
}
func (f *fakeFunc) FileLine(_ uintptr) (string, int) {
	return f.file, f.lineno
}

// TestRuntimeFunctionTrimming tests if symbolize trims runtime functions as intended.
func TestRuntimeRunctionTrimming(t *testing.T) {
	fakeFuncMap := map[uintptr]*fakeFunc{
		0x10: &fakeFunc{"runtime.goexit", "runtime.go", 10},
		0x20: &fakeFunc{"runtime.other", "runtime.go", 20},
		0x30: &fakeFunc{"foo", "foo.go", 30},
		0x40: &fakeFunc{"bar", "bar.go", 40},
	}
	backupFuncForPC := funcForPC
	funcForPC = func(pc uintptr) function {
		return fakeFuncMap[pc]
	}
	defer func() {
		funcForPC = backupFuncForPC
	}()
	testLoc := []*profile.Location{
		{ID: 1, Address: 0x10},
		{ID: 2, Address: 0x20},
		{ID: 3, Address: 0x30},
		{ID: 4, Address: 0x40},
	}
	testProfile := &profile.Profile{
		Sample: []*profile.Sample{
			{Location: []*profile.Location{testLoc[0], testLoc[1], testLoc[3], testLoc[2]}},
			{Location: []*profile.Location{testLoc[1], testLoc[3], testLoc[2]}},
			{Location: []*profile.Location{testLoc[3], testLoc[2], testLoc[1]}},
			{Location: []*profile.Location{testLoc[3], testLoc[2], testLoc[0]}},
			{Location: []*profile.Location{testLoc[0], testLoc[1], testLoc[3], testLoc[0]}},
		},
		Location: testLoc,
	}
	testProfiles := make([]*profile.Profile, 2)
	testProfiles[0] = testProfile.Copy()
	testProfiles[1] = testProfile.Copy()
	// Test case for profilez.
	testProfiles[0].PeriodType = &profile.ValueType{Type: "cpu", Unit: "nanoseconds"}
	// Test case for heapz.
	testProfiles[1].PeriodType = &profile.ValueType{Type: "space", Unit: "bytes"}
	wantFunc := []*profile.Function{
		{ID: 1, Name: "runtime.goexit", SystemName: "runtime.goexit", Filename: "runtime.go"},
		{ID: 2, Name: "runtime.other", SystemName: "runtime.other", Filename: "runtime.go"},
		{ID: 3, Name: "foo", SystemName: "foo", Filename: "foo.go"},
		{ID: 4, Name: "bar", SystemName: "bar", Filename: "bar.go"},
	}
	wantLoc := []*profile.Location{
		{ID: 1, Address: 0x10, Line: []profile.Line{{Function: wantFunc[0], Line: 10}}},
		{ID: 2, Address: 0x20, Line: []profile.Line{{Function: wantFunc[1], Line: 20}}},
		{ID: 3, Address: 0x30, Line: []profile.Line{{Function: wantFunc[2], Line: 30}}},
		{ID: 4, Address: 0x40, Line: []profile.Line{{Function: wantFunc[3], Line: 40}}},
	}
	wantProfiles := []*profile.Profile{
		{
			PeriodType: &profile.ValueType{Type: "cpu", Unit: "nanoseconds"},
			Sample: []*profile.Sample{
				{Location: []*profile.Location{wantLoc[1], wantLoc[3], wantLoc[2]}},
				{Location: []*profile.Location{wantLoc[1], wantLoc[3], wantLoc[2]}},
				{Location: []*profile.Location{wantLoc[3], wantLoc[2], wantLoc[1]}},
				{Location: []*profile.Location{wantLoc[3], wantLoc[2]}},
				{Location: []*profile.Location{wantLoc[1], wantLoc[3]}},
			},
			Location: wantLoc,
			Function: wantFunc,
		},
		{
			PeriodType: &profile.ValueType{Type: "space", Unit: "bytes"},
			Sample: []*profile.Sample{
				{Location: []*profile.Location{wantLoc[3], wantLoc[2]}},
				{Location: []*profile.Location{wantLoc[3], wantLoc[2]}},
				{Location: []*profile.Location{wantLoc[3], wantLoc[2], wantLoc[1]}},
				{Location: []*profile.Location{wantLoc[3], wantLoc[2]}},
				{Location: []*profile.Location{wantLoc[3]}},
			},
			Location: wantLoc,
			Function: wantFunc,
		},
	}
	for i := 0; i < 2; i++ {
		symbolize(testProfiles[i])
		if !reflect.DeepEqual(testProfiles[i], wantProfiles[i]) {
			t.Errorf("incorrect trimming (testcase = %d): got {%v}, want {%v}", i, testProfiles[i], wantProfiles[i])
		}
	}
}

// Tests that goroutine profiles are parsed correctly
func TestGoroutineProfile(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			p, err := translateGoroutineProfile([]byte(goroutineTestProfile), startTime)
			if err != nil {
				t.Fatalf("translate failed: %v", err)
			}
			if err := p.Write(w); err != nil {
				t.Fatalf("write failed: %v", err)
			}
		}))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	p, err := profile.Parse(resp.Body)
	if err != nil {
		t.Fatalf("Could not parse Profile profile: %v", err)
	}
	// Expected PeriodType and SampleType.
	expectedPeriodType := &profile.ValueType{Type: "goroutine", Unit: "count"}
	expectedSampleType := []*profile.ValueType{
		{Type: "goroutine", Unit: "count"},
	}
	if !reflect.DeepEqual(p.PeriodType, expectedPeriodType) {
		t.Fatalf("Period types do not match")
	}
	if !reflect.DeepEqual(p.SampleType, expectedSampleType) {
		t.Fatalf("Sample types do not match")
	}
	var count, value int64
	for _, s := range p.Sample {
		count++
		value += s.Value[0]
	}
	if count != 5 || value != 24 {
		t.Errorf("Got %d samples, total %d, want %d samples, total %d", count, value, 5, 24)
	}
}

var goroutineTestProfile = `
# See /helpz and /debug/pprof for alternate output formats.
goroutine profile: total 24
20 @ 0x4306b3 0x43feb7 0x43f412 0x525f42 0x4625e1
1 @ 0x4306b3 0x430774 0x44cab9 0x70b6fc 0x4625e1
1 @ 0x4306b3 0x430774 0x44cab9 0x70bfa2 0x4625e1
1 @ 0x4625e1
1 @ 0x798fc8 0x798da3 0x794074 0x4adfd7 0x638b9a 0x667078 0x66acf4 0x6677c7 0x667ca9 0x72f541 0x638b9a 0x63aede 0x6378fe 0x4625e1
`
