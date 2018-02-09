// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run genzfunc.go

// Package sort provides primitives for sorting slices and user-defined
// collections.
package sort

// A type, typically a collection, that satisfies sort.Interface can be
// sorted by the routines in this package. The methods require that the
// elements of the collection be enumerated by an integer index.
type Interface interface {
	// Len is the number of elements in the collection.
	Len() int
	// Less reports whether the element with
	// index i should sort before the element with index j.
	Less(i, j int) bool
	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
}

// Insertion sort
func insertionSort(data Interface, a, b int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && data.Less(j, j-1); j-- {
			data.Swap(j, j-1)
		}
	}
}

// siftDown implements the heap property on data[lo, hi).
// first is an offset into the array where the root of the heap lies.
func siftDown(data Interface, lo, hi, first int) {
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

func heapSort(data Interface, a, b int) {
	first := a
	lo := 0
	hi := b - a

	// Build heap with greatest element at top.
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown(data, i, hi, first)
	}

	// Pop elements, largest first, into end of data.
	for i := hi - 1; i >= 0; i-- {
		data.Swap(first, first+i)
		siftDown(data, lo, i, first)
	}
}

// Quicksort, loosely following Bentley and McIlroy,
// ``Engineering a Sort Function,'' SP&E November 1993.

// medianOfThree moves the median of the three values data[m0], data[m1], data[m2] into data[m1].
func medianOfThree(data Interface, m1, m0, m2 int) {
	// sort 3 elements
	if data.Less(m1, m0) {
		data.Swap(m1, m0)
	}
	// data[m0] <= data[m1]
	if data.Less(m2, m1) {
		data.Swap(m2, m1)
		// data[m0] <= data[m2] && data[m1] < data[m2]
		if data.Less(m1, m0) {
			data.Swap(m1, m0)
		}
	}
	// now data[m0] <= data[m1] <= data[m2]
}

func swapRange(data Interface, a, b, n int) {
	for i := 0; i < n; i++ {
		data.Swap(a+i, b+i)
	}
}

func doPivot(data Interface, lo, hi int) (midlo, midhi int) {
	m := int(uint(lo+hi) >> 1) // Written like this to avoid integer overflow.
	if hi-lo > 40 {
		// Tukey's ``Ninther,'' median of three medians of three.
		s := (hi - lo) / 8
		medianOfThree(data, lo, lo+s, lo+2*s)
		medianOfThree(data, m, m-s, m+s)
		medianOfThree(data, hi-1, hi-1-s, hi-1-2*s)
	}
	medianOfThree(data, lo, m, hi-1)

	// Invariants are:
	//	data[lo] = pivot (set up by ChoosePivot)
	//	data[lo < i < a] < pivot
	//	data[a <= i < b] <= pivot
	//	data[b <= i < c] unexamined
	//	data[c <= i < hi-1] > pivot
	//	data[hi-1] >= pivot
	pivot := lo
	a, c := lo+1, hi-1

	for ; a < c && data.Less(a, pivot); a++ {
	}
	b := a
	for {
		for ; b < c && !data.Less(pivot, b); b++ { // data[b] <= pivot
		}
		for ; b < c && data.Less(pivot, c-1); c-- { // data[c-1] > pivot
		}
		if b >= c {
			break
		}
		// data[b] > pivot; data[c-1] <= pivot
		data.Swap(b, c-1)
		b++
		c--
	}
	// If hi-c<3 then there are duplicates (by property of median of nine).
	// Let be a bit more conservative, and set border to 5.
	protect := hi-c < 5
	if !protect && hi-c < (hi-lo)/4 {
		// Lets test some points for equality to pivot
		dups := 0
		if !data.Less(pivot, hi-1) { // data[hi-1] = pivot
			data.Swap(c, hi-1)
			c++
			dups++
		}
		if !data.Less(b-1, pivot) { // data[b-1] = pivot
			b--
			dups++
		}
		// m-lo = (hi-lo)/2 > 6
		// b-lo > (hi-lo)*3/4-1 > 8
		// ==> m < b ==> data[m] <= pivot
		if !data.Less(m, pivot) { // data[m] = pivot
			data.Swap(m, b-1)
			b--
			dups++
		}
		// if at least 2 points are equal to pivot, assume skewed distribution
		protect = dups > 1
	}
	if protect {
		// Protect against a lot of duplicates
		// Add invariant:
		//	data[a <= i < b] unexamined
		//	data[b <= i < c] = pivot
		for {
			for ; a < b && !data.Less(b-1, pivot); b-- { // data[b] == pivot
			}
			for ; a < b && data.Less(a, pivot); a++ { // data[a] < pivot
			}
			if a >= b {
				break
			}
			// data[a] == pivot; data[b-1] < pivot
			data.Swap(a, b-1)
			a++
			b--
		}
	}
	// Swap pivot into middle
	data.Swap(pivot, b-1)
	return b - 1, c
}

func quickSort(data Interface, a, b, maxDepth int) {
	for b-a > 12 { // Use ShellSort for slices <= 12 elements
		if maxDepth == 0 {
			heapSort(data, a, b)
			return
		}
		maxDepth--
		mlo, mhi := doPivot(data, a, b)
		// Avoiding recursion on the larger subproblem guarantees
		// a stack depth of at most lg(b-a).
		if mlo-a < b-mhi {
			quickSort(data, a, mlo, maxDepth)
			a = mhi // i.e., quickSort(data, mhi, b)
		} else {
			quickSort(data, mhi, b, maxDepth)
			b = mlo // i.e., quickSort(data, a, mlo)
		}
	}
	if b-a > 1 {
		// Do ShellSort pass with gap 6
		// It could be written in this simplified form cause b-a <= 12
		for i := a + 6; i < b; i++ {
			if data.Less(i, i-6) {
				data.Swap(i, i-6)
			}
		}
		insertionSort(data, a, b)
	}
}

// Sort sorts data.
// It makes one call to data.Len to determine n, and O(n*log(n)) calls to
// data.Less and data.Swap. The sort is not guaranteed to be stable.
func Sort(data Interface) {
	n := data.Len()
	quickSort(data, 0, n, maxDepth(n))
}

// maxDepth returns a threshold at which quicksort should switch
// to heapsort. It returns 2*ceil(lg(n+1)).
func maxDepth(n int) int {
	var depth int
	for i := n; i > 0; i >>= 1 {
		depth++
	}
	return depth * 2
}

// lessSwap is a pair of Less and Swap function for use with the
// auto-generated func-optimized variant of sort.go in
// zfuncversion.go.
type lessSwap struct {
	Less func(i, j int) bool
	Swap func(i, j int)
}

type reverse struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	Interface
}

// Less returns the opposite of the embedded implementation's Less method.
func (r reverse) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}

// Reverse returns the reverse order for data.
func Reverse(data Interface) Interface {
	return &reverse{data}
}

// IsSorted reports whether data is sorted.
func IsSorted(data Interface) bool {
	n := data.Len()
	for i := n - 1; i > 0; i-- {
		if data.Less(i, i-1) {
			return false
		}
	}
	return true
}

// Convenience types for common cases

// IntSlice attaches the methods of Interface to []int, sorting in increasing order.
type IntSlice []int

func (p IntSlice) Len() int           { return len(p) }
func (p IntSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p IntSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p IntSlice) Sort() { Sort(p) }

// Float64Slice attaches the methods of Interface to []float64, sorting in increasing order
// (not-a-number values are treated as less than other values).
type Float64Slice []float64

func (p Float64Slice) Len() int           { return len(p) }
func (p Float64Slice) Less(i, j int) bool { return p[i] < p[j] || isNaN(p[i]) && !isNaN(p[j]) }
func (p Float64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// isNaN is a copy of math.IsNaN to avoid a dependency on the math package.
func isNaN(f float64) bool {
	return f != f
}

// Sort is a convenience method.
func (p Float64Slice) Sort() { Sort(p) }

// StringSlice attaches the methods of Interface to []string, sorting in increasing order.
type StringSlice []string

func (p StringSlice) Len() int           { return len(p) }
func (p StringSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p StringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// Sort is a convenience method.
func (p StringSlice) Sort() { Sort(p) }

// Convenience wrappers for common cases

// Ints sorts a slice of ints in increasing order.
func Ints(a []int) { Sort(IntSlice(a)) }

// Float64s sorts a slice of float64s in increasing order
// (not-a-number values are treated as less than other values).
func Float64s(a []float64) { Sort(Float64Slice(a)) }

// Strings sorts a slice of strings in increasing order.
func Strings(a []string) { Sort(StringSlice(a)) }

// IntsAreSorted tests whether a slice of ints is sorted in increasing order.
func IntsAreSorted(a []int) bool { return IsSorted(IntSlice(a)) }

// Float64sAreSorted tests whether a slice of float64s is sorted in increasing order
// (not-a-number values are treated as less than other values).
func Float64sAreSorted(a []float64) bool { return IsSorted(Float64Slice(a)) }

// StringsAreSorted tests whether a slice of strings is sorted in increasing order.
func StringsAreSorted(a []string) bool { return IsSorted(StringSlice(a)) }

// In the sorted array data[a:b], finds the least member from which data[x] is
// lesser.
func searchLess(data Interface, a, b, x int) int {
	return a + Search(b-a, func(i int) bool { return data.Less(x, a+i) })
}

// Finds the index (relative to a) of a maximum of the array data[a:b].
func max(data Interface, a, b int) int {
	m := b
	for b--; a <= b; b-- {
		if data.Less(m, b) {
			m = b
		}
	}
	return m - a
}

// Integer parts of square roots of successive powers of 2, rounded.
//
// IEEE 754 64-bit floating point can exactly represent integers up to only
// 1<<53, hence a proper integer algorithm for isqrt is needed.
//
// (2^n)^½ = (2^½)^n
// Branch with regards to the parity of n:
// Half the powers of 2^½ are just integer powers of two, and the others
// are (2^½)*(2^n), which can be constructed from the digits of 2^½.
func isqrter() func() int {
	const (
		// 63 binary digits of 2^½, ⌊2^62.5⌋ = ⌊2^½ << 62⌋.
		// The greatest power of 2^½ representable in int64.
		sqrt2 = 6521908912666391106

		// The greatest even power of 2^½ representable in int64.
		even = 1 << 62

		// The initial blockSize's square root is 4 = 2^2
		initialPower = 2

		z = (62-initialPower)<<1 + 1

		// 32 if uint is 32 bits wide, 0 if uint is 64 bits wide.
		thirtyTwo = 32 - (((^uint(0)) >> 63) << 5)
	)

	// If uint is 32-bit, fit the constants into the 32 bits.
	bits := [2]int{even >> thirtyTwo, sqrt2 >> thirtyTwo}
	pow := uint(z - (thirtyTwo << 1))

	return func() int {
		shift := pow >> 1
		pow--
		// For floor(sqrt()) we would just return bits[pow&1] >> shift
		// This variant instead gives rounded(sqrt()).
		return (bits[pow&1] + (1 << (shift - 1))) >> shift
	}
}

// Counts distinct elements in the sorted block data[a:b], stopping
// if the count reaches max. It returns the count and the index of the last
// distinct element counted.
//
// Eg. 12223345556 has 6 distinct elements.
func distinctElementCount(data Interface, a, b, soughtSize int) (int, int) {
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

// Counts distinct elements for an internal buffer for merges and movement
// imitation while trying to find block distribution storage candidates. The
// distinct elements are counted in the high-index half of BDS. All of the
// aforementioned is done only in data[a:b], which must be sorted. Not more
// than soughtSize distinct elements will be counted for the buffer, and each
// half of the BDS is supposed to be soughtSize<<1 array members long. b-a must
// be greater than or equal to soughtSize<<2 (the minimal size of an entire
// BDS).
//
// bds and backupBDS are either -1 or the high index edge of the low index half
// of a BDS. backupBDS refers to an undersized BDS, see earlier in the comment.
//
// bufferFound is true if a soughtSize-d buffer of distinct element can be
// extracted from data[a:b], and false otherwise. If it is true, lastDiEl is
// the index of the final distinct element found, to be the last element in the
// extracted buffer.
func findBDSAndCountDistinctElementsNear(data Interface, a, b, soughtSize int) (
	bds, backupBDS int, bufferFound bool, lastDiEl int) {

	// If not -1, represents the high-index edge of the low-index half of
	// the candidate BDS (in which case the high-index edge of the
	// high-index half is b).
	bds = -1

	// To prevent Less calls later on, record the distinct element for
	// a smaller BDS for backup.
	backupBDS = -1

	// Distinct element count. The 1 we initialize to is for data[b-1].
	dECnt := 1

	// Last distinct element for buffer, so it would not have to be
	// searched for, doing the same Less calls again.
	lastDiEl = b - 1

	// First see if there can be BDS without padding, or with less than
	// soughtSize array elements of padding. Also count distinct elements.
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
			// BDS found.
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

// Counts distinct elements for an internal buffer for merges and movement
// imitation while trying to find block distribution storage candidates.
// Candidates for BDS edges are only searched for in data[a:b-(soughtSize<<2)].
// All of the aforementioned is done only in data[a:b], which must be sorted.
// Not more than soughtSize distinct elements will be counted for the buffer,
// and each half of the BDS is supposed to be soughtSize<<1 array members long.
// b-a must be greater than or equal to soughtSize<<2.
//
// bds and backupBDS are either -1 or the high index edge of the low index half
// of a BDS. backupBDS refers to an undersized BDS, see earlier in the comment.
//
// dECnt is the count of distinct elements in data[a:b], or soughtSize if there
// are more than soughtSize distinct elements. lastDiEl is the index of the
// last distinct member accounted for in dECnt. dECntAfterBDS and
// lastDiElAfterBDS are the same, but they are only searched for if a full sized
// BDS is found (as returned by bds) after the space that would be taken up by
// that BDS.
func findBDSFarAndCountDistinctElements(data Interface, a, b, soughtSize int) (
	bds, backupBDS, dECnt, lastDiEl, dECntAfterBDS, lastDiElAfterBDS int) {
	bds = -1
	backupBDS = -1

	dECnt = 1
	lastDiEl = b - 1

	// TODO: could be replaced with a bool?
	dECntAfterBDS = 1

	lastDiElAfterBDS = -1

	i := b - 1

	// TODO: check for undersized/backup BDS here, too.

	// Count distinct elements.
	for ; b-(soughtSize<<2) < i; i-- {
		if data.Less(i-1, i) {
			dECnt++
			if dECnt <= soughtSize {
				lastDiEl = i - 1
			}
		}
	}

	// Count distinct elements and search for distinct element appropriate for
	// assigning BDS (BDS padding).
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

// data[i,j] is a sequence of identical elements with the maximal extent, with
// the added condition j<=e. equalRange returns j-1.
func equalRange(data Interface, i, e int) int {
	for ; i+1 != e && !data.Less(i, i+1); i++ {
	}
	return i
}

// Pulls out mutually distinct elements from the sorted sequence data[a:e] to
// data[e-c:e], where c is the number of distinct elements in data[a:e]
// (a <= e-c < e).
//
// Eg. 12223345556 is stably transformed to 22355123456 in 3 rotations.
func extractDist(data Interface, a, e, c int) {
	// data[a:m] is the sequence of distinct elements that we are
	// rotating through data[a:e] and progressively expanding.
	m := a + 1
	// c is used to prevent unneccessary data.Less calls when there is a
	// contiguous block of distinct elements in the end.
	for e-a != c {
		t := equalRange(data, m, e)
		if t == m {
			m++
			continue
		}
		rotate(data, a, m, t)
		a += t - m
		m = t + 1
	}
}

// Rotate two consecutives blocks u = data[a:m] and v = data[m:b] in data:
// Data of the form 'x u v y' is changed to 'x v u y'.
// Rotate performs at most b-a many calls to data.Swap.
// Rotate assumes non-degenerate arguments: a < m && m < b.
func rotate(data Interface, a, m, b int) {
	i := m - a
	j := b - m

	for i != j {
		if j < i {
			swapRange(data, m-i, m, j)
			i -= j
		} else {
			swapRange(data, m-i, m+j-i, i)
			j -= i
		}
	}
	// i == j
	swapRange(data, m-i, m, i)
}

// Returns the count of how many blocks after m are < member. Searches up to e.
func lessBlocks(data Interface, member, m, bS, e int) int {
	i := 0
	for ; e <= m-i*bS && data.Less(member, m-i*bS); i++ {
	}
	return i
}

// Stably merges data[a:m] with data[m:b] using the buffer data[buf-(b-m):buf].
// The buffer is left in a disordered state after the merge. Assumes a < m.
//
// Used for direct merging of merge blocks in the "merge" function when a buffer
// is available.
func simpleMergeBufBigSmall(data Interface, a, m, b, buf int) {
	// Exhaust largest data[m:b] members before swapping to the buffer.
	for m != b && !data.Less(b-1, m-1) {
		b--
	}
	if m == b {
		return
	}
	swapRange(data, m, buf-(b-m), b-m)
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

	// Swap whatever is left in the buffer to its final position.
	if m != b {
		swapRange(data, m, buf-(b-m), b-m)
	}
}

// Stably merges data[a:m] with data[m:b], which has maximally dist distinct
// elements. Assumes a < m.
//
// Used for direct merging of merge blocks in the "merge" function.
func simpleMerge(data Interface, a, m, b, dist int) {
	// This is the Hwang and Lin merging algorithm.
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
		j := searchLess(data, lowEdge, m-step*(i-1), b-1)

		if j != m {
			rotate(data, j, m, b)
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

// Stably merges the two sorted subsequences data[a:m] and data[m:b].
//
// The aux Interface is for the MI buffer and BDS.
//
// The input is rearranged in blocks using the movement imitation buffer,
// aux[bufEnd-bS:bufEnd]. The appropriate blocks are then merged using the
// simpleMerge and simpleMergeBufBigSmall functions. The latter is used if
// mergingBuf is true, and uses the local merge buffer
// data[locMergBufEnd-bS:locMergBufEnd]. bds0, bds1 are the high-index edges
// of the BDS halves in aux, each 2*bS long. maxDECnt is the maximum possible
// number of distinct elements in data[m:b].
func merge(data, aux Interface, a, m, b, locMergBufEnd, bS,
	bufEnd, bds0, bds1, maxDECnt int, mergingBuf bool) {
	// We divide data[a:b] into equally sized blocks so they could be
	// exchanged with swapRange instead of rotate. The partition starts
	// from m, leaving 2 possible smaller blocks at each end of data[a:b].
	// They will be merged after the equally sized blocks in the middle.

	// A block index k represents the block data[k-bS:k]. (Applies to m, f,
	// F, maxBl, prevBl ...)

	// Edges of the fullsized-block area.
	//f := b - (b-m)%bufSz
	f := b - (b-m)%bS
	e := a + (m-a)%bS
	F := f

	// The maximal block of data[m:f].
	maxBl := f

	// The low-index edge of the movement imitation buffer.
	buf := bufEnd - (b-m)/bS

	bds0--
	bds1--

	// Block rearrangement
	for ; e != m && m != f; f -= bS {
		// How many blocks?
		t := lessBlocks(data, maxBl-1, m-bS, bS, e)

		if d := t % (bufEnd - buf); d != 0 {
			// Record the changes we will now make to the
			// ordering of data[m:f].
			rotate(aux, buf, bufEnd-d, bufEnd)
		}

		// Roll data[m:f] through data[e:m].
		for ; 0 < t; t-- {
			swapRange(data, m-bS, f-bS, bS)
			if maxBl == f {
				maxBl = m
			}
			m -= bS
			f -= bS
			bds0--
			bds1--
		}

		// Place the block to its final position before merging, while
		// tracking the ordering in the movement imitation
		// buffer. Find the new maxBl.
		bufEnd--
		if f != maxBl {
			aux.Swap(bufEnd, bufEnd-(f-maxBl)/bS)
			swapRange(data, maxBl-bS, f-bS, bS)
		}
		maxBl = m + bS*(1+max(aux, buf, bufEnd-1))

		aux.Swap(bds0, bds1)
		bds0--
		bds1--
	}

	for ; m != f; f -= bS {
		// Place the block to its final position before merging, while
		// tracking the ordering in the movement imitation
		// buffer. Find the new maxBl.
		bufEnd--
		if f != maxBl {
			aux.Swap(bufEnd, bufEnd-(f-maxBl)/bS)
			swapRange(data, maxBl-bS, f-bS, bS)
		}
		maxBl = m + bS*(1+max(aux, buf, bufEnd-1))

		aux.Swap(bds0, bds1)
		bds0--
		bds1--
	}

	for ; e != m; m -= bS {
		bds0--
		bds1--
	}

	// Local merges
	for ; e != F; e += bS {
		bds0++
		bds1++
		if aux.Less(bds0, bds1) {
			aux.Swap(bds0, bds1)
			if a != e {
				rotat := searchLess(data, a, e, e-1+bS)
				if rotat != e {
					rotate(data, rotat, e, e+bS)
				}

				// Local merge
				if a != rotat {
					if mergingBuf {
						simpleMergeBufBigSmall(data, a, rotat, rotat+bS,
							locMergBufEnd)
					} else {
						simpleMerge(data, a, rotat, rotat+bS,
							maxDECnt)
					}
				}
				a = rotat + bS
			}
		}
	}

	if a != e {
		// Locally merge the undersized block
		if mergingBuf {
			simpleMergeBufBigSmall(data, a, e, b,
				locMergBufEnd)
		} else {
			simpleMerge(data, a, e, b,
				maxDECnt)
		}
	}

	if mergingBuf {
		// Sort the "buffer" we just used for local merges. An unstable
		// sorting routine can be used here because the buffer consists
		// of mutually distinct elements.
		quickSort(data, locMergBufEnd-bS, locMergBufEnd, maxDepth(bS))
	}
}

// The size of the movement imitation buffer in compTimeMIBuffer.
const compTimeMIBufSize = 1 << 8

type (
	// The number of values representable in compTimeMIBufMemb must be at
	// least compTimeMIBufSize. Needs to be an unsigned integer type.
	compTimeMIBufMemb uint8
	// Each half of BDS is 2*compTimeMIBufSize, 1+2+2=5. This array contains
	// a BDS and a MI buffer. Used in stable() as an optimization.
	compTimeMIBuffer [5 * compTimeMIBufSize]compTimeMIBufMemb
)

func (b *compTimeMIBuffer) Less(i, j int) bool {
	return b[i] < b[j]
}
func (b *compTimeMIBuffer) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b *compTimeMIBuffer) Len() int {
	// Should not come here.
	panic(nil)
}

// SymMerge merges the two sorted subsequences data[a:m] and data[m:b] using
// the SymMerge algorithm from Pok-Son Kim and Arne Kutzner, "Stable Minimum
// Storage Merging by Symmetric Comparisons", in Susanne Albers and Tomasz
// Radzik, editors, Algorithms - ESA 2004, volume 3221 of Lecture Notes in
// Computer Science, pages 714-723. Springer, 2004.
//
// Let M = m-a and N = b-n. Wolog M < N.
// The recursion depth is bound by ceil(log(N+M)).
// The algorithm needs O(M*log(N/M + 1)) calls to data.Less.
// The algorithm needs O((M+N)*log(M)) calls to data.Swap.
//
// The paper gives O((M+N)*log(M)) as the number of assignments assuming a
// rotation algorithm which uses O(M+N+gcd(M+N)) assignments. The argumentation
// in the paper carries through for Swap operations, especially as the block
// swapping rotate uses only O(M+N) Swaps.
//
// symMerge assumes non-degenerate arguments: a < m && m < b.
// Having the caller check this condition eliminates many leaf recursion calls,
// which improves performance.
func symMerge(data Interface, a, m, b int) {
	// Avoid unnecessary recursions of symMerge
	// by direct insertion of data[a] into data[m:b]
	// if data[a:m] only contains one element.
	if m-a == 1 {
		// Use binary search to find the lowest index i
		// such that data[i] >= data[a] for m <= i < b.
		// Exit the search loop with i == b in case no such index exists.
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
		// Swap values until data[a] reaches the position before i.
		for k := a; k < i-1; k++ {
			data.Swap(k, k+1)
		}
		return
	}

	// Avoid unnecessary recursions of symMerge
	// by direct insertion of data[m] into data[a:m]
	// if data[m:b] only contains one element.
	if b-m == 1 {
		// Use binary search to find the lowest index i
		// such that data[i] > data[m] for a <= i < m.
		// Exit the search loop with i == m in case no such index exists.
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
		// Swap values until data[m] reaches the position i.
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
		rotate(data, start, m, end)
	}
	if a < start && start < mid {
		symMerge(data, a, start, mid)
	}
	if mid < end && end < b {
		symMerge(data, mid, end, b)
	}
}

func stable(data Interface, n int) {
	// A merge sort with the merge algorithm based on Kim & Kutzner 2008
	// (Ratio based stable in-place merging).
	//
	// Differences from the paper:
	//
	// We use rounded integer square root (see isqrter) where the authors
	// use the floor of the square root.
	//
	// Only searching for buffer and block distribution storage once in
	// merge sort level, instead of as part of every merge.
	//
	// We have a constant size compile time assignable buffer whose portions
	// are to be used in the block rearrangement part of merging when
	// possible as a movement imitation buffer and a block distribution
	// storage.

	// MI buffers and buffers for local merges must consist of distinct
	// elements. The MI buffer must be sorted before being used in block
	// rearrangement and will be sorted after the block rearrangement.
	//
	// BDS consists of two equally sized arrays (each usually double the MI
	// buffer size), and a padding between the halves, if necessary. If
	// bds0 and bds1 are indices of high index edges of BDS halves, the size
	// of a bds half is bdsSize, and the halves reside in the Interface
	// data; for each i from 1 to bdsSize, inclusive;
	// data.Less(bds1-i, bds0-i) must be true before the BDS is used in
	// block rearrangement and will be true after the local merges following
	// the block rearrangement.

	// As an optimization, reverse ranges of more than two elements in reverse order.
	a := 0
	for i := 0; i < n; i++ {
		// Find reversed ranges.
		for ; i+1 < n && data.Less(i+1, i); i++ {
		}

		if 1 < i-a {
			half := (i + 1 - a) >> 1
			for j := 0; j < half; j++ {
				data.Swap(a+j, i-j)
			}
		}
		a = i + 1
	}

	// We will need square roots of blockSize values, and calculating
	// square roots of successive powers of 2 is easy. The initialization
	// value is coupled with isqrter.
	blockSize := 1 << 4

	// Insertion sort blocks of input.
	a = 0
	for a+blockSize < n {
		insertionSort(data, a, a+blockSize)
		a += blockSize
	}
	insertionSort(data, a, n)

	if n <= blockSize {
		return
	}

	// Use symMerge for small arrays.
	if n < 2000 {
		for ; blockSize < n; blockSize <<= 1 {
			for a = blockSize << 1; a <= n; a += blockSize << 1 {
				symMerge(data, a-blockSize<<1, a-blockSize, a)
			}
			if a-blockSize < n {
				symMerge(data, a-blockSize<<1, a-blockSize, n)
			}
		}
		return
	}

	// As an optimization, use a MI buffer and BDS assignable at
	// compilation time, instead of extracting from input data.
	var compTimeMovImBuf compTimeMIBuffer
	// Initialize the compile-time assigned movement imitation buffer.
	for i := 0; i < compTimeMIBufSize; i++ {
		compTimeMovImBuf[i] = compTimeMIBufMemb(i)
	}
	// The first half of BDS stays filled with zeros, the other half we fill
	// with anything greater than zero; 0xffff can presumably be memset more
	// easily than 0x0001, so we go with the former.
	for i := 3 * compTimeMIBufSize; i < len(compTimeMovImBuf); i++ {
		compTimeMovImBuf[i] = ^compTimeMIBufMemb(0)
	}

	isqrt := isqrter()
	for ; blockSize < n; blockSize <<= 1 {
		// The square root of the blockSize.
		// May be used as the number of subblocks of the merge sort
		// block that will be rearranged and locally merged. We will
		// search for 2*sqrt distinct elements for 2 internal buffers.
		sqrt := isqrt()

		compTimeMIBIsEnough := sqrt <= compTimeMIBufSize

		// TODO: as an optimization, add logic and integer square root
		// routine for the case when the only w-block is undersized.

		// For the merging algorithms we need "buffers" of distinct
		// elements, extracted from the input array and used for movement
		// imitation of merge blocks and local merges of said blocks.

		// TODO: Instead of every other block, every block could be
		// searched for distinct elements.

		bds0, bds1, backupBDS0, backupBDS1, buf, bufLastDiEl :=
			-1, -1, -1, -1, -1, -1

		// For backup BDS.
		bdsSize := 0

		for a = blockSize << 1; a <= n && (buf == -1 || ((bds0 == -1 || bds0 == buf) && !compTimeMIBIsEnough)); a += blockSize << 1 {
			tmpBDS1, tmpBackupBDS1, bufferFound, tmpBufLastDiEl :=
				findBDSAndCountDistinctElementsNear(data,
					a-blockSize, a, sqrt)

			if tmpBDS1 != -1 && (bds0 == -1 || bds0 == buf) {
				bds0 = a
				bds1 = tmpBDS1
			}

			if bufferFound && (buf == -1 || bds0 == buf) {
				buf = a
				bufLastDiEl = tmpBufLastDiEl
			}

			tmpBDSSize := a - tmpBackupBDS1
			if tmpBackupBDS1 != -1 && bdsSize < tmpBDSSize {
				bdsSize = tmpBDSSize
				backupBDS0 = a
				backupBDS1 = tmpBackupBDS1
			}
		}
		if sqrt<<2 <= n-a+blockSize && (buf == -1 || ((bds0 == -1 || bds0 == buf) && !compTimeMIBIsEnough)) {
			tmpBDS1, tmpBackupBDS1, bufferFound, tmpBufLastDiEl :=
				findBDSAndCountDistinctElementsNear(data,
					a-blockSize, n, sqrt)

			if tmpBDS1 != -1 && (bds0 == -1 || bds0 == buf) {
				bds0 = n
				bds1 = tmpBDS1
			}

			if bufferFound && (buf == -1 || bds0 == buf) {
				buf = n
				bufLastDiEl = tmpBufLastDiEl
			}

			tmpBDSSize := n - tmpBackupBDS1
			if tmpBackupBDS1 != -1 && bdsSize < tmpBDSSize {
				bdsSize = tmpBDSSize
				backupBDS0 = n
				backupBDS1 = tmpBackupBDS1
			}
		}

		// The count of distinct elements in the block containing the buffer (or
		// which will contain the buffer), up to sqrt.
		bufDiElCnt := 1

		// Maximal count of distinct elements over all blocks searched, up to sqrt.
		// We need this variable for the case when the maximal block goes to BDS,
		// as opposed to the buffer.
		maxDiElCnt := 1

		if buf != -1 {
			bufDiElCnt = sqrt
			maxDiElCnt = blockSize
		}
		if bds0 == -1 || buf == -1 || bds0 == buf {
			for a = blockSize << 1; a <= n && (bufDiElCnt < sqrt || ((bds0 == -1 || bds0 == buf) && !compTimeMIBIsEnough)); a += blockSize << 1 {
				tmpBDS1, tmpBackupBDS1, dECnt, tmpBufLastDiEl, dECnt0, tBLDiEl0 :=
					findBDSFarAndCountDistinctElements(data,
						a-blockSize, a, sqrt)

				if tmpBDS1 != -1 && (bds0 == -1 || bds0 == buf) {
					bds0 = a
					bds1 = tmpBDS1
				}

				if maxDiElCnt < dECnt {
					maxDiElCnt = dECnt
				}

				if bufDiElCnt <= dECnt && (bufDiElCnt < sqrt || bds0 == buf) {
					buf = a
					bufDiElCnt = dECnt
					bufLastDiEl = tmpBufLastDiEl

					if !compTimeMIBIsEnough && dECnt0 == sqrt && (bds0 == -1 || bds0 == buf) {
						bds0 = a
						bds1 = tmpBDS1
						buf = tmpBDS1 - (sqrt << 1)
						bufDiElCnt = sqrt
						bufLastDiEl = tBLDiEl0
					}
				}

				tmpBDSSize := tmpBackupBDS1 - (a - blockSize)
				if tmpBackupBDS1 != -1 && bdsSize < tmpBDSSize {
					bdsSize = tmpBDSSize
					backupBDS0 = a
					backupBDS1 = tmpBackupBDS1
				}
			}
			if sqrt<<2 <= n-a+blockSize && (bufDiElCnt < sqrt || ((bds0 == -1 || bds0 == buf) && !compTimeMIBIsEnough)) {
				tmpBDS1, tmpBackupBDS1, dECnt, tmpBufLastDiEl, dECnt0, tBLDiEl0 :=
					findBDSFarAndCountDistinctElements(data,
						a-blockSize, n, sqrt)

				if tmpBDS1 != -1 && (bds0 == -1 || bds0 == buf) {
					bds0 = n
					bds1 = tmpBDS1
				}

				if maxDiElCnt < dECnt {
					maxDiElCnt = dECnt
				}

				if bufDiElCnt <= dECnt && (bufDiElCnt < sqrt || bds0 == buf) {
					buf = n
					bufDiElCnt = dECnt
					bufLastDiEl = tmpBufLastDiEl

					if !compTimeMIBIsEnough && dECnt0 == sqrt && (bds0 == -1 || bds0 == buf) {
						bds0 = n
						bds1 = tmpBDS1
						buf = tmpBDS1 - (sqrt << 1)
						bufDiElCnt = dECnt0
						bufLastDiEl = tBLDiEl0
					}
				}

				tmpBDSSize := tmpBackupBDS1 - (a - blockSize)
				if tmpBackupBDS1 != -1 && bdsSize < tmpBDSSize {
					bdsSize = tmpBDSSize
					backupBDS0 = n
					backupBDS1 = tmpBackupBDS1
				}
			} else if bufDiElCnt < sqrt || bds0 == buf && !compTimeMIBIsEnough {
				dECnt, tmpBufLastDiEl := distinctElementCount(data,
					a-blockSize, n, sqrt)

				if maxDiElCnt < dECnt {
					maxDiElCnt = dECnt
				}

				if bufDiElCnt < dECnt {
					buf = n
					bufDiElCnt = dECnt
					bufLastDiEl = tmpBufLastDiEl
				}
			}
		}

		if maxDiElCnt == sqrt {
			maxDiElCnt = blockSize
		}
		if buf == -1 {
			bufDiElCnt = -1
		}

		// Collapse bds* and backupBDS* into bDS*, meaning the former
		// two will not be used after this block.
		bDS0 := -1
		bDS1 := -1
		if bds0 != buf && bds0 != -1 {
			bDS0 = bds0
			bDS1 = bds1
			bdsSize = sqrt << 1
		} else if backupBDS0 != buf {
			bDS0 = backupBDS0
			bDS1 = backupBDS1
		} else {
			bdsSize = -1
		}

		// Movement imitation (MI) buffer size.
		// The maximal number of blocks handled by BDS is movImitBufSize*2.
		movImitBufSize := bufDiElCnt

		movImData, movImDataIsInputData := data, true
		mergeBlockSize := sqrt
		if sqrt <= compTimeMIBufSize {
			movImitBufSize = sqrt
			movImData, movImDataIsInputData = &compTimeMovImBuf, false
		} else {
			if bdsSize>>1 < bufDiElCnt {
				movImitBufSize = bdsSize >> 1
			}
			if movImitBufSize <= compTimeMIBufSize {
				movImitBufSize = compTimeMIBufSize
				movImData, movImDataIsInputData = &compTimeMovImBuf, false
			}
			mergeBlockSize = blockSize / movImitBufSize
		}
		movImBufI := buf
		if !movImDataIsInputData {
			// MIB and BDS are the ones from the compilation-time
			// assigned array.
			movImBufI = compTimeMIBufSize
			bDS1 = 3 * compTimeMIBufSize
			bDS0 = len(compTimeMovImBuf)
		}

		// Will there be a helping buffer for local merges.
		bufferForMerging := false
		if movImitBufSize == sqrt && sqrt == bufDiElCnt {
			bufferForMerging = true
		}

		// Notes for optimizing: we could have a constant size cache
		// recording counts of distinct elements, to obviate some Less
		// calls later on.
		// A smarter way to search for a buffer would be to start from
		// the one we had in the previous merge sort level.
		// Also, when an array is found to contain only a few distinct
		// members it could maybe be merged right away (there would be
		// a cache of equal member range positions within an array).

		// Search in the last (possibly undersized) high-index block.
		//
		// TODO: could add logic for the case (n<2*blockSize &&
		// n-(a-blockSize)<16), to save Less and Swap calls. (We do not
		// need a buffer if there are just 2 small blocks.)
		// Perhaps more general logic for small blocks should be added,
		// instead.

		// Sift the bufDiEl distinct elements from data[bufLast:buf]
		// through to data[buf-bufDiEl:buf], making a buffer to be
		// used by the merging routines.
		if bufDiElCnt == sqrt || movImDataIsInputData {
			extractDist(data, bufLastDiEl, buf, bufDiElCnt)
			bufLastDiEl = buf - bufDiElCnt
		} else {
			buf = -1
			bufLastDiEl = buf
			bufDiElCnt = 0
		}

		// Merge blocks in pairs.
		for a = blockSize << 1; a <= n; a += blockSize << 1 {
			if !data.Less(a-blockSize, a-blockSize-1) {
				// As an optimization skip this merge sort block
				// pair because they are already merged/sorted.
				continue
			}

			e := a

			if movImDataIsInputData && e == bDS0 {
				e = bDS1 - bdsSize
			}
			if e == buf {
				e = bufLastDiEl
			}

			if bufferForMerging || 5 < maxDiElCnt {
				merge(data, movImData,
					a-(blockSize<<1), a-blockSize, e, buf, mergeBlockSize,
					movImBufI, bDS0, bDS1, maxDiElCnt, bufferForMerging)
			} else {
				simpleMerge(data, a-(blockSize<<1), a-blockSize, e, maxDiElCnt)
			}
		}

		smallBS := n - a + blockSize
		// Merge the possible last pair (with the undersized block).
		if 0 < smallBS {
			e := n

			if movImDataIsInputData && e == bDS0 {
				e = bDS1 - bdsSize
			}
			if e == buf {
				e = bufLastDiEl
			}

			smallBS = e - (a - blockSize)
			if smallBS <= bufDiElCnt {
				// If the whole block can be contained
				// in the buffer, merge it directly.
				simpleMergeBufBigSmall(data,
					a-(blockSize<<1), a-blockSize, e,
					buf)

				// Sort the "buffer" we just used for local merging. An unstable
				// sorting routine can be used here because the buffer consists
				// of mutually distinct elements.
				quickSort(data, buf-smallBS, buf, maxDepth(smallBS))
			} else if bufferForMerging || 5 < maxDiElCnt {
				merge(data, movImData,
					a-(blockSize<<1), a-blockSize, e, buf, mergeBlockSize,
					movImBufI, bDS0, bDS1, maxDiElCnt, bufferForMerging)
			} else {
				simpleMerge(data, a-(blockSize<<1), a-blockSize, e, maxDiElCnt)
			}
		}

		// Merge the possible buffer and BDS with the rest of their block(s).
		// (pair).
		if bufDiElCnt == sqrt || movImDataIsInputData {
			if !movImDataIsInputData {
				// Merge the buffer with the rest of its block,
				// unless it takes up the whole block.
				b := buf - (blockSize << 1)
				if k := buf % (blockSize << 1); k != 0 {
					b = buf - k
				}
				if b != buf-bufDiElCnt {
					simpleMerge(data, b, buf-bufDiElCnt, buf, maxDiElCnt)
				}
			} else if buf == bDS1-bdsSize {
				// Merge BDS and buffer with the rest of their block,
				// unless they take up the whole block.
				b := bDS0 - (blockSize << 1)
				if k := bDS0 % (blockSize << 1); k != 0 {
					b = bDS0 - k
				}
				if b != buf-bufDiElCnt {
					simpleMerge(data, b, buf-bufDiElCnt, bDS0, maxDiElCnt)
				}
			} else {
				// Merge BDS and buffer with the rest of their blocks,
				// unless they take up whole blocks.
				b := bDS0 - (blockSize << 1)
				if k := bDS0 % (blockSize << 1); k != 0 {
					b = bDS0 - k
				}
				if b != bDS1-bdsSize {
					simpleMerge(data, b, bDS1-bdsSize, bDS0, maxDiElCnt)
				}
				b = buf - (blockSize << 1)
				if k := buf % (blockSize << 1); k != 0 {
					b = buf - k
				}
				if b != buf-bufDiElCnt {
					simpleMerge(data, b, buf-bufDiElCnt, buf, maxDiElCnt)
				}
			}
		}
	}
}

// Stable sorts data while keeping the original order of equal elements.
//
// It makes one call to data.Len to determine n, O(n*log(n)) calls to both
// data.Less and data.Swap, and it uses only O(lg(sqrt(n)+1)) additional space.
func Stable(data Interface) {
	// NB: for changing the bounds on space usage, even making it O(1); it is
	// simply necessary to modify the recursion level bounding in the
	// quickSort calls in "merge" and "stable".

	stable(data, data.Len())
}
