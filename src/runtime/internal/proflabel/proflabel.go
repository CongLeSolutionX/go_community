// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package proflabel provides procedures for accessing goroutine profile labels.
// It exists to avoid creating a circular dependency between the runtime and
// runtime/pprof packages.
package proflabel

// A Label is a key, value string pair.
type Label struct {
	Key, Value string
}

// Labels is an immutable linked-list of labels. If multiple labels with the
// same key appear in a single List, or throughout the linked-list nodes,
// all but the first are ignored.
type Labels struct {
	List []Label
	Next *Labels
}

// runtime_getProfLabel is defined in runtime/proflabel.go.
func runtime_setProfLabel(*Labels)

// Set attaches the labels to the running goroutine.
func Set(labels *Labels) {
	runtime_setProfLabel(labels)
}

// runtime_getProfLabel is defined in runtime/proflabel.go.
func runtime_getProfLabel() *Labels

// Get returns the labels set on the running goroutine.
// Since Get could potentially be used with SetProfLabel to create
// a goroutine-local-storage mechanism, it should be used with care.
func Get() *Labels {
	return runtime_getProfLabel()
}
