// Excluded from sort.go because of not playing well with genzfunc.go, because
// of a variable that implements Interface being used in the code. Sources
// zfuncversion-stable.go.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort

func stable(data Interface, n int) {
	// A merge sort with the merge algorithm based on Kim & Kutzner 2008
	// (Ratio based stable in-place merging).
	//
	// Differences from the paper:
	//
	// We use rounded integer square root where they tell to use the floor
	// (a small mistake on their part).
	//
	// Only searching for buffer and block distribution storage once on merge
	// sort level.
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

		// TODO: In the case when we can not have a buffer for local
		// merges, maybe block rearrangement should be skipped in favor
		// of just doing a Hwang-Lin merge of the merge sort block pair.
		// That would presumably enable better utilisation of
		// maxDiElCnt.

		// TODO: try changing the bds==-1 condition to bds0==-1

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

		for a = blockSize << 1; a <= n && (buf == -1 || ((bds1 == -1 || bds0 == buf) && !compTimeMIBIsEnough)); a += blockSize << 1 {
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
		if sqrt<<2 <= n-a+blockSize && (buf == -1 || ((bds1 == -1 || bds0 == buf) && !compTimeMIBIsEnough)) {
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
		if bds1 == -1 || buf == -1 || bds0 == buf {
			for a = blockSize << 1; a <= n && (bufDiElCnt < sqrt || ((bds1 == -1 || bds0 == buf) && !compTimeMIBIsEnough)); a += blockSize << 1 {
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
			if sqrt<<2 <= n-a+blockSize && (bufDiElCnt < sqrt || ((bds1 == -1 || bds0 == buf) && !compTimeMIBIsEnough)) {
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

		// Collapse bds* and backupBDS* into bDS*.
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
				hL(data, a-(blockSize<<1), a-blockSize, e, maxDiElCnt)
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
				hLBufBigSmall(data,
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
				hL(data, a-(blockSize<<1), a-blockSize, e, maxDiElCnt)
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
					hL(data, b, buf-bufDiElCnt, buf, maxDiElCnt)
				}
			} else if buf == bDS1-bdsSize {
				// Merge BDS and buffer with the rest of their block,
				// unless they take up the whole block.
				b := bDS0 - (blockSize << 1)
				if k := bDS0 % (blockSize << 1); k != 0 {
					b = bDS0 - k
				}
				if b != buf-bufDiElCnt {
					hL(data, b, buf-bufDiElCnt, bDS0, maxDiElCnt)
				}
			} else {
				// Merge BDS and buffer with the rest of their blocks,
				// unless they take up whole blocks.
				b := bDS0 - (blockSize << 1)
				if k := bDS0 % (blockSize << 1); k != 0 {
					b = bDS0 - k
				}
				if b != bDS1-bdsSize {
					hL(data, b, bDS1-bdsSize, bDS0, maxDiElCnt)
				}
				b = buf - (blockSize << 1)
				if k := buf % (blockSize << 1); k != 0 {
					b = buf - k
				}
				if b != buf-bufDiElCnt {
					hL(data, b, buf-bufDiElCnt, buf, maxDiElCnt)
				}
			}
		}
	}
}
