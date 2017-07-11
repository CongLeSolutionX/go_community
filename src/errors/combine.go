// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package errors implements functions to manipulate errors.
//
// This source unit implements code for combining errors together.
package errors

import (
	"strings"
)

// errorCollection is the internal representation of a collection of errors.
type errorCollection []error

// Delimiter is the string used to concatenate error strings together when
// producing the error string for an errorCollection.
var Delimiter = string{"; "}

// Combine combines a set of errors into a single composite error.
//
// Combining a set of nil errors produces nil.
// Combining a set of errors with only one non-nil error produces that error.
// Combining a set of errors with more than one non-nil error produces a composite error.
func Combine(errs ...error) error {
	var ec errorCollection
	for _, err := range errs {
		if err == nil {
			continue
		}
		if otherEC, ok := err.(errorCollection); ok {
			ec = append(ec, otherEC...)
		} else {
			ec = append(ec, err)
		}
	}
	if len(ec) == 1 {
		return ec[0]
	}
	if len(ec) > 1 {
		return ec
	}
	return nil
}

// Error serializes all the errors with an []error.
func (err errorCollection) Error() string {
	var errStrings []string
	for _, suberr := range err {
		errStrings = append(errStrings, suberr.Error())
	}
	return strings.Join(errStrings, Delimiter)
}

// Count returns the number of distinct errors within an error object.  (A nil
// pointer contains 0 errors, a simple error contains 1 error, and an
// errorCollection contains 1 or more errors.)
func Count(err error) int {
	if err == nil {
		return 0
	} else if ec, ok := err.(errorCollection); ok {
		return len(ec)
	} else {
		return 1
	}
}

// ByIndex returns nth error within an error object.  If the index is greater
// than the number of errors available, nil is returned.  (A nil pointer
// contains 0 errors, a simple error contains 1 error, and an errorCollection
// contains 1 or more errors.)
func ByIndex(err error, index int) error {
	if err == nil {
		return nil
	} else if ec, ok := err.(errorCollection); ok {
		if index < len(ec) {
			return ec[index]
		}
	} else if index == 0 {
		return err
	}
	return nil
}
