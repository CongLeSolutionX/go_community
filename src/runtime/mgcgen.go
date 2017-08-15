package runtime

/****/
import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
	"unsafe"
)

var grandTotalCardMarks = uintptr(0)
var grandTotalScanCards = uintptr(0)
var grandTotalNoScanCards = uintptr(0)
var grandTotalPointerCount = uintptr(0)
var grandTotalMatureToYoungPointerCount = uintptr(0)
var grandTotalIgnoredCards = uintptr(0)

var totalCardMarks = uintptr(0)
var totalScanCards = uintptr(0)
var totalNoScanCards = uintptr(0)
var totalPointerCount = uintptr(0)
var totalMatureToYoungPointerCount = uintptr(0)
var totalIgnoredCards = uintptr(0)

// Counts to get a better idea about what can be filtered during the write barrier.
var totalMatureToNilWrites = uintptr(0)
var totalYoungToNilWrites = uintptr(0)
var totalMatureToYoungWrites = uintptr(0)
var totalMatureToMatureWrites = uintptr(0)
var totalYoungToMatureWrites = uintptr(0)
var totalYoungToYoungWrites = uintptr(0)
var totalMatureNonNilNonHeapWrites = uintptr(0)
var totalYoungNonNilNonHeapWrites = uintptr(0)

// Counts to get a better idea about what can be filtered during the write barrier.
var grandTotalMatureToNilWrites = uintptr(0)
var grandTotalYoungToNilWrites = uintptr(0)
var grandTotalMatureToYoungWrites = uintptr(0)
var grandTotalMatureToMatureWrites = uintptr(0)
var grandTotalYoungToMatureWrites = uintptr(0)
var grandTotalYoungToYoungWrites = uintptr(0)
var grandTotalMatureNonNilNonHeapWrites = uintptr(0)
var grandTotalYoungNonNilNonHeapWrites = uintptr(0)

var totalMarkingMarkedCard = uintptr(0) // This are marking of filtered (already mature-to-young filtered cards)
var grandTotalMarkingMarkedCard = uintptr(0)

// var totalMarkingUnfilteredMarkedCard = uintptr(0) // Marking as mature-to-young cards with an unfiltered mark
// var grandTotalMarkingUnfilteredMarkedCard = uintptr(0)
var totalWriteBarrierCalls = uintptr(0)
var grandTotalWriteBarrierCalls = uintptr(0)
var totalMarkedAllYoungCount = uintptr(0)
var grandTotalMarkedAllYoungCount = uintptr(0)

var totalUnfilteredMarks = uintptr(0) // The slot/card is in a heap span with a _MSpanInUse state
var grandTotalUnfilteredMarks = uintptr(0)

const matureToYoungMark = byte(66)
const unfilteredMark = byte(42)

//go:nosplit
func gatherCardMarkWBInfo(src, dst uintptr) {
	atomic.Xadduintptr(&totalWriteBarrierCalls, 1)
	if inheap(src) {
		if isMature(src) {
			if dst == 0 {
				// write of nil
				atomic.Xadduintptr(&totalMatureToNilWrites, 1)
			} else if inheap(dst) {
				if isYoung(dst) {
					if !isMatureToYoung(src, dst) {
						println("runtime !isMatureToYoun(src,dst) src=", hex(src), "dst=", hex(dst),
							"isMature(src)=", isMature(src), "isYoung(dst)=", isYoung(dst))
						throw("why")
					}
					offset := src - uintptr(unsafe.Pointer(mheap_.arena_start))
					index := offset / _CardBytes // turns into a shift.
					if *addb(mheap_.cardMarks, index) != 0 {
						if !(*addb(mheap_.cardMarks, index) == 66 || *addb(mheap_.cardMarks, index) == 42) {
							println("card mark index is not 0 and not 66")
							throw("why")
						}
						if *addb(mheap_.cardMarks, index) == 66 {
							atomic.Xadduintptr(&totalMarkingMarkedCard, 1) // Marking a filtered card.
						}
						//if *addb(mheap_.cardMarks, index) == 42 {
						//	atomic.Xadduintptr(&totalMarkingUnfilteredMarkedCard, 1) // Marking a filtered card.
						//}
					}
					atomic.Xadduintptr(&totalMatureToYoungWrites, 1)
				} else if isMature(dst) {
					atomic.Xadduintptr(&totalMatureToMatureWrites, 1)
				}
			} else {
				atomic.Xadduintptr(&totalMatureNonNilNonHeapWrites, 1)
			}
		} else {
			// Source is in the heap but not mature so it is young.
			if dst == 0 {
				// write of nil
				atomic.Xadduintptr(&totalYoungToNilWrites, 1)
			} else if inheap(dst) {
				if isYoung(dst) {
					atomic.Xadduintptr(&totalYoungToYoungWrites, 1)
				} else if isMature(dst) {
					atomic.Xadduintptr(&totalYoungToMatureWrites, 1)
				}
			} else {
				atomic.Xadduintptr(&totalYoungNonNilNonHeapWrites, 1)
			}
		}
	}
}

func (aspan *mspan) xxgatherSpanCardInfo() (cardMarks, scanCards, noScanCards, pointerCount, matureToYoungPointerCount uintptr) {
	return
}

// gatherCardShardInfo take a shard passed as an index starting at shard*rootCardMarkShardSize and
// ended rootCardMarkShardSize later.
// It returns various counters intended to be used to quantify heap makeup.
func gatherCardShardInfo(shard int) (cardMarkCount, scanCards, noScanCards, pointerCount,
	matureToYoungPointerCount, ignoredCards, markedAllYoungCount uintptr) {
	trace := false
	cardTableIndex := uintptr(shard * rootCardMarkShardSize)
	cardMarks := addb(mheap_.cardMarks, cardTableIndex)
	if trace {
		println("cardTableIndex", cardTableIndex, "rootCardMarkShardSize", rootCardMarkShardSize,
			"cardMarks=", cardMarks, "mheap_.cardMarks=", mheap_.cardMarks)
	}
	if debug.gcgen == 0 {
		throw("gatherCardShardInfo debug.gcgen == 0 || debug.gctrace == 0")
	}

	var nilPtr, slotToNonHeap uintptr
	if trace {
		println("scanning shard", shard, "card from ", cardTableIndex, "to ", cardTableIndex+rootCardMarkShardSize)
	}
	for ind := cardTableIndex; ind < cardTableIndex+rootCardMarkShardSize; ind++ {
		markVal := *(addb(mheap_.cardMarks, ind))

		cardStart := mheap_.arena_start + uintptr(ind)*_CardBytes
		aspan := spanOf(cardStart)
		if aspan.state == _MSpanManual {
			noScanCards++
			continue
		}
		if markVal == 0 {
			if aspan == nil {
				continue
			}
			if aspan.spanclass.noscan() {
				noScanCards++
			} else {
				scanCards++
			}
			continue
		}
		if !(markVal == 66 || markVal == 42) && memstats.numgc != 0 {
			println("markVal=", markVal, "ind=", ind, "memstats.numgc=", memstats.numgc)
			throw("why")
		}
		if aspan == nil {
			throw("mark found but span is not valid")
		}
		if aspan.state != _MSpanInUse {
			println("runtime: BAD_WHY Perhaps _MSpanManual issue. ------------- runtime: aspan.state is not _MspanInUse is ", mSpanStateNames[aspan.state])
			throw("aspan.state != _MSpanInUse has card mark set.")
		}
		if aspan.spanclass.noscan() {
			continue
			// println("aspan is noscan() and card is marked ", mSpanStateNames[aspan.state], "markVal=", markVal, "shard=", shard)
			throw("Marked card in noscan span")
		}
		if trace {
			println("SpanStateNames[aspan.state]=", mSpanStateNames[aspan.state])
		}

		scanCards++
		cardMarkCount++
		*(addb(mheap_.cardMarks, ind)) = 0 // Clear the card mark
		objSize := aspan.elemsize
		objIndex := aspan.objIndex(cardStart)
		objStart := objIndex*objSize + aspan.base()
		cardEnd := cardStart + _CardBytes

		if trace {
			println("150: objIndex=", objIndex, "objStart=", hex(objStart), "aspan.base()", hex(aspan.base()))
		}
		matureObjectFound := false
		for ; objStart < cardEnd; objStart += objSize {
			// Only scan mature object or we risk scanning objects with bogus bit maps.

			// if !isMature(objStart) {

			if trace {
				println("objIndex=", objIndex, "objStart=", hex(objStart), "aspan.base()", hex(aspan.base()),
					"aspan.allocBits=", aspan.allocBits)
			}
			if !aspan.allocBitsForIndex(objIndex).isMarked() {
				if isMature(objStart) {
					throw("why")
				}
				objIndex++
				continue
			}
			matureObjectFound = true
			objIndex++
			promoteReferents(objStart, cardStart, cardEnd, objSize)
			if debug.gctrace >= 1 {
				tcount, tnilPtr, tmatureToYoungPointerCount, tslotToNonHeap, tmatureObjectFound :=
					countObject(objStart, cardStart, cardEnd, objSize)
				matureObjectFound = tmatureObjectFound
				pointerCount += tcount
				nilPtr += tnilPtr
				matureToYoungPointerCount += tmatureToYoungPointerCount
				slotToNonHeap += tslotToNonHeap
			}
		}

		// Count the number of cards that did not contain a mature object.
		if !matureObjectFound {
			markedAllYoungCount++
		}
	}

	if trace {
		println("Done 175")
	}
	return
}

/**
func (aspan *mspan) gatherSpanCardInfoxx() (cardMarkCount, scanCards, noScanCards, pointerCount,
	matureToYoungPointerCount, ignoredCards uintptr) {

	if aspan != nil {
		println("gatherSpanCardInfo 210 aspan.base() = ", hex(aspan.base()), "to aspan.limit=", hex(aspan.limit),
			"aspan.npages=", aspan.npages)
	}
	if debug.gcgen == 0 {
		throw("gatherSpanCardInfo debug.gcgen == 0 || debug.gctrace == 0")
	}
	var nilPtr, slotToNonHeap uintptr

	cardsPerPage := uintptr(8192 / _CardBytes) // 16
	if aspan.state == _MSpanInUse {
		if aspan.spanclass.noscan() {
			noScanCards = aspan.npages * cardsPerPage
		} else {
			//	baseCardIndex := cardIndex(aspan.base())
			//	maxCardIndex := cardIndex(aspan.limit)
			scanCards = aspan.npages * cardsPerPage // Thats 8K pages and 512 bytes / card....
			objSize := aspan.elemsize
			baseCardIndex := cardIndex(aspan.base())
			// Count the number of marked cards.
			for i := uintptr(0); i < scanCards; i++ {
				if *addb(mheap_.cardMarks, baseCardIndex+i) != 0 {
					if *addb(mheap_.cardMarks, baseCardIndex+i) != 66 {
						println("gatherSpanCardInfo card at index ", i, "with baseCardIndex of ", baseCardIndex,
							" is set to", *addb(mheap_.cardMarks, baseCardIndex+i), "instead of 66.")
					}
					// Clear the mark...
					// Move clearing the mark to the sweep routine so we can look a young object cards.
					*addb(mheap_.cardMarks, baseCardIndex+i) = 0
					//println("-----card start ---------")
					cardMarkCount++
					// Count the number of pointers in the card.
					// Find the start of the first object in the card. The object may
					// start in a previous card.
					objStart := aspan.base() + ((i*_CardBytes)/objSize)*objSize
					cardStart := aspan.base() + _CardBytes*i
					cardEnd := cardStart + _CardBytes
					objIndex := aspan.objIndex(objStart)
					for ; objStart < cardEnd; objStart += objSize {
						// Only scan mature object or we risk scanning objects with bogus bit maps.

						// if !isMature(objStart) {
						if !aspan.allocBitsForIndex(objIndex).isMarked() {
							if isMature(objStart) {
								throw("why")
							}
							objIndex++
							continue
						}
						objIndex++
						promoteReferents(objStart, cardStart, cardEnd, objSize)
						if debug.gctrace >= 1 {
							tcount, tnilPtr, tmatureToYoungPointerCount, tslotToNonHeap, matureObjectFound :=
								countObject(objStart, cardStart, cardEnd, objSize)
							pointerCount += tcount
							nilPtr += tnilPtr
							matureToYoungPointerCount += tmatureToYoungPointerCount
							slotToNonHeap += tslotToNonHeap
						}
					}
					// Note that it is possible to have a marked card without finding a slot in it. This can happen
					// since we don't inspect young objects and a write into a young object may well cause the card
					// to be marked.
					//					println("-----card end -----------")
				} else {
					// We have an unmarked card. If we know that the object is alive due to the fact that the allocation pointer
					// is past the object then, for debugging purposes and statistic gathering, we can look at the object
					// and if it contains a pointer then we can throw since an object was allocated and a pointer was written to
					// it since the last GC but the card has not been marked....
				}
			}
		}
	} else {
		ignoredCards = aspan.npages * cardsPerPage
	}
	return
}
**/
// Some heap characterization stuff.
func printHeapCardInfo() {

	if !cardMarkOn {
		return
	}
	if atomic.Load(&gcphase) == _GCoff {
		throw("why is gcphase == _GCoff? Perhaps a card marking bug.")
	}

	if debug.gcgen == 0 || debug.gctrace == 0 {
		throw("printHeapCardInfo debug.gcgen == 0 || debug.gctrace == 0")
	}
	if !cardMarkOn {
		return
	}
	grandTotalCardMarks += totalCardMarks
	grandTotalScanCards += totalScanCards
	grandTotalNoScanCards += totalNoScanCards
	grandTotalPointerCount += totalPointerCount
	grandTotalMatureToYoungPointerCount += totalMatureToYoungPointerCount
	grandTotalIgnoredCards += totalIgnoredCards

	grandTotalMatureToNilWrites += totalMatureToNilWrites
	grandTotalYoungToNilWrites += totalYoungToNilWrites
	grandTotalMatureToYoungWrites += totalMatureToYoungWrites
	grandTotalMatureToMatureWrites += totalMatureToMatureWrites
	grandTotalYoungToMatureWrites += totalYoungToMatureWrites
	grandTotalYoungToYoungWrites += totalYoungToYoungWrites
	grandTotalMatureNonNilNonHeapWrites += totalMatureNonNilNonHeapWrites
	grandTotalYoungNonNilNonHeapWrites += totalYoungNonNilNonHeapWrites

	grandTotalWriteBarrierCalls += totalWriteBarrierCalls
	grandTotalMarkingMarkedCard += totalMarkingMarkedCard
	grandTotalUnfilteredMarks += totalUnfilteredMarks
	totalMarkedAllYoungCount += grandTotalMarkedAllYoungCount

	println("Metric,\t\t\t\t\tGC Cycle,\tTotal",
		"\n Cards in scannable spans=,\t\t", totalScanCards, ",\t", grandTotalScanCards,
		"\n Cards in NoScan spans,\t\t\t", totalNoScanCards, ",\t", grandTotalNoScanCards,
		"\n Other ignored cards,\t\t\t", totalIgnoredCards, ",\t", grandTotalIgnoredCards,
		"\n Cards,\t\t\t\t\t", totalScanCards+totalNoScanCards, ",\t", grandTotalScanCards+grandTotalNoScanCards,
		"\n Marked cards(MC),\t\t\t", totalCardMarks, ",\t", grandTotalCardMarks,
		"\n Pointers found in MCs,\t\t\t", totalPointerCount, ",\t", grandTotalPointerCount,
		"\n Mature to Young found in MCs,\t\t", totalMatureToYoungPointerCount, ",\t", grandTotalMatureToYoungPointerCount,
		"\n Percent Total Mature to Young in MCs,,\t", ((float64(grandTotalMatureToYoungPointerCount) / float64(grandTotalPointerCount)) * float64(100.0)),
		"\n Percent Scannable cards marked,,\t", ((float64(grandTotalCardMarks) / float64(grandTotalScanCards)) * float64(100.0)),
		"\n Percent Total of cards marked,,\t\t", ((float64(grandTotalCardMarks) / float64(grandTotalScanCards+grandTotalNoScanCards)) * float64(100.0)),
		"\n\n Stats from write barrier",
		"\n Write barrier calls,\t\t\t", totalWriteBarrierCalls, ",\t", grandTotalWriteBarrierCalls,
		"\n Marking already marked card,\t\t", totalMarkingMarkedCard, ",\t", grandTotalMarkingMarkedCard,
		"\n Unfiltered marked card,\t\t", totalUnfilteredMarks, ",\t", grandTotalUnfilteredMarks,
		"\n Marked Cards without Mature Objects,\t\t", totalMarkedAllYoungCount, ",\t", grandTotalMarkedAllYoungCount,
		"\n Mature to nil writes,\t\t\t", totalMatureToNilWrites, ",\t", grandTotalMatureToNilWrites,
		"\n Young to nil writes,\t\t\t", totalYoungToNilWrites, ",\t", grandTotalYoungToNilWrites,
		"\n Mature to Young writes,\t\t", totalMatureToYoungWrites, ",\t", grandTotalMatureToYoungWrites,
		"\n Mature to Mature writes,\t\t", totalMatureToMatureWrites, ",\t", grandTotalMatureToMatureWrites,
		"\n Young to Mature writes,\t\t", totalYoungToMatureWrites, ",\t", grandTotalYoungToMatureWrites,
		"\n Young to Young writes,\t\t\t", totalYoungToYoungWrites, ",\t", grandTotalYoungToYoungWrites,
		"\n Mature to non-nil/non-heap writes,\t", totalMatureNonNilNonHeapWrites, ",\t", grandTotalMatureNonNilNonHeapWrites,
		"\n Young to non-nil/non-heap writes,\t", totalYoungNonNilNonHeapWrites, ",\t", grandTotalYoungNonNilNonHeapWrites,
		"\n Heap size,\t\t\t\t\t", mheap_.arena_used-mheap_.arena_start,
		",\tnumber of cards, should be used-start/512", (mheap_.arena_used-mheap_.arena_start)/_CardBytes,
	)

	// Clear for next GC cycle
	if totalPointerCount < totalMatureToYoungPointerCount {
		throw("totalPointerCount < totalMatureToYoungPointerCount")
	}

	totalCardMarks = uintptr(0)
	totalScanCards = uintptr(0)
	totalNoScanCards = uintptr(0)
	totalPointerCount = uintptr(0)
	totalMatureToYoungPointerCount = uintptr(0)
	totalIgnoredCards = uintptr(0)

	totalMatureToNilWrites = uintptr(0)
	totalYoungToNilWrites = uintptr(0)
	totalMatureToYoungWrites = uintptr(0)
	totalMatureToMatureWrites = uintptr(0)
	totalYoungToMatureWrites = uintptr(0)
	totalYoungToYoungWrites = uintptr(0)
	totalMatureNonNilNonHeapWrites = uintptr(0)
	totalYoungNonNilNonHeapWrites = uintptr(0)

	totalWriteBarrierCalls = uintptr(0)
	totalMarkingMarkedCard = uintptr(0)
	totalUnfilteredMarks = uintptr(0)
	totalMarkedAllYoungCount = uintptr(0)

	// At this point there should be no marked cards. Check that this is true
	cardBase := mheap_.cardMarks
	markedCardCount := uintptr(0)
	for i := uintptr(0); i < mheap_.cardMarksMapped; i++ {
		if *(addb(cardBase, i)) != 0 {
			markedCardCount++
			heapAddr := mheap_.arena_start + i*_CardBytes
			// get the associated span.
			s := spanOf(heapAddr)
			if s == nil {
				println("------------- span associated with marked card is nil.")
			} else {
				println("------------- stray card mark span state:", mSpanStateNames[s.state],
					", start address", hex(s.startAddr), ", s.npages:", s.npages, ", s.sweepgen:", s.sweepgen,
					"s.elemsize:", s.elemsize, "*(addb(cardBase, i)), should be 66 =", *(addb(cardBase, i)))
			}
			*(addb(cardBase, i)) = 0
		}
	}
	if markedCardCount != 0 && memstats.numgc != 0 {
		println("runtime:bad -- markedCardCount (should be 0) = ", markedCardCount, "-------------BAD------------")
		//throw("why")
	}
	if atomic.Load(&gcphase) == _GCoff {
		throw("why printHeapCardInfo bottom")
	}
}

// Some debug stuff.
func scanSpan() {
	spanLen := len(mheap_.spans)
	println("len(mheap.spans)", spanLen)

	cardsPerPage := uintptr(16)
	for i := 0; i < spanLen; {
		aspan := mheap_.spans[i]
		if aspan == nil {
			i++
			continue
		}
		if aspan.state == _MSpanInUse {
			if aspan.spanclass.noscan() {
			} else {
				baseCardIndex := cardIndex(aspan.base())
				numCardsInSpan := aspan.npages * cardsPerPage // Thats 8K pages and 512 bytes / card....
				aspan.processCard(baseCardIndex, numCardsInSpan)
				// assuming 512 bytes that 16 cards per page.
			}
		}
		i = i + int(aspan.npages)
	}
}
func (aspan *mspan) processCard(baseCardIndex, numCardsInSpan uintptr) {

}

// Card mark
// card = (ptr - runtime.cardMarks.arenaStart) / cardBytes
// if card < runtime.cardMarks.mapped {
//   *(runtime.cardMarks.base+card) = 1
// }

func (aspan *mspan) scanCards(baseCardIndex, count uintptr) uintptr {
	markCount := uintptr(0)
	cardBase := uintptr(unsafe.Pointer(mheap_.cardMarks)) + baseCardIndex
	for i := uintptr(0); i < count; i++ {
		if *((*byte)(unsafe.Pointer(cardBase + i))) != 0 {
			markCount++
			// and clear it for next GC.
			*((*byte)(unsafe.Pointer(cardBase + i))) = byte(0)
		}
	}

	//if markCount != 0 {
	//	throw("cardmark write barrier is off but card is marked.")
	//}
	return markCount
}

// markUnfilteredCard marks a card if a _any_ pointer is written
// into the card. Compare this to the normal markCard below that only
// marks a card if a young pointer is being written into a mature slot.
//
// This is not called since the marking would race between filtered
// and unfiltered marks. A count is maintained in gcmarkwb_m of
// cards that would have been marked if there were
// It is possible that a card is not marked but before the 42 is written
// another goroutine may mark it as a 62 indicating there is a mature
// to young pointer present. This is OK but requires that we inspect
// all marked cards for mature to young pointer, not just those marked 62.
func unfilteredMarkCard(ptr uintptr) {
	if atomic.Load(&gcphase) != _GCoff {
		throw("markCard top how")
	}

	offset := ptr - uintptr(unsafe.Pointer(mheap_.arena_start))
	index := offset / _CardBytes // turns into a shift.

	// Only mark unmarked cards with 42 thus preserving
	// any marks made by the markCard routine below.
	if *addb(mheap_.cardMarks, index) == 0 {
		*addb(mheap_.cardMarks, index) = 42
		atomic.Xadduintptr(&totalUnfilteredMarks, 1)
	}

	if atomic.Load(&gcphase) != _GCoff {
		throw("markCard bottom how")
	}
}

// The card table is starts at mheap_.cardTable which
// is a the base of the memory reserved for the heap.
// The intent is that we can go from a heap address to the
// card table base without having to refer to mheap_.cardTable
// This seems reasonable if we have a continuous heap area or we
// can afford to reserve a card table large enough to cover any
// potential heap pages.
// Nosplit because marks shouldn't happen during mark and mark termination
// and there is a check later that determines if a card was marked
// during mark or mark termination, which is check atomically
// by loading gcphase and comparing it with _GCoff.
//go:nosplit
func markCard(ptr uintptr) {
	if atomic.Load(&gcphase) != _GCoff {
		throw("markCard top how")
	}

	offset := ptr - uintptr(unsafe.Pointer(mheap_.arena_start))
	index := offset / _CardBytes // turns into a shift.
	// Use the slice so we get bounds checking and consider
	// using address hacking for speed.
	//	if uintptr(unsafe.Pointer(&mheap_.cards[index])) != uintptr(unsafe.Pointer(mheap_.cardMarks))+index {
	//		throw("unsafe.Pointer(mheap_.cards[index])) != mheap_.cardMarks+index")
	//	}
	//	*((*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(mheap_.cardMarks)) + index))) = 1
	*addb(mheap_.cardMarks, index) = 66
	//	mheap_.cards[index] = byte(255)

	if atomic.Load(&gcphase) != _GCoff {
		throw("markCard bottom how")
	}
}

//go:nosplit
func isAllocBitSet(ptr uintptr) bool {
	aBits := allocBitsForAddr(ptr)
	return aBits.isMarked()
}

// isMature takes an object and returns true if it is marked.
//go:nosplit
func isMature(ptr uintptr) bool {
	if !inheap(ptr) {
		// println("asking about ptr not in heap. ptr = ", hex(ptr))
		return false
	}
	return isAllocBitSet(ptr)
}

//go:nosplit
func isYoung(ptr uintptr) bool {
	return !isMature(ptr)
}

// matureToYoung returns true if both the slot and
// the dst are in the heap, the dst is a young object,
// and slot is in a mature object.
// otherwise it returns false.
//go:nosplit
func isMatureToYoung(slot, dst uintptr) bool {
	if inheap(dst) {
		if isYoung(dst) {
			if inheap(slot) {
				return isMature(slot)
			}
		}
	}
	return false
}

// RLH Note: markCard is actually what is not inlined by the compiler
// the isMatureToYoung should be done during the scan.
//go:nosplit
func cardMarkWB(src, dst uintptr) {
	if debug.gctrace >= 1 {
		gatherCardMarkWBInfo(src, dst)
	}
	if isMatureToYoung(src, dst) {
		//		println("marking card at ", src)
		markCard(src)
	}
}

func (s *mspan) objAddr(p uintptr) uintptr {
	return s.elemsize * s.objIndex(p)
}

// scanCard scans the ith card, which is known to
// be in span s.
// TBD RLH: Scan only the part of the object in the card.
// For now we scan an object repeatedly for each
// card that is marked that the object is a part of.
func scanCard(i uintptr, s *mspan, gcw *gcWork) {
	start := mheap_.arena_start + i*_CardBytes // Address of the start of the card.
	objBase := s.objAddr(start)                // Start of object occupying the start of the card.
	ceiling := start + _CardBytes
	foundMatureObj := false
	for ; objBase <= ceiling; objBase = objBase + s.elemsize {
		if isMature(objBase) {
			// You don't have to check to see if the slot hold pointers
			// to young object explicitly. Unmarked (same as young) gets
			// marked (same as mature) so we just scan the object.
			scanobject(objBase, gcw)
			foundMatureObj = true
		}
	}
	if !foundMatureObj {
		throw("card marked but no mature objects found")
	}
}

// countCard scans the ith card, which is known to
// be in span s, counting the number of pointers
// it finds.
// For now we scan an object repeatedly for each
// card that is marked that the object is a part of.
// We only scan mature objects.
/***
func countCard(i uintptr, s *mspan, gcw *gcWork) uintptr {
	count := uintptr(0)
	start := mheap_.arena_start + i*_CardBytes
	objBase := s.objAddr(start)
	ceiling := start + _CardBytes
	foundMatureObj := false
	if s.elemsize == 0 {
		throw("s.elemsize == 0")
	}
	for ; objBase <= ceiling; objBase = objBase + s.elemsize {
		if isMature(objBase) {
			count += countObject(objBase, start, ceiling) // counts the number of pointers in the object.
			foundMatureObj = true
		}
	}
	if !foundMatureObj {
		throw("card marked but no mature objects found")
	}
	return count
}
***/

var onePrintCall = true

var onePrintFound = true

// just a place holder for setting the alloc bits.
var matureToYoungDoIt = uintptr(0)

func promoteReferents(b, cardStart, cardEnd, objSize uintptr) {
	// Find the bits for b and the size of the object at b.
	//
	// b is either the beginning of an object, in which case this
	// is the size of the object to scan, or it points to an
	// oblet, in which case we compute the size to scan below.

	var hbits heapBits
	var scanStart uintptr
	scanEnd := b + objSize
	if b < cardStart {
		hbits = heapBitsForAddr(cardStart)
		scanStart = cardStart
		// If object starts in an earlier card skip all pointers before this card.
	} else {
		hbits = heapBitsForAddr(b)
		scanStart = b
	}
	if scanEnd > cardEnd {
		scanEnd = cardEnd
	}
	i := uintptr(0)
	x := uintptr(0)
	for i = scanStart; i < scanEnd; i += sys.PtrSize {
		x++
		// i is offset into object.
		// Find bits for this word.
		if i != scanStart {
			// Avoid needless hbits.next() on last iteration.
			hbits = hbits.next()
		}
		// Load bits once. See CL 22712 and issue 16973 for discussion.
		bits := hbits.bits()
		// During checkmarking, 1-word objects store the checkmark
		// in the type bit for the one word. The only one-word objects
		// are pointers, or else they'd be merged with other non-pointer
		// data into larger allocations.
		if i != (b+1*sys.PtrSize) && bits&bitScan == 0 {
			break // no more pointers in this object
		}
		if bits&bitPointer == 0 {
			//			println("continue at 352")
			continue // not a pointer
		}
		// b is the base of the source object, determine if it is a mature object.
		// Be careful I might be scanning garbage objects that point to freed up objects.....
		// How do I determine if this object is alive or dead. If it is dead I want to skip it.
		// Work here is duplicated in scanblock and above.
		// If you make changes here, make changes there too.
		slot := (*uintptr)(unsafe.Pointer(i))
		obj := *slot

		//if !inheap(obj) {
		// nil is not in the heap.
		//	continue
		//}
		// It could be a pointer to the interior of an object. Need the base to check
		// the allocation bit.
		//dst := objBase(obj)
		dst, _, dstSpan, dstIndex := heapBitsForObject(obj, b, i-b)
		if dst == 0 {
			// dst not in heap (possibly nil)
			continue
		}
		if !dstSpan.allocBitsForIndex(dstIndex).isMarked() {
			//		if !isAllocBitSet(dst) {
			// Ultimately this where the payload of making dst mature is done.
			// set alloc bit is done here in order to promote dst to mature.
			matureToYoungDoIt++
		}
		// Ignore slots that are not in mature objects since we don't know if they are valid.
		// In anycase they are uninteresting w.r.t. generational GC.
	}
	return
}

// countObject scans the object starting at b accumulating the number
// of pointers in the object.
// b must point to the beginning of a heap object or an oblet.
// b must point to a mature object.
// scanobject consults the GC bitmap for the pointer mask and the
// spans for the size of the object.
//
//go:nowritebarrier
func countObject(b, cardStart, cardEnd, objSize uintptr) (count, nilPtr, matureToYoung, slotToNonHeap uintptr, matureObjectFound bool) {
	//if onePrintCall {
	//	onePrintCall = false
	//	println("*********called countObject**********")
	//}

	// Note that arena_used may change concurrently during
	// scanobject and hence scanobject may encounter a pointer to
	// a newly allocated heap object that is *not* in
	// [start,used). It will not mark this object; however, we
	// know that it was just installed by a mutator, which means
	// that mutator will execute a write barrier and take care of
	// marking it. This is even more pronounced on relaxed memory
	// architectures since we access arena_used without barriers
	// or synchronization, but the same logic applies.

	//arena_start := mheap_.arena_start
	//arena_used := mheap_.arena_used

	// Find the bits for b and the size of the object at b.
	//
	// b is either the beginning of an object, in which case this
	// is the size of the object to scan, or it points to an
	// oblet, in which case we compute the size to scan below.

	var hbits heapBits
	var scanStart uintptr
	scanEnd := b + objSize
	if b < cardStart {
		hbits = heapBitsForAddr(cardStart)
		scanStart = cardStart
		// If object starts in an earlier card skip all pointers before this card.
	} else {
		hbits = heapBitsForAddr(b)
		scanStart = b
	}
	if scanEnd > cardEnd {
		scanEnd = cardEnd
	}
	i := uintptr(0)
	x := uintptr(0)
	for i = scanStart; i < scanEnd; i += sys.PtrSize {
		x++
		// i is offset into object.
		// Find bits for this word.
		if i != scanStart {
			// Avoid needless hbits.next() on last iteration.
			hbits = hbits.next()
		}
		// Load bits once. See CL 22712 and issue 16973 for discussion.
		bits := hbits.bits()
		// During checkmarking, 1-word objects store the checkmark
		// in the type bit for the one word. The only one-word objects
		// are pointers, or else they'd be merged with other non-pointer
		// data into larger allocations.
		if i != (b+1*sys.PtrSize) && bits&bitScan == 0 {
			//			println("break at 340")
			//			println(" cardStart=", hex(cardStart), " cardEnd=", hex(cardEnd), " i=", hex(i), " s.elemsize=", s.elemsize, "b=", hex(b))
			break // no more pointers in this object
		}
		/**
		if i >= cardEnd {
			//			println("break at 344")
			break // there are no more pointers in this card
		}
		if i < cardStart {
			throw("i < cardStart")
			//			println("continue at 348")
			continue // ignore pointers before the start of the card in question
		}
		*/
		if bits&bitPointer == 0 {
			//			println("continue at 352")
			continue // not a pointer
		}
		//		println("found pointer at 356 count is now", count)
		// if onePrintFound {
		//	onePrintFound = false
		//	println("*********called countObject, found object with a slot **********")
		// }
		count++

		// b is the base of the source object, determine if it is a mature object.
		// Be careful I might be scanning garbage objects that point to freed up objects.....
		// How do I determine if this object is alive or dead. If it is dead I want to skip it.
		// Work here is duplicated in scanblock and above.
		// If you make changes here, make changes there too.
		slot := (*uintptr)(unsafe.Pointer(i))
		obj := *slot
		if obj == 0 {
			nilPtr++
			continue
		}

		if !inheap(obj) {
			slotToNonHeap++
			continue
		}
		dst, _, _, _ := heapBitsForObject(obj, b, i-b)
		if !isAllocBitSet(dst) {
			// Ultimately this where the payload of making dst mature is done.
			// set alloc bit is done here in order to promote dst to mature.
			matureToYoung++
		}
		// Ignore slots that are not in mature objects since we don't know if they are valid.
		// In anycase they are uninteresting w.r.t. generational GC.
	}
	//	println("count=", count)
	//println("countObject returns ", b, cardStart, cardEnd, count)
	return
}

// countObject scans the object starting at b accumulating the number
// of pointers in the object.
// b must point to the beginning of a heap object or an oblet.
// scanobject consults the GC bitmap for the pointer mask and the
// spans for the size of the object.
//
//go:nowritebarrier
func countObjectFull(b, cardStart, cardEnd uintptr) (count, nilPtr, matureToYoung, matureToMature, youngToYoung, youngToMature, slotToNonHeap uintptr) {
	// Note that arena_used may change concurrently during
	// scanobject and hence scanobject may encounter a pointer to
	// a newly allocated heap object that is *not* in
	// [start,used). It will not mark this object; however, we
	// know that it was just installed by a mutator, which means
	// that mutator will execute a write barrier and take care of
	// marking it. This is even more pronounced on relaxed memory
	// architectures since we access arena_used without barriers
	// or synchronization, but the same logic applies.

	//arena_start := mheap_.arena_start
	//arena_used := mheap_.arena_used

	// Find the bits for b and the size of the object at b.
	//
	// b is either the beginning of an object, in which case this
	// is the size of the object to scan, or it points to an
	// oblet, in which case we compute the size to scan below.
	hbits := heapBitsForAddr(b)
	s := spanOfUnchecked(b)
	n := s.elemsize
	if n == 0 {
		throw("scanobject n == 0")
	}
	// start := cardStart - s.base()
	// end := cardEnd - s.base()
	i := uintptr(0)
	for i = uintptr(0); i < n; i += sys.PtrSize {
		// i is offset into object.
		// Find bits for this word.
		if i != 0 {
			// Avoid needless hbits.next() on last iteration.
			hbits = hbits.next()
		}
		// Load bits once. See CL 22712 and issue 16973 for discussion.
		bits := hbits.bits()
		// During checkmarking, 1-word objects store the checkmark
		// in the type bit for the one word. The only one-word objects
		// are pointers, or else they'd be merged with other non-pointer
		// data into larger allocations.
		if i != 1*sys.PtrSize && bits&bitScan == 0 {
			//			println("break at 340")
			//			println(" cardStart=", hex(cardStart), " cardEnd=", hex(cardEnd), " i=", hex(i), " s.elemsize=", s.elemsize, "b=", hex(b))
			break // no more pointers in this object
		}
		if b+i >= cardEnd {
			//			println("break at 344")
			break // there are no more pointers in this card
		}
		if b+i < cardStart {
			//			println("continue at 348")
			continue // ignore pointers before the start of the card in question
		}
		if bits&bitPointer == 0 {
			//			println("continue at 352")
			continue // not a pointer
		}
		//		println("found pointer at 356 count is now", count)
		count++

		// b is the base of the source object, determine if it is a mature object.
		// Be careful I might be scanning garbage objects that point to freed up objects.....
		// How do I determine if this object is alive or dead. If it is dead I want to skip it.
		// Work here is duplicated in scanblock and above.
		// If you make changes here, make changes there too.
		slot := (*uintptr)(unsafe.Pointer(b + i))
		obj := *slot
		if obj == 0 {
			nilPtr++
			continue
		}

		if !inheap(obj) {
			slotToNonHeap++
			continue
		}
		dst, _, _, _ := heapBitsForObject(obj, b, i)
		if isMature(b) {
			if isMature(dst) {
				matureToMature++
			} else {
				matureToYoung++
			}
			continue
		}
		if isMature(dst) {
			youngToMature++
		} else {
			youngToYoung++
		}
	}
	//	println("count=", count)
	//println("countObject returns ", b, cardStart, cardEnd, count)
	return
}

// countObject scans the object starting at b accumulating the number
// of pointers in the object.
// b must point to the beginning of a heap object or an oblet.
// scanobject consults the GC bitmap for the pointer mask and the
// spans for the size of the object.
//
//go:nowritebarrier
func countObjectDebug(b, cardStart, cardEnd uintptr) (count, nilPtr, matureToYoung, matureToMature, youngToYoung, youngToMature, slotToNonHeap uintptr) {
	// Note that arena_used may change concurrently during
	// scanobject and hence scanobject may encounter a pointer to
	// a newly allocated heap object that is *not* in
	// [start,used). It will not mark this object; however, we
	// know that it was just installed by a mutator, which means
	// that mutator will execute a write barrier and take care of
	// marking it. This is even more pronounced on relaxed memory
	// architectures since we access arena_used without barriers
	// or synchronization, but the same logic applies.

	//arena_start := mheap_.arena_start
	//arena_used := mheap_.arena_used

	// Find the bits for b and the size of the object at b.
	//
	// b is either the beginning of an object, in which case this
	// is the size of the object to scan, or it points to an
	// oblet, in which case we compute the size to scan below.

	hbits := heapBitsForAddr(b)
	s := spanOfUnchecked(b)
	n := s.elemsize
	if b >= cardEnd {
		println("424: b >= cardEnd, b=", hex(b), " cardEnd=", hex(cardEnd), "cardStart=", hex(cardStart))
		println("425: s.elemsize=", s.elemsize, "s.base()=", s.base(), "b+elemsize=", hex(b+s.elemsize))
		throw("b>=cardEnd")
	}
	if n == 0 {
		throw("scanobject n == 0")
	}
	// start := cardStart - s.base()
	// end := cardEnd - s.base()
	i := uintptr(0)
	for i = uintptr(0); i < n; i += sys.PtrSize {
		println("At top of loop i=", hex(i), "b=", hex(b), "b+i =", hex(b+i), " cardStart", hex(cardStart),
			"   cardEnd=", hex(cardEnd), "n aka s.elemsize=", hex(s.elemsize))
		// i is offset into object.
		// Find bits for this word.
		if i != 0 {
			// Avoid needless hbits.next() on last iteration.
			hbits = hbits.next()
		}
		// Load bits once. See CL 22712 and issue 16973 for discussion.
		bits := hbits.bits()
		// During checkmarking, 1-word objects store the checkmark
		// in the type bit for the one word. The only one-word objects
		// are pointers, or else they'd be merged with other non-pointer
		// data into larger allocations.
		if i != 1*sys.PtrSize && bits&bitScan == 0 {
			println("break at 446 b+i, no more pointers in obj ", hex(b+i), " < cardStart", hex(cardStart), "s.elemsize=", hex(s.elemsize))
			break // no more pointers in this object
		}
		if b+i >= cardEnd {
			println("break at 451")
			break // there are no more pointers in this card
		}
		if b+i < cardStart {
			println("continue at 455 b+i ", hex(b+i), " < cardStart", hex(cardStart), "s.elemsize=", hex(s.elemsize))
			continue // ignore pointers before the start of the card in question
		}
		if bits&bitPointer == 0 {
			println("continue at 549")
			continue // not a pointer
		}
		println("found pointer at 462 count is now", count)
		count++

		// b is the base of the source object, determine if it is a mature object.
		// Be careful I might be scanning garbage objects that point to freed up objects.....
		// How do I determine if this object is alive or dead. If it is dead I want to skip it.
		// Work here is duplicated in scanblock and above.
		// If you make changes here, make changes there too.
		slot := (*uintptr)(unsafe.Pointer(b + i))
		obj := *slot
		if obj == 0 {
			nilPtr++
			continue
		}

		if !inheap(obj) {
			slotToNonHeap++
			continue
		}
		if isMature(b) {
			dst, _, _, _ := heapBitsForObject(obj, b, i)
			if isMature(dst) {
				matureToMature++
			} else {
				matureToYoung++
			}
			continue
		}
		if true {
			// This should be youngToMumble since we don't know if young is valid so we can't look at the dst.
			youngToMature++
		} else {
			youngToYoung++
		}
	}
	println("519: countObject returns ", hex(b), hex(cardStart), hex(cardEnd), count)
	return
}

//go:nosplit
func cardIndex(addr uintptr) uintptr {
	return (addr - mheap_.arena_start) / _CardBytes
}

/****/

/*******

func createCardTable () {

}

var cardLeftShiftCount = bitsInAddress -
var cardSize 1 << cardRightShiftCount
var cardRightShiftCount = 9
var numCards = 42

func scanCardTableRoots() {
	for i = 0; i < numCards; i++ {
		if cardTable[i] != 0 {
			scanCard(i)
		}
	}
}

var cardTable []byte

func scanCards(uintptr(i)) {
	for i, v := range mheap_.cards {

	if v != 0 {
		if v != 255 {
			throw("card table has invalid mark")
		}
		cardStart := mheap_.arena_start+i*_CardBytes
		scanCard(i)
	}

	}
}

func scanCard(i uintptr) {
	start := mheap_.arena_start+i*_CardBytes
	_ = start
}

func minorGC () {
	// similar to fullGC.
	sync goroutines
	scanStacksandGlobals
	scanCardTableRoots
	deal with specials including finalized objects.
	sync goroutines and declare victory
}

***/
