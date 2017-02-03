// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
  if { // ERROR "missing condition in if statement"
  }

  if true; { // ERROR "missing condition in if statement"
  }

  if true
{ // ERROR "missing \{ after if clause"
}

  /*
  // Uncomment and enable this test case when #18915 is fixed
  if () { // "unexpected \).*expecting expression"
  }
  */
}
