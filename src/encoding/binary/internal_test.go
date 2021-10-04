// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package binary

import (
	"reflect"
	"sync"
)

// This file exposes unexported definitions to the external test package.
// Previously, internal tests accessed these definitions directly, but that
// causes an import cycle when because testing imports "runtime/debug", which
// imports "encoding/binary". So the internal test package must not import
// "testing".

var Internal = struct {
	StructSize *sync.Map
	DataSize   func(reflect.Value) int
	Overflow   error
}{
	&structSize,
	dataSize,
	overflow,
}
