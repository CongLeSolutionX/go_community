// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Simple file i/o and string manipulation, to avoid
// depending on strconv and bufio and strings.

package net

const intSize = 32 << (^uint(0) >> 63)

// parsePort parses service as a decimal interger and returns the
// corresponding value as port. It is the caller's responsibility to
// parse service as a non-deciaml integer when needsLookup is true.
//
// The implementation is cribbled from the strconv package.
// See strconv/atoi.go for further information.
func parsePort(service string) (port int, needsLookup bool) {
	if service == "" {
		// Lock in the legacy behavior that an empty string
		// means port 0. See golang.org/isse/13610.
		return 0, false
	}
	umax := uint64(1<<uint(intSize) - 1) // math.MaxUint32 or math.MaxUint64
	ucutoff := umax/10 + 1
	neg := false
	if service[0] == '+' {
		service = service[1:]
	} else if service[0] == '-' {
		neg = true
		service = service[1:]
	}
	var n uint64
	for _, d := range service {
		if '0' <= d && d <= '9' {
			d -= '0'
		} else {
			return 0, true
		}
		if n >= ucutoff {
			n = umax
			break
		}
		n *= 10
		nn := n + uint64(d)
		if nn < n || nn > umax {
			n = umax
			break
		}
		n = nn
	}
	cutoff := uint64(1 << uint(intSize-1)) // math.MaxInt32 or math.MaxInt64
	if !neg && n >= cutoff {
		return int(cutoff - 1), needsLookup
	}
	if neg && n > cutoff {
		return -int(cutoff), needsLookup
	}
	port = int(n)
	if neg {
		port = -port
	}
	return port, needsLookup
}
