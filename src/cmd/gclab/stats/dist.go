// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stats

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

type Dist[T ~int | ~int64 | ~float64] struct {
	vals []T
	sum  T
}

type DistCommon interface {
	String() string
	Plot(pngPath, x, y string)
}

func (d *Dist[T]) Add(val T) {
	d.vals = append(d.vals, val)
	d.sum += val
}

func (d Dist[T]) Quantiles(qs ...float64) []T {
	slices.Sort(d.vals)
	n := len(d.vals)
	res := make([]T, len(qs))
	for i, q := range qs {
		if q < 0 || q > 1 {
			panic("quantile out of range")
		}
		qi := int((float64(n) + 0.5) * q)
		qi = min(qi, len(d.vals)-1)
		res[i] = d.vals[qi]
	}
	return res
}

func (d Dist[T]) InvCDF(thresh T) float64 {
	slices.Sort(d.vals)
	for i, v := range d.vals {
		if v >= thresh {
			return float64(i) / float64(len(d.vals)-1)
		}
	}
	return 1
}

func (d Dist[T]) String() string {
	if len(d.vals) == 0 {
		return "  n=0\n"
	}
	qs := []float64{0, 0.01, 0.25, 0.5, 0.75, 0.99, 1}
	vs := d.Quantiles(qs...)
	var buf strings.Builder
	fmt.Fprintf(&buf, "  n=%d sum=%v mean=%f\n", len(d.vals), d.sum, float64(d.sum)/float64(len(d.vals)))
	for i, q := range qs {
		fmt.Fprintf(&buf, "  %4s %v\n", fmt.Sprintf("p%d", int(q*100)), vs[i])
	}
	return buf.String()
}

// Call cb for each dist in dists, which must be a pointer to a struct whose
// fields are dists.
func ForEachDist(dists interface{}, cb func(dist DistCommon, tag reflect.StructTag)) {
	rv := reflect.ValueOf(dists).Elem()
	for i := range rv.NumField() {
		f := rv.Field(i)
		tag := rv.Type().Field(i).Tag
		cb(f.Interface().(DistCommon), tag)
	}
}
