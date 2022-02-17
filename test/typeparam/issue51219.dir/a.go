// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

type JsonRaw []byte

type MyStruct struct {
	x *I[JsonRaw]
}

type IConstraint interface {
	JsonRaw | MyStruct
}

type I[T IConstraint] struct {
}

type Message struct {
	Interaction *Interaction[JsonRaw] `json:"interaction,omitempty"`
}

type ResolvedDataConstraint interface {
	User | Message
}

type Snowflake uint64

type ResolvedData[T ResolvedDataConstraint] map[Snowflake]T

type User struct {
}

type Resolved struct {
	Users ResolvedData[User] `json:"users,omitempty"`
}

type resolvedInteractionWithOptions struct {
	Resolved Resolved `json:"resolved,omitempty"`
}

type UserCommandInteractionData struct {
	resolvedInteractionWithOptions
}

type InteractionDataConstraint interface {
	JsonRaw | UserCommandInteractionData
}

type Interaction[DataT InteractionDataConstraint] struct {
}
