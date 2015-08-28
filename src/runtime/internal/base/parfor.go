// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Parallel for algorithm.

package base

// A parfor holds state for the parallel for operation.
type Parfor struct {
	Body   func(*Parfor, uint32) // executed for each element
	Done   uint32                // number of idle threads
	Nthr   uint32                // total number of threads
	Thrseq uint32                // thread id sequencer
	Cnt    uint32                // iteration space [0, cnt)
	Wait   bool                  // if true, wait while all threads finish processing,
	// otherwise parfor may return while other threads are still working

	Thr []Parforthread // thread descriptors

	// stats
	Nsteal     uint64
	Nstealcnt  uint64
	Nprocyield uint64
	Nosyield   uint64
	Nsleep     uint64
}

// A parforthread holds state for a single thread in the parallel for.
type Parforthread struct {
	// the thread's iteration space [32lsb, 32msb)
	Pos uint64
	// stats
	nsteal     uint64
	nstealcnt  uint64
	nprocyield uint64
	nosyield   uint64
	nsleep     uint64
	pad        [CacheLineSize]byte
}

func Parfordo(desc *Parfor) {
	// Obtain 0-based thread index.
	tid := Xadd(&desc.Thrseq, 1) - 1
	if tid >= desc.Nthr {
		print("tid=", tid, " nthr=", desc.Nthr, "\n")
		Throw("parfor: invalid tid")
	}

	// If single-threaded, just execute the for serially.
	body := desc.Body
	if desc.Nthr == 1 {
		for i := uint32(0); i < desc.Cnt; i++ {
			body(desc, i)
		}
		return
	}

	me := &desc.Thr[tid]
	mypos := &me.Pos
	for {
		for {
			// While there is local work,
			// bump low index and execute the iteration.
			pos := Xadd64(mypos, 1)
			begin := uint32(pos) - 1
			end := uint32(pos >> 32)
			if begin < end {
				body(desc, begin)
				continue
			}
			break
		}

		// Out of work, need to steal something.
		idle := false
		for try := uint32(0); ; try++ {
			// If we don't see any work for long enough,
			// increment the done counter...
			if try > desc.Nthr*4 && !idle {
				idle = true
				Xadd(&desc.Done, 1)
			}

			// ...if all threads have incremented the counter,
			// we are done.
			extra := uint32(0)
			if !idle {
				extra = 1
			}
			if desc.Done+extra == desc.Nthr {
				if !idle {
					Xadd(&desc.Done, 1)
				}
				goto exit
			}

			// Choose a random victim for stealing.
			var begin, end uint32
			victim := Fastrand1() % (desc.Nthr - 1)
			if victim >= tid {
				victim++
			}
			victimpos := &desc.Thr[victim].Pos
			for {
				// See if it has any work.
				pos := Atomicload64(victimpos)
				begin = uint32(pos)
				end = uint32(pos >> 32)
				if begin+1 >= end {
					end = 0
					begin = end
					break
				}
				if idle {
					Xadd(&desc.Done, -1)
					idle = false
				}
				begin2 := begin + (end-begin)/2
				newpos := uint64(begin) | uint64(begin2)<<32
				if Cas64(victimpos, pos, newpos) {
					begin = begin2
					break
				}
			}
			if begin < end {
				// Has successfully stolen some work.
				if idle {
					Throw("parfor: should not be idle")
				}
				Atomicstore64(mypos, uint64(begin)|uint64(end)<<32)
				me.nsteal++
				me.nstealcnt += uint64(end) - uint64(begin)
				break
			}

			// Backoff.
			if try < desc.Nthr {
				// nothing
			} else if try < 4*desc.Nthr {
				me.nprocyield++
				Procyield(20)
			} else if !desc.Wait {
				// If a caller asked not to wait for the others, exit now
				// (assume that most work is already done at this point).
				if !idle {
					Xadd(&desc.Done, 1)
				}
				goto exit
			} else if try < 6*desc.Nthr {
				me.nosyield++
				Osyield()
			} else {
				me.nsleep++
				Usleep(1)
			}
		}
	}

exit:
	Xadd64(&desc.Nsteal, int64(me.nsteal))
	Xadd64(&desc.Nstealcnt, int64(me.nstealcnt))
	Xadd64(&desc.Nprocyield, int64(me.nprocyield))
	Xadd64(&desc.Nosyield, int64(me.nosyield))
	Xadd64(&desc.Nsleep, int64(me.nsleep))
	me.nsteal = 0
	me.nstealcnt = 0
	me.nprocyield = 0
	me.nosyield = 0
	me.nsleep = 0
}
