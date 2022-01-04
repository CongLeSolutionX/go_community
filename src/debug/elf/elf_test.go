// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elf

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

type nameTest struct {
	val any
	str string
}

var nameTests = []nameTest{
	{ELFOSABI_LINUX, "ELFOSABI_LINUX"},
	{ET_EXEC, "ET_EXEC"},
	{EM_860, "EM_860"},
	{SHN_LOPROC, "SHN_LOPROC"},
	{SHT_PROGBITS, "SHT_PROGBITS"},
	{SHF_MERGE + SHF_TLS, "SHF_MERGE+SHF_TLS"},
	{PT_LOAD, "PT_LOAD"},
	{PF_W + PF_R + 0x50, "PF_W+PF_R+0x50"},
	{DT_SYMBOLIC, "DT_SYMBOLIC"},
	{DF_BIND_NOW, "DF_BIND_NOW"},
	{NT_FPREGSET, "NT_FPREGSET"},
	{STB_GLOBAL, "STB_GLOBAL"},
	{STT_COMMON, "STT_COMMON"},
	{STV_HIDDEN, "STV_HIDDEN"},
	{R_X86_64_PC32, "R_X86_64_PC32"},
	{R_ALPHA_OP_PUSH, "R_ALPHA_OP_PUSH"},
	{R_ARM_THM_ABS5, "R_ARM_THM_ABS5"},
	{R_386_GOT32, "R_386_GOT32"},
	{R_PPC_GOT16_HI, "R_PPC_GOT16_HI"},
	{R_SPARC_GOT22, "R_SPARC_GOT22"},
	{ET_LOOS + 5, "ET_LOOS+5"},
	{ProgFlag(0x50), "0x50"},
}

func TestNames(t *testing.T) {
	for i, tt := range nameTests {
		s := fmt.Sprint(tt.val)
		if s != tt.str {
			t.Errorf("#%d: Sprint(%d) = %q, want %q", i, tt.val, s, tt.str)
		}
	}
}

func TestNobitsSection(t *testing.T) {
	const wrongContents = "wrong"
	s := &Section{
		SectionHeader: SectionHeader{
			Type:   SHT_NOBITS,
			Addr:   0x1000,
			Offset: 0x1000,
			Size:   uint64(len(wrongContents))},
		sr: io.NewSectionReader(bytes.NewReader([]byte(wrongContents)), 0, 1<<63-1),
	}
	buf, err := s.Data()
	if err != nil {
		t.Fatalf("error reading section data: %v", err)
	}
	if len(buf) != len(wrongContents) {
		t.Errorf("wrong  length %d", len(buf))
	}
	for i := range buf {
		if buf[i] != 0x00 {
			t.Errorf("wrong byte in section data at %d: %#x\n", i, buf[i])
		}
	}
}
