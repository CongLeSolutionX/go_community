// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fmt_test

import "fmt"

type Person struct {
	Name string
	Age  int
}

func (p Person) GoString() string {
	return fmt.Sprintf("<< Name:%s Age:%d >>", p.Name, p.Age)
}

func (p Person) String() string {
	return fmt.Sprintf("%s %d", p.Name, p.Age)
}

func Example_stringers() {
	p := &Person{
		Name: "Gopher",
		Age:  5,
	}

	fmt.Printf("%s %#v", p, p)
}
