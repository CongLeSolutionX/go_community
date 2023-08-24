// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// reflect.StructOf and closures inside it must not be flagged with ReflectMethod.

package main

import "reflect"

func main() {
	t := reflect.StructOf([]reflect.StructField{
		{
			Name: "X",
			Type: reflect.TypeOf(int(0)),
		},
	})
	println(t.Name())
}
