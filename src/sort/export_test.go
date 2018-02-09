// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort

func Heapsort(data Interface) {
	heapSort(data, 0, data.Len())
}

func ExtractDist(data Interface, c int) {
	extractDist(data, 0, data.Len(), c)
}

func HL(data Interface, m, b, z int) {
	hL(data, 0, m, b, z)
}

func HLBufBigSmall(data Interface, m, b, buf int) {
	hLBufBigSmall(data, 0, m, b, buf)
}
