// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loader

// RelocAllocator helps to temporarily parcel out slices of
// loader.Reloc's from a larger reloc slab, helping to reduce
// allocation overhead. Intended to be helpful in cases where clients
// need what amounts to a stack of relocation slices.
//
// Intended usage:
//
//    allocator := loader.MakeRelocAllocator(ldr)
//    ...
//    relocs := ldr.Relocs(sym)
//    rsl := allocator.Alloc(relocs.Count)
//    relocs.ReadAll(rsl)
//    for i := range rsl {
//       <do something with reloc i>
//    }
//    allocator.Release(relocs.Count)
//
// Alloc and Release method calls need to be properly paired; once a
// call to Release is made, the storage underlying the slice returned
// from the alloc will be reclaimed and made available for new
// allocations.
type RelocAllocator struct {
	alloc *relChunk // allocated chunks
	free  *relChunk // free chunks
	ldr   *Loader   // the loader we're working with
	csize int       // chunk/slab size for new allocations
}

// relChunk is a chunk of loader.Relocs; subslices from a chunk
// are parceled out by Alloc and then reclaimed by Release.
type relChunk struct {
	prev *relChunk
	buf  []Reloc
	used int
}

const relocAllocatorChunkSize = 1024

// MakeRelocAllocator creates an initially empty allocator.
func MakeRelocAllocator(l *Loader) *RelocAllocator {
	return &RelocAllocator{ldr: l, csize: relocAllocatorChunkSize}
}

// Alloc method returns a subslice of loader.Reloc's taken from the
// specified RelocAllocator. It allocates a new chunk if need be, and
// does book-keeping to record what's been used so far.
func (ra *RelocAllocator) Alloc(count int) []Reloc {
	if count == 0 {
		return []Reloc{}
	}
	if count < 0 {
		panic("bad reloc count")
	}

	// Check to see if the slice in the existing chunk has enough
	// capacity for this new request.
	rc := ra.alloc
	avail := 0
	if rc != nil {
		avail = len(rc.buf) - rc.used
	}

	if count > avail {
		// not enough space.
		reqSize := ra.csize
		if count > reqSize {
			reqSize = count
		}
		var chunk *relChunk

		// If there is a sufficiently large chunk at top of free list, use it.
		if ra.free != nil && reqSize <= len(ra.free.buf) {
			chunk = ra.free
			ra.free = chunk.prev
			chunk.prev = nil
		} else {
			// ... otherwise allocate a new chunk.
			chunk = &relChunk{
				prev: rc,
				buf:  make([]Reloc, reqSize),
			}
		}
		chunk.prev = ra.alloc
		ra.alloc = chunk
		rc = chunk
	}

	// Carve off and return new slice from the slab.
	rval := rc.buf[rc.used : rc.used+count : count+rc.used]
	rc.used += count
	return rval[:0]
}

// Release returns the sub-slice of loader.Reloc's requested by
// a previous call to RelocAllocator.Alloc back to the allocator.
func (ra *RelocAllocator) Release(count int) {
	if count == 0 {
		return
	}
	rc := ra.alloc

	if rc == nil || count > rc.used {
		panic("should never happen")
	}
	rc.used -= count

	if rc.used == 0 {

		// Remove the top-of-stack chunk and place it onto the free list
		freed := rc
		rc = rc.prev
		ra.alloc = rc

		freed.prev = ra.free
		freed.used = 0
		ra.free = freed
	}
}

// SanityCheck is a helper to verify to see whether a loader
// client invoked Alloc/Release correctly (one release for each
// allocation).
func (ra *RelocAllocator) SanityCheck() {
	if ra.alloc != nil {
		panic("reloc allocator insanity")
	}
}
