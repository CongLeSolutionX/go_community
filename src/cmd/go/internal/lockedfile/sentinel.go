// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lockedfile

import (
	"fmt"
	"os"
)

// A Sentinel is a file that is made nonempty to signal that some event has
// completed.
type Sentinel struct {
	f *File
}

// NewSentinel opens and locks a sentinel file for writing.
//
// If the sentinel already exists and is non-empty — that is, if the event it
// signals has already completed — NewSentinel returns a nil *Sentinel and an
// error matching os.IsExist.
//
// The returned Sentinel must be closed before the program exits.
func NewSentinel(name string) (_ *Sentinel, err error) {
	f, err := OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			f.Close()
		}
	}()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.Size() > 0 {
		return nil, &os.PathError{
			Op:   "NewSentinel",
			Path: f.Name(),
			Err:  os.ErrExist,
		}
	}

	return &Sentinel{f: f}, nil
}

// Name returns the name of the underlying file.
func (s *Sentinel) Name() string { return s.f.Name() }

func (s *Sentinel) String() string {
	return fmt.Sprintf("%T(%s)", s, s.f.Name())
}

// WriteString writes content to the sentinel file.
// If content is non-empty, this marks the sentinel's event as complete.
func (s *Sentinel) WriteString(content string) (int, error) {
	return s.f.WriteString(content)
}

// Close closes and unlocks (but does not delete) the sentinel file.
//
// It cannot safely delete the sentinel file, even if it is empty: the
// file-locking APIs are based on file descriptors, so when we close the file,
// the descriptor becomes invalid and some other process may lock (and write to)
// the file. If we concurrently delete that file we may undo the work of the
// other process.
func (s *Sentinel) Close() error {
	return s.f.Close()
}
