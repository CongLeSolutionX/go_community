// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/atomic"
	"runtime/internal/sys"
)

const (
	// For the time histogram type, we use an HDR histogram
	// with a maximum error of 1/timeHistNumSubBuckets*100%.
	//
	// timeHistNumBuckets defines the range of the histogram.
	// 48 buckets corresponds to a maximum supported duration
	// of approx. 3 days, more than enough to handle any
	// useful duration in the runtime.
	timeHistSubBucketBits = 4
	timeHistNumSubBuckets = 1 << timeHistSubBucketBits
	timeHistNumBuckets    = 48 - (timeHistSubBucketBits - 1)
	timeHistTotalBuckets  = timeHistNumBuckets*timeHistNumSubBuckets + 1
)

// timeHistogram represents a distribution of durations in
// nanoseconds.
//
// The accuracy and range of the histogram is defined by the
// timeHistSubBucketBits and timeHistNumBuckets constants.
//
// It is an HDR histogram with exponentially-distributed
// buckets and linearly distributed sub-buckets.
//
// Counts in the histogram are updated atomically, so it is safe
// for concurrent use. It is also safe to read all the values
// atomically.
type timeHistogram struct {
	counts   [timeHistNumBuckets * timeHistNumSubBuckets]uint64
	overflow uint64
}

// record adds the given duration to the distribution.
//
// Although the duration is an int64 to facilitate ease-of-use
// with e.g. nanotime, the duration must be non-negative.
//
// Disallow preemptions and stack growths because this function
// may run in sensitive locations.
//go:nosplit
func (h *timeHistogram) record(duration int64) {
	if duration < 0 {
		throw("timeHistogram encountered negative duration")
		return
	}
	// The index of the exponential bucket is just the index
	// of the highest set bit adjusted for how many bits we
	// use for the subbucket. Note that it's subBucketsBits-1
	// because we use the 0th bucket to hold values < timeHistNumSubBuckets.
	var bucket, subBucket uint
	if duration >= timeHistNumSubBuckets {
		// At this point, we know the duration value will always be
		// at least timeHistSubBucketsBits long.
		bucket = uint(sys.Len64(uint64(duration))) - (timeHistSubBucketBits - 1)
		if bucket >= uint(len(h.counts)) {
			// The bucket index we got is larger than what we support, so
			// add into the special overflow bucket.
			atomic.Xadd64(&h.overflow, 1)
			return
		}
		// The linear subbucket index is just the timeHistSubBucketsBits
		// bits after the top bit. To extract that value, shift down
		// the duration such that we leave the top bit and the next bits
		// intact, then extract the index.
		subBucket = uint((duration >> (bucket - 1)) % timeHistNumSubBuckets)
	} else {
		subBucket = uint(duration)
	}
	atomic.Xadd64(&h.counts[bucket*timeHistNumSubBuckets+subBucket], 1)
}
