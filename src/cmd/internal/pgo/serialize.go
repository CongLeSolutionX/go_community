// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pgo

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Serialization of a Profile allows go tool preprofile to construct the edge
// map only once (rather than once per compile process). The compiler processes
// then parse the pre-processed data directly from the serialized format.
//
// The format of the serialized output is as follows.
//
//      GO PREPROFILE V1
//      caller_name
//      callee_name
//      "call site offset" "call edge weight"
//      ...
//      caller_name
//      callee_name
//      "call site offset" "call edge weight"
//
// Entries are sorted by "call edge weight", from highest to lowest.

const serializationHeader = "GO PREPROFILE V1\n"

const licoToBlocksEnd = "!LCTBLKE"

// WriteTo writes a serialized representation of Profile to w.
//
// FromSerialized can parse the format back to Profile.
//
// WriteTo implements io.WriterTo.Write.
func (d *Profile) WriteTo(w io.Writer) (int64, error) {
	bw := bufio.NewWriter(w)

	var written int64

	// Header
	n, err := bw.WriteString(serializationHeader)
	written += int64(n)
	if err != nil {
		return written, err
	}

	for _, edge := range d.NamedEdgeMap.ByWeight {
		weight := d.NamedEdgeMap.Weight[edge]

		n, err = fmt.Fprintln(bw, edge.CallerName)
		written += int64(n)
		if err != nil {
			return written, err
		}

		n, err = fmt.Fprintln(bw, edge.CalleeName)
		written += int64(n)
		if err != nil {
			return written, err
		}

		n, err = fmt.Fprintf(bw, "%d %d\n", edge.CallSiteOffset, weight)
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	n, err = bw.WriteString(licoToBlocksEnd + "\n")
	written += int64(n)
	if err != nil {
		return written, err
	}
	for fName, m1 := range d.InlineProfile {
		fline := fName + "\n"
		for inlT, m2 := range m1 {
			iline := inlT + "\n"
			for l, m3 := range m2 {
				for c, freq := range m3 {
					n, err = bw.WriteString(fline)
					written += int64(n)
					if err != nil {
						return written, err
					}
					n, err = bw.WriteString(iline)
					written += int64(n)
					if err != nil {
						return written, err
					}
					line := strconv.FormatInt(int64(l), 10) + " " + strconv.FormatInt(int64(c), 10) + " " + strconv.FormatInt(int64(freq), 10) + "\n"
					n, err = bw.WriteString(line)
					written += int64(n)
					if err != nil {
						return written, err
					}
				}
			}
		}
	}

	if err := bw.Flush(); err != nil {
		return written, err
	}

	// No need to serialize TotalWeight, it can be trivially recomputed
	// during parsing.

	return written, nil
}
