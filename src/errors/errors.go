// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors implements functions to manipulate errors.
//
// Use the New function to create errors whose only content is a text message.
//
// The Unwrap, Is and As functions work on errors that may wrap other errors.
// An error type wraps another error if its type has the method
//
//	Unwrap() error
//
// If e.Unwrap() returns a non-nil error w, then we say that e wraps w.
//
// fmt.Errorf creates wrapped errors when called with the %w verb and a corresponding
// error argument.
//
//	fmt.Errorf("... %w ...", ..., w, ...).Unwrap()
//
// returns w.
//
// The Unwrap function gets wrapped errors. If its argument's type has an
// Unwrap method, it calls the method once. Otherwise, it returns nil.
//
// The Is function compares errors for equality. It compares
// its arguments using ==, but if that fails it calls Unwrap on its first argument
// and repeats the comparison, looping until a comparison succeeds or Unwrap returns
// nil. Is should be used to compare an error to a particular error value. Writing
//
//	if errors.Is(err, os.ErrExist)
//
// is preferable to
//
//	if err == os.ErrExist
//
// because the former will succeed if err wraps os.ErrExist.
//
// The As function determines if an error is of a given type. The second
// argument must be a pointer of some type PE pointing to some value E. As behaves
// much like Is, but instead of testing for equality is tests for assignability. If
// the first argument is assignable to *PE, As sets E to the value. If that fails, As
// calls Unwrap on the first argument and repeats the test, looping until
// assignability succeeds or Unwrap returns nil. As reports whether it succeeded in
// making an assignment. Writing
//
//	var perr *os.PathError
//	if errors.As(err, &perr) {
//		fmt.Println(perr.Path)
//  }
//
// is preferable to
//
//	if perr, ok := err.(*os.PathError); ok {
//		fmt.Println(perr.Path)
//	}
//
// because the former will succeed if err wraps an *os.PathError.
package errors

// New returns an error that formats as the given text.
// Each call to new returns a distinct error value even if the text is identical.
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
