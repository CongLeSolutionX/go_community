// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

type schedAffinity int

func (a schedAffinity) print() {
	print(hex(uintptr(a)))
}

func (a schedAffinity) add(a1 schedAffinity) schedAffinity {
	return a | a1
}

/*
func setProcessAffinity(a schedAffinity) {
        nr := uintptr(0) // NR_SCHED_SETAFFINITY
        switch runtime.GOARCH {
        case "amd64":
                nr = 203
        case "386":
                nr = 241
        default:
		gothrow("not implemented")
        }
        runtime.LockOSThread()
        defer runtime.UnlockOSThread()
        _, _, errno := syscall.Syscall(nr, uintptr(syscall.Getpid()), uintptr(unsafe.Sizeof(v)), uintptr(unsafe.Pointer(&v)))
}
*/

type affThreadDesc struct {
	proc int
	core int
	node int
}

func sched_getaffinity(what int, size uintptr, buf *uintptr) uintptr {
	*buf = ^uintptr(0)
	if false {
		*buf ^= 1 << 2
		*buf ^= 1 << 3
		*buf ^= 1 << 5
		*buf ^= 1 << 21
		*buf ^= 1 << 22
		*buf ^= 1 << 23
	}
	return 16
}

func affinityInitOS() *schedDomain {
	threads := affinityGetCpuinfo()

	nodes := affinityGetNumaNodes()
	for ni, cores := range nodes {
		for _, c := range cores {
			if c >= 0 && c < len(threads) {
				threads[c].node = ni
			}
		}
	}

	distance := affinityGetNumaDistance()

	world := &schedDomain{kind: schedDomainWorld}

	nn := make([]*schedDomain, len(nodes))
	for ni := range nodes {
		n := &schedDomain{kind: schedDomainNode, distance: distance[ni]}
		nn[ni] = n
		world.sub = append(world.sub, n)
	}

	var aff [16]uintptr
	sched_getaffinity(0, unsafe.Sizeof(aff), &aff[0])

	procToNode := make([][]int, len(nodes))

	for ti, t := range threads {
		//!!! check bounds
		if aff[ti/(ptrSize*8)]&(1<<uint(ti%(ptrSize*8))) == 0 {
			continue
		}

		n := world.sub[t.node]
		var proc *schedDomain
		for i, pi := range procToNode[t.node] {
			if pi == t.proc {
				proc = n.sub[i]
				break
			}
		}
		if proc == nil {
			proc = &schedDomain{kind: schedDomainProcessor}
			procToNode[t.node] = append(procToNode[t.node], t.proc)
			n.sub = append(n.sub, proc)
		}

		for len(proc.sub) <= t.core {
			proc.sub = append(proc.sub, &schedDomain{kind: schedDomainCore})
		}

		core := proc.sub[t.core]
		core.sub = append(core.sub, &schedDomain{
			kind:     schedDomainThread,
			affinity: 1 << uint(ti),
		})
	}

	return world
}

// affinityGetNumaNodes parses /sys/devices/system/node/node*/cpulist files
// to obtain number of numa nodes as well as numbers of cores on nodes.
// If anything goes wrong, the function returns nil.
// File contents look like "0-7,16-23".
func affinityGetNumaNodes() [][]int {
	var nodes [][]int
	//!!! use noescape to prevent escaping
	var tmp [128]byte
	for ni := 0; ; ni++ {
		//fn := "/sys/devices/system/node/node" + string(itoa(tmp, uint64(ni))) + "/cpulist"
		fn := "/Users/dvyukov/src/go10/misc/node" + string(itoa(tmp[:], uint64(ni))) + "/cpulist"
		fd := open(*(**byte)(unsafe.Pointer(&fn)), _O_RDONLY, 0)
		if fd == -_ENOENT {
			break
		}
		if fd < 0 {
			return nil
		}
		n := read(fd, noescape(unsafe.Pointer(&tmp[0])), int32(len(tmp)))
		close(fd)
		if n <= 0 || n >= int32(len(tmp)) {
			return nil
		}
		var cores []int
		core := -1
		core1 := -1
		flush := func() bool {
			if core == -1 || core1 > core {
				return false
			}
			if core1 == -1 {
				core1 = core
			}
			for i := core1; i <= core; i++ {
				cores = append(cores, i)
			}
			core = -1
			core1 = -1
			return true
		}
		for i := 0; i < int(n); i++ {
			switch c := tmp[i]; {
			case c == ',' || c == '\n':
				if !flush() {
					return nil
				}
			case c == '-':
				if core1 != -1 {
					return nil
				}
				core1 = core
				core = -1
			case c >= '0' && c <= '9':
				if core == -1 {
					core = int(c) - '0'
				} else {
					core = core*10 + int(c) - '0'
				}
			default:
				return nil
			}
		}
		if core != -1 {
			if !flush() {
				return nil
			}
		}
		nodes = append(nodes, cores)
	}
	return nodes
}

func affinityGetNumaDistance() [][]int {
	var distance [][]int
	//!!! use noescape to prevent escaping
	var tmp [128]byte
	for ni := 0; ; ni++ {
		//fn := "/sys/devices/system/node/node" + string(itoa(tmp, uint64(ni))) + "/distance"
		fn := "/Users/dvyukov/src/go10/misc/node" + string(itoa(tmp[:], uint64(ni))) + "/distance"
		fd := open(*(**byte)(unsafe.Pointer(&fn)), _O_RDONLY, 0)
		if fd == -_ENOENT {
			break
		}
		if fd < 0 {
			return nil
		}
		n := read(fd, noescape(unsafe.Pointer(&tmp[0])), int32(len(tmp)))
		close(fd)
		if n <= 0 || n >= int32(len(tmp)) {
			return nil
		}
		var dist []int
		for data := tmp[:n]; len(data) > 0; {
			dist = append(dist, atoi(slicebytetostringtmp(data)))
			for data[0] >= '0' && data[0] <= '9' {
				data = data[1:]
			}
			for len(data) > 0 && (data[0] < '0' || data[0] > '9') {
				data = data[1:]
			}
		}
		distance = append(distance, dist)
	}
	return distance
}

func affinityGetCpuinfo() []affThreadDesc {
	fn := "/Users/dvyukov/src/go10/misc/cpuinfo"
	fd := open(*(**byte)(unsafe.Pointer(&fn)), _O_RDONLY, 0)
	if fd < 0 {
		return nil
	}
	var threads []affThreadDesc
	var buf [1024]byte
	offset := 0
	processor := -1
	coreid := -1
	for {
		n := read(fd, noescape(unsafe.Pointer(&buf[offset])), int32(len(buf)-offset))
		if n == 0 {
			break
		}
		if n < 0 {
			close(fd)
			return nil
		}
		data := buf[:offset+int(n)]
		for {
			nl := findChar(data, '\n')
			if nl < 0 {
				break
			}
			//print("LINE: '", slicebytetostringtmp(data[:16]), "'\n")
			if s := "processor"; memeq(unsafe.Pointer(&data[0]), *(*unsafe.Pointer)(unsafe.Pointer(&s)), uintptr(len(s))) {
				if processor != -1 && coreid != -1 {
					threads = append(threads, affThreadDesc{processor, coreid, 0})
				}
				processor = -1
				coreid = -1
			} else if s := "physical id"; memeq(unsafe.Pointer(&data[0]), *(*unsafe.Pointer)(unsafe.Pointer(&s)), uintptr(len(s))) {
				data = data[len(s):]
				for data[0] == ' ' || data[0] == '\t' || data[0] == ':' {
					data = data[1:]
				}
				processor = atoi(string(data))
				//print("found phys id: ", processor, "\n")
			} else if s := "core id"; memeq(unsafe.Pointer(&data[0]), *(*unsafe.Pointer)(unsafe.Pointer(&s)), uintptr(len(s))) {
				data = data[len(s):]
				for data[0] == ' ' || data[0] == '\t' || data[0] == ':' {
					data = data[1:]
				}
				coreid = atoi(string(data))
				//print("found core id: ", coreid, "\n")
			}
			for data[0] != '\n' {
				data = data[1:]
			}
			data = data[1:]
		}
		copy(buf[:], data)
		offset = len(data)
	}
	if processor != -1 && coreid != -1 {
		threads = append(threads, affThreadDesc{processor, coreid, 0})
	}
	close(fd)
	return threads
}

func findChar(b []byte, c byte) int {
	for i, v := range b {
		if v == c {
			return i
		}
	}
	return -1
}
