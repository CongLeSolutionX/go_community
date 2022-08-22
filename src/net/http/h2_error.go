// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !nethttpomithttp2
// +build !nethttpomithttp2

package http

import (
	"reflect"
)

var (
	typeUint32 = reflect.TypeOf(uint32(0))
	typeError  = reflect.TypeOf((*error)(nil)).Elem()
)

func (e http2StreamError) As(target any) bool {
	const (
		nameStreamID = "StreamID"
		nameCode     = "Code"
		nameCause    = "Cause"
	)
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		return false
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return false
	}
	typ := val.Type()
	if typ.NumField() != 3 {
		return false
	}
	if !hasNamedFieldOfType(typ, nameStreamID, typeUint32) {
		return false
	}
	if !hasNamedFieldOfKind(typ, nameCode, reflect.Uint32) {
		return false
	}
	if !hasNamedFieldOfType(typ, nameCause, typeError) {
		return false
	}
	val.FieldByName(nameStreamID).SetUint(uint64(e.StreamID))
	code := val.FieldByName(nameCode)
	code.Set(reflect.ValueOf(e.Code).Convert(code.Type()))
	cause := val.FieldByName(nameCause)
	if e.Cause != nil {
		cause.Set(reflect.ValueOf(e.Cause))
	} else if !cause.IsNil() {
		cause.Set(reflect.Zero(cause.Type()))
	}
	return true
}

func hasNamedFieldOfType(typ reflect.Type, name string, want reflect.Type) bool {
	field, ok := typ.FieldByName(name)
	if !ok {
		return false
	}
	return field.Type == want
}

func hasNamedFieldOfKind(typ reflect.Type, name string, want reflect.Kind) bool {
	field, ok := typ.FieldByName(name)
	if !ok {
		return false
	}
	return field.Type.Kind() == want
}
