package runtime

import "runtime/internal/sys"

func matchEmptyOrDeleted(tophash [bucketCnt]uint8) bitmask64 {
	// The high bit is set for both empty slot and deleted slot.
	tophashs := littleEndianBytesToUint64(tophash)
	return bitmask64(emptyOrDeletedMask & tophashs)
}

func matchEmpty(tophash [bucketCnt]uint8) bitmask64 {
	// Same as matchTopHash(tophash, emptySlot), but faster.
	//
	// The high bit is set for both empty slot and deleted slot.
	// (tophashs & emptyOrDeletedMask) get all empty or deleted slots.
	// (tophashs << 1) clears the high bit for deletedSlot.
	// ANDing them we can get all the empty slots.
	tophashs := littleEndianBytesToUint64(tophash)
	return bitmask64((tophashs << 1) & tophashs & emptyOrDeletedMask)
}

// matchTopHash returns a bitmask indicating all bytes in the tophash which *may*
// have the given value.
//
// For the technique, see:
// http://graphics.stanford.edu/~seander/bithacks.html##ValueInWord
// (Determine if a word has a byte equal to n).
//
// Caveat: there are false positives but:
// - they only occur if there is a real match
// - they will be handled gracefully by subsequent checks in code
func matchTopHash(tophash [bucketCnt]uint8, top uint8) bitmask64 {
	tophashs := littleEndianBytesToUint64(tophash)
	cmp := tophashs ^ (uint64(0x0101_0101_0101_0101) * uint64(top))
	return bitmask64((cmp - 0x0101_0101_0101_0101) & ^cmp & 0x8080_8080_8080_8080)
}

func matchFull(tophash [bucketCnt]uint8) bitmask64 {
	// If a slot is neither empty nor deleted, then it must be FUll.
	tophashs := littleEndianBytesToUint64(tophash)
	return bitmask64(emptyOrDeletedMask & ^tophashs)
}

func prepareSameSizeGrow(tophash *[bucketCnt]uint8) {
	// For all slots:
	// Mark DELETED as EMPTY
	// Mark FULL as DELETED
	tophashs := littleEndianBytesToUint64(*tophash)
	full := ^tophashs & emptyOrDeletedMask
	full = ^full + (full >> 7)
	*tophash = littleEndianUint64ToBytes(full)
}

// bitmask64 contains the result of a `match` operation on a `bucket`.
//
// The bitmask is arranged so that low-order bits represent lower memory
// addresses for group match results.
//
// For generic version, the bits in the bitmask is sparsely packed, so
// that there is only one bit-per-byte used (the high bit, 7).
type bitmask64 uint64

// NextMatch returns the lowest bit's byte index(low-order).
// If there is no match, it returns bucketCnt.
func (b *bitmask64) NextMatch() uintptr {
	return uintptr(sys.TrailingZeros64(uint64(*b)) / bucketCnt)
}

// RemoveNextMatch remove the lowest bit in the bitmask.
func (b *bitmask64) RemoveNextMatch() {
	*b = *b & (*b - 1)
}

func littleEndianBytesToUint64(v [8]uint8) uint64 {
	return uint64(v[0]) | uint64(v[1])<<8 | uint64(v[2])<<16 | uint64(v[3])<<24 | uint64(v[4])<<32 | uint64(v[5])<<40 | uint64(v[6])<<48 | uint64(v[7])<<56
}

func littleEndianUint64ToBytes(v uint64) [8]uint8 {
	return [8]uint8{uint8(v), uint8(v >> 8), uint8(v >> 16), uint8(v >> 24), uint8(v >> 32), uint8(v >> 40), uint8(v >> 48), uint8(v >> 56)}
}
