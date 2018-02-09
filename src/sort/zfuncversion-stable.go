// Code meant to be used with zfuncversion.go, but fixed by hand instead of
// auto-generated. Sourced from sort-stable.go.

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort

// Func-optimized variant of sort.go:stable
func stable_func(data lessSwap, n int) {
	a := 0
	for i := 0; i < n; i++ {
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
	blockSize := 1 << 4
	a = 0
	for a+blockSize < n {
		insertionSort_func(data, a, a+blockSize)
		a += blockSize
	}
	insertionSort_func(data, a, n)
	if n <= blockSize {
		return
	}
	if n < 2000 {
		for ; blockSize < n; blockSize <<= 1 {
			for a = blockSize << 1; a <= n; a += blockSize << 1 {
				symMerge_func(data, a-blockSize<<1, a-blockSize, a)
			}
			if a-blockSize < n {
				symMerge_func(data, a-blockSize<<1, a-blockSize, n)
			}
		}
		return
	}
	var compTimeMovImBuf compTimeMIBuffer
	for i := 0; i < compTimeMIBufSize; i++ {
		compTimeMovImBuf[i] = compTimeMIBufMemb(i)
	}
	for i := 3 * compTimeMIBufSize; i < len(compTimeMovImBuf); i++ {
		compTimeMovImBuf[i] = ^compTimeMIBufMemb(0)
	}
	isqrt := isqrter()
	for ; blockSize < n; blockSize <<= 1 {
		sqrt := isqrt()
		compTimeMIBIsEnough := sqrt <= compTimeMIBufSize
		bds0, bds1, backupBDS0, backupBDS1, buf, bufLastDiEl :=
			-1, -1, -1, -1, -1, -1
		bdsSize := 0
		for a = blockSize << 1; a <= n && (buf == -1 || ((bds1 == -1 || bds0 == buf) && !compTimeMIBIsEnough)); a += blockSize << 1 {
			tmpBDS1, tmpBackupBDS1, bufferFound, tmpBufLastDiEl :=
				findBDSAndCountDistinctElementsNear_func(data,
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
				findBDSAndCountDistinctElementsNear_func(data,
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
		bufDiElCnt := 1
		maxDiElCnt := 1
		if buf != -1 {
			bufDiElCnt = sqrt
			maxDiElCnt = blockSize
		}
		if bds1 == -1 || buf == -1 || bds0 == buf {
			for a = blockSize << 1; a <= n && (bufDiElCnt < sqrt || ((bds1 == -1 || bds0 == buf) && !compTimeMIBIsEnough)); a += blockSize << 1 {
				tmpBDS1, tmpBackupBDS1, dECnt, tmpBufLastDiEl, dECnt0, tBLDiEl0 :=
					findBDSFarAndCountDistinctElements_func(data,
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
					findBDSFarAndCountDistinctElements_func(data,
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
				dECnt, tmpBufLastDiEl := distinctElementCount_func(data,
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
		movImitBufSize := bufDiElCnt
		movImData, movImDataIsInputData := data, true
		mergeBlockSize := sqrt

		// Fixed by hand.
		compTimeMovImBuf_func := lessSwap{Less: func(i, j int) bool {
			return compTimeMovImBuf[i] < compTimeMovImBuf[j]
		}, Swap: func(i, j int) {
			compTimeMovImBuf[i], compTimeMovImBuf[j] = compTimeMovImBuf[j], compTimeMovImBuf[i]
		}}

		if sqrt <= compTimeMIBufSize {
			movImitBufSize = sqrt
			movImData, movImDataIsInputData = compTimeMovImBuf_func, false
		} else {
			if bdsSize>>1 < bufDiElCnt {
				movImitBufSize = bdsSize >> 1
			}
			if movImitBufSize <= compTimeMIBufSize {
				movImitBufSize = compTimeMIBufSize
				movImData, movImDataIsInputData = compTimeMovImBuf_func, false
			}
			mergeBlockSize = blockSize / movImitBufSize
		}

		movImBufI := buf
		if !movImDataIsInputData {
			movImBufI = compTimeMIBufSize
			bDS1 = 3 * compTimeMIBufSize
			bDS0 = len(compTimeMovImBuf)
		}
		bufferForMerging := false
		if movImitBufSize == sqrt && sqrt == bufDiElCnt {
			bufferForMerging = true
		}
		if bufDiElCnt == sqrt || movImDataIsInputData {
			extractDist_func(data, bufLastDiEl, buf, bufDiElCnt)
			bufLastDiEl = buf - bufDiElCnt
		} else {
			buf = -1
			bufLastDiEl = buf
			bufDiElCnt = 0
		}
		for a = blockSize << 1; a <= n; a += blockSize << 1 {
			if !data.Less(a-blockSize, a-blockSize-1) {
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
				merge_func(data, movImData,
					a-(blockSize<<1), a-blockSize, e, buf, mergeBlockSize,
					movImBufI, bDS0, bDS1, maxDiElCnt, bufferForMerging)
			} else {
				hL_func(data, a-(blockSize<<1), a-blockSize, e, maxDiElCnt)
			}
		}
		smallBS := n - a + blockSize
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
				hLBufBigSmall_func(data,
					a-(blockSize<<1), a-blockSize, e,
					buf)
				quickSort_func(data, buf-smallBS, buf, maxDepth(smallBS))
			} else if bufferForMerging || 5 < maxDiElCnt {
				merge_func(data, movImData,
					a-(blockSize<<1), a-blockSize, e, buf, mergeBlockSize,
					movImBufI, bDS0, bDS1, maxDiElCnt, bufferForMerging)
			} else {
				hL_func(data, a-(blockSize<<1), a-blockSize, e, maxDiElCnt)
			}
		}
		if bufDiElCnt == sqrt || movImDataIsInputData {
			if !movImDataIsInputData {
				b := buf - (blockSize << 1)
				if k := buf % (blockSize << 1); k != 0 {
					b = buf - k
				}
				if b != buf-bufDiElCnt {
					hL_func(data, b, buf-bufDiElCnt, buf, maxDiElCnt)
				}
			} else if buf == bDS1-bdsSize {
				b := bDS0 - (blockSize << 1)
				if k := bDS0 % (blockSize << 1); k != 0 {
					b = bDS0 - k
				}
				if b != buf-bufDiElCnt {
					hL_func(data, b, buf-bufDiElCnt, bDS0, maxDiElCnt)
				}
			} else {
				b := bDS0 - (blockSize << 1)
				if k := bDS0 % (blockSize << 1); k != 0 {
					b = bDS0 - k
				}
				if b != bDS1-bdsSize {
					hL_func(data, b, bDS1-bdsSize, bDS0, maxDiElCnt)
				}
				b = buf - (blockSize << 1)
				if k := buf % (blockSize << 1); k != 0 {
					b = buf - k
				}
				if b != buf-bufDiElCnt {
					hL_func(data, b, buf-bufDiElCnt, buf, maxDiElCnt)
				}
			}
		}
	}
}
