// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iface

// A TypeAssertionError explains a failed type assertion.
type TypeAssertionError struct {
	InterfaceString string
	ConcreteString  string
	AssertedString  string
	MissingMethod   string // one method needed by Interface, missing from Concrete
}

func (*TypeAssertionError) RuntimeError() {}

func (e *TypeAssertionError) Error() string {
	inter := e.InterfaceString
	if inter == "" {
		inter = "interface"
	}
	if e.ConcreteString == "" {
		return "interface conversion: " + inter + " is nil, not " + e.AssertedString
	}
	if e.MissingMethod == "" {
		return "interface conversion: " + inter + " is " + e.ConcreteString +
			", not " + e.AssertedString
	}
	return "interface conversion: " + e.ConcreteString + " is not " + e.AssertedString +
		": missing method " + e.MissingMethod
}
