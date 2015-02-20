// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import "unsafe"

type schedAffinity int

func affinityInitOS() *schedDomain {
	nodes := affinityGetNumaNodes()
	for ni, cores := range nodes {
		print("node ", ni, ":")
		for _, c := range cores {
			print(" ", c)
		}
		print("\n")
	}

	procs, cores := affinityGetCpuinfo()
	for i, p := range procs {
		print("core ", i, ": proc=", p, " core=", cores[i], "\n")
	}
	return nil
	/*
		var buf [16]uintptr
		r := sched_getaffinity(0, unsafe.Sizeof(buf), &buf[0])
		n := int32(0)
		for _, v := range buf[:r/ptrSize] {
			for i := 0; i < 64; i++ {
				n += int32(v & 1)
				v >>= 1
			}
		}
		if n == 0 {
			n = 1
		}
		return n
	*/
}

// affinityGetNumaNodes parses /sys/devices/system/node/node*/cpulist files
// to obtain number of numa nodes as well as numbers of cores on nodes.
// If anything goes wrong, the function returns nil.
// File contents look like "0-7,16-23".
func affinityGetNumaNodes() [][]int {
	var nodes [][]int
	//!!! use noescape to prevent escaping
	tmp := make([]byte, 128)
	for ni := 0; ; ni++ {
		fn := "/sys/devices/system/node/node" + string(itoa(tmp, uint64(ni))) + "/cpulist"
		fd := open(*(**byte)(unsafe.Pointer(&fn)), _O_RDONLY|_O_CLOEXEC, 0)
		if fd == -_ENOENT {
			break
		}
		if fd < 0 {
			return nil
		}
		n := read(fd, unsafe.Pointer(&tmp[0]), int32(len(tmp)))
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

func affinityGetCpuinfo() (procs []int, cores []int) {
	fn := "/proc/cpuinfo"
	fd := open(*(**byte)(unsafe.Pointer(&fn)), _O_RDONLY|_O_CLOEXEC, 0)
	if fd < 0 {
		return
	}
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
			return nil, nil
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
					procs = append(procs, processor)
					cores = append(cores, coreid)
				}
				processor = -1
				coreid = -1
			} else if s := "physical id"; memeq(unsafe.Pointer(&data[0]), *(*unsafe.Pointer)(unsafe.Pointer(&s)), uintptr(len(s))) {
				data = data[len(s):]
				for data[0] == ' ' || data[0] == '\t' || data[0] == ':' {
					data = data[1:]
				}
				processor = atoi(string(data))
				print("found phys id: ", processor, "\n")
			} else if s := "core id"; memeq(unsafe.Pointer(&data[0]), *(*unsafe.Pointer)(unsafe.Pointer(&s)), uintptr(len(s))) {
				data = data[len(s):]
				for data[0] == ' ' || data[0] == '\t' || data[0] == ':' {
					data = data[1:]
				}
				coreid = atoi(string(data))
				print("found core id: ", coreid, "\n")
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
		procs = append(procs, processor)
		cores = append(cores, coreid)
	}
	close(fd)
	return
}

func findChar(b []byte, c byte) int {
	for i, v := range b {
		if v == c {
			return i
		}
	}
	return -1
}
