// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tar implements access to tar archives.
// It aims to cover most of the variations, including those produced
// by GNU and BSD tars.
//
// References:
//	https://www.freebsd.org/cgi/man.cgi?query=tar&sektion=5
//	https://www.gnu.org/software/tar/manual/html_node/Standard.html
//	http://pubs.opengroup.org/onlinepubs/9699919799/utilities/pax.html
package tar

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"time"
)

// Additional references:
//	http://manpages.ubuntu.com/manpages/intrepid/man5/star.5.html
//
// Popular implementations:
//	http://git.savannah.gnu.org/cgit/tar.git
//	https://github.com/libarchive/libarchive
//	http://sourceforge.net/projects/s-tar/

// BUG: Use of the Uid and Gid fields in Header could overflow on 32-bit
// architectures. If a large value is encountered when decoding, the result
// stored in Header will be the truncated version. Similarly, the truncated
// version will be used when encoding.

const (
	// The format is unknown.
	formatUnknown = iota

	// The format of the original Unix V7 tar tool prior to standardization.
	/*
		type headerV7 struct {
			name     [100]byte //   0
			mode     [8]byte   // 100
			uid      [8]byte   // 108
			gid      [8]byte   // 116
			size     [12]byte  // 124
			mtime    [12]byte  // 136
			chksum   [8]byte   // 148
			typeflag [1]byte   // 156
			linkname [100]byte // 157
			                   // 257
		}
	*/
	formatV7

	// The old and new GNU formats, which are incompatible with USTAR.
	// This does cover the old GNU sparse extension.
	// This does not cover the GNU sparse extensions using PAX headers,
	// versions 0.0, 0.1, and 1.0; these fall under the PAX format.
	/*
		// The magic and version forms "ustar  \x00"
		type headerGNU struct {
			headerV7                   //   0
			magic      [6]byte         // 257
			version    [2]byte         // 263
			uname      [32]byte        // 265
			gname      [32]byte        // 297
			devmajor   [8]byte         // 329
			devminor   [8]byte         // 337
			atime      [12]byte        // 345
			ctime      [12]byte        // 357
			offset     [12]byte        // 369
			longnames  [4]byte         // 381
			unused     [1]byte         // 385
			sparse     [4]headerSparse // 386
			isextended [1]byte         // 482
			realsize   [12]byte        // 483
			                           // 495
		}

		type headerSparse struct {
			offset   [12]byte //  0
			numbytes [12]byte // 12
			                  // 24
		}
	*/
	formatGNU

	// Schily's tar format, which is incompatible with USTAR.
	// This does not cover STAR extensions to the PAX format; these fall under
	// the PAX format.
	/*
		// The magic forms "ustar\x00".
		// The trailer forms "tar\x00".
		type headerSTAR struct {
			headerV7           //   0
			magic    [6]byte   // 257
			version  [2]byte   // 263
			uname    [32]byte  // 265
			gname    [32]byte  // 297
			devmajor [8]byte   // 329
			devminor [8]byte   // 337
			prefix   [131]byte // 345
			atime    [12]byte  // 476
			ctime    [12]byte  // 488
			                   // 500
		}
	*/
	formatSTAR

	// USTAR is the former standardization of tar defined in POSIX.1-1988.
	// This is an extension of USTAR and is "backwards compatible" with it.
	// This is incompatible with the GNU and STAR formats.
	/*
		// The magic forms "ustar\x00".
		type headerUSTAR struct {
			headerV7           //   0
			magic    [6]byte   // 257
			version  [2]byte   // 263
			uname    [32]byte  // 265
			gname    [32]byte  // 297
			devmajor [8]byte   // 329
			devminor [8]byte   // 337
			prefix   [155]byte // 345
			                   // 500
		}
	*/
	formatUSTAR

	// PAX is the latest standardization of tar defined in POSIX.1-2001.
	// It adds a new header format that allows for arbitrarily long fields and
	// is an extension of the previous USTAR format.
	//
	// Some formats add their own extensions to PAX, such as GNU sparse files
	// and SCHILY extended attributes.
	formatPAX
)

// Magics used to identify various formats.
const (
	magicSTAR  = "tar\x00"
	magicUSTAR = "ustar\x00"
	magicGNU   = "ustar  \x00"
)

// Header type flags.
const (
	TypeReg           = '0'    // regular file
	TypeRegA          = '\x00' // regular file
	TypeLink          = '1'    // hard link
	TypeSymlink       = '2'    // symbolic link
	TypeChar          = '3'    // character device node
	TypeBlock         = '4'    // block device node
	TypeDir           = '5'    // directory
	TypeFifo          = '6'    // fifo node
	TypeCont          = '7'    // reserved
	TypeXHeader       = 'x'    // extended header
	TypeXGlobalHeader = 'g'    // global extended header
	TypeGNULongName   = 'L'    // Next file has a long name
	TypeGNULongLink   = 'K'    // Next file symlinks to a file w/ a long name
	TypeGNUSparse     = 'S'    // sparse file
)

// A Header represents a single header in a tar archive.
// Some fields may not be populated.
type Header struct {
	Name       string    // name of header file entry
	Mode       int64     // permission and mode bits
	Uid        int       // user id of owner
	Gid        int       // group id of owner
	Size       int64     // length in bytes
	ModTime    time.Time // modified time
	Typeflag   byte      // type of header entry
	Linkname   string    // target name of link
	Uname      string    // user name of owner
	Gname      string    // group name of owner
	Devmajor   int64     // major number of character or block device
	Devminor   int64     // minor number of character or block device
	AccessTime time.Time // access time
	ChangeTime time.Time // status change time
	Xattrs     map[string]string
}

// Size constants from the tar spec.
const (
	blockSize          = 512
	fileNameSize       = 100 // Maximum number of bytes in a standard tar name.
	fileNamePrefixSize = 155 // Maximum number of ustar extension bytes.
)

// FileInfo returns an os.FileInfo for the Header.
func (h *Header) FileInfo() os.FileInfo {
	return headerFileInfo{h}
}

// headerFileInfo implements os.FileInfo.
type headerFileInfo struct {
	h *Header
}

func (fi headerFileInfo) Size() int64        { return fi.h.Size }
func (fi headerFileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi headerFileInfo) ModTime() time.Time { return fi.h.ModTime }
func (fi headerFileInfo) Sys() interface{}   { return fi.h }

// Name returns the base name of the file.
func (fi headerFileInfo) Name() string {
	if fi.IsDir() {
		return path.Base(path.Clean(fi.h.Name))
	}
	return path.Base(fi.h.Name)
}

// Mode returns the permission and mode bits for the headerFileInfo.
func (fi headerFileInfo) Mode() (mode os.FileMode) {
	// Set file permission bits.
	mode = os.FileMode(fi.h.Mode).Perm()

	// Set setuid, setgid and sticky bits.
	if fi.h.Mode&c_ISUID != 0 {
		// setuid
		mode |= os.ModeSetuid
	}
	if fi.h.Mode&c_ISGID != 0 {
		// setgid
		mode |= os.ModeSetgid
	}
	if fi.h.Mode&c_ISVTX != 0 {
		// sticky
		mode |= os.ModeSticky
	}

	// Set file mode bits.
	// clear perm, setuid, setgid and sticky bits.
	m := os.FileMode(fi.h.Mode) &^ 07777
	if m == c_ISDIR {
		// directory
		mode |= os.ModeDir
	}
	if m == c_ISFIFO {
		// named pipe (FIFO)
		mode |= os.ModeNamedPipe
	}
	if m == c_ISLNK {
		// symbolic link
		mode |= os.ModeSymlink
	}
	if m == c_ISBLK {
		// device file
		mode |= os.ModeDevice
	}
	if m == c_ISCHR {
		// Unix character device
		mode |= os.ModeDevice
		mode |= os.ModeCharDevice
	}
	if m == c_ISSOCK {
		// Unix domain socket
		mode |= os.ModeSocket
	}

	switch fi.h.Typeflag {
	case TypeSymlink:
		// symbolic link
		mode |= os.ModeSymlink
	case TypeChar:
		// character device node
		mode |= os.ModeDevice
		mode |= os.ModeCharDevice
	case TypeBlock:
		// block device node
		mode |= os.ModeDevice
	case TypeDir:
		// directory
		mode |= os.ModeDir
	case TypeFifo:
		// fifo node
		mode |= os.ModeNamedPipe
	}

	return mode
}

// sysStat, if non-nil, populates h from system-dependent fields of fi.
var sysStat func(fi os.FileInfo, h *Header) error

// Mode constants from the tar spec.
const (
	c_ISUID  = 04000   // Set uid
	c_ISGID  = 02000   // Set gid
	c_ISVTX  = 01000   // Save text (sticky bit)
	c_ISDIR  = 040000  // Directory
	c_ISFIFO = 010000  // FIFO
	c_ISREG  = 0100000 // Regular file
	c_ISLNK  = 0120000 // Symbolic link
	c_ISBLK  = 060000  // Block special file
	c_ISCHR  = 020000  // Character special file
	c_ISSOCK = 0140000 // Socket
)

// Keywords for the PAX Extended Header.
const (
	paxNone     = ""
	paxPath     = "path"
	paxLinkpath = "linkpath"
	paxUname    = "uname"
	paxGname    = "gname"
	paxUid      = "uid"
	paxGid      = "gid"
	paxAtime    = "atime"
	paxMtime    = "mtime"
	paxCtime    = "ctime"
	paxSize     = "size"
	paxXattr    = "SCHILY.xattr."
)

// FileInfoHeader creates a partially-populated Header from fi.
// If fi describes a symlink, FileInfoHeader records link as the link target.
// If fi describes a directory, a slash is appended to the name.
// Because os.FileInfo's Name method returns only the base name of
// the file it describes, it may be necessary to modify the Name field
// of the returned header to provide the full path name of the file.
func FileInfoHeader(fi os.FileInfo, link string) (*Header, error) {
	if fi == nil {
		return nil, errors.New("tar: FileInfo is nil")
	}
	fm := fi.Mode()
	h := &Header{
		Name:    fi.Name(),
		ModTime: fi.ModTime(),
		Mode:    int64(fm.Perm()), // or'd with c_IS* constants later
	}
	switch {
	case fm.IsRegular():
		h.Mode |= c_ISREG
		h.Typeflag = TypeReg
		h.Size = fi.Size()
	case fi.IsDir():
		h.Typeflag = TypeDir
		h.Mode |= c_ISDIR
		h.Name += "/"
	case fm&os.ModeSymlink != 0:
		h.Typeflag = TypeSymlink
		h.Mode |= c_ISLNK
		h.Linkname = link
	case fm&os.ModeDevice != 0:
		if fm&os.ModeCharDevice != 0 {
			h.Mode |= c_ISCHR
			h.Typeflag = TypeChar
		} else {
			h.Mode |= c_ISBLK
			h.Typeflag = TypeBlock
		}
	case fm&os.ModeNamedPipe != 0:
		h.Typeflag = TypeFifo
		h.Mode |= c_ISFIFO
	case fm&os.ModeSocket != 0:
		h.Mode |= c_ISSOCK
	default:
		return nil, fmt.Errorf("archive/tar: unknown file mode %v", fm)
	}
	if fm&os.ModeSetuid != 0 {
		h.Mode |= c_ISUID
	}
	if fm&os.ModeSetgid != 0 {
		h.Mode |= c_ISGID
	}
	if fm&os.ModeSticky != 0 {
		h.Mode |= c_ISVTX
	}
	// If possible, populate additional fields from OS-specific
	// FileInfo fields.
	if sys, ok := fi.Sys().(*Header); ok {
		// This FileInfo came from a Header (not the OS). Use the
		// original Header to populate all remaining fields.
		h.Uid = sys.Uid
		h.Gid = sys.Gid
		h.Uname = sys.Uname
		h.Gname = sys.Gname
		h.AccessTime = sys.AccessTime
		h.ChangeTime = sys.ChangeTime
		if sys.Xattrs != nil {
			h.Xattrs = make(map[string]string)
			for k, v := range sys.Xattrs {
				h.Xattrs[k] = v
			}
		}
		if sys.Typeflag == TypeLink {
			// hard link
			h.Typeflag = TypeLink
			h.Size = 0
			h.Linkname = sys.Linkname
		}
	}
	if sysStat != nil {
		return h, sysStat(fi, h)
	}
	return h, nil
}

var zeroBlock = make([]byte, blockSize)

// isHeaderOnlyType checks if the given type flag is of the type that has no
// data section even if a size is specified.
func isHeaderOnlyType(tf byte) bool {
	switch tf {
	case TypeLink, TypeSymlink, TypeChar, TypeBlock, TypeDir, TypeFifo:
		return true
	default:
		return false
	}
}

// POSIX specifies a sum of the unsigned byte values, but the Sun tar uses signed byte values.
// We compute and return both.
func checksum(header []byte) (unsigned int64, signed int64) {
	for i := 0; i < len(header); i++ {
		if i == 148 {
			// The chksum field (header[148:156]) is special: it should be treated as space bytes.
			unsigned += ' ' * 8
			signed += ' ' * 8
			i += 7
			continue
		}
		unsigned += int64(header[i])
		signed += int64(int8(header[i]))
	}
	return
}

type slicer []byte

func (sp *slicer) next(n int) (b []byte) {
	s := *sp
	b, *sp = s[0:n], s[n:]
	return
}

func isASCII(s string) bool {
	for _, c := range s {
		if c >= 0x80 {
			return false
		}
	}
	return true
}

func toASCII(s string) string {
	if isASCII(s) {
		return s
	}
	var buf bytes.Buffer
	for _, c := range s {
		if c < 0x80 {
			buf.WriteByte(byte(c))
		}
	}
	return buf.String()
}
