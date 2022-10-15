// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pkg

import "fmt"

// Package example
func Example() {
	fmt.Println("Package example output")
	// Output: Package example output
}

// Function example
func ExampleExportedFunc() {
	fmt.Println("Function example output")
	// Output: Function example output
}

// Type example
func ExampleExportedType() {
	fmt.Println("Type example output")
	// Output: Type example output
}

// Method on type example
func ExampleExportedType_ExportedMethod() {
	fmt.Println("Method on type example output")
	// Output: Method on type example output
}
