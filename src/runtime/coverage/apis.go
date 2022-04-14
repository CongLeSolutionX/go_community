// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package coverage

import (
	"fmt"
	"io"
)

// CoverageMetaDataEmitToDir writes a coverage meta-data file for the
// currently running program to the directory specified in 'dir'. An
// error will be returned if operation can't be completed successfully
// (for example, if the currently running program was not built with
// "-cover", or if the directory does not exist).
func CoverageMetaDataEmitToDir(dir string) error {
	if !finalHashComputed {
		return fmt.Errorf("error: no meta-data available (binary not built with -cover?)")
	}
	return emitMetaDataToDirectory(dir)
}

// CoverageMetaDataEmitToWriter writes the meta-data content (the
// payload that would normally be emitted to a meta-data file) for
// currently running program to the the writer 'w'. An error will be
// returned if operation can't be completed successfully (for example,
// if the currently running program was not built with "-cover", or if
// a write fails).
func CoverageMetaDataEmitToWriter(w io.Writer) error {
	if !finalHashComputed {
		return fmt.Errorf("error: no meta-data available (binary not built with -cover?)")
	}
	ml := getCovMetaList()
	return writeMetaData(w, ml, cmode, finalHash)
}

// CoverageCounterDataEmitToDir writes a coverage counter-data file for the
// currently running program to the directory specified in 'dir'. An
// error will be returned if operation can't be completed successfully
// (for example, if the currently running program was not built with
// "-cover", or if the directory does not exist). The counter data written
// will be a snapshot taken at the point of the invocation.
func CoverageCounterDataEmitToDir(dir string) error {
	return emitCounterDataToDirectory(dir)
}

// CoverageCounterDataEmitToWriter writes coverage counter-data
// content for the currently running program to the writer 'w'. An
// error will be returned if operation can't be completed successfully
// (for example, if the currently running program was not built with
// "-cover", or if a write fails). The counter data written will be a
// snapshot taken at the point of the invocation.
func CoverageCounterDataEmitToWriter(w io.Writer) error {
	// Ask the runtime for the list of coverage counter symbols.
	cl := getCovCounterList()
	if len(cl) == 0 {
		return fmt.Errorf("program not built with -coverage")
	}
	if !finalHashComputed {
		return fmt.Errorf("meta-data not written yet, unable to write counter data")
	}

	pm := getCovPkgMap()
	s := &emitState{
		counterlist: cl,
		pkgmap:      pm,
	}
	return s.emitCounterDataToWriter(w)
}
