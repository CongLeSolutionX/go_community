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
	// 45 buckets corresponds to a maximum supported duration
	// of approx. 3 days. This is more than enough to handle
	// any useful duration in the runtime.
	//
	// As an example, consider 45 buckets with 16 sub-buckets.
	//
	// Buckets are counted and indexed like so:
	//
	//           Bucket #1 contains values with a zero 5th bit
	//    00000          and no higher 1 bits
	//    ^
	//           Bucket #2 contains values with a set 5th bit
	//    10000          and no higher 1 bits
	//    ^
	//        _  Bucket #3 contains values with a set 6th bit
	//   100000          and the 1st bit is no longer used.
	//   ^
	//        _  Bucket #4 contains values with a set 7th bit, and the 0th bit is no
	//  1000000          and the 2nd bit and lower is no longer used.
	//  ^
	//
	// Following this pattern, bucket #45 will have the 48th bit set. We
	// don't have any buckets for higher values, so the highest sub-bucket
	// will contain values of 2^48-1 nanoseconds or approx. 3 days.
	timeHistSubBucketBits = 4
	timeHistNumSubBuckets = 1 << timeHistSubBucketBits
	timeHistNumBuckets    = 45
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
	}
	// The index of the exponential bucket is just the index
	// of the highest set bit adjusted for how many bits we
	// use for the subbucket. Note that it's timeHistSubBucketsBits-1
	// because we use the 0th bucket to hold values < timeHistNumSubBuckets.
	var bucket, subBucket uint
	if duration >= timeHistNumSubBuckets {
		// At this point, we know the duration value will always be
		// at least timeHistSubBucketsBits long.
		bucket = uint(sys.Len64(uint64(duration))) - timeHistSubBucketBits
		if bucket*timeHistNumSubBuckets >= uint(len(h.counts)) {
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

// timeHistogramMetricsBuckets generates a slice of boundaries for
// the timeHistogram. These boundaries are represented in seconds,
// not nanoseconds like the timeHistogram represents durations.
func timeHistogramMetricsBuckets() []float64 {
	b := make([]float64, timeHistTotalBuckets-1)
	for i := 0; i < timeHistNumBuckets; i++ {
		// Throughout this code as we refer to "buckets" we mean
		// buckets as defined by this file (exponentially spaced
		// with sub-buckets) NOT the way the metrics package defines
		// it.
		bucketMin := uint64(0)
		// The (inclusive) minimum for the first bucket is 0.
		if i > 0 {
			// The minimum for the second bucket will be
			// 1 << timeHistSubBucketBits, indicating that all
			// sub-buckets are represented by the next timeHistSubBucketBits
			// bits.
			// Thereafter, we shift up by 1 each time, so we can represent
			// this pattern as (i-1)+timeHistSubBucketBits.
			bucketMin = uint64(1) << uint(i-1+timeHistSubBucketBits)
		}
		// subBucketShift is the amount that we need to shift the sub-bucket
		// index to combine it with the bucketMin.
		subBucketShift := uint(0)
		if i > 1 {
			// The first two buckets are exact with respect to integers,
			// so we'll never have to shift the sub-bucket index. Thereafter,
			// we shift up by 1 with each subsequent bucket.
			subBucketShift = uint(i - 2)
		}
		for j := 0; j < timeHistNumSubBuckets; j++ {
			// j is the sub-bucket index. By shifting the index into position to
			// combine with the bucket minimum, we obtain the minimum value for that
			// sub-bucket.
			subBucketMin := bucketMin + (uint64(j) << subBucketShift)

			// Convert the subBucketMin which is in nanoseconds to a float64 seconds value.
			// These values will all be exactly representable by a float64.
			b[i*timeHistNumSubBuckets+j] = float64(subBucketMin) / 1e9
		}
	}
	return b
}
