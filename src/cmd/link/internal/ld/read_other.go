// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !darwin,!linux

package ld

import "io/ioutil"

// newObjReader loads an *objReader from a source file.
func newObjReader(file string) (*objReader, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	input := &objReader{
		file: file,
		data: data,
	}
	return input, nil
}
