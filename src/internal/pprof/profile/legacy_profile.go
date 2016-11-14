// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements parsers to convert legacy profiles into the
// profile.proto format.

package profile

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	countStartRE = regexp.MustCompile(`\A(\w+) profile: total \d+\z`)
	countRE      = regexp.MustCompile(`\A(\d+) @(( 0x[0-9a-f]+)+)\z`)

	heapHeaderRE = regexp.MustCompile(`heap profile: *(\d+): *(\d+) *\[ *(\d+): *(\d+) *\] *@ *(heap[_a-z0-9]*)/?(\d*)`)
	heapSampleRE = regexp.MustCompile(`(-?\d+): *(-?\d+) *\[ *(\d+): *(\d+) *] @([ x0-9a-f]*)`)

	contentionSampleRE = regexp.MustCompile(`(\d+) *(\d+) @([ x0-9a-f]*)`)

	hexNumberRE = regexp.MustCompile(`0x[0-9a-f]+`)

	growthHeaderRE = regexp.MustCompile(`heap profile: *(\d+): *(\d+) *\[ *(\d+): *(\d+) *\] @ growthz?`)

	fragmentationHeaderRE = regexp.MustCompile(`heap profile: *(\d+): *(\d+) *\[ *(\d+): *(\d+) *\] @ fragmentationz?`)

	threadzStartRE = regexp.MustCompile(`--- threadz \d+ ---`)
	threadStartRE  = regexp.MustCompile(`--- Thread ([[:xdigit:]]+) \(name: (.*)/(\d+)\) stack: ---`)

	procMapsRE = regexp.MustCompile(`([[:xdigit:]]+)-([[:xdigit:]]+)\s+([-rwxp]+)\s+([[:xdigit:]]+)\s+([[:xdigit:]]+):([[:xdigit:]]+)\s+([[:digit:]]+)\s*(\S+)?`)

	briefMapsRE = regexp.MustCompile(`\s*([[:xdigit:]]+)-([[:xdigit:]]+):\s*(\S+)(\s.*@)?([[:xdigit:]]+)?`)
)

func isSpaceOrComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	return len(trimmed) == 0 || trimmed[0] == '#'
}

// parseGoCount parses a Go count profile (e.g., threadcreate or
// goroutine) and returns a new Profile.
func parseGoCount(b []byte) (*Profile, error) {
	s := bufio.NewScanner(bytes.NewBuffer(b))
	// Skip comments at the beginning of the file.
	for s.Scan() && isSpaceOrComment(s.Text()) {
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	m := countStartRE.FindStringSubmatch(s.Text())
	if m == nil {
		return nil, errUnrecognized
	}
	profileType := m[1]
	p := &Profile{
		PeriodType: &ValueType{Type: profileType, Unit: "count"},
		Period:     1,
		SampleType: []*ValueType{{Type: profileType, Unit: "count"}},
	}
	locations := make(map[uint64]*Location)
	for s.Scan() {
		line := s.Text()
		if isSpaceOrComment(line) {
			continue
		}
		if strings.HasPrefix(line, "---") {
			break
		}
		m := countRE.FindStringSubmatch(line)
		if m == nil {
			return nil, errMalformed
		}
		n, err := strconv.ParseInt(m[1], 0, 64)
		if err != nil {
			return nil, errMalformed
		}
		fields := strings.Fields(m[2])
		locs := make([]*Location, 0, len(fields))
		for _, stk := range fields {
			addr, err := strconv.ParseUint(stk, 0, 64)
			if err != nil {
				return nil, errMalformed
			}
			// Adjust all frames by -1 to land on top of the call instruction.
			addr--
			loc := locations[addr]
			if loc == nil {
				loc = &Location{
					Address: addr,
				}
				locations[addr] = loc
				p.Location = append(p.Location, loc)
			}
			locs = append(locs, loc)
		}
		p.Sample = append(p.Sample, &Sample{
			Location: locs,
			Value:    []int64{n},
		})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	if err := parseAdditionalSections(s, p); err != nil {
		return nil, err
	}
	return p, nil
}

// remapLocationIDs ensures there is a location for each address
// referenced by a sample, and remaps the samples to point to the new
// location ids.
func (p *Profile) remapLocationIDs() {
	seen := make(map[*Location]bool, len(p.Location))
	var locs []*Location

	for _, s := range p.Sample {
		for _, l := range s.Location {
			if seen[l] {
				continue
			}
			l.ID = uint64(len(locs) + 1)
			locs = append(locs, l)
			seen[l] = true
		}
	}
	p.Location = locs
}

func (p *Profile) remapFunctionIDs() {
	seen := make(map[*Function]bool, len(p.Function))
	var fns []*Function

	for _, l := range p.Location {
		for _, ln := range l.Line {
			fn := ln.Function
			if fn == nil || seen[fn] {
				continue
			}
			fn.ID = uint64(len(fns) + 1)
			fns = append(fns, fn)
			seen[fn] = true
		}
	}
	p.Function = fns
}

// remapMappingIDs matches location addresses with existing mappings
// and updates them appropriately. This is O(N*M), if this ever shows
// up as a bottleneck, evaluate sorting the mappings and doing a
// binary search, which would make it O(N*log(M)).
func (p *Profile) remapMappingIDs() {
	// Some profile handlers will incorrectly set regions for the main
	// executable if its section is remapped. Fix them through heuristics.

	if len(p.Mapping) > 0 {
		// Remove the initial mapping if named '/anon_hugepage' and has a
		// consecutive adjacent mapping.
		if m := p.Mapping[0]; strings.HasPrefix(m.File, "/anon_hugepage") {
			if len(p.Mapping) > 1 && m.Limit == p.Mapping[1].Start {
				p.Mapping = p.Mapping[1:]
			}
		}
	}

	// Subtract the offset from the start of the main mapping if it
	// ends up at a recognizable start address.
	if len(p.Mapping) > 0 {
		const expectedStart = 0x400000
		if m := p.Mapping[0]; m.Start-m.Offset == expectedStart {
			m.Start = expectedStart
			m.Offset = 0
		}
	}

	// Associate each location with an address to the corresponding
	// mapping. Create fake mapping if a suitable one isn't found.
	var fake *Mapping
nextLocation:
	for _, l := range p.Location {
		a := l.Address
		if l.Mapping != nil || a == 0 {
			continue
		}
		for _, m := range p.Mapping {
			if m.Start <= a && a < m.Limit {
				l.Mapping = m
				continue nextLocation
			}
		}
		// Work around legacy handlers failing to encode the first
		// part of mappings split into adjacent ranges.
		for _, m := range p.Mapping {
			if m.Offset != 0 && m.Start-m.Offset <= a && a < m.Start {
				m.Start -= m.Offset
				m.Offset = 0
				l.Mapping = m
				continue nextLocation
			}
		}
		// If there is still no mapping, create a fake one.
		// This is important for the Go legacy handler, which produced
		// no mappings.
		if fake == nil {
			fake = &Mapping{
				ID:    1,
				Limit: ^uint64(0),
			}
			p.Mapping = append(p.Mapping, fake)
		}
		l.Mapping = fake
	}

	// Reset all mapping IDs.
	for i, m := range p.Mapping {
		m.ID = uint64(i + 1)
	}
}

var cpuInts = []func([]byte) (uint64, []byte){
	get32l,
	get32b,
	get64l,
	get64b,
}

func get32l(b []byte) (uint64, []byte) {
	if len(b) < 4 {
		return 0, nil
	}
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24, b[4:]
}

func get32b(b []byte) (uint64, []byte) {
	if len(b) < 4 {
		return 0, nil
	}
	return uint64(b[3]) | uint64(b[2])<<8 | uint64(b[1])<<16 | uint64(b[0])<<24, b[4:]
}

func get64l(b []byte) (uint64, []byte) {
	if len(b) < 8 {
		return 0, nil
	}
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56, b[8:]
}

func get64b(b []byte) (uint64, []byte) {
	if len(b) < 8 {
		return 0, nil
	}
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 | uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56, b[8:]
}

// parseCPU parses a profilez legacy profile and returns a newly
// populated Profile.
//
// The general format for profilez samples is a sequence of words in
// binary format. The first words are a header with the following data:
//   1st word -- 0
//   2nd word -- 3
//   3rd word -- 0 if a c++ application, 1 if a java application.
//   4th word -- Sampling period (in microseconds).
//   5th word -- Padding.
func parseCPU(b []byte) (*Profile, error) {
	var parse func([]byte) (uint64, []byte)
	var n1, n2, n3, n4, n5 uint64
	for _, parse = range cpuInts {
		var tmp []byte
		n1, tmp = parse(b)
		n2, tmp = parse(tmp)
		n3, tmp = parse(tmp)
		n4, tmp = parse(tmp)
		n5, tmp = parse(tmp)

		if tmp != nil && n1 == 0 && n2 == 3 && n3 == 0 && n4 > 0 && n5 == 0 {
			b = tmp
			return cpuProfile(b, int64(n4), parse)
		}
		if tmp != nil && n1 == 0 && n2 == 3 && n3 == 1 && n4 > 0 && n5 == 0 {
			b = tmp
			return javaCPUProfile(b, int64(n4), parse)
		}
	}
	return nil, errUnrecognized
}

// cpuProfile returns a new Profile from C++ profilez data.
// b is the profile bytes after the header, period is the profiling
// period, and parse is a function to parse 8-byte chunks from the
// profile in its native endianness.
func cpuProfile(b []byte, period int64, parse func(b []byte) (uint64, []byte)) (*Profile, error) {
	p := &Profile{
		Period:     period * 1000,
		PeriodType: &ValueType{Type: "cpu", Unit: "nanoseconds"},
		SampleType: []*ValueType{
			{Type: "samples", Unit: "count"},
			{Type: "cpu", Unit: "nanoseconds"},
		},
	}
	var err error
	if b, _, err = parseCPUSamples(b, parse, true, p); err != nil {
		return nil, err
	}

	// If *most* samples have the same second-to-the-bottom frame, it
	// strongly suggests that it is an uninteresting artifact of
	// measurement -- a stack frame pushed by the signal handler. The
	// bottom frame is always correct as it is picked up from the signal
	// structure, not the stack. Check if this is the case and if so,
	// remove.

	// Remove up to two frames.
	maxiter := 2
	// Allow one different sample for this many samples with the same
	// second-to-last frame.
	similarSamples := 32
	margin := len(p.Sample) / similarSamples

	for iter := 0; iter < maxiter; iter++ {
		addr1 := make(map[uint64]int)
		for _, s := range p.Sample {
			if len(s.Location) > 1 {
				a := s.Location[1].Address
				addr1[a] = addr1[a] + 1
			}
		}

		for id1, count := range addr1 {
			if count >= len(p.Sample)-margin {
				// Found uninteresting frame, strip it out from all samples
				for _, s := range p.Sample {
					if len(s.Location) > 1 && s.Location[1].Address == id1 {
						s.Location = append(s.Location[:1], s.Location[2:]...)
					}
				}
				break
			}
		}
	}

	if err := p.ParseMemoryMap(bytes.NewBuffer(b)); err != nil {
		return nil, err
	}

	cleanupDuplicateLocations(p)
	return p, nil
}

func cleanupDuplicateLocations(p *Profile) {
	// The profile handler may duplicate the leaf frame, because it gets
	// its address both from stack unwinding and from the signal
	// context. Detect this and delete the duplicate, which has been
	// adjusted by -1. The leaf address should not be adjusted as it is
	// not a call.
	for _, s := range p.Sample {
		if len(s.Location) > 1 && s.Location[0].Address == s.Location[1].Address+1 {
			s.Location = append(s.Location[:1], s.Location[2:]...)
		}
	}
}

// parseCPUSamples parses a collection of profilez samples from a
// profile.
//
// profilez samples are a repeated sequence of stack frames of the
// form:
//    1st word -- The number of times this stack was encountered.
//    2nd word -- The size of the stack (StackSize).
//    3rd word -- The first address on the stack.
//    ...
//    StackSize + 2 -- The last address on the stack
// The last stack trace is of the form:
//   1st word -- 0
//   2nd word -- 1
//   3rd word -- 0
//
// Addresses from stack traces may point to the next instruction after
// each call. Optionally adjust by -1 to land somewhere on the actual
// call (except for the leaf, which is not a call).
func parseCPUSamples(b []byte, parse func(b []byte) (uint64, []byte), adjust bool, p *Profile) ([]byte, map[uint64]*Location, error) {
	locs := make(map[uint64]*Location)
	for len(b) > 0 {
		var count, nstk uint64
		count, b = parse(b)
		nstk, b = parse(b)
		if b == nil || nstk > uint64(len(b)/4) {
			return nil, nil, errUnrecognized
		}
		var sloc []*Location
		addrs := make([]uint64, nstk)
		for i := 0; i < int(nstk); i++ {
			addrs[i], b = parse(b)
		}

		if count == 0 && nstk == 1 && addrs[0] == 0 {
			// End of data marker
			break
		}
		for i, addr := range addrs {
			if adjust && i > 0 {
				addr--
			}
			loc := locs[addr]
			if loc == nil {
				loc = &Location{
					Address: addr,
				}
				locs[addr] = loc
				p.Location = append(p.Location, loc)
			}
			sloc = append(sloc, loc)
		}
		p.Sample = append(p.Sample,
			&Sample{
				Value:    []int64{int64(count), int64(count) * p.Period},
				Location: sloc,
			})
	}
	// Reached the end without finding the EOD marker.
	return b, locs, nil
}

// parseHeap parses a heapz legacy or a growthz profile and
// returns a newly populated Profile.
func parseHeap(b []byte) (p *Profile, err error) {
	s := bufio.NewScanner(bytes.NewBuffer(b))
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}
		return nil, errUnrecognized
	}
	p = &Profile{}

	sampling := ""
	hasAlloc := false

	line := s.Text()
	p.PeriodType = &ValueType{Type: "space", Unit: "bytes"}
	if header := heapHeaderRE.FindStringSubmatch(line); header != nil {
		sampling, p.Period, hasAlloc, err = parseHeapHeader(line)
		if err != nil {
			return nil, err
		}
	} else if header = growthHeaderRE.FindStringSubmatch(line); header != nil {
		p.Period = 1
	} else if header = fragmentationHeaderRE.FindStringSubmatch(line); header != nil {
		p.Period = 1
	} else {
		return nil, errUnrecognized
	}

	if hasAlloc {
		// Put alloc before inuse so that default pprof selection
		// will prefer inuse_space.
		p.SampleType = []*ValueType{
			{Type: "alloc_objects", Unit: "count"},
			{Type: "alloc_space", Unit: "bytes"},
			{Type: "inuse_objects", Unit: "count"},
			{Type: "inuse_space", Unit: "bytes"},
		}
	} else {
		p.SampleType = []*ValueType{
			{Type: "objects", Unit: "count"},
			{Type: "space", Unit: "bytes"},
		}
	}

	locs := make(map[uint64]*Location)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())

		if isSpaceOrComment(line) {
			continue
		}

		if sectionTrigger(line) != unrecognizedSection {
			break
		}

		value, blocksize, addrs, err := parseHeapSample(line, p.Period, sampling, hasAlloc)
		if err != nil {
			return nil, err
		}

		var sloc []*Location
		for _, addr := range addrs {
			// Addresses from stack traces point to the next instruction after
			// each call. Adjust by -1 to land somewhere on the actual call.
			addr--
			loc := locs[addr]
			if locs[addr] == nil {
				loc = &Location{
					Address: addr,
				}
				p.Location = append(p.Location, loc)
				locs[addr] = loc
			}
			sloc = append(sloc, loc)
		}

		p.Sample = append(p.Sample, &Sample{
			Value:    value,
			Location: sloc,
			NumLabel: map[string][]int64{"bytes": {blocksize}},
		})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if err := parseAdditionalSections(s, p); err != nil {
		return nil, err
	}
	return p, nil
}

func parseHeapHeader(line string) (sampling string, period int64, hasAlloc bool, err error) {
	header := heapHeaderRE.FindStringSubmatch(line)
	if header == nil {
		return "", 0, false, errUnrecognized
	}

	if len(header[6]) > 0 {
		if period, err = strconv.ParseInt(header[6], 10, 64); err != nil {
			return "", 0, false, errUnrecognized
		}
	}

	if (header[3] != header[1] && header[3] != "0") || (header[4] != header[2] && header[4] != "0") {
		hasAlloc = true
	}

	switch header[5] {
	case "heapz_v2", "heap_v2":
		return "v2", period, hasAlloc, nil
	case "heapprofile":
		return "", 1, hasAlloc, nil
	case "heap":
		return "v2", period / 2, true, nil
	default:
		return "", 0, false, errUnrecognized
	}
}

// parseHeapSample parses a single row from a heap profile into a new Sample.
func parseHeapSample(line string, rate int64, sampling string, includeAlloc bool) (value []int64, blocksize int64, addrs []uint64, err error) {
	sampleData := heapSampleRE.FindStringSubmatch(line)
	if len(sampleData) != 6 {
		return nil, 0, nil, fmt.Errorf("unexpected number of sample values: got %d, want 6", len(sampleData))
	}

	// This is a local-scoped helper function to avoid needing to pass
	// around rate, sampling and many return parameters.
	addValues := func(countString, sizeString string, label string) error {
		count, err := strconv.ParseInt(countString, 10, 64)
		if err != nil {
			return fmt.Errorf("malformed sample: %s: %v", line, err)
		}
		size, err := strconv.ParseInt(sizeString, 10, 64)
		if err != nil {
			return fmt.Errorf("malformed sample: %s: %v", line, err)
		}
		if count == 0 && size != 0 {
			return fmt.Errorf("%s count was 0 but %s bytes was %d", label, label, size)
		}
		if count != 0 {
			blocksize = size / count
			if sampling == "v2" {
				count, size = scaleHeapSample(count, size, rate)
			}
		}
		value = append(value, count, size)
		return nil
	}

	if includeAlloc {
		if err := addValues(sampleData[3], sampleData[4], "allocation"); err != nil {
			return nil, 0, nil, err
		}
	}

	if err := addValues(sampleData[1], sampleData[2], "inuse"); err != nil {
		return nil, 0, nil, err
	}

	addrs = parseHexAddresses(sampleData[5])

	return value, blocksize, addrs, nil
}

// extractHexAddresses extracts hex numbers from a string and returns
// them, together with their numeric value, in a slice.
func extractHexAddresses(s string) ([]string, []uint64) {
	hexStrings := hexNumberRE.FindAllString(s, -1)
	var ids []uint64
	for _, s := range hexStrings {
		if id, err := strconv.ParseUint(s, 0, 64); err == nil {
			ids = append(ids, id)
		} else {
			// Do not expect any parsing failures due to the regexp matching.
			panic("failed to parse hex value:" + s)
		}
	}
	return hexStrings, ids
}

// parseHexAddresses parses hex numbers from a string and returns them
// in a slice.
func parseHexAddresses(s string) []uint64 {
	_, ids := extractHexAddresses(s)
	return ids
}

// scaleHeapSample adjusts the data from a heapz Sample to
// account for its probability of appearing in the collected
// data. heapz profiles are a sampling of the memory allocations
// requests in a program. We estimate the unsampled value by dividing
// each collected sample by its probability of appearing in the
// profile. heapz v2 profiles rely on a poisson process to determine
// which samples to collect, based on the desired average collection
// rate R. The probability of a sample of size S to appear in that
// profile is 1-exp(-S/R).
func scaleHeapSample(count, size, rate int64) (int64, int64) {
	if count == 0 || size == 0 {
		return 0, 0
	}

	if rate <= 1 {
		// if rate==1 all samples were collected so no adjustment is needed.
		// if rate<1 treat as unknown and skip scaling.
		return count, size
	}

	avgSize := float64(size) / float64(count)
	scale := 1 / (1 - math.Exp(-avgSize/float64(rate)))

	return int64(float64(count) * scale), int64(float64(size) * scale)
}

// parseContention parses a contentionz profile and returns a newly
// populated Profile.
func parseContention(b []byte) (p *Profile, err error) {
	s := bufio.NewScanner(bytes.NewBuffer(b))
	if !s.Scan() {
		if err := s.Err(); err != nil {
			return nil, err
		}
		return nil, errUnrecognized
	}
	line := s.Text()

	if !strings.HasPrefix(line, "--- contention") {
		return nil, errUnrecognized
	}

	p = &Profile{
		PeriodType: &ValueType{Type: "contentions", Unit: "count"},
		Period:     1,
		SampleType: []*ValueType{
			{Type: "contentions", Unit: "count"},
			{Type: "delay", Unit: "nanoseconds"},
		},
	}

	var cpuHz int64
	// Parse text of the form "attribute = value" before the samples.
	const delimiter = "="
	for s.Scan() {
		line := s.Text()

		if line = strings.TrimSpace(line); line == "" {
			continue
		}

		if strings.HasPrefix(line, "---") {
			break
		}

		attr := strings.SplitN(line, delimiter, 2)
		if len(attr) != 2 {
			break
		}
		key, val := strings.TrimSpace(attr[0]), strings.TrimSpace(attr[1])
		var err error
		switch key {
		case "cycles/second":
			if cpuHz, err = strconv.ParseInt(val, 0, 64); err != nil {
				return nil, errUnrecognized
			}
		case "sampling period":
			if p.Period, err = strconv.ParseInt(val, 0, 64); err != nil {
				return nil, errUnrecognized
			}
		case "ms since reset":
			ms, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				return nil, errUnrecognized
			}
			p.DurationNanos = ms * 1000 * 1000
		case "format":
			// CPP contentionz profiles don't have format.
			return nil, errUnrecognized
		case "resolution":
			// CPP contentionz profiles don't have resolution.
			return nil, errUnrecognized
		case "discarded samples":
		default:
			return nil, errUnrecognized
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	locs := make(map[uint64]*Location)
	for {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "---") {
			break
		}
		value, addrs, err := parseContentionSample(line, p.Period, cpuHz)
		if err != nil {
			return nil, err
		}
		var sloc []*Location
		for _, addr := range addrs {
			// Addresses from stack traces point to the next instruction after
			// each call. Adjust by -1 to land somewhere on the actual call.
			addr--
			loc := locs[addr]
			if locs[addr] == nil {
				loc = &Location{
					Address: addr,
				}
				p.Location = append(p.Location, loc)
				locs[addr] = loc
			}
			sloc = append(sloc, loc)
		}
		p.Sample = append(p.Sample, &Sample{
			Value:    value,
			Location: sloc,
		})

		if !s.Scan() {
			break
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	if err = parseAdditionalSections(s, p); err != nil {
		return nil, err
	}

	return p, nil
}

// parseContentionSample parses a single row from a contention profile
// into a new Sample.
func parseContentionSample(line string, period, cpuHz int64) (value []int64, addrs []uint64, err error) {
	sampleData := contentionSampleRE.FindStringSubmatch(line)
	if sampleData == nil {
		return value, addrs, errUnrecognized
	}

	v1, err := strconv.ParseInt(sampleData[1], 10, 64)
	if err != nil {
		return value, addrs, fmt.Errorf("malformed sample: %s: %v", line, err)
	}
	v2, err := strconv.ParseInt(sampleData[2], 10, 64)
	if err != nil {
		return value, addrs, fmt.Errorf("malformed sample: %s: %v", line, err)
	}

	// Unsample values if period and cpuHz are available.
	// - Delays are scaled to cycles and then to nanoseconds.
	// - Contentions are scaled to cycles.
	if period > 0 {
		if cpuHz > 0 {
			cpuGHz := float64(cpuHz) / 1e9
			v1 = int64(float64(v1) * float64(period) / cpuGHz)
		}
		v2 = v2 * period
	}

	value = []int64{v2, v1}
	addrs = parseHexAddresses(sampleData[3])

	return value, addrs, nil
}

// parseThread parses a Threadz profile and returns a new Profile.
func parseThread(b []byte) (*Profile, error) {
	s := bufio.NewScanner(bytes.NewBuffer(b))
	// Skip past comments and empty lines seeking a real header.
	for s.Scan() && isSpaceOrComment(s.Text()) {
	}

	line := s.Text()
	if m := threadzStartRE.FindStringSubmatch(line); m != nil {
		// Advance over initial comments until first stack trace.
		for s.Scan() {
			if line = s.Text(); sectionTrigger(line) != unrecognizedSection || strings.HasPrefix(line, "-") {
				break
			}
		}
	} else if t := threadStartRE.FindStringSubmatch(line); len(t) != 4 {
		return nil, errUnrecognized
	}

	p := &Profile{
		SampleType: []*ValueType{{Type: "thread", Unit: "count"}},
		PeriodType: &ValueType{Type: "thread", Unit: "count"},
		Period:     1,
	}

	locs := make(map[uint64]*Location)
	// Recognize each thread and populate profile samples.
	for sectionTrigger(line) == unrecognizedSection {
		if strings.HasPrefix(line, "---- no stack trace for") {
			line = ""
			break
		}
		if t := threadStartRE.FindStringSubmatch(line); len(t) != 4 {
			return nil, errUnrecognized
		}

		var addrs []uint64
		var err error
		line, addrs, err = parseThreadSample(s)
		if err != nil {
			return nil, errUnrecognized
		}
		if len(addrs) == 0 {
			// We got a --same as previous threads--. Bump counters.
			if len(p.Sample) > 0 {
				s := p.Sample[len(p.Sample)-1]
				s.Value[0]++
			}
			continue
		}

		var sloc []*Location
		for i, addr := range addrs {
			// Addresses from stack traces point to the next instruction after
			// each call. Adjust by -1 to land somewhere on the actual call
			// (except for the leaf, which is not a call).
			if i > 0 {
				addr--
			}
			loc := locs[addr]
			if locs[addr] == nil {
				loc = &Location{
					Address: addr,
				}
				p.Location = append(p.Location, loc)
				locs[addr] = loc
			}
			sloc = append(sloc, loc)
		}

		p.Sample = append(p.Sample, &Sample{
			Value:    []int64{1},
			Location: sloc,
		})
	}

	if err := parseAdditionalSections(s, p); err != nil {
		return nil, err
	}

	cleanupDuplicateLocations(p)
	return p, nil
}

// parseThreadSample parses a symbolized or unsymbolized stack trace.
// Returns the first line after the traceback, the sample (or nil if
// it hits a 'same-as-previous' marker) and an error.
func parseThreadSample(s *bufio.Scanner) (nextl string, addrs []uint64, err error) {
	var line string
	sameAsPrevious := false
	for s.Scan() {
		line = strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "---") {
			break
		}
		if strings.Contains(line, "same as previous thread") {
			sameAsPrevious = true
			continue
		}

		addrs = append(addrs, parseHexAddresses(line)...)
	}
	if err := s.Err(); err != nil {
		return "", nil, err
	}
	if sameAsPrevious {
		return line, nil, nil
	}
	return line, addrs, nil
}

// parseAdditionalSections parses any additional sections in the
// profile, ignoring any unrecognized sections.
func parseAdditionalSections(s *bufio.Scanner, p *Profile) error {
	for sectionTrigger(s.Text()) != memoryMapSection && s.Scan() {
	}
	if err := s.Err(); err != nil {
		return err
	}
	return p.parseMemoryMapFromScanner(s)
}

// ParseProcMaps parses a memory map in the format of /proc/self/maps.
// ParseMemoryMap should be called after setting on a profile to
// associate locations to the corresponding mapping based on their
// address.
func ParseProcMaps(rd io.Reader) ([]*Mapping, error) {
	s := bufio.NewScanner(rd)
	return parseProcMapsFromScanner(s)
}

func parseProcMapsFromScanner(s *bufio.Scanner) ([]*Mapping, error) {
	var mapping []*Mapping

	var attrs []string
	var r *strings.Replacer
	const delimiter = "="
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		if r != nil {
			line = r.Replace(line)
		}
		m, err := parseMappingEntry(line)
		if err != nil {
			if err == errUnrecognized {
				// Recognize assignments of the form: attr=value, and replace
				// $attr with value on subsequent mappings.
				if attr := strings.SplitN(line, delimiter, 2); len(attr) == 2 {
					attrs = append(attrs, "$"+strings.TrimSpace(attr[0]), strings.TrimSpace(attr[1]))
					r = strings.NewReplacer(attrs...)
				}
				// Ignore any unrecognized entries
				continue
			}
			return nil, err
		}
		if m == nil {
			continue
		}
		mapping = append(mapping, m)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return mapping, nil
}

// ParseMemoryMap parses a memory map in the format of
// /proc/self/maps, and overrides the mappings in the current profile.
// It renumbers the samples and locations in the profile correspondingly.
func (p *Profile) ParseMemoryMap(rd io.Reader) error {
	return p.parseMemoryMapFromScanner(bufio.NewScanner(rd))
}

func (p *Profile) parseMemoryMapFromScanner(s *bufio.Scanner) error {
	mapping, err := parseProcMapsFromScanner(s)
	if err != nil {
		return err
	}
	p.Mapping = append(p.Mapping, mapping...)
	p.massageMappings()
	p.remapLocationIDs()
	p.remapFunctionIDs()
	p.remapMappingIDs()
	return nil
}

func parseMappingEntry(l string) (*Mapping, error) {
	mapping := &Mapping{}
	var err error
	if me := procMapsRE.FindStringSubmatch(l); len(me) == 9 {
		if !strings.Contains(me[3], "x") {
			// Skip non-executable entries.
			return nil, nil
		}
		if mapping.Start, err = strconv.ParseUint(me[1], 16, 64); err != nil {
			return nil, errUnrecognized
		}
		if mapping.Limit, err = strconv.ParseUint(me[2], 16, 64); err != nil {
			return nil, errUnrecognized
		}
		if me[4] != "" {
			if mapping.Offset, err = strconv.ParseUint(me[4], 16, 64); err != nil {
				return nil, errUnrecognized
			}
		}
		mapping.File = me[8]
		return mapping, nil
	}

	if me := briefMapsRE.FindStringSubmatch(l); len(me) == 6 {
		if mapping.Start, err = strconv.ParseUint(me[1], 16, 64); err != nil {
			return nil, errUnrecognized
		}
		if mapping.Limit, err = strconv.ParseUint(me[2], 16, 64); err != nil {
			return nil, errUnrecognized
		}
		mapping.File = me[3]
		if me[5] != "" {
			if mapping.Offset, err = strconv.ParseUint(me[5], 16, 64); err != nil {
				return nil, errUnrecognized
			}
		}
		return mapping, nil
	}

	return nil, errUnrecognized
}

type sectionType int

const (
	unrecognizedSection sectionType = iota
	memoryMapSection
)

var memoryMapTriggers = []string{
	"--- Memory map: ---",
	"MAPPED_LIBRARIES:",
}

func sectionTrigger(line string) sectionType {
	for _, trigger := range memoryMapTriggers {
		if strings.Contains(line, trigger) {
			return memoryMapSection
		}
	}
	return unrecognizedSection
}

func (p *Profile) addLegacyFrameInfo() {
	switch {
	case isProfileType(p, heapzSampleTypes):
		p.DropFrames, p.KeepFrames = allocRxStr, allocSkipRxStr
	case isProfileType(p, contentionzSampleTypes):
		p.DropFrames, p.KeepFrames = lockRxStr, ""
	default:
		p.DropFrames, p.KeepFrames = cpuProfilerRxStr, ""
	}
}

var heapzSampleTypes = [][]string{
	{"allocations", "size"}, // early Go pprof profiles
	{"objects", "space"},
	{"inuse_objects", "inuse_space"},
	{"alloc_objects", "alloc_space"},
}
var contentionzSampleTypes = [][]string{
	{"contentions", "delay"},
}

func isProfileType(p *Profile, types [][]string) bool {
	st := p.SampleType
nextType:
	for _, t := range types {
		if len(st) != len(t) {
			continue
		}

		for i := range st {
			if st[i].Type != t[i] {
				continue nextType
			}
		}
		return true
	}
	return false
}

var allocRxStr = strings.Join([]string{
	// POSIX entry points.
	`calloc`,
	`cfree`,
	`malloc`,
	`free`,
	`memalign`,
	`do_memalign`,
	`(__)?posix_memalign`,
	`pvalloc`,
	`valloc`,
	`realloc`,

	// TC malloc.
	`tcmalloc::.*`,
	`tc_calloc`,
	`tc_cfree`,
	`tc_malloc`,
	`tc_free`,
	`tc_memalign`,
	`tc_posix_memalign`,
	`tc_pvalloc`,
	`tc_valloc`,
	`tc_realloc`,
	`tc_new`,
	`tc_delete`,
	`tc_newarray`,
	`tc_deletearray`,
	`tc_new_nothrow`,
	`tc_newarray_nothrow`,

	// Memory-allocation routines on OS X.
	`malloc_zone_malloc`,
	`malloc_zone_calloc`,
	`malloc_zone_valloc`,
	`malloc_zone_realloc`,
	`malloc_zone_memalign`,
	`malloc_zone_free`,

	// Go runtime
	`runtime\..*`,

	// Other misc. memory allocation routines
	`BaseArena::.*`,
	`(::)?do_malloc_no_errno`,
	`(::)?do_malloc_pages`,
	`(::)?do_malloc`,
	`DoSampledAllocation`,
	`MallocedMemBlock::MallocedMemBlock`,
	`_M_allocate`,
	`__builtin_(vec_)?delete`,
	`__builtin_(vec_)?new`,
	`__gnu_cxx::new_allocator::allocate`,
	`__libc_malloc`,
	`__malloc_alloc_template::allocate`,
	`allocate`,
	`cpp_alloc`,
	`operator new(\[\])?`,
	`simple_alloc::allocate`,
}, `|`)

var allocSkipRxStr = strings.Join([]string{
	// Preserve Go runtime frames that appear in the middle/bottom of
	// the stack.
	`runtime\.panic`,
}, `|`)

var cpuProfilerRxStr = strings.Join([]string{
	`ProfileData::Add`,
	`ProfileData::prof_handler`,
	`CpuProfiler::prof_handler`,
	`__pthread_sighandler`,
	`__restore`,
}, `|`)

var lockRxStr = strings.Join([]string{
	`RecordLockProfileData`,
	`(base::)?RecordLockProfileData.*`,
	`(base::)?SubmitMutexProfileData.*`,
	`(base::)?SubmitSpinLockProfileData.*`,
	`(base::Mutex::)?AwaitCommon.*`,
	`(base::Mutex::)?Unlock.*`,
	`(base::Mutex::)?UnlockSlow.*`,
	`(base::Mutex::)?ReaderUnlock.*`,
	`(base::MutexLock::)?~MutexLock.*`,
	`(Mutex::)?AwaitCommon.*`,
	`(Mutex::)?Unlock.*`,
	`(Mutex::)?UnlockSlow.*`,
	`(Mutex::)?ReaderUnlock.*`,
	`(MutexLock::)?~MutexLock.*`,
	`(SpinLock::)?Unlock.*`,
	`(SpinLock::)?SlowUnlock.*`,
	`(SpinLockHolder::)?~SpinLockHolder.*`,
}, `|`)
