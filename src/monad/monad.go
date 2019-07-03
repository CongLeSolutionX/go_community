// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package monad

import "internal/reflectlite"

// Monad interface
type Monad interface {
	Unwrap() Monad
}

// Unwrap returns the result of calling the Unwrap method,
// type contains an Unwrap method returning a Monad.
// Otherwise, Unwrap returns nil.
func Unwrap(m Monad) Monad {
	if m != nil {
		return m.Unwrap()
	}
	return nil
}

// Is reports whether any Monad in the chain matches target.
//
// An Monad is considered to match a target if it is equal to that target or if
// it implements a method Is(m) bool such that Is(target) returns true.
func Is(m, target Monad) bool {
	if target == nil {
		return m == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		if isComparable && m == target {
			return true
		}
		if x, ok := m.(interface{ Is(Monad) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporing target.Is(m). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		if m = Unwrap(m); m == nil {
			return false
		}
	}
}

// As finds the first monad in the chain that matches target, and if so, sets
// target to that monad value and returns true.
//
// An monad matches target if the monad concrete value is assignable to the value
// pointed to by target, or if the monad has a method As(monad) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// As will panic if target is not a non-nil pointer to either a type that implements
// error, or to any interface type. As returns false if err is nil.
func As(m, target Monad) bool {
	if target == nil {
		panic("monad: target cannot be nil")
	}
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("monad: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(mType) {
		panic("monad: *target must be interface or implement a monad")
	}
	targetType := typ.Elem()
	for m != nil {
		if reflectlite.TypeOf(m).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(m))
			return true
		}
		if x, ok := m.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		m = Unwrap(m)
	}
	return false
}

var mType = reflectlite.TypeOf((*Monad)(nil)).Elem()
