// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"a"
	"fmt"
	"reflect"
	"strconv"
)

type myint int

func (i myint) String() string {
	return strconv.Itoa(int(i))
}

func main() {
	x := []myint{myint(1), myint(2), myint(3)}

	got := a.Stringify(x)
	want := []string{"1", "2", "3"}
	if !reflect.DeepEqual(got, want) {
		panic(fmt.Sprintf("got %s, want %s", got, want))
	}

	//y := []int{1, 2, 3}
	//println(clos.Stringify(y))  // error: int does not satisfy Stringer
}
