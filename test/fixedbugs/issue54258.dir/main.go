// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "./b"

type I0 interface {
	M0() (w struct{ f string })
}

var _ I0 = b.S{} // ERROR "cannot use b.S.*as I0 value.*wrong type for method M0.*\n.* package test/b .*\n.* package main "

type I1 interface {
	M1() (x struct{ string })
}

var _ I1 = b.S{} // ERROR "cannot use b.S.*as I1 value.*wrong type for method M1.*\n.* package test/b .*\n.* package main "

type I2 interface {
	M2() (y struct{ f struct{ f string } })
}

var _ I2 = b.S{} // ERROR "cannot use b.S.*as I2 value.*wrong type for method M2.*\n.* package test/b .*\n.* package main "

type I3 interface {
	M3() (z struct{ F struct{ f string } })
}

var _ I3 = b.S{} // ERROR "cannot use b.S.*as I3 value.*wrong type for method M3.*\n.* package test/b .*\n.* package main "

func main() {}

// Original error messages
// ./main.go:13:12: cannot use b.S{} (value of type b.S) as I0 value in variable declaration: b.S does not implement I0 (wrong type for method M0)
// 		have M0() (struct{f string /* package test/b */ })
// 		want M0() (struct{f string /* package main */ })
// ./main.go:19:12: cannot use b.S{} (value of type b.S) as I1 value in variable declaration: b.S does not implement I1 (wrong type for method M1)
// 		have M1() (struct{string /* package test/b */ })
// 		want M1() (struct{string /* package main */ })
// ./main.go:25:12: cannot use b.S{} (value of type b.S) as I2 value in variable declaration: b.S does not implement I2 (wrong type for method M2)
// 		have M2() (struct{f struct{f string} /* package test/b */ })
// 		want M2() (struct{f struct{f string} /* package main */ })
// ./main.go:31:12: cannot use b.S{} (value of type b.S) as I3 value in variable declaration: b.S does not implement I3 (wrong type for method M3)
// 		have M3() (struct{F struct{f string /* package test/b */ }})
// 		want M3() (struct{F struct{f string /* package main */ }})
