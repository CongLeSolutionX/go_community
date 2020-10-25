// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package user

import (
	"fmt"
)

func ExampleCurrent() {
	user, err := Current()
	if err != nil {
		panic(err)
	}
	fmt.Println("The user ID is:", user.Uid)
	// Gid is the primary group ID.
	fmt.Println("User Gid is:", user.Gid)
	// Name is the user's real or display name.
	fmt.Println("Current Username is:", user.Username)
	fmt.Println("Current User is:", user.Name)
	// HomeDir is the path to the user's home directory.
	fmt.Println("User Home dir is:", user.HomeDir)
}
