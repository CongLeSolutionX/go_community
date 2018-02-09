// Code generated from sort.go using genzfunc.go; DO NOT EDIT.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort

// Auto-generated variant of sort.go:insertionSort
func insertionSort_func(data lessSwap, a, b int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

// Auto-generated variant of sort.go:siftDown
func siftDown_func(data lessSwap, lo, hi, first int) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}
		if child+1 < hi && data.Less(first+child, first+child+1) {
			child++
		}
		if !data.Less(first+root, first+child) {
			return
		}
		data.Swap(first+root, first+child)
		root = child
	}
}

// Auto-generated variant of sort.go:heapSort
func heapSort_func(data lessSwap, a, b int) {
	first := a
	lo := 0
	hi := b - a
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown_func(data, i, hi, first)
	}
	for i := hi - 1; i >= 0; i-- {
		data.Swap(first, first+i)
		siftDown_func(data, lo, i, first)
	}
}

// Auto-generated variant of sort.go:medianOfThree
func medianOfThree_func(data lessSwap, m1, m0, m2 int) {
	if data.Less(m1, m0) {
		data.Swap(m1, m0)
	}
	if data.Less(m2, m1) {
		data.Swap(m2, m1)
		if data.Less(m1, m0) {
			data.Swap(m1, m0)
		}
	}
}

// Auto-generated variant of sort.go:swapRange
func swapRange_func(data lessSwap, a, b, n int) {
	for i := 0; i < n; i++ {
		data.Swap(a+i, b+i)
	}
}

// Auto-generated variant of sort.go:doPivot
func doPivot_func(data lessSwap, lo, hi int) (midlo, midhi int) {
	m := int(uint(lo+hi) >> 1)
	if hi-lo > 40 {
		s := (hi - lo) / 8
		medianOfThree_func(data, lo, lo+s, lo+2*s)
		medianOfThree_func(data, m, m-s, m+s)
		medianOfThree_func(data, hi-1, hi-1-s, hi-1-2*s)
	}
	medianOfThree_func(data, lo, m, hi-1)
	pivot := lo
	a, c := lo+1, hi-1
	for ; a < c && data.Less(a, pivot); a++ {
	}
	b := a
	for {
		for ; b < c && !data.Less(pivot, b); b++ {
		}
		for ; b < c && data.Less(pivot, c-1); c-- {
		}
		if b >= c {
			break
		}
		data.Swap(b, c-1)
		b++
		c--
	}
	protect := hi-c < 5
	if !protect && hi-c < (hi-lo)/4 {
		dups := 0
		if !data.Less(pivot, hi-1) {
			data.Swap(c, hi-1)
			c++
			dups++
		}
		if !data.Less(b-1, pivot) {
			b--
			dups++
		}
		if !data.Less(m, pivot) {
			data.Swap(m, b-1)
			b--
			dups++
		}
		protect = dups > 1
	}
	if protect {
		for {
			for ; a < b && !data.Less(b-1, pivot); b-- {
			}
			for ; a < b && data.Less(a, pivot); a++ {
			}
			if a >= b {
				break
			}
			data.Swap(a, b-1)
			a++
			b--
		}
	}
	data.Swap(pivot, b-1)
	return b - 1, c
}

// Auto-generated variant of sort.go:quickSort
func quickSort_func(data lessSwap, a, b, maxDepth int) {
	for b-a > 12 {
		if maxDepth == 0 {
			heapSort_func(data, a, b)
			return
		}
		maxDepth--
		mlo, mhi := doPivot_func(data, a, b)
		if mlo-a < b-mhi {
			quickSort_func(data, a, mlo, maxDepth)
			a = mhi
		} else {
			quickSort_func(data, mhi, b, maxDepth)
			b = mlo
		}
	}
	if b-a > 1 {
		for i := a + 6; i < b; i++ {
			if data.Less(i, i-6) {
				data.Swap(i, i-6)
			}
		}
		insertionSort_func(data, a, b)
	}
}

// Auto-generated variant of sort.go:searchLess
func searchLess_func(data lessSwap, a, b, x int) int {
	return a + Search(b-a, func(i int) bool { return data.Less(x, a+i) })
}

// Auto-generated variant of sort.go:max
func max_func(data lessSwap, a, b int) int {
	m := b
	for b--; a <= b; b-- {
		if data.Less(m, b) {
			m = b
		}
	}
	return m - a
}

// Auto-generated variant of sort.go:distinctElementCount
func distinctElementCount_func(data lessSwap, a, b, soughtSize int) (int, int) {
	dECnt := 1
	lastDiEl := b - 1
	for i := lastDiEl; a < i; i-- {
		if data.Less(i-1, i) {
			lastDiEl = i - 1
			dECnt++
			if dECnt == soughtSize {
				break
			}
		}
	}
	return dECnt, lastDiEl
}

// Auto-generated variant of sort.go:findBDSAndCountDistinctElementsNear
func findBDSAndCountDistinctElementsNear_func(data lessSwap, a, b, soughtSize int) (
	int, int, bool, int) {
	bds := -1
	backupBDS := -1
	dECnt := 1
	lastDiEl := b - 1
	pad := -1
	for i := lastDiEl; (bds == -1 || dECnt < soughtSize) &&
		b-(soughtSize<<1) < i; i-- {
		cnt := 0
		if data.Less(i-1, i) {
			cnt++
			if dECnt < soughtSize {
				dECnt++
				lastDiEl = i - 1
			}
			backupBDS = i
		}
		if data.Less(i-(soughtSize<<1)-1, i-(soughtSize<<1)) {
			cnt++
			if pad == -1 && a <= i-(soughtSize<<2) {
				pad = i - (soughtSize << 1)
			}
		}
		if cnt == 2 {
			bds = b - (soughtSize << 1)
		}
	}
	if (bds == -1 || dECnt < soughtSize) &&
		data.Less(b-(soughtSize<<1)-1, b-(soughtSize<<1)) {
		if dECnt < soughtSize {
			dECnt++
			lastDiEl = b - (soughtSize << 1) - 1
		}
		bds = b - (soughtSize << 1)
	}
	if bds == -1 && pad != -1 {
		bds = pad
	}
	return bds, backupBDS, dECnt == soughtSize, lastDiEl
}

// Auto-generated variant of sort.go:findBDSFarAndCountDistinctElements
func findBDSFarAndCountDistinctElements_func(data lessSwap, a, b, soughtSize int) (
	bds, backupBDS, dECnt, lastDiEl, dECntAfterBDS, lastDiElAfterBDS int) {
	bds = -1
	backupBDS = -1
	dECnt = 1
	lastDiEl = b - 1
	dECntAfterBDS = 1
	lastDiElAfterBDS = -1
	i := b - 1
	for ; b-(soughtSize<<2) < i; i-- {
		if data.Less(i-1, i) {
			dECnt++
			if dECnt <= soughtSize {
				lastDiEl = i - 1
			}
		}
	}
	for ; a < i; i-- {
		if data.Less(i-1, i) {
			dECnt++
			if dECnt <= soughtSize {
				lastDiEl = i - 1
			}
			if a <= i-(soughtSize<<1) {
				if bds == -1 {
					bds = i
				}
			} else {
				if backupBDS == -1 {
					backupBDS = i
				}
			}
			if i < bds-soughtSize<<1 {
				dECntAfterBDS++
				lastDiElAfterBDS = i - 1
				if dECntAfterBDS == soughtSize {
					break
				}
			}
		}
	}
	if soughtSize < dECnt {
		dECnt = soughtSize
	}
	return bds, backupBDS, dECnt, lastDiEl, dECntAfterBDS, lastDiElAfterBDS
}

// Auto-generated variant of sort.go:equalRange
func equalRange_func(data lessSwap, i, e int) int {
	for ; i+1 != e && !data.Less(i, i+1); i++ {
	}
	return i
}

// Auto-generated variant of sort.go:extractDist
func extractDist_func(data lessSwap, a, e, c int) {
	m := a + 1
	for e-a != c {
		t := equalRange_func(data, m, e)
		if t == m {
			m++
			continue
		}
		rotate_func(data, a, m, t)
		a += t - m
		m = t + 1
	}
}

// Auto-generated variant of sort.go:rotate
func rotate_func(data lessSwap, a, m, b int) {
	i := m - a
	j := b - m
	for i != j {
		if j < i {
			swapRange_func(data, m-i, m, j)
			i -= j
		} else {
			swapRange_func(data, m-i, m+j-i, i)
			j -= i
		}
	}
	swapRange_func(data, m-i, m, i)
}

// Auto-generated variant of sort.go:lessBlocks
func lessBlocks_func(data lessSwap, member, m, bS, e int) int {
	i := 0
	for ; e <= m-i*bS && data.Less(member, m-i*bS); i++ {
	}
	return i
}

// Auto-generated variant of sort.go:hLBufBigSmall
func hLBufBigSmall_func(data lessSwap, a, m, b, buf int) {
	for m != b && !data.Less(b-1, m-1) {
		b--
	}
	if m == b {
		return
	}
	swapRange_func(data, m, buf-(b-m), b-m)
	data.Swap(m-1, b-1)
	m--
	b--
	for a != m && m != b {
		if data.Less(buf-1, m-1) {
			data.Swap(m-1, b-1)
			m--
			b--
		} else {
			data.Swap(buf-1, b-1)
			buf--
			b--
		}
	}
	if m != b {
		swapRange_func(data, m, buf-(b-m), b-m)
	}
}

// Auto-generated variant of sort.go:hL
func hL_func(data lessSwap, a, m, b, dist int) {
	for m != b {
		step := (m - a) / (b - m)
		if step == 0 {
			step = 1
		}
		i := 1
		for ; a <= m-step*i && data.Less(b-1, m-step*i); i++ {
		}
		lowEdge := m - step*i + 1
		if lowEdge < a {
			lowEdge = a
		}
		j := searchLess_func(data, lowEdge, m-step*(i-1), b-1)
		if j != m {
			rotate_func(data, j, m, b)
			b -= m - j
			m = j
		}
		if dist == 1 || a == m || m == b-1 {
			break
		}
		dist--
		for !data.Less(b-2, b-1) {
			b--
			if m == b-1 {
				break
			}
		}
		b--
	}
}

// Auto-generated variant of sort.go:merge
func merge_func(data, aux lessSwap, a, m, b, locMergBufEnd, bS,
	bufEnd, bds0, bds1, maxDECnt int, mergingBuf bool) {
	f := b - (b-m)%bS
	e := a + (m-a)%bS
	F := f
	maxBl := f
	buf := bufEnd - (b-m)/bS
	bds0--
	bds1--
	for ; e != m && m != f; f -= bS {
		t := lessBlocks_func(data, maxBl-1, m-bS, bS, e)
		if d := t % (bufEnd - buf); d != 0 {
			rotate_func(aux, buf, bufEnd-d, bufEnd)
		}
		for ; 0 < t; t-- {
			swapRange_func(data, m-bS, f-bS, bS)
			if maxBl == f {
				maxBl = m
			}
			m -= bS
			f -= bS
			bds0--
			bds1--
		}
		bufEnd--
		if f != maxBl {
			aux.Swap(bufEnd, bufEnd-(f-maxBl)/bS)
			swapRange_func(data, maxBl-bS, f-bS, bS)
		}
		maxBl = m + bS*(1+max_func(aux, buf, bufEnd-1))
		aux.Swap(bds0, bds1)
		bds0--
		bds1--
	}
	for ; m != f; f -= bS {
		bufEnd--
		if f != maxBl {
			aux.Swap(bufEnd, bufEnd-(f-maxBl)/bS)
			swapRange_func(data, maxBl-bS, f-bS, bS)
		}
		maxBl = m + bS*(1+max_func(aux, buf, bufEnd-1))
		aux.Swap(bds0, bds1)
		bds0--
		bds1--
	}
	for ; e != m; m -= bS {
		bds0--
		bds1--
	}
	for ; e != F; e += bS {
		bds0++
		bds1++
		if aux.Less(bds0, bds1) {
			aux.Swap(bds0, bds1)
			if a != e {
				rotat := searchLess_func(data, a, e, e-1+bS)
				if rotat != e {
					rotate_func(data, rotat, e, e+bS)
				}
				if a != rotat {
					if mergingBuf {
						hLBufBigSmall_func(data, a, rotat, rotat+bS,
							locMergBufEnd)
					} else {
						hL_func(data, a, rotat, rotat+bS,
							maxDECnt)
					}
				}
				a = rotat + bS
			}
		}
	}
	if a != e {
		if mergingBuf {
			hLBufBigSmall_func(data, a, e, b,
				locMergBufEnd)
		} else {
			hL_func(data, a, e, b,
				maxDECnt)
		}
	}
	if mergingBuf {
		quickSort_func(data, locMergBufEnd-bS, locMergBufEnd, maxDepth(bS))
	}
}

// Auto-generated variant of sort.go:symMerge
func symMerge_func(data lessSwap, a, m, b int) {
	if m-a == 1 {
		i := m
		j := b
		for i < j {
			h := int(uint(i+j) >> 1)
			if data.Less(h, a) {
				i = h + 1
			} else {
				j = h
			}
		}
		for k := a; k < i-1; k++ {
			data.Swap(k, k+1)
		}
		return
	}
	if b-m == 1 {
		i := a
		j := m
		for i < j {
			h := int(uint(i+j) >> 1)
			if !data.Less(m, h) {
				i = h + 1
			} else {
				j = h
			}
		}
		for k := m; k > i; k-- {
			data.Swap(k, k-1)
		}
		return
	}
	mid := int(uint(a+b) >> 1)
	n := mid + m
	var start, r int
	if m > mid {
		start = n - b
		r = mid
	} else {
		start = a
		r = m
	}
	p := n - 1
	for start < r {
		c := int(uint(start+r) >> 1)
		if !data.Less(p-c, c) {
			start = c + 1
		} else {
			r = c
		}
	}
	end := n - start
	if start < m && m < end {
		rotate_func(data, start, m, end)
	}
	if a < start && start < mid {
		symMerge_func(data, a, start, mid)
	}
	if mid < end && end < b {
		symMerge_func(data, mid, end, b)
	}
}
