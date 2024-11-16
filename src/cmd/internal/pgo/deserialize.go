// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pgo

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// IsSerialized returns true if r is a serialized Profile.
//
// IsSerialized only peeks at r, so seeking back after calling is not
// necessary.
func IsSerialized(r *bufio.Reader) (bool, error) {
	hdr, err := r.Peek(len(serializationHeader))
	if err == io.EOF {
		// Empty file.
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error reading profile header: %w", err)
	}

	return string(hdr) == serializationHeader, nil
}

// FromSerialized parses a profile from serialization output of Profile.WriteTo.
func FromSerialized(r io.Reader) (*Profile, error) {
	d := emptyProfile()

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading preprocessed profile: %w", err)
		}
		return nil, fmt.Errorf("preprocessed profile missing header")
	}
	if gotHdr := scanner.Text() + "\n"; gotHdr != serializationHeader {
		return nil, fmt.Errorf("preprocessed profile malformed header; got %q want %q", gotHdr, serializationHeader)
	}

	for scanner.Scan() {
		readStr := scanner.Text()
		if readStr == licoToBlocksEnd {
			break
		}

		callerName := readStr

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading preprocessed profile: %w", err)
			}
			return nil, fmt.Errorf("preprocessed profile entry missing callee")
		}
		calleeName := scanner.Text()

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading preprocessed profile: %w", err)
			}
			return nil, fmt.Errorf("preprocessed profile entry missing weight")
		}
		readStr = scanner.Text()

		split := strings.Split(readStr, " ")

		if len(split) != 2 {
			return nil, fmt.Errorf("preprocessed profile entry got %v want 2 fields", split)
		}

		co, err := strconv.Atoi(split[0])
		if err != nil {
			return nil, fmt.Errorf("preprocessed profile error processing call line: %w", err)
		}

		edge := NamedCallEdge{
			CallerName:     callerName,
			CalleeName:     calleeName,
			CallSiteOffset: co,
		}

		weight, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("preprocessed profile error processing call weight: %w", err)
		}

		if _, ok := d.NamedEdgeMap.Weight[edge]; ok {
			return nil, fmt.Errorf("preprocessed profile contains duplicate edge %+v", edge)
		}

		d.NamedEdgeMap.ByWeight = append(d.NamedEdgeMap.ByWeight, edge) // N.B. serialization is ordered.
		d.NamedEdgeMap.Weight[edge] += weight
		d.TotalWeight += weight
	}

	d.InlineProfile = make(map[string]map[string]map[int]map[int]int)
	for scanner.Scan() {
		fName := scanner.Text()
		if !scanner.Scan() {
			return nil, fmt.Errorf("preprocessed profile error processing inline profile, missing inline tree")
		}
		inlT := scanner.Text()
		if !scanner.Scan() {
			return nil, fmt.Errorf("preprocessed profile error processing inline profile, missing lico freq map entry")
		}
		readStr := scanner.Text()
		split := strings.Split(readStr, " ")
		if len(split) != 3 {
			return nil, fmt.Errorf("preprocessed profile error processing inline profile, lico freq map entry format error, expect 'line col freq', got %s", readStr)
		}
		if d.InlineProfile[fName] == nil {
			d.InlineProfile[fName] = make(map[string]map[int]map[int]int)
		}
		if d.InlineProfile[fName][inlT] == nil {
			d.InlineProfile[fName][inlT] = make(map[int]map[int]int)
		}
		l, _ := strconv.ParseInt(split[0], 10, 64)
		c, _ := strconv.ParseInt(split[1], 10, 64)
		freq, _ := strconv.ParseInt(split[2], 10, 64)
		if d.InlineProfile[fName][inlT][int(l)] == nil {
			d.InlineProfile[fName][inlT][int(l)] = make(map[int]int)
		}
		d.InlineProfile[fName][inlT][int(l)][int(c)] += int(freq)
	}

	return d, nil

}
