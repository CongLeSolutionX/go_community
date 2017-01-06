package proflabel

import _ "unsafe"

type Label struct {
	Key, Value string
}

// Labels is an immutable linked-list of labels.
type Labels struct {
	List []Label
	Next *Labels
}

//go:linkname runtime_setProfLabel runtime.runtime_setProfLabel
func runtime_setProfLabel(*Labels)

// Set attaches the labels to the running goroutine.
func Set(labels *Labels) {
	runtime_setProfLabel(labels)
}

//go:linkname runtime_getProfLabel runtime.runtime_getProfLabel
func runtime_getProfLabel() *Labels

// Get returns the labels set on the running goroutine.
// Since Get could potentially be used with SetProfLabel to create
// a goroutine-local-storage mechanism, it should be used with care.
func Get() *Labels {
	return runtime_getProfLabel()
}
