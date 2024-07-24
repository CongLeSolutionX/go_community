// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package secret_test

// Note: it is probably necessary to run
//   ulimit -c unlimited
// before running this test.

import (
	"bytes"
	"debug/elf"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/secret"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

// A secret value that should never appear in a core dump,
// except for this global variable itself.
// The first byte of the secret is variable, to track
// different instances of it.
// TODO: this is little-endian specific.
var secretStore = [8]byte{
	0x00,
	0x81,
	0xa0,
	0xc6,
	0xb3,
	0x01,
	0x66,
	0x53,
}

func TestCrash(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skip("unsupported arch")
	}
	if runtime.GOOS != "linux" {
		t.Skip("unsupported os")
	}
	if os.Getenv("GOTRACEBACK") == "" {
		outer(t)
		return
	}
	inner()
}

type violation struct {
	id  byte   // secret ID
	off uint64 // offset in core dump
}

func outer(t *testing.T) {
	os.Remove("./core")

	tmpDir := t.TempDir()

	// Fork ourselves with GOTRACEBACK=crash set.
	// This runs inner() in a subprocess.
	cmd := exec.Command(os.Args[0], "-test.run=Crash")
	cmd.Env = append(os.Environ(), "GOTRACEBACK=crash")
	cmd.Env = append(cmd.Env, "GODEBUG=asyncpreemptoff=1") // TODO: remove!
	cmd.Dir = tmpDir                                       // put core file in tempdir
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	if err := cmd.Start(); err != nil {
		t.Fatalf("couldn't fork copy of test: %v", err)
	}

	// Wait until the subprocess is done and dumped core.
	cmd.Wait()

	// For debugging.
	t.Logf("\n\n\n--- START SUBPROCESS ---\n\n\n" + stdout.String() + "\n\n--- END SUBPROCESS ---\n\n\n")

	// Load core into memory.
	core, err := os.Open(filepath.Join(tmpDir, "core"))
	if err != nil {
		t.Fatalf("core file not found. Maybe you need to run \"ulimit -c unlimited\"? %v", err)
	}
	b, err := io.ReadAll(core)
	if err != nil {
		t.Fatalf("can't read core file: %v", err)
	}

	// Open elf view onto core file.
	coreElf, err := elf.NewFile(core)
	if err != nil {
		t.Fatalf("can't parse core file: %v", err)
	}

	// Find where in core to expect the one allowed secret.
	// TODO: will this work with ASLR?
	secretStoreAddr := uint64(uintptr(unsafe.Pointer(&secretStore[0])))
	var secretStoreOffset uint64 // Core file offset of global secret variable.
	for _, p := range coreElf.Progs {
		if secretStoreAddr >= p.Vaddr && secretStoreAddr < p.Vaddr+p.Memsz {

			secretStoreOffset = p.Off + (secretStoreAddr - p.Vaddr)
		}
	}

	// Look for any other places that have the secret.
	var violations []violation // core file offsets where we found a secret
	foundAllowed := false
	i := 0
	for {
		j := bytes.Index(b[i:], secretStore[1:])
		if j < 0 {
			break
		}
		j--
		i += j
		if uint64(i) == secretStoreOffset {
			foundAllowed = true
		} else {
			t.Errorf("secret %d found at offset %x in core file", b[i], i)
			violations = append(violations, violation{
				id:  b[i],
				off: uint64(i),
			})
		}
		i += len(secretStore)
	}
	if !foundAllowed {
		t.Errorf("didn't find allowed secret")
	}

	// Get more specific data about where in the core we found the secrets.
	regions := elfRegions(t, core, coreElf)
	for _, r := range regions {
		for _, v := range violations {
			if v.off >= r.min && v.off < r.max {
				var addr string
				if r.addrMin != 0 {
					addr = fmt.Sprintf(" addr=%x", r.addrMin+(v.off-r.min))
				}
				t.Logf("additional info: secret %d at offset %x in %s%s", v.id, v.off-r.min, r.name, addr)
			}
		}
	}
}

// inner does 5 seconds of work, using the secret value
// inside secret.Do in a bunch of ways.
func inner() {
	stop := make(chan bool)
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(1 * time.Second)
			for {
				select {
				case <-stop:
					wg.Done()
					return
				default:
					secret.Do(func() {
						// Copy key into a variable-sized heap allocation.
						// This both puts secrets in heap objects,
						// and more generally just causes allocation,
						// which forces garbage collection, which
						// requires interrupts and the like.
						s := bytes.Repeat(secretStore[:], 1+i*2)
						// Also spam the secret across all registers.
						secret.UseSecret(s)
					})
				}
			}
		}()
	}

	// Send some allocations over a channel. This does 2 things:
	// 1) forces some GCs to happen
	// 2) causes more scheduling noise (Gs moving between Ms, etc.)
	c := make(chan []byte)
	wg.Add(2)
	go func() {
		for {
			select {
			case <-stop:
				wg.Done()
				return
			case c <- make([]byte, 256):
			}
		}
	}()
	go func() {
		for {
			select {
			case <-stop:
				wg.Done()
				return
			case <-c:
			}
		}
	}()

	time.Sleep(5 * time.Second)
	close(stop)
	wg.Wait()
	runtime.GC() // GC should clear any secret heap objects.

	// Panic to dump core.
	// TODO: does panic in test work?
	panic("terminate")
}

// TODO: linux only?
const NT_X86_XSTATE = 0x202

// hasViolation reports whether the address range [min:max] in the
// inferior has any addresses of secrets we found.
func hasViolation(violations []uint64, min, max uint64) bool {
	for _, v := range violations {
		if v >= min && v < max {
			return true
		}
	}
	return false
}

type elfRegion struct {
	name             string
	min, max         uint64 // core file offset range
	addrMin, addrMax uint64 // inferior address range (or 0,0 if no address, like registers)
}

func elfRegions(t *testing.T, core *os.File, coreElf *elf.File) []elfRegion {
	var regions []elfRegion
	for _, p := range coreElf.Progs {
		regions = append(regions, elfRegion{
			name:    fmt.Sprintf("%s[%s]", p.Type, p.Flags),
			min:     p.Off,
			max:     p.Off + min(p.Filesz, p.Memsz),
			addrMin: p.Vaddr,
			addrMax: p.Vaddr + min(p.Filesz, p.Memsz),
		})
	}

	regions = append(regions, threadRegions(t, core, coreElf)...)

	for i, r1 := range regions {
		for j, r2 := range regions {
			if i == j {
				continue
			}
			if r1.max <= r2.min || r2.max <= r1.min {
				continue
			}
			t.Fatalf("overlapping regions %v %v", r1, r2)
		}
	}

	return regions
}

func threadRegions(t *testing.T, core *os.File, coreElf *elf.File) []elfRegion {
	var regions []elfRegion

	for _, prog := range coreElf.Progs {
		if prog.Type != elf.PT_NOTE {
			continue
		}

		b := make([]byte, prog.Filesz)
		_, err := core.ReadAt(b, int64(prog.Off))
		if err != nil {
			t.Fatalf("can't read core file %v", err)
		}
		prefix := "unk"
		b0 := b
		for len(b) > 0 {
			namesz := coreElf.ByteOrder.Uint32(b)
			b = b[4:]
			descsz := coreElf.ByteOrder.Uint32(b)
			b = b[4:]
			typ := elf.NType(coreElf.ByteOrder.Uint32(b))
			b = b[4:]
			name := string(b[:namesz-1])
			b = b[(namesz+3)/4*4:]
			off := prog.Off + uint64(len(b0)-len(b))
			desc := b[:descsz]
			b = b[(descsz+3)/4*4:]

			if name != "CORE" && name != "LINUX" {
				continue
			}
			end := off + uint64(len(desc))
			// Note: amd64 specific
			// See /usr/include/x86_64-linux-gnu/bits/sigcontext.h
			//
			//   struct _fpstate
			switch typ {
			case elf.NT_PRSTATUS:
				pid := coreElf.ByteOrder.Uint32(desc[32:36])
				prefix = fmt.Sprintf("thread%d: ", pid)
				regions = append(regions, elfRegion{
					name: prefix + "prstatus header",
					min:  off,
					max:  off + 112,
				})
				off += 112
				greg := []string{
					"r15",
					"r14",
					"r13",
					"r12",
					"rbp",
					"rbx",
					"r11",
					"r10",
					"r9",
					"r8",
					"rax",
					"rcx",
					"rdx",
					"rsi",
					"rdi",
					"orig_rax",
					"rip",
					"cs",
					"eflags",
					"rsp",
					"ss",
					"fs_base",
					"gs_base",
					"ds",
					"es",
					"fs",
					"gs",
				}
				for _, r := range greg {
					regions = append(regions, elfRegion{
						name: prefix + r,
						min:  off,
						max:  off + 8,
					})
					off += 8
				}
				regions = append(regions, elfRegion{
					name: prefix + "prstatus footer",
					min:  off,
					max:  off + 8,
				})
				off += 8
			case elf.NT_FPREGSET:
				regions = append(regions, elfRegion{
					name: prefix + "fpregset header",
					min:  off,
					max:  off + 32,
				})
				off += 32
				for i := 0; i < 8; i++ {
					regions = append(regions, elfRegion{
						name: prefix + fmt.Sprintf("mmx%d", i),
						min:  off,
						max:  off + 16,
					})
					off += 16
					// They are long double (10 bytes), but
					// stored in 16-byte slots.
				}
				for i := 0; i < 16; i++ {
					regions = append(regions, elfRegion{
						name: prefix + fmt.Sprintf("xmm%d", i),
						min:  off,
						max:  off + 16,
					})
					off += 16
				}
				regions = append(regions, elfRegion{
					name: prefix + "fpregset footer",
					min:  off,
					max:  off + 96,
				})
				off += 96
				/*
					case NT_X86_XSTATE: // aka NT_PRPSINFO+511
						// legacy: 512 bytes
						// xsave header: 64 bytes
						fmt.Printf("hdr %v\n", desc[512:][:64])
						// ymm high128: 256 bytes

						println(len(desc))
						fallthrough
				*/
			default:
				regions = append(regions, elfRegion{
					name: fmt.Sprintf("%s/%s", name, typ),
					min:  off,
					max:  off + uint64(len(desc)),
				})
				off += uint64(len(desc))
			}
			if off != end {
				t.Fatalf("note section incomplete")
			}
		}
	}
	return regions
}
