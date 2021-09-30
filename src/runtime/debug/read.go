// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
)

var (
	// errUnrecognizedFormat is returned when a given executable file doesn't
	// appear to be in a known format, or it breaks the rules of that format,
	// or when there are I/O errors reading the file.
	errUnrecognizedFormat = errors.New("unrecognized file format")

	// errNotGoExe is returned when a given executable file is valid but does
	// not contain Go build information.
	errNotGoExe = errors.New("not a Go executable")

	// The build info blob left by the linker is identified by
	// a 16-byte header, consisting of buildInfoMagic (14 bytes),
	// the binary's pointer size (1 byte),
	// and whether the binary is big endian (1 byte).
	buildInfoMagic = []byte("\xff Go buildinf:")
)

// ReadBuildInfoFromFile returns build information embedded in a Go binary
// file at the given path. Most information is only available for binaries built
// with module support.
func ReadBuildInfoFromFile(name string) (info *BuildInfo, err error) {
	defer func() {
		if pathErr := (*fs.PathError)(nil); errors.As(err, &pathErr) {
			err = fmt.Errorf("could not read Go build info: %w", err)
		} else if err != nil {
			err = fmt.Errorf("could not read Go build info from %s: %w", name, err)
		}
	}()

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadBuildInfoFrom(f)
}

// ReadBuildInfoFrom returns build information embedded in a Go binary file
// accessed through the given ReaderAt. Most information is only available for
// binaries built with module support.
func ReadBuildInfoFrom(r io.ReaderAt) (*BuildInfo, error) {
	vers, mod, err := readRawBuildInfo(r)
	if err != nil {
		return nil, err
	}
	info, ok := parseBuildInfo(mod)
	if !ok {
		return nil, fmt.Errorf("could not parse Go build info")
	}
	info.GoVersion = vers
	return info, nil
}

// readRawBuildInfo extracts the Go toolchain version and module information
// strings from a Go binary. On success, vers should be non-empty. mod
// is empty if the binary was not built with modules enabled.
//
// This function has its own minimal helper functions for parsing binaries.
// This avoids the need to import debug/elf and so on from this package.
func readRawBuildInfo(r io.ReaderAt) (vers, mod string, err error) {
	// Read the first bytes of the file to identify the format, then delegate to
	// a format-specific function to load segment and section headers.
	// This buffer is large enough to fit the initial/ file header for all
	// supported formats, so it's passed to the functions/ to avoid reading these
	// bytes twice. Valid executables may be smaller than this buffer,
	// but certainly not Go executables.
	hdr := make([]byte, 96)
	if err := readFullAt(r, 0, hdr); err != nil {
		return "", "", errUnrecognizedFormat
	}
	var x *exe
	switch {
	case bytes.HasPrefix(hdr, []byte("\x7FELF")):
		x, err = readELFHeaders(r, hdr)
	case bytes.HasPrefix(hdr, []byte("MZ")):
		x, err = readPEHeaders(r, hdr)
	case binary.LittleEndian.Uint32(hdr[:4])&^1 == 0xFEEDFACE:
		x, err = readMachoHeaders(r, hdr)
	case bytes.HasPrefix(hdr, []byte{0x01, 0xDF}) || bytes.HasPrefix(hdr, []byte{0x01, 0xF7}):
		x, err = readXCOFFHeaders(r, hdr)
	default:
		return "", "", errUnrecognizedFormat
	}
	if err != nil {
		return "", "", err
	}

	// Find the build info blob.
	// Its location depends on the executable format, but it should be in the
	// ".go.buildinfo" section (ELF), "__go_buildinfo" section (Macho-O),
	// or the first 64KiB of the first data segment.
	var data []byte
	for _, sec := range x.sections {
		if sec.name == ".go.buildinfo" || sec.name == "__go_buildinfo" {
			data = make([]byte, sec.size)
			if err := readFullAt(r, sec.offset, data); err != nil {
				return "", "", errUnrecognizedFormat
			}
			break
		}
	}
	if data == nil {
		for _, seg := range x.segments {
			if !seg.isData {
				continue
			}
			readSize := seg.size
			if readSize > 64*1024 {
				readSize = 64 * 1024
			}
			data = make([]byte, readSize)
			if err := readFullAt(r, seg.offset, data); err != nil {
				return "", "", errUnrecognizedFormat
			}
			break
		}
	}
	for ; !bytes.HasPrefix(data, buildInfoMagic); data = data[32:] {
		if len(data) < 32 {
			return "", "", errNotGoExe
		}
	}

	// Decode the blob.
	// It starts with a 14-byte magic prefix, followed by two bytes indicating
	// pointer size and endianness which must match the executable.
	// Following that are two pointers (virtual addresses) to the version string
	// and module info string. Both point to Go string headers.
	if x.ptrSize == 0 {
		// PE doesn't explicitly set pointer size.
		x.ptrSize = int(data[14])
	} else if x.ptrSize != int(data[14]) {
		return "", "", errUnrecognizedFormat
	}
	if bigEndian := data[15] != 0; bigEndian != (x.bo == binary.BigEndian) {
		return "", "", errUnrecognizedFormat
	}
	vers, err = x.readGoString(r, x.readPtr(data[16:]))
	if err != nil {
		return "", "", err
	}
	mod, err = x.readGoString(r, x.readPtr(data[16+x.ptrSize:]))
	if err != nil {
		return "", "", err
	}
	if len(mod) >= 33 && mod[len(mod)-17] == '\n' {
		// Strip module framing.
		mod = mod[16 : len(mod)-16]
	} else {
		mod = ""
	}
	return vers, mod, nil
}

// exe is a generic representation of an executable file.
type exe struct {
	ptrSize  int
	bo       binary.ByteOrder
	segments []segment
	sections []section
	base     uint64
}

// segment is a generic representation of a part of an executable file that
// can be mapped into memory. Different formats use different terminology: this
// is an ELF program, a Mach-O segment, or a PE or XCOFF section.
type segment struct {
	addr, offset, size uint64
	isData             bool
}

// section is a generic representation of a named part of an executable file.
// Only used for ELF and Mach-O.
type section struct {
	name         string
	offset, size uint64
}

// offsetForAddr maps a virtual address to an offset within the executable file.
// If the address points outside any segment with data in the file,
// offsetForAddr returns 0. This happens for addresses in BSS segments.
func (x *exe) offsetForAddr(addr uint64) uint64 {
	addr -= x.base
	for _, seg := range x.segments {
		if seg.addr == 0 {
			continue
		}
		if seg.addr <= addr && addr <= seg.addr+seg.size-1 {
			return seg.offset + addr - seg.addr
		}
	}
	return 0
}

func (x *exe) readPtr(b []byte) uint64 {
	if x.ptrSize == 4 {
		return uint64(x.bo.Uint32(b))
	} else {
		return x.bo.Uint64(b)
	}
}

func readCString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i < 0 {
		return string(b)
	} else {
		return string(b[:i])
	}
}

// readGoString returns the string at the given virtual address.
// addr must point to a Go string header.
func (x *exe) readGoString(r io.ReaderAt, addr uint64) (string, error) {
	offset := x.offsetForAddr(addr)
	if offset == 0 {
		// Likely a pointer to a zero-initialized BSS segment.
		return "", nil
	}
	hdr := make([]byte, 2*x.ptrSize)
	if err := readFullAt(r, offset, hdr); err != nil {
		return "", errUnrecognizedFormat
	}
	dataAddr := x.readPtr(hdr)
	dataLen := x.readPtr(hdr[x.ptrSize:])
	if dataLen == 0 {
		return "", nil
	}
	dataOffset := x.offsetForAddr(dataAddr)
	if dataOffset == 0 {
		return "", errUnrecognizedFormat
	}
	data := make([]byte, dataLen)
	if err := readFullAt(r, dataOffset, data); err != nil {
		return "", errUnrecognizedFormat
	}
	return string(data), nil
}

// Compare with debug/elf.Header32.
type elfFileHeader32 struct {
	Ident     [16]byte /* File identification. */
	Type      uint16   /* File type. */
	Machine   uint16   /* Machine architecture. */
	Version   uint32   /* ELF format version. */
	Entry     uint32   /* Entry point. */
	Phoff     uint32   /* Program header file offset. */
	Shoff     uint32   /* Section header file offset. */
	Flags     uint32   /* Architecture-specific flags. */
	Ehsize    uint16   /* Size of ELF header in bytes. */
	Phentsize uint16   /* Size of program header entry. */
	Phnum     uint16   /* Number of program header entries. */
	Shentsize uint16   /* Size of section header entry. */
	Shnum     uint16   /* Number of section header entries. */
	Shstrndx  uint16   /* Section name strings section. */
}

// Compare with debug/elf.Header64.
type elfFileHeader64 struct {
	Ident     [16]byte /* File identification. */
	Type      uint16   /* File type. */
	Machine   uint16   /* Machine architecture. */
	Version   uint32   /* ELF format version. */
	Entry     uint64   /* Entry point. */
	Phoff     uint64   /* Program header file offset. */
	Shoff     uint64   /* Section header file offset. */
	Flags     uint32   /* Architecture-specific flags. */
	Ehsize    uint16   /* Size of ELF header in bytes. */
	Phentsize uint16   /* Size of program header entry. */
	Phnum     uint16   /* Number of program header entries. */
	Shentsize uint16   /* Size of section header entry. */
	Shnum     uint16   /* Number of section header entries. */
	Shstrndx  uint16   /* Section name strings section. */
}

// Compare with debug/elf.Section32.
type elfSectionHeader32 struct {
	Name      uint32 /* Section name (index into the section header string table). */
	Type      uint32 /* Section type. */
	Flags     uint32 /* Section flags. */
	Addr      uint32 /* Address in memory image. */
	Off       uint32 /* Offset in file. */
	Size      uint32 /* Size in bytes. */
	Link      uint32 /* Index of a related section. */
	Info      uint32 /* Depends on section type. */
	Addralign uint32 /* Alignment in bytes. */
	Entsize   uint32 /* Size of each entry in section. */
}

// Compare with debug/elf.Section64.
type elfSectionHeader64 struct {
	Name      uint32 /* Section name (index into the section header string table). */
	Type      uint32 /* Section type. */
	Flags     uint64 /* Section flags. */
	Addr      uint64 /* Address in memory image. */
	Off       uint64 /* Offset in file. */
	Size      uint64 /* Size in bytes. */
	Link      uint32 /* Index of a related section. */
	Info      uint32 /* Depends on section type. */
	Addralign uint64 /* Alignment in bytes. */
	Entsize   uint64 /* Size of each entry in section. */
}

// Compare with debug/elf.Prog32.
type elfProgramHeader32 struct {
	Type   uint32 /* Entry type. */
	Off    uint32 /* File offset of contents. */
	Vaddr  uint32 /* Virtual address in memory image. */
	Paddr  uint32 /* Physical address (not used). */
	Filesz uint32 /* Size of contents in file. */
	Memsz  uint32 /* Size of contents in memory. */
	Flags  uint32 /* Access permission flags. */
	Align  uint32 /* Alignment in memory and file. */
}

// Compare with debug/elf.Prog64
type elfProgramHeader64 struct {
	Type   uint32 /* Entry type. */
	Flags  uint32 /* Access permission flags. */
	Off    uint64 /* File offset of contents. */
	Vaddr  uint64 /* Virtual address in memory image. */
	Paddr  uint64 /* Physical address (not used). */
	Filesz uint64 /* Size of contents in file. */
	Memsz  uint64 /* Size of contents in memory. */
	Align  uint64 /* Alignment in memory and file. */
}

// Compare with debug/elf.NewFile.
func readELFHeaders(r io.ReaderAt, hdr []byte) (*exe, error) {
	const (
		PT_LOAD = 1 // loadable segment
		PF_X    = 1 // executable
		PF_W    = 2 // writable
	)

	// Read file header.
	x := &exe{}
	switch hdr[4] {
	case 1:
		x.ptrSize = 4
	case 2:
		x.ptrSize = 8
	default:
		return nil, errUnrecognizedFormat
	}
	switch hdr[5] {
	case 1:
		x.bo = binary.LittleEndian
	case 2:
		x.bo = binary.BigEndian
	default:
		return nil, errUnrecognizedFormat
	}

	var shnum, shentsize, shstrndx uint16
	var shoff uint64
	var phnum, phentsize uint16
	var phoff uint64
	if x.ptrSize == 4 {
		var fh elfFileHeader32
		if err := binary.Read(bytes.NewReader(hdr), x.bo, &fh); err != nil {
			return nil, errUnrecognizedFormat
		}
		shnum, shentsize, shstrndx = fh.Shnum, fh.Shentsize, fh.Shstrndx
		shoff = uint64(fh.Shoff)
		phnum, phentsize = fh.Phnum, fh.Phentsize
		phoff = uint64(fh.Phoff)
	} else {
		var fh elfFileHeader64
		if err := binary.Read(bytes.NewReader(hdr), x.bo, &fh); err != nil {
			return nil, errUnrecognizedFormat
		}
		shnum, shentsize, shstrndx = fh.Shnum, fh.Shentsize, fh.Shstrndx
		shoff = fh.Shoff
		phnum, phentsize = fh.Phnum, fh.Phentsize
		phoff = fh.Phoff
	}
	if shstrndx >= shnum {
		return nil, errUnrecognizedFormat
	}

	// Read section headers.
	secData := make([]byte, shnum*shentsize)
	if err := readFullAt(r, shoff, secData); err != nil {
		return nil, errUnrecognizedFormat
	}
	x.sections = make([]section, shnum)
	nameOffsets := make([]uint32, shnum)
	for i := uint16(0); i < shnum; i++ {
		sr := bytes.NewReader(secData[i*shentsize : (i+1)*shentsize])
		if x.ptrSize == 4 {
			var sh elfSectionHeader32
			if err := binary.Read(sr, x.bo, &sh); err != nil {
				return nil, errUnrecognizedFormat
			}
			nameOffsets[i] = sh.Name
			x.sections[i] = section{
				offset: uint64(sh.Off),
				size:   uint64(sh.Size),
			}
		} else {
			var sh elfSectionHeader64
			if err := binary.Read(sr, x.bo, &sh); err != nil {
				return nil, errUnrecognizedFormat
			}
			nameOffsets[i] = sh.Name
			x.sections[i] = section{
				offset: sh.Off,
				size:   sh.Size,
			}
		}
	}
	shstrtab := make([]byte, x.sections[shstrndx].size)
	if err := readFullAt(r, x.sections[shstrndx].offset, shstrtab); err != nil {
		return nil, errUnrecognizedFormat
	}
	for i := uint16(0); i < shnum; i++ {
		if uint64(nameOffsets[i]) >= uint64(len(shstrtab)) {
			return nil, errUnrecognizedFormat
		}
		x.sections[i].name = readCString(shstrtab[nameOffsets[i]:])
	}

	// Read program (segment) headers.
	progData := make([]byte, phnum*phentsize)
	if err := readFullAt(r, phoff, progData); err != nil {
		return nil, errUnrecognizedFormat
	}
	x.segments = make([]segment, phnum)
	for i := uint16(0); i < phnum; i++ {
		pr := bytes.NewReader(progData[i*phentsize : (i+1)*phentsize])
		if x.ptrSize == 4 {
			var ph elfProgramHeader32
			if err := binary.Read(pr, x.bo, &ph); err != nil {
				return nil, errUnrecognizedFormat
			}
			x.segments[i] = segment{
				addr:   uint64(ph.Vaddr),
				offset: uint64(ph.Off),
				size:   uint64(ph.Filesz),
				isData: ph.Type == PT_LOAD && ph.Flags&(PF_X|PF_W) == PF_W,
			}
		} else {
			var ph elfProgramHeader64
			if err := binary.Read(pr, x.bo, &ph); err != nil {
				return nil, errUnrecognizedFormat
			}
			x.segments[i] = segment{
				addr:   ph.Vaddr,
				offset: ph.Off,
				size:   ph.Filesz,
				isData: ph.Type == PT_LOAD && ph.Flags&(PF_X|PF_W) == PF_W,
			}
		}
	}
	return x, nil
}

// Compare with debug/macho.FileHeader.
type machoFileHeader struct {
	Magic  uint32
	Cpu    uint32
	SubCpu uint32
	Type   uint32
	Ncmd   uint32
	Cmdsz  uint32
	Flags  uint32
}

// Compare with debug/macho.Segment32.
type machoSegmentCmd32 struct {
	Cmd     uint32
	Len     uint32
	Name    [16]byte
	Addr    uint32
	Memsz   uint32
	Offset  uint32
	Filesz  uint32
	Maxprot uint32
	Prot    uint32
	Nsect   uint32
	Flag    uint32
}

// Compare with debug/macho.Segment64.
type machoSegmentCmd64 struct {
	Cmd     uint32
	Len     uint32
	Name    [16]byte
	Addr    uint64
	Memsz   uint64
	Offset  uint64
	Filesz  uint64
	Maxprot uint32
	Prot    uint32
	Nsect   uint32
	Flag    uint32
}

// Compare with debug/macho.Section32.
type machoSection32 struct {
	Name     [16]byte
	Seg      [16]byte
	Addr     uint32
	Size     uint32
	Offset   uint32
	Align    uint32
	Reloff   uint32
	Nreloc   uint32
	Flags    uint32
	Reserve1 uint32
	Reserve2 uint32
}

// Compare with debug/macho.Section64.
type machoSection64 struct {
	Name     [16]byte
	Seg      [16]byte
	Addr     uint64
	Size     uint64
	Offset   uint32
	Align    uint32
	Reloff   uint32
	Nreloc   uint32
	Flags    uint32
	Reserve1 uint32
	Reserve2 uint32
	Reserve3 uint32
}

// Compare with debug/macho.NewFile.
func readMachoHeaders(r io.ReaderAt, hdr []byte) (*exe, error) {
	const (
		fileHeaderSize32 uint64 = 7 * 4
		fileHeaderSize64 uint64 = 8 * 4
		loadCmdSegment32 uint32 = 0x1
		loadCmdSegment64 uint32 = 0x19
	)

	// Read the file header.
	// Only support little-endian. Go doesn't target any big-endian platforms
	// where Mach-O is used.
	x := &exe{bo: binary.LittleEndian}
	var fileHeaderSize uint64
	magic := binary.LittleEndian.Uint32(hdr[:4])
	if magic == 0xfeedface {
		x.ptrSize = 4
		fileHeaderSize = fileHeaderSize32
	} else if magic == 0xfeedfacf {
		x.ptrSize = 8
		fileHeaderSize = fileHeaderSize64
	} else {
		return nil, errUnrecognizedFormat
	}

	var fh machoFileHeader
	if err := binary.Read(bytes.NewReader(hdr), x.bo, &fh); err != nil {
		return nil, errUnrecognizedFormat
	}

	// Read loader commands.
	// Each command begins with a 32-bit enum and 32-bit size.
	// We only care about segment commands.
	// Each segment command is immediately followed by its sections.
	cmdData := make([]byte, fh.Cmdsz)
	if err := readFullAt(r, fileHeaderSize, cmdData); err != nil {
		return nil, errUnrecognizedFormat
	}
	cmdIdx := uint32(0)
	for len(cmdData) >= 8 {
		cmd := x.bo.Uint32(cmdData[:4])
		sz := x.bo.Uint32(cmdData[4:8])
		if sz < 8 || sz > uint32(len(cmdData)) {
			return nil, errUnrecognizedFormat
		}
		switch cmd {
		case loadCmdSegment32:
			var scmd machoSegmentCmd32
			sr := bytes.NewReader(cmdData[:sz])
			if err := binary.Read(sr, x.bo, &scmd); err != nil {
				return nil, errUnrecognizedFormat
			}
			x.segments = append(x.segments, segment{
				addr:   uint64(scmd.Addr),
				offset: uint64(scmd.Offset),
				size:   uint64(scmd.Filesz),
				isData: scmd.Prot&6 == 2, // writable, not executable
			})
			for i := uint32(0); i < scmd.Nsect; i++ {
				var sh machoSection32
				if err := binary.Read(sr, x.bo, &sh); err != nil {
					return nil, errUnrecognizedFormat
				}
				x.sections = append(x.sections, section{
					name:   readCString(sh.Name[:]),
					offset: uint64(sh.Offset),
					size:   uint64(sh.Size),
				})
			}

		case loadCmdSegment64:
			var scmd machoSegmentCmd64
			sr := bytes.NewReader(cmdData[:sz])
			if err := binary.Read(sr, x.bo, &scmd); err != nil {
				return nil, errUnrecognizedFormat
			}
			x.segments = append(x.segments, segment{
				addr:   scmd.Addr,
				offset: scmd.Offset,
				size:   scmd.Filesz,
				isData: scmd.Prot&6 == 2, // writable, not executable
			})
			for i := uint32(0); i < scmd.Nsect; i++ {
				var sec machoSection64
				if err := binary.Read(sr, x.bo, &sec); err != nil {
					return nil, errUnrecognizedFormat
				}
				x.sections = append(x.sections, section{
					name:   readCString(sec.Name[:]),
					offset: uint64(sec.Offset),
					size:   sec.Size,
				})
			}

		default:
			break
		}

		cmdData = cmdData[sz:]
		cmdIdx++
	}
	if len(cmdData) != 0 || cmdIdx != fh.Ncmd {
		return nil, errUnrecognizedFormat
	}

	return x, nil
}

// Compare with debug/pe.FileHeader.
type peFileHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

// Compare with debug/pe.SectionHeader32.
// There's no SectionHeader64.
type peSectionHeader32 struct {
	Name                 [8]uint8
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLineNumbers uint32
	NumberOfRelocations  uint16
	NumberOfLineNumbers  uint16
	Characteristics      uint32
}

// Compare with debug/pe.NewFile.
func readPEHeaders(r io.ReaderAt, hdr []byte) (*exe, error) {
	const peCOFFSymbolSize = 18

	// Read file header.
	if hdr[0] != 'M' || hdr[1] != 'Z' {
		return nil, errUnrecognizedFormat
	}
	signoff := int64(binary.LittleEndian.Uint32(hdr[0x3c:]))
	var sign [4]byte
	if err := readFullAt(r, uint64(signoff), sign[:]); err != nil || sign != [4]byte{'P', 'E', 0, 0} {
		return nil, errUnrecognizedFormat
	}
	base := signoff + 4
	sr := io.NewSectionReader(r, 0, 1<<63-1)
	sr.Seek(base, io.SeekStart)

	x := &exe{bo: binary.LittleEndian}
	var fh peFileHeader
	if err := binary.Read(sr, x.bo, &fh); err != nil {
		return nil, errUnrecognizedFormat
	}

	// Read optional header. We only need the image base field.
	if fh.SizeOfOptionalHeader < 2 {
		return nil, errUnrecognizedFormat
	}
	optHdr := make([]byte, fh.SizeOfOptionalHeader)
	if _, err := io.ReadFull(sr, optHdr); err != nil {
		return nil, errUnrecognizedFormat
	}
	switch binary.LittleEndian.Uint16(optHdr) {
	case 0x10b:
		// PE32
		x.ptrSize = 4
		if fh.SizeOfOptionalHeader != 224 {
			return nil, errUnrecognizedFormat
		}
		x.base = uint64(binary.LittleEndian.Uint32(optHdr[28:]))

	case 0x20b:
		// PE64
		x.ptrSize = 8
		if fh.SizeOfOptionalHeader != 240 {
			return nil, errUnrecognizedFormat
		}
		x.base = binary.LittleEndian.Uint64(optHdr[24:])

	default:
		return nil, errUnrecognizedFormat
	}

	// Read string table.
	var strtab []byte
	if fh.PointerToSymbolTable > 0 {
		strtabOffset := int64(fh.PointerToSymbolTable) + peCOFFSymbolSize*int64(fh.NumberOfSymbols)
		sr.Seek(strtabOffset, io.SeekStart)
		var strtabLen uint32
		if err := binary.Read(sr, binary.LittleEndian, &strtabLen); err != nil {
			return nil, errUnrecognizedFormat
		}
		// string table length includes itself
		if strtabLen <= 4 {
			return nil, errUnrecognizedFormat
		}
		strtabLen -= 4
		strtab = make([]byte, strtabLen)
		if _, err := io.ReadFull(sr, strtab); err != nil {
			return nil, errUnrecognizedFormat
		}
	}

	// Seek past headers and read sections.
	sr.Seek(base+int64(binary.Size(fh))+int64(fh.SizeOfOptionalHeader), io.SeekStart)
	for i := uint16(0); i < fh.NumberOfSections; i++ {
		var sh peSectionHeader32
		if err := binary.Read(sr, binary.LittleEndian, &sh); err != nil {
			return nil, errUnrecognizedFormat
		}

		const (
			IMAGE_SCN_MEM_EXECUTE = 0x20000000
			IMAGE_SCN_MEM_WRITE   = 0x80000000
		)
		x.segments = append(x.segments, segment{
			addr:   uint64(sh.VirtualAddress),
			offset: uint64(sh.PointerToRawData),
			size:   uint64(sh.SizeOfRawData),
			isData: sh.Characteristics&IMAGE_SCN_MEM_EXECUTE == 0 && sh.Characteristics&IMAGE_SCN_MEM_WRITE != 0,
		})
	}

	return x, nil
}

// Compare with internal/xcoff.FileHeader32.
type xcoffFileHeader32 struct {
	Fmagic   uint16 // Target machine
	Fnscns   uint16 // Number of sections
	Ftimedat int32  // Time and date of file creation
	Fsymptr  uint32 // Byte offset to symbol table start
	Fnsyms   int32  // Number of entries in symbol table
	Fopthdr  uint16 // Number of bytes in optional header
	Fflags   uint16 // Flags
}

// Compare with internal/xcoff.FileHeader64.
type xcoffFileHeader64 struct {
	Fmagic   uint16 // Target machine
	Fnscns   uint16 // Number of sections
	Ftimedat int32  // Time and date of file creation
	Fsymptr  uint64 // Byte offset to symbol table start
	Fopthdr  uint16 // Number of bytes in optional header
	Fflags   uint16 // Flags
	Fnsyms   int32  // Number of entries in symbol table
}

// Compare with internal/xcoff.SectionHeader32.
type xcoffSectionHeader32 struct {
	Sname    [8]byte // Section name
	Spaddr   uint32  // Physical address
	Svaddr   uint32  // Virtual address
	Ssize    uint32  // Section size
	Sscnptr  uint32  // Offset in file to raw data for section
	Srelptr  uint32  // Offset in file to relocation entries for section
	Slnnoptr uint32  // Offset in file to line number entries for section
	Snreloc  uint16  // Number of relocation entries
	Snlnno   uint16  // Number of line number entries
	Sflags   uint32  // Flags to define the section type
}

// Compare with internal/xcoff.SectionHeader64.
type xcoffSectionHeader64 struct {
	Sname    [8]byte // Section name
	Spaddr   uint64  // Physical address
	Svaddr   uint64  // Virtual address
	Ssize    uint64  // Section size
	Sscnptr  uint64  // Offset in file to raw data for section
	Srelptr  uint64  // Offset in file to relocation entries for section
	Slnnoptr uint64  // Offset in file to line number entries for section
	Snreloc  uint32  // Number of relocation entries
	Snlnno   uint32  // Number of line number entries
	Sflags   uint32  // Flags to define the section type
	Spad     uint32  // Needs to be 72 bytes long
}

// Compare with internal/xcoff.NewFile.
func readXCOFFHeaders(r io.ReaderAt, hdr []byte) (*exe, error) {
	const (
		XCOFF_U802TOCMAGIC = 0737
		XCOFF_U64_TOCMAGIC = 0767
	)

	// Read file header.
	x := &exe{bo: binary.BigEndian}
	sr := io.NewSectionReader(r, 0, 1<<63-1)
	var headerSize, sectionHeaderSize int
	var optHeaderSize uint16
	var numSections uint16
	switch binary.BigEndian.Uint16(hdr) {
	case XCOFF_U802TOCMAGIC:
		// AIX 32-bit XCOFF
		x.ptrSize = 4
		var fh xcoffFileHeader32
		if err := binary.Read(sr, binary.BigEndian, &fh); err != nil {
			return nil, err
		}
		headerSize = binary.Size(fh)
		sectionHeaderSize = binary.Size(xcoffSectionHeader32{})
		optHeaderSize = fh.Fopthdr
		numSections = fh.Fnscns
	case XCOFF_U64_TOCMAGIC:
		// AIX 64-bit XCOFF
		x.ptrSize = 8
		var fh xcoffFileHeader64
		if err := binary.Read(sr, binary.BigEndian, &fh); err != nil {
			return nil, err
		}
		headerSize = binary.Size(fh)
		sectionHeaderSize = binary.Size(xcoffSectionHeader64{})
		optHeaderSize = fh.Fopthdr
		numSections = fh.Fnscns
	default:
		return nil, errUnrecognizedFormat
	}
	if optHeaderSize < 0 {
		return nil, errUnrecognizedFormat
	}

	// Read section headers
	const (
		STYP_TEXT = 0x0020
		STYP_DATA = 0x0040
	)
	sr.Seek(int64(headerSize)+int64(optHeaderSize), io.SeekStart)
	secData := make([]byte, int(numSections)*sectionHeaderSize)
	if _, err := io.ReadFull(sr, secData); err != nil {
		return nil, errUnrecognizedFormat
	}
	shr := bytes.NewReader(secData)
	x.segments = make([]segment, 0, numSections)
	for i := uint16(0); i < numSections; i++ {
		if x.ptrSize == 4 {
			var sh xcoffSectionHeader32
			if err := binary.Read(shr, binary.BigEndian, &sh); err != nil {
				return nil, errUnrecognizedFormat
			}
			x.segments = append(x.segments, segment{
				addr:   uint64(sh.Svaddr),
				offset: uint64(sh.Sscnptr),
				size:   uint64(sh.Ssize),
				isData: sh.Sflags == STYP_DATA,
			})
			if sh.Sflags != STYP_TEXT && sh.Sflags != STYP_DATA {
				x.segments[i].size = 0 // BSS or something else not relevant.
			}
		} else {
			var sh xcoffSectionHeader64
			if err := binary.Read(shr, binary.BigEndian, &sh); err != nil {
				return nil, errUnrecognizedFormat
			}
			x.segments = append(x.segments, segment{
				addr:   sh.Svaddr,
				offset: sh.Sscnptr,
				size:   sh.Ssize,
				isData: sh.Sflags == STYP_DATA,
			})
			if sh.Sflags != STYP_TEXT && sh.Sflags != STYP_DATA {
				x.segments[i].size = 0 // BSS or something else not relevant.
			}
		}
	}
	return x, nil
}

// readFullAt reads len(b) bytes from r at offset off.
// readFullAt is like io.ReadFull but with io.ReaderAt.
func readFullAt(r io.ReaderAt, off uint64, b []byte) error {
	if int64(off) < 0 {
		return errUnrecognizedFormat
	}
	for len(b) > 0 {
		n, err := r.ReadAt(b, int64(off))
		if err == io.EOF {
			return io.ErrUnexpectedEOF
		} else if err != nil && n < len(b) {
			return err
		}
		b = b[n:]
		off += uint64(n)
	}
	return nil
}
