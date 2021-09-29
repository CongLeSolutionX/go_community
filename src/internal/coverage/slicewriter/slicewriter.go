// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slicewriter

import (
	"fmt"
	"os"
)

// WriteSeeker is a helper object that implements the io.WriteSeeker
// interface. Clients can create a WriteSeeker, make a series of Write
// calls to add data to it (and possibly Seek calls to update
// previously written portions), then finally invoke BytesWritten() to
// get a pointer to the constructed byte slice.
type WriteSeeker struct {
	payload []byte
	off     int64
}

func (sws *WriteSeeker) Write(p []byte) (n int, err error) {
	amt := len(p)
	towrite := sws.payload[sws.off:]
	if len(towrite) < amt {
		sws.payload = append(sws.payload, make([]byte, amt-len(towrite))...)
		towrite = sws.payload[sws.off:]
	}
	copy(towrite, p)
	sws.off += int64(amt)
	return amt, nil
}

func (sws *WriteSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_SET:
		sws.off = offset
		return offset, nil
	case os.SEEK_CUR:
		sws.off += offset
		return sws.off, nil
	case os.SEEK_END:
		newoff := int64(len(sws.payload)) + offset
		if newoff < 0 || newoff > int64(len(sws.payload)) {
			return 0, fmt.Errorf("invalid seek: new offset %d (out of range [0 %d]", newoff, len(sws.payload))
		}
		sws.off = newoff
		return sws.off, nil
	}
	// other modes not supported
	panic(fmt.Sprintf("unsupported seek mode %d", whence))
}

// BytesWritten returns the underlying byte slice for the WriteSeeker,
// containing the data written to it via Write/Seek calls.
func (sws *WriteSeeker) BytesWritten() []byte {
	return sws.payload
}

func (sws *WriteSeeker) Read(p []byte) (n int, err error) {
	amt := len(p)
	toread := sws.payload[sws.off:]
	if len(toread) < amt {
		amt = len(toread)
	}
	copy(p, toread)
	sws.off += int64(amt)
	return amt, nil
}
