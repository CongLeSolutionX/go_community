// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package a

type Provider struct{}

func (p *Provider) ProvideType(a ...any) {
	for _, a := range a {
		_ = a
	}
}
func ProvideTypeG[T any](p *Provider) {
	p.ProvideType(new(T), new([]T), new([1]T), new([2]T), new([3]T), new([4]T), new([5]T), new([6]T), new([7]T), new([8]T))
}

func Init() {
	p := &Provider{}
	ProvideTypeG[int32](p)
}
