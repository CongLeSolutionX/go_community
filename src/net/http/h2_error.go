// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !nethttpomithttp2
// +build !nethttpomithttp2

package http

import (
	"reflect"
)

func (e http2StreamError) As(target any) bool {
	return asHTTP2StreamError(e, target, "golang.org/x/net/http2", "StreamError")
}

func asHTTP2StreamError(err http2StreamError, target any, pkgPath, name string) bool {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Pointer {
		return false
	}
	val = val.Elem()
	if t := val.Type(); t.PkgPath() != pkgPath || t.Name() != name {
		return false
	}
	val.FieldByName("StreamID").SetUint(uint64(err.StreamID))
	code := val.FieldByName("Code")
	code.Set(reflect.ValueOf(err.Code).Convert(code.Type()))
	cause := val.FieldByName("Cause")
	if err.Cause != nil {
		cause.Set(reflect.ValueOf(err.Cause))
	} else if !cause.IsNil() {
		cause.Set(reflect.Zero(cause.Type()))
	}
	return true
}
