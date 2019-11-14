// +build staticlockranking

package runtime

import (
	"unsafe"
)

// Mutual exclusion locks.  In the uncontended case,
// as fast as spin locks (just a few user-level instructions),
// but on the contention path they sleep in the kernel.
// A zeroed Mutex is unlocked (no need to initialize each lock).
// Initialization is helpful for static lock ranking, but not required.
type mutex struct {
	// Futex-based impl treats it as uint32 key,
	// while sema-based impl as M* waitm.
	// Used to be a union, but unions break precise GC.
	key uintptr
	// static lock ranking of the lock
	rank int
	pad  int
}

// Use lockRankAcquire() for static lock ranking if lockInit() was not called.
func lock(l *mutex) {
	lockRankAcquire(l, l.rank)
}

func unlock(l *mutex) {
	lockRankRelease(l)
}

func lockInit(l *mutex, rank int) {
	l.rank = rank
}

// The following functions are the entry-points to record lock
// operations.
// All of these are nosplit and switch to the system stack immediately
// to avoid stack growths. Since a stack growth could itself have lock
// operations, this prevents re-entrant calls.

//
// lockRankAcquire is like lock(l), but records the lock class and rank
// for a non-static lock acquisition.
//go:nosplit
func lockRankAcquire(l *mutex, rank int) {
	if l == &debuglock {
		lock2(l)
		return
	}
	if rank == 0 {
		rank = _LleafRank
	}
	gp := getg()
	// Log the new class.
	systemstack(func() {
		i := gp.m.lockIndex
		if i >= 10 {
			throw("overflow")
		}
		gp.m.locksHeld[i].rank = rank
		gp.m.locksHeld[i].l = uintptr(unsafe.Pointer(l))
		gp.m.lockIndex++
		i++

		if i > 1 {
			found := false
			prevRank := gp.m.locksHeld[i-2].rank
			if prevRank > rank {
				// The quick static lock ranking check
				found = false
			} else if rank == _LleafRank {
				if prevRank != _LleafRank {
					found = true
				}
			} else {
				// We could do a static lock ranking check here, but
				// instead we are doing the partial-ordering (arcs) check
				// to make sure the arcs table is valid and up-to-date.
				list := arcs[rank]
				for _, entry := range list {
					if entry == 0 {
						break
					}
					if entry == prevRank {
						found = true
						break
					}
				}
			}
			if !found {
				println(gp.m.procid, " ======")
				for j := 0; j < i; j++ {
					lockName := "LEAF"
					jrank := gp.m.locksHeld[j].rank
					if jrank != _LleafRank {
						lockName = lockNames[jrank]
					}
					println(j, ":", lockName, " ", jrank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
				}
				throw("lock ordering problem")
			}
		}
		lock2(l)
	})
}

//go:nosplit
func lockRankRelease(l *mutex) {
	if l == &debuglock {
		unlock2(l)
		return
	}
	gp := getg()
	systemstack(func() {
		found := false
		for i := gp.m.lockIndex - 1; i >= 0; i-- {
			if gp.m.locksHeld[i].l == uintptr(unsafe.Pointer(l)) {
				found = true
				for j := i; j < gp.m.lockIndex-1; j++ {
					gp.m.locksHeld[j] = gp.m.locksHeld[j+1]
				}
				gp.m.lockIndex--
			}
		}
		if !found {
			println(gp.m.procid, ": unmatching lock", l)
			throw("unmatching lock")
		}
		//println("release", uint64(uintptr(unsafe.Pointer(l))))
		unlock2(l)
	})
}

//go:nosplit
func lockLogMayAcquire(l *mutex, rank int) {
	gp := getg()

	systemstack(func() {
		i := gp.m.lockIndex
		if i > 0 && rank < gp.m.locksHeld[i-1].rank {
			println(gp.m.procid, " ======")
			for j := 0; j < i; j++ {
				println(j, ":", lockNames[gp.m.locksHeld[j].rank], " ", gp.m.locksHeld[j].rank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
			}
			println(i, ":", lockNames[rank], " ", unsafe.Pointer(l))
			throw("lock ordering problem, maybe")
		}
		if i > 0 && rank != _LleafRank {
			found := false
			list := arcs[rank]
			for j := 0; j < 25; j++ {
				if list[j] == 0 {
					break
				}
				if list[j] == gp.m.locksHeld[i-1].rank {
					found = true
					break
				}
			}
			if !found {
				println(gp.m.procid, " ======")
				for j := 0; j < i; j++ {
					println(j, ":", lockNames[gp.m.locksHeld[j].rank], " ", gp.m.locksHeld[j].rank, " ", unsafe.Pointer(gp.m.locksHeld[j].l))
				}
				println(i, ":", lockNames[rank], " ", unsafe.Pointer(l))
				throw("lock ordering problem")
			}
		}
	})
}
