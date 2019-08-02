// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors implements functions to manipulate errors.
//
// Use the New function to create errors whose only content is a text message.
//
// The Unwrap, Is and As functions work on errors that may wrap other errors.
// An error type wraps another error if the type has the method
//
//	Unwrap() error
//
// If e.Unwrap() returns a non-nil error w, then we say that e wraps w.
//
// You can create your own error types that wrap errors, or you can call
// fmt.Errorf with the %w verb and a corresponding error argument.
//
//	fmt.Errorf("... %w ...", ..., w, ...).Unwrap()
//
// returns w.
//
// Use the Unwrap function to get wrapped errors. If its argument's type has an
// Unwrap method, it calls the method once. Otherwise, it returns nil.
//
// Use the Is function to compare errors for equality. It compares
// its arguments using ==, but if that fails it calls Unwrap on its first argument
// and repeats the comparison, looping until a comparison succeeds or Unwrap returns
// nil. Use Is to compare an error to a particular error value. Writing
//
//	if errors.Is(err, os.ErrExist)
//
// is preferable to
//
//	if err == os.ErrExist
//
// because the former will succeed even if err wraps os.ErrExist.
//
// (To preserve compatability, the function os.IsExist, and other, similar functions
// in the os package and elsewhere in the standard library, do not unwrap their
// argument.)
//
// Use the As function to determine if an error is of a given type. It behaves much
// like Is, but instead of testing for equality it tests whether its first argument
// is assignable to the type of its second. It also sets its second argument, which
// must be a pointer, to the matching error for subsequent inspection. Writing
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
// because the former will succeed even if err wraps an *os.PathError.
package errors

// New returns an error that formats as the given text.
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
