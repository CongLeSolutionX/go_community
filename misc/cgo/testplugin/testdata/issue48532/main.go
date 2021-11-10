// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"plugin"
	"testplugin/issue48532/base"
)

func main() {
	var a interface{}
	a = &base.ImplB{}
	switch x := a.(type) {
	case base.InterA:
		x.Method1()
	default:
		log.Fatalln("assert x to A failed")
	}

	p, err := plugin.Open("issue48532.so")
	if err != nil {
		log.Fatalln(err)
	}
	sym, err := p.Lookup("TestFunc")
	if err != nil {
		log.Fatalln(err)
	}

	testFunc := sym.(func(base.InterA))
	switch x := a.(type) {
	case base.InterA:
		testFunc(x)
	default:
		log.Fatalln("assert x to A failed")
	}
}
