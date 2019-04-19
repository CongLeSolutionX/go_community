// compile

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(...*int) {}

func g(p *int) {
	go f()
	go f(p)
	go f(p, p)

	go f(nil...)
	go f([]*int{}...)
	go f([]*int{p}...)
	go f([]*int{p, p}...)

	defer f()
	defer f(p)
	defer f(p, p)

	defer f(nil...)
	defer f([]*int{}...)
	defer f([]*int{p}...)
	defer f([]*int{p, p}...)

	for {
		go f()
		go f(p)
		go f(p, p)

		go f(nil...)
		go f([]*int{}...)
		go f([]*int{p}...)
		go f([]*int{p, p}...)

		defer f()
		defer f(p)
		defer f(p, p)

		defer f(nil...)
		defer f([]*int{}...)
		defer f([]*int{p}...)
		defer f([]*int{p, p}...)
	}
}
