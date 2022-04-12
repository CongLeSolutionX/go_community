// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NOTE: If you change this file you must run "go generate"
// to update prototype.go. This is not done automatically
// to avoid depending on having a working compiler binary.

//go:build ignore
// +build ignore

package runtime

import "unsafe"

type stringStructDWARF struct {
	str *byte
	len int
}

type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}

type hmap struct {
	count     int
	flags     uint8
	B         uint8
	noverflow uint16
	hash0     uint32

	buckets    unsafe.Pointer
	oldbuckets unsafe.Pointer
	nevacuate  uintptr

	extra *mapextra
}

type bmap struct {
	tophash [8]uint8
}

type sudog struct {
	g           *g // todo: g is too complex
	next        *sudog
	prev        *sudog
	elem        unsafe.Pointer
	acquiretime int64
	releasetime int64
	ticket      uint32
	isSelect    bool
	success     bool
	parent      *sudog
	waitlink    *sudog
	waittail    *sudog
	c           *hchan
}

type waitq struct {
	first *sudog
	last  *sudog
}

type hchan struct {
	qcount   uint
	dataqsiz uint
	buf      unsafe.Pointer
	elemsize uint16
	closed   uint32
	elemtype *_type
	sendx    uint
	recvx    uint
	recvq    waitq
	sendq    waitq
	lock     mutex
}

// for static link, only need name.
// for dynamic link, not get a better way now.
type g struct{}

type mapextra struct{}

type mutex struct {
	// todo how to deal with lockRankStruct?
	key uintptr
}

type _type struct{}
