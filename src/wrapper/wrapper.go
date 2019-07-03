// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wrapper

import "internal/reflectlite"

// Wrapper interface
type Wrapper interface {
	Unwrap() Wrapper
}

// Unwrap returns the result of calling the Unwrap method,
// type contains an Unwrap method returning a Wrapper.
// Otherwise, Unwrap returns nil.
func Unwrap(w Wrapper) Wrapper {
	if w != nil {
		return w.Unwrap()
	}
	return nil
}

// Is reports whether any Wrapper in the chain matches target.
//
// An Wrapper is considered to match a target if it is equal to that target or if
// it implements a method Is(m) bool such that Is(target) returns true.
func Is(w, target Wrapper) bool {
	if target == nil {
		return w == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		if isComparable && w == target {
			return true
		}
		if x, ok := w.(interface{ Is(Wrapper) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporing target.Is(m). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		if w = Unwrap(w); w == nil {
			return false
		}
	}
}

// As finds the first wrapper in the chain that matches target, and if so, sets
// target to that wrapper value and returns true.
//
// An wrapper matches target if the wrapper concrete value is assignable to the value
// pointed to by target, or if the wrapper has a method As(wrapper) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// As will panic if target is not a non-nil pointer to either a type that implements
// error, or to any interface type. As returns false if err is nil.
func As(w, target Wrapper) bool {
	if target == nil {
		panic("wrapper: target cannot be nil")
	}
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("wrapper: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(wType) {
		panic("wrapper: *target must be interface or implement a wrapper")
	}
	targetType := typ.Elem()
	for w != nil {
		if reflectlite.TypeOf(w).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(w))
			return true
		}
		if x, ok := w.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		w = Unwrap(w)
	}
	return false
}

var wType = reflectlite.TypeOf((*Wrapper)(nil)).Elem()
