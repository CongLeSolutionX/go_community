// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors implements functions to manipulate errors.
package errors

// New returns an error that formats as the given text.
func New(text string) error {
	return &errorString{text, nil}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
	w wrapper
}

func (e *errorString) Error() string {
	return e.s
}

func (e *errorString) Unwrap() wrapper {
	return e.w
}
