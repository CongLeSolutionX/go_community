// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !cgo,!windows,!plan9,!android

package user

import "errors"

func init() {
	userImplemented = false
	groupImplemented = false
}

func current() (*User, error) {
	return nil, errors.New("user: Current requires cgo")
}

func lookupUser(username string) (*User, error) {
	return nil, errors.New("user: Lookup requires cgo")
}

func lookupUserId(uid string) (*User, error) {
	return nil, errors.New("user: LookupId requires cgo")
}

func currentGroup() (*Group, error) {
	return nil, errors.New("user: CurrentGroup requires cgo")
}

func lookupGroup(groupname string) (*Group, error) {
	return nil, errors.New("user: LookupGroup requires cgo")
}

func lookupGroupId(string) (*Group, error) {
	return nil, errors.New("user: LookupGroupId requires cgo")
}

func userInGroup(u *User, g *Group) (bool, error) {
	return false, errors.New("user: User.In requires cgo")
}
