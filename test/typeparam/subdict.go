package main

import (
	"fmt"
)

type value[T any] struct {
	val T
}

func (v *value[T]) get(def T) T {
	var c value[int]
	if c.get(3) == 0 {
		return v.val
	} else {
		return def
	}
}


func main() {
	var s value[string]
	if got, want := s.get("ab"), ""; got != want {
		panic(fmt.Sprintf("get() == %d, want %d", got, want))
	}
}
