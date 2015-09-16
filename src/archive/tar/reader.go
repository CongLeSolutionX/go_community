// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tar

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// Implementation specific notes about the decoding process.
//
// At a high level, this package supports reading the following formats:
//	V7, GNU, STAR, USTAR, and PAX.
//
// In order to accept a wide range of formats, this library is intentionally
// liberal with the values and formats it accepts. Thus, it will even accept
// formats that are a hybrid mix of one of the formats mentioned above.
//
// Some specific features supported:
//	* GNU extended types including GNUTYPE_LONGNAME(L) and GNUTYPE_LONGLINK(K).
//	This allows for arbitrarily long paths and linkpaths for the GNU format.
//	* GNU sparse formats. All known formats including the old GNU format, the
//	PAX 0.0, 0.1, and 1.0 formats are supported.
//	* GNU base-256 encoding. Most numeric fields in headers may use base-256
//	encoding even if the overall format is not GNU. This increases the effective
//	bit-width of those fields by 2.667x bits.
//	* Negative timestamps. The use of base-256 provides the means to read
//	negative values from fields. This allows for timestamps occurring before the
//	Unix epoch. This works for all formats.
//	* PAX local headers. This allows for arbitrarily large fields for nearly
//	every field that Header contains and allows for the storage of UTF-8
//	characters in path names. PAX headers are decoded even if the underlying
//	format is not PAX or USTAR.
//	* PAX sub-second time resolution. Any time field encoded using PAX will
//	support exact time resolution into the nanoseconds.
//
// Some specific features *not* supported:
//	* GNU base-64 encoding. For a brief period, GNU used base-64 to extend the
//	range of numeric fields.
//	* GNU extended types including GNUTYPE_DUMPDIR(D), GNUTYPE_MULTIVOL(M), and
//	GNUTYPE_VOLHDR(V). Access to these will be provided to the user as a
//	readable file with the appropriate type flag set.
//	* STAR sparse format. The STAR format has provisions for a sparse format
//	similar to the GNU sparse format and even uses the same type flag. The only
//	way to tell them apart is a magic value in the last 4 bytes of the block.
//	* PAX global headers. Inclusion to the PAX standard itself was controversial
//	and this feature is neither widely supported nor used. If this is
//	encountered, the header's contents can still be read as a normal file.
//	It will be the user's responsibility to parse and apply them.
//	* PAX extended headers like "comment", "charset", "hdrcharset", etc.
//	These headers are currently ignored and there is no way to access them.
//
// In the degenerate case where multiple PAX and GNU headers are used together
// for a given archive member, then it is not specified which header will
// ultimately be used to populate the fields. Any archive that does this is
// neither compliant with the PAX nor GNU formats and the behavior is undefined.
// The GNU and BSD tar programs operate in different ways in this situation, but
// none of them crash. Most importantly, multiple special headers preceding a
// normal archive member should be coalesced together, so that parsing of
// subsequent members occur without issue. For more information, see the
// TestMultipleHeaders unit test.
//
// Normally, a tar file is terminated using 2 zero blocks after reading the data
// and padding for a archive member. However, to support various legacy and
// buggy tar writers, this library will consider a tar stream to have ended
// without error under the following conditions:
//	* At least 2 blocks of zeros is found (normal case).
//	* Exactly 1 block of zeros is found and EOF is hit in input io.Reader.
//	* EOF is hit in the input io.Reader in the padding after the data.
//
// Edge case 1: An empty stream is considered a "valid" tar file.
// This behavior is consistent with BSD, but inconsistent with GNU.
//
// Edge case 2: The special types TypeXHeader, TypeGNULongName, and
// TypeGNULongLink are technically "files" that modify the fields of the next
// archive member. Normally, another member would be expected, but according
// to the conditions above, this will stop the stream without error.
// This behavior is consistent with GNU, but inconsistent with BSD.

// Errors returned while decoding.
var (
	ErrHeader = errors.New("archive/tar: invalid tar header")
)

// A Reader provides sequential access to the contents of a tar archive, which
// represents a sequence of files.
//
// The Next method advances to the next file in the archive (including the
// first), and then it can be treated as an io.Reader to access the file's data.
type Reader struct {
	r       io.Reader
	err     error
	pad     int64           // amount of padding (ignored) after current file entry
	curr    numBytesReader  // reader for current file entry
	hdrBuff [blockSize]byte // buffer to use in readHeader
}

// A numBytesReader is an io.Reader with a numBytes method, returning the number
// of bytes remaining in the underlying encoded data.
type numBytesReader interface {
	io.Reader
	numBytes() int64
}

// A regFileReader is a numBytesReader for reading file data from a tar archive.
type regFileReader struct {
	r  io.Reader // underlying reader
	nb int64     // number of unread bytes for current file entry
}

// A sparseFileReader is a numBytesReader for reading sparse file data from a
// tar archive.
type sparseFileReader struct {
	rfr numBytesReader // reads the sparse-encoded file data
	sp  []sparseEntry  // the sparse map for the file
	pos int64          // keeps track of file position
	tot int64          // total size of the file
}

// Keywords for GNU sparse files in a PAX extended header.
const (
	paxGNUSparseNumBlocks = "GNU.sparse.numblocks"
	paxGNUSparseOffset    = "GNU.sparse.offset"
	paxGNUSparseNumBytes  = "GNU.sparse.numbytes"
	paxGNUSparseMap       = "GNU.sparse.map"
	paxGNUSparseName      = "GNU.sparse.name"
	paxGNUSparseMajor     = "GNU.sparse.major"
	paxGNUSparseMinor     = "GNU.sparse.minor"
	paxGNUSparseSize      = "GNU.sparse.size"
	paxGNUSparseRealSize  = "GNU.sparse.realsize"
)

// NewReader creates a new Reader reading from r.
func NewReader(r io.Reader) *Reader { return &Reader{r: r} }

// Next advances to the next entry in the tar archive. Any remaining data in
// the current file will be automatically discarded.
//
// io.EOF is returned at the end of the archive stream.
//
// Certain special types like TypeLink, TypeSymLink, TypeChar, TypeBlock,
// TypeDir, and TypeFifo may have the Header.Size field set to a non-zero value.
// Even if this happens, attempting to call Read while on these "files" will
// cause io.EOF to be immediately returned.
func (tr *Reader) Next() (*Header, error) {
	var format int
	var hdr *Header
	var rawHdr []byte
	var extHdrs map[string]string

	// Externally, Next iterates through the tar archive as if it is a series of
	// files. Internally, the tar format often uses fake "files" to add meta
	// data that describes the next file. These meta data "files" should not
	// normally be visible to the outside. As such, this loop iterates through
	// one or more "header files" until it finds a "normal file".
loop:
	for {
		tr.skipUnread()
		if tr.err != nil {
			return nil, tr.err
		}

		format, hdr, rawHdr = tr.readHeader()
		if tr.err != nil {
			return nil, tr.err
		}

		tr.err = tr.handleRegularFile(hdr)
		if tr.err != nil {
			return nil, tr.err
		}

		// Check for PAX/GNU special headers and files.
		switch hdr.Typeflag {
		case TypeXHeader:
			extHdrs, tr.err = parsePAX(tr, extHdrs)
			if tr.err != nil {
				return nil, tr.err
			}
			continue loop // This is a meta header affecting the next header
		case TypeGNULongName, TypeGNULongLink:
			// Do not assert that the format is GNU here since there is no
			// structure to the file data.

			var realname []byte
			realname, tr.err = ioutil.ReadAll(tr)
			if tr.err != nil {
				return nil, tr.err
			}

			// Convert GNU extensions to use PAX headers.
			if extHdrs == nil {
				extHdrs = make(map[string]string)
			}
			switch hdr.Typeflag {
			case TypeGNULongName:
				extHdrs[paxPath] = cString(realname)
			case TypeGNULongLink:
				extHdrs[paxLinkpath] = cString(realname)
			}
			continue loop // This is a meta header affecting the next header
		default:
			tr.err = mergePAX(hdr, extHdrs)
			if tr.err != nil {
				return nil, tr.err
			}

			// The extended headers may have updated the size. Thus, we must
			// setup the regFileReader again after merging the PAX headers.
			tr.err = tr.handleRegularFile(hdr)
			if tr.err != nil {
				return nil, tr.err
			}

			// Sparse formats rely on being able to read from the logical data
			// section; there must be a preceding call to handleRegularFile.
			tr.err = tr.handleSparseFile(format, hdr, rawHdr, extHdrs)
			if tr.err != nil {
				return nil, tr.err
			}
			break loop // This is a file, so stop
		}
	}
	return hdr, nil
}

// handleRegularFile sets up the current file reader and padding such that it
// can only read the following logical data section. It will properly handle
// special headers that contain no data section.
func (tr *Reader) handleRegularFile(hdr *Header) error {
	nb := hdr.Size
	if isHeaderOnlyType(hdr.Typeflag) {
		nb = 0
	}

	tr.pad = -nb & (blockSize - 1) // blockSize is a power of two
	tr.curr, tr.err = newRegFileReader(tr.r, nb)
	return tr.err
}

// handleSparseFile checks if the current file is a sparse format of any type
// and sets the curr reader appropriately.
func (tr *Reader) handleSparseFile(format int, hdr *Header, rawHdr []byte, extHdrs map[string]string) error {
	var sp []sparseEntry
	if hdr.Typeflag == TypeGNUSparse {
		sp, tr.err = tr.readOldGNUSparseMap(tr.r, format, hdr, rawHdr)
	} else {
		sp, tr.err = tr.checkForGNUSparsePAXHeaders(hdr, extHdrs)
	}
	if tr.err != nil {
		return tr.err
	}

	if len(sp) > 0 {
		// Sparse files do not make sense when applied to the special header
		// types that never have a data section.
		if isHeaderOnlyType(hdr.Typeflag) {
			tr.err = ErrHeader
			return tr.err
		}
		tr.curr, tr.err = newSparseFileReader(tr.curr, sp, hdr.Size)
	}
	return tr.err
}

// checkForGNUSparsePAXHeaders checks the PAX headers for GNU sparse headers. If they are found, then
// this function reads the sparse map and returns it. Unknown sparse formats are ignored, causing the file to
// be treated as a regular file.
func (tr *Reader) checkForGNUSparsePAXHeaders(hdr *Header, extHdrs map[string]string) ([]sparseEntry, error) {
	var sparseFormat string

	// Check for sparse format indicators.
	major, majorOk := extHdrs[paxGNUSparseMajor]
	minor, minorOk := extHdrs[paxGNUSparseMinor]
	sparseName, sparseNameOk := extHdrs[paxGNUSparseName]
	_, sparseMapOk := extHdrs[paxGNUSparseMap]
	sparseSize, sparseSizeOk := extHdrs[paxGNUSparseSize]
	sparseRealSize, sparseRealSizeOk := extHdrs[paxGNUSparseRealSize]

	// Identify which sparse format applies based on which PAX headers are set.
	if majorOk && minorOk {
		sparseFormat = major + "." + minor
	} else if sparseNameOk && sparseMapOk {
		sparseFormat = "0.1"
	} else if sparseSizeOk {
		sparseFormat = "0.0"
	} else {
		// Not a PAX format GNU sparse file, so quit without error.
		return nil, nil
	}

	// Check for unknown sparse format.
	if sparseFormat != "0.0" && sparseFormat != "0.1" && sparseFormat != "1.0" {
		return nil, ErrHeader
	}

	// Update hdr from GNU sparse PAX headers.
	if sparseNameOk {
		hdr.Name = sparseName
	}
	switch {
	case sparseSizeOk:
		hdr.Size, tr.err = strconv.ParseInt(sparseSize, 10, 64)
	case sparseRealSizeOk:
		hdr.Size, tr.err = strconv.ParseInt(sparseRealSize, 10, 64)
	}
	if tr.err != nil || hdr.Size < 0 {
		tr.err = ErrHeader
		return nil, tr.err
	}

	// Set up the sparse map, according to the particular sparse format in use.
	var sp []sparseEntry
	switch sparseFormat {
	case "0.0", "0.1":
		sp, tr.err = readGNUSparseMap0x1(extHdrs)
	case "1.0":
		sp, tr.err = readGNUSparseMap1x0(tr.curr)
	}
	return sp, tr.err
}

// mergePAX merges well known headers according to PAX standard.
// In general values found in the headers map overwrite those with the same
// name as those found in Header.
func mergePAX(hdr *Header, extHdrs map[string]string) error {
	for k, v := range extHdrs {
		switch k {
		case paxPath:
			hdr.Name = v
		case paxLinkpath:
			hdr.Linkname = v
		case paxUname:
			hdr.Uname = v
		case paxGname:
			hdr.Gname = v
		case paxUid:
			uid, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return ErrHeader
			}
			hdr.Uid = int(uid) // Integer overflow possible
		case paxGid:
			gid, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return ErrHeader
			}
			hdr.Gid = int(gid) // Integer overflow possible
		case paxAtime:
			t, err := parsePAXTime(v)
			if err != nil {
				return err
			}
			hdr.AccessTime = t
		case paxMtime:
			t, err := parsePAXTime(v)
			if err != nil {
				return err
			}
			hdr.ModTime = t
		case paxCtime:
			t, err := parsePAXTime(v)
			if err != nil {
				return err
			}
			hdr.ChangeTime = t
		case paxSize:
			size, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return ErrHeader
			}
			hdr.Size = size
		default:
			if strings.HasPrefix(k, paxXattr) {
				if hdr.Xattrs == nil {
					hdr.Xattrs = make(map[string]string)
				}
				hdr.Xattrs[k[len(paxXattr):]] = v
			}
		}
	}
	return nil
}

// parsePAXTime takes a string of the form %d.%d as described in the PAX
// specification. Note that this implementation allows for negative timestamps,
// which is allowed for by the PAX specification, but not always portable.
func parsePAXTime(s string) (ts time.Time, err error) {
	var ss, sn string // Strings for seconds and nanoseconds parts
	var secs, nsecs int64

	// Split string into two components.
	ss, sn = s, ""
	pos := strings.IndexByte(s, '.')
	if pos != -1 {
		ss, sn = s[:pos], s[pos+1:]
	}

	// Parse the seconds.
	if len(ss) > 0 {
		secs, err = strconv.ParseInt(ss, 10, 64)
		if err != nil {
			return ts, ErrHeader
		}
	}

	// Parse the nanoseconds.
	if len(sn) > 0 {
		// String must be entirely comprised of digits. This ensures both that
		// the number is positive and that there are no invalid characters
		// following the truncation point.
		if strings.Trim(sn, "0123456789") != "" {
			return ts, ErrHeader
		}

		// Pad with zeros if too short and truncate extra digits if too long.
		//
		// Example:
		//	0.3              =>  0.300000000  =>  300000000ns
		//	0.1234567890123  =>  0.123456789  =>  123456789ns
		const maxNanoSecondDigits = 9
		if len(sn) < maxNanoSecondDigits {
			sn += strings.Repeat("0", maxNanoSecondDigits-len(sn))
		} else {
			sn = sn[:maxNanoSecondDigits]
		}

		// If seconds is negative, correct the nanoseconds.
		nsecs, _ = strconv.ParseInt(sn, 10, 64) // Never fails
		if secs < 0 {
			nsecs *= -1
		}
	}

	return time.Unix(secs, nsecs), nil
}

// parsePAX parses PAX headers. As an input, it takes in existing extHdrs and
// updates it accordingly. It is not standard PAX behavior for multiple PAX
// headers to precede a given file, but many tar programs accept this practice.
func parsePAX(r io.Reader, extHdrs map[string]string) (map[string]string, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return extHdrs, err
	}
	sbuf := string(buf)

	// For GNU PAX sparse format 0.0 support.
	// This function transforms the sparse format 0.0 headers into format 0.1
	// headers since 0.0 headers were not PAX compliant.
	var sparseHeaders = [2]string{paxGNUSparseOffset, paxGNUSparseNumBytes}
	var sparseMap []string

	// Each record is constructed in the following format:
	//	"%d %s=%s\n", length, keyword, value
	//
	// Where the keyword and value are both encoded using UTF-8, while all other
	// fields and tokens use ASCII. The key and value may be comprised of any
	// arbitrary UTF-8 string except that the key may not contain an '='.
	if extHdrs == nil {
		extHdrs = make(map[string]string)
	}
	for len(sbuf) > 0 {
		// The size field ends at the first space.
		sp := strings.IndexByte(sbuf, ' ')
		if sp == -1 {
			return extHdrs, ErrHeader
		}

		// Parse the first token as a decimal integer.
		n, err := strconv.ParseInt(sbuf[:sp], 10, 0) // Intentionally parse as native int
		if err != nil || n < 5 || int64(len(sbuf)) < n {
			return extHdrs, ErrHeader
		}

		// Extract everything between the space and the final newline.
		var record, endline string
		record, endline, sbuf = sbuf[sp+1:n-1], sbuf[n-1:n], sbuf[n:]
		if endline != "\n" {
			return extHdrs, ErrHeader
		}

		// The first equals separates the key from the value.
		eq := strings.IndexByte(record, '=')
		if eq == -1 {
			return extHdrs, ErrHeader
		}
		key, value := record[:eq], record[eq+1:]

		// Store the key, value pair.
		switch key {
		case paxGNUSparseOffset, paxGNUSparseNumBytes:
			// Check that the GNU sparse headers are in the right order.
			if key != sparseHeaders[len(sparseMap)%2] {
				return extHdrs, ErrHeader
			}
			sparseMap = append(sparseMap, value)
		default:
			// According to PAX specification, a value is stored only if it is
			// non-empty. Otherwise, the key is deleted.
			if len(value) > 0 {
				extHdrs[key] = value
			} else {
				delete(extHdrs, key)
			}
		}
	}

	if len(sparseMap) > 0 {
		extHdrs[paxGNUSparseMap] = strings.Join(sparseMap, ",")
	}
	return extHdrs, nil
}

// cString parses bytes as a NULL-terminated C-style string.
// If a NULL byte is not found then the whole slice is returned as a string.
func cString(b []byte) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	return string(b[:n])
}

// numeric parses the input as being encoded in either base-256 or octal.
// This function may return negative numbers.
// If parsing fails or an integer overflow occurs, err will be set.
func (tr *Reader) numeric(b []byte) int64 {
	// Check for base-256 format first.
	// If the first bit is set, then all following bits constitute a two's
	// complement encoded number in big-endian byte order.
	if len(b) > 0 && b[0]&0x80 != 0 {
		// Handling negative numbers relies on the following identity:
		//	-a-1 == ^a
		//
		// If the number is negative, we use an inversion mask to invert the
		// data bytes and treat the value as an unsigned number. We then use
		// sign and incr to perform the identity above.
		var inv, sign, incr = byte(0x00), int64(+1), int64(0)
		if b[0]&0x40 != 0 {
			inv, sign, incr = 0xff, -1, 1
		}

		var x uint64
		for i, c := range b {
			c ^= inv
			if i == 0 {
				c &= 0x7f // Ignore signal bit in first byte
			}
			if (x >> 56) > 0 {
				tr.err = ErrHeader // Integer overflow
			}
			x = x<<8 | uint64(c)
		}
		if (x >> 63) > 0 {
			tr.err = ErrHeader // Integer overflow
		}
		return sign*int64(x) - incr
	}

	// Normal case is octal format.
	return tr.octal(b)
}

// octal parses the input as an octal encoded value.
// This function may return negative numbers.
// If parsing fails or an integer overflow occurs, err will be set.
func (tr *Reader) octal(b []byte) int64 {
	// Because unused fields are filled with NULLs, we need to skip leading
	// NULLs. Fields may also be padded with spaces or NULLs.
	// So remove leading and trailing NULLs and spaces to be sure.
	b = bytes.Trim(b, " \x00")
	if len(b) == 0 {
		return 0
	}
	x, err := strconv.ParseInt(cString(b), 8, 64)
	if err != nil {
		tr.err = ErrHeader
	}
	return x
}

// skipUnread skips any unread bytes in the existing file entry, as well as any
// alignment padding. It will set err to io.ErrUnexpectedEOF if any io.EOF is
// encountered in the data portion; it is okay to hit io.EOF in the padding.
//
// Note that this function still works properly even when sparse files are being
// used since numBytes returns the bytes remaining in the underlying io.Reader.
func (tr *Reader) skipUnread() {
	if tr.err != nil {
		return
	}

	var nd, nb, nr1, nr2 int64
	nd = tr.numBytes()
	nb = nd + tr.pad // Total number of bytes to skip
	tr.curr, tr.pad = nil, 0

	// If possible, Seek to the last byte before the end of the data section.
	// Do this because Seek is often lazy about reporting errors; this will mask
	// the fact that the tar stream may be truncated. We can rely on the
	// io.CopyN done shortly afterwards to trigger any IO errors.
	if sr, ok := tr.r.(io.Seeker); ok && nd > 1 {
		pos1, _ := sr.Seek(0, os.SEEK_CUR)
		pos2, _ := sr.Seek(nd-1, os.SEEK_CUR)
		nr1 = pos2 - pos1 // Number of bytes skipped via Seek
	}

	nr2, tr.err = io.CopyN(ioutil.Discard, tr.r, nb-nr1)
	if tr.err == io.EOF && nr1+nr2 < nd {
		tr.err = io.ErrUnexpectedEOF
	}
}

func (tr *Reader) verifyChecksum(rawHdr []byte) bool {
	if tr.err != nil {
		return false
	}

	given := tr.numeric(rawHdr[148:156])
	unsigned, signed := checksum(rawHdr)
	return given == unsigned || given == signed
}

// readHeader reads the next block header and assumes that the underlying reader
// is already aligned to a block boundary.
//
// This function returns the format, the parsed Header, and also the raw header
// block for further format specific parsing.
// This method is guaranteed to set err if no Header is returned.
//
// The err will be set to io.EOF only when the following occurs:
//	* Exactly 0 bytes are read and EOF is hit.
//	* Exactly 1 block of zeros is read and EOF is hit.
//	* At least 2 blocks of zeros are read.
func (tr *Reader) readHeader() (format int, hdr *Header, rawHdr []byte) {
	rawHdr = tr.hdrBuff[:]
	copy(rawHdr, zeroBlock)
	if _, tr.err = io.ReadFull(tr.r, rawHdr); tr.err != nil {
		return formatUnknown, nil, nil // io.EOF is okay here
	}

	// Two blocks of zero bytes marks the normal end of an archive.
	if bytes.Equal(rawHdr, zeroBlock) {
		if _, tr.err = io.ReadFull(tr.r, rawHdr); tr.err != nil {
			return formatUnknown, nil, nil // io.EOF is okay here
		}
		if bytes.Equal(rawHdr, zeroBlock) {
			tr.err = io.EOF
		} else {
			tr.err = ErrHeader // Zero block and then non-zero block
		}
		return formatUnknown, nil, nil
	}

	if !tr.verifyChecksum(rawHdr) {
		tr.err = ErrHeader
		return formatUnknown, nil, nil
	}

	hdr = new(Header)
	s := slicer(rawHdr)

	// Parse the V7 header.
	format = formatV7
	hdr.Name = cString(s.next(100))
	hdr.Mode = tr.octal(s.next(8))
	hdr.Uid = int(tr.numeric(s.next(8))) // Integer overflow possible
	hdr.Gid = int(tr.numeric(s.next(8))) // Integer overflow possible
	hdr.Size = tr.numeric(s.next(12))
	hdr.ModTime = time.Unix(tr.numeric(s.next(12)), 0)
	s.next(8) // chksum
	hdr.Typeflag = s.next(1)[0]
	hdr.Linkname = cString(s.next(100))

	// Make sure mode is encodable by USTAR in octal format, which is composed
	// of 7 ASCII characters without trailing NULL.
	if hdr.Mode != hdr.Mode&07777777 {
		tr.err = ErrHeader
		return formatUnknown, nil, nil
	}

	// The remainder of the header depends on the value of magic.
	// The original (V7) version of tar had no explicit magic field,
	// so its magic bytes, like the rest of the block, are NULLs.
	magic := string(s.next(8)) // Contains version field as well
	switch {
	case magic[:6] == magicUSTAR:
		if string(rawHdr[508:512]) == magicSTAR {
			format = formatSTAR
		} else {
			format = formatUSTAR
		}
	case magic == magicGNU:
		format = formatGNU
	}

	switch format {
	case formatUSTAR, formatGNU, formatSTAR:
		hdr.Uname = cString(s.next(32))
		hdr.Gname = cString(s.next(32))
		devmajor := s.next(8)
		devminor := s.next(8)
		if hdr.Typeflag == TypeChar || hdr.Typeflag == TypeBlock {
			hdr.Devmajor = tr.numeric(devmajor)
			hdr.Devminor = tr.numeric(devminor)
		}

		var prefix string
		switch format {
		case formatUSTAR:
			prefix = cString(s.next(155))
		case formatSTAR, formatGNU:
			if format == formatSTAR {
				prefix = cString(s.next(131))
			}

			atime := s.next(12)
			ctime := s.next(12)

			// The atime and ctime fields are often left unused. The Header
			// struct should reflect this fact by using the zero values.
			if !bytes.Equal(atime, zeroBlock[:12]) {
				hdr.AccessTime = time.Unix(tr.numeric(atime), 0)
			}
			if !bytes.Equal(ctime, zeroBlock[:12]) {
				hdr.ChangeTime = time.Unix(tr.numeric(ctime), 0)
			}
		}
		if len(prefix) > 0 {
			hdr.Name = prefix + "/" + hdr.Name
		}
	}

	// Check for any numeric parsing errors at the end.
	if tr.err != nil {
		tr.err = ErrHeader
		return formatUnknown, nil, nil
	}
	return format, hdr, rawHdr
}

// A sparseEntry holds a single entry in a sparse file's sparse map.
// A sparse entry indicates the offset and size in a sparse file of a
// block of data.
type sparseEntry struct {
	offset   int64
	numBytes int64
}

// readOldGNUSparseMap reads the sparse map from the old GNU sparse format.
// The sparse map is stored in the tar header if it's small enough.
// If it's larger than four entries, then one or more extension headers are used
// to store the rest of the sparse map.
//
// The Header.Size does not reflect the size of any extended headers used.
// Thus, this function will read from the raw io.Reader to fetch extra headers.
func (tr *Reader) readOldGNUSparseMap(r io.Reader, format int, hdr *Header, rawHdr []byte) ([]sparseEntry, error) {
	// Constants relevant to parsing old GNU headers.
	const (
		mainHeaderNumEntries       = 4
		mainHeaderArrayOffset      = 386
		mainHeaderIsExtendedOffset = 482
		mainHeaderSizeOffset       = 483

		extendedHeaderNumEntries       = 21
		extendedHeaderArrayOffset      = 0
		extendedHeaderIsExtendedOffset = 504

		numericSize = 12
	)

	// Make sure that the input format is GNU.
	// Unfortunately, the STAR format also has a sparse header format that uses
	// the same type flag but has a completely different layout.
	if format != formatGNU {
		tr.err = ErrHeader
		return nil, tr.err
	}

	// Get the real size of the file.
	sizeStr := rawHdr[mainHeaderSizeOffset : mainHeaderSizeOffset+numericSize]
	hdr.Size = tr.numeric(sizeStr)
	if tr.err != nil || hdr.Size < 0 {
		tr.err = ErrHeader
		return nil, tr.err
	}

	// First loop treats header as the GNU "main" header.
	var sp []sparseEntry
	var blk = append([]byte(nil), rawHdr...) // Copy of header
	var numEntries = mainHeaderNumEntries
	var extOffset = mainHeaderIsExtendedOffset
	var s = slicer(blk[mainHeaderArrayOffset:])
loop:
	for {
		// Parse each sparse entry.
		for i := 0; i < numEntries; i++ {
			offset := tr.numeric(s.next(numericSize))
			numBytes := tr.numeric(s.next(numericSize))
			if tr.err != nil {
				tr.err = ErrHeader
				return nil, tr.err
			}
			if offset == 0 && numBytes == 0 {
				break loop
			}
			sp = append(sp, sparseEntry{offset: offset, numBytes: numBytes})
		}

		if isExtended := blk[extOffset] > 0; isExtended {
			// Fetch the next block.
			if _, tr.err = io.ReadFull(r, blk); tr.err != nil {
				if tr.err == io.EOF {
					tr.err = io.ErrUnexpectedEOF
				}
				return nil, tr.err
			}

			// All subsequent loops treat the header as a GNU "extended" header.
			numEntries = extendedHeaderNumEntries
			extOffset = extendedHeaderIsExtendedOffset
			s = slicer(blk[extendedHeaderArrayOffset:])
		} else {
			break loop
		}
	}
	return sp, nil
}

// readGNUSparseMap1x0 reads the sparse map as stored in GNU's PAX sparse format
// version 1.0. The sparse map is stored just before the file data and padded
// out to the nearest block boundary.
//
// Note that the GNU manual says that numeric values should be encoded in octal
// format. However, the GNU tar utility itself outputs these values in decimal.
// As such, this library treats values as being encoded in decimal.
func readGNUSparseMap1x0(r io.Reader) ([]sparseEntry, error) {
	var cntNewline int64
	var buf bytes.Buffer
	var blk = make([]byte, blockSize)

	// feedTokens guarantees that at least cnt newlines exist in buf if there
	// are no errors encountered.
	var feedTokens = func(cnt int64) error {
		for cntNewline < cnt {
			if _, err := io.ReadFull(r, blk); err != nil {
				if err == io.EOF {
					err = io.ErrUnexpectedEOF
				}
				return err
			}
			buf.Write(blk)
			cntNewline += int64(bytes.Count(blk, []byte("\n")))
		}
		return nil
	}

	// Get the next token delimited by a newline. This assumes that at least one
	// newline exists in the buffer.
	var nextToken = func() string {
		cntNewline--
		tok, _ := buf.ReadString('\n')
		return tok[:len(tok)-1] // Cut off newline
	}

	// Parse for the number of entries.
	// Use integer overflow resistant math to check this.
	if err := feedTokens(1); err != nil {
		return nil, err
	}
	numEntries, err := strconv.ParseInt(nextToken(), 10, 0) // Intentionally parse as native int
	if err != nil || numEntries < 0 || int(numEntries*2) < int(numEntries) {
		return nil, ErrHeader
	}

	// Parse for all member entries.
	if err := feedTokens(numEntries * 2); err != nil {
		return nil, err
	}
	sp := make([]sparseEntry, 0, numEntries) // numEntries is trusted now
	for i := int64(0); i < numEntries; i++ {
		offset, err := strconv.ParseInt(nextToken(), 10, 64)
		if err != nil {
			return nil, ErrHeader
		}
		numBytes, err := strconv.ParseInt(nextToken(), 10, 64)
		if err != nil {
			return nil, ErrHeader
		}
		sp = append(sp, sparseEntry{offset: offset, numBytes: numBytes})
	}
	return sp, nil
}

// readGNUSparseMap0x1 reads the sparse map as stored in GNU's PAX sparse format
// version 0.1. The sparse map is stored in the PAX headers.
func readGNUSparseMap0x1(extHdrs map[string]string) ([]sparseEntry, error) {
	// Get number of entries.
	// Use integer overflow resistant math to check this.
	numEntriesStr := extHdrs[paxGNUSparseNumBlocks]
	numEntries, err := strconv.ParseInt(numEntriesStr, 10, 0) // Intentionally parse as native int
	if err != nil || numEntries < 0 || int(numEntries*2) < int(numEntries) {
		return nil, ErrHeader
	}

	// There should be two numbers in sparseMap for each entry.
	sparseMap := strings.Split(extHdrs[paxGNUSparseMap], ",")
	if len(sparseMap) != int(numEntries*2) {
		return nil, ErrHeader
	}

	// Loop through the entries in the sparse map.
	sp := make([]sparseEntry, 0, numEntries) // numEntries is trusted now
	for i := int64(0); i < numEntries; i++ {
		offset, err := strconv.ParseInt(sparseMap[2*i], 10, 64)
		if err != nil {
			return nil, ErrHeader
		}
		numBytes, err := strconv.ParseInt(sparseMap[2*i+1], 10, 64)
		if err != nil {
			return nil, ErrHeader
		}
		sp = append(sp, sparseEntry{offset: offset, numBytes: numBytes})
	}
	return sp, nil
}

// numBytes returns the number of bytes left to read in the current file's entry
// in the tar archive, or 0 if there is no current file.
func (tr *Reader) numBytes() int64 {
	if tr.curr == nil {
		return 0 // No current file, so no bytes
	}
	return tr.curr.numBytes()
}

// Read reads from the current entry in the tar archive.
// It returns 0, io.EOF when it reaches the end of that entry,
// until Next is called to advance to the next entry.
func (tr *Reader) Read(b []byte) (n int, err error) {
	if tr.curr == nil {
		return 0, io.EOF
	}
	n, err = tr.curr.Read(b)
	if err != nil && err != io.EOF {
		tr.err = err
	}
	return
}

func newRegFileReader(r io.Reader, nb int64) (*regFileReader, error) {
	if nb < 0 {
		return nil, ErrHeader
	}
	return &regFileReader{r: r, nb: nb}, nil
}

func (rfr *regFileReader) Read(b []byte) (n int, err error) {
	if rfr.nb == 0 {
		return 0, io.EOF // File consumed
	}
	if int64(len(b)) > rfr.nb {
		b = b[:rfr.nb]
	}
	n, err = rfr.r.Read(b)
	rfr.nb -= int64(n)

	if err == io.EOF && rfr.nb > 0 {
		err = io.ErrUnexpectedEOF
	}
	return
}

// numBytes returns the number of bytes left to read from the underlying
// archive stream. This is not the perceived size from sparse files.
func (rfr *regFileReader) numBytes() int64 {
	return rfr.nb
}

// newSparseFileReader creates a new sparseFileReader, but validates all of the
// sparse entries before doing so.
func newSparseFileReader(rfr numBytesReader, sp []sparseEntry, tot int64) (*sparseFileReader, error) {
	if tot < 0 {
		return nil, ErrHeader // Total size cannot be negative
	}

	// Validate all sparse entries. These are the same checks as performed by
	// the BSD tar utility.
	for i, s := range sp {
		switch {
		case s.offset < 0 || s.numBytes < 0:
			return nil, ErrHeader // Negative values are never okay
		case s.offset > math.MaxInt64-s.numBytes:
			return nil, ErrHeader // Integer overflow with large length
		case s.offset+s.numBytes > tot:
			return nil, ErrHeader // Region extends beyond the "real" size
		case i > 0 && sp[i-1].offset+sp[i-1].numBytes > s.offset:
			return nil, ErrHeader // Regions can't overlap
		}
	}
	return &sparseFileReader{rfr: rfr, sp: sp, tot: tot}, nil
}

// readHole reads a sparse hole ending at endOffset.
func (sfr *sparseFileReader) readHole(b []byte, endOffset int64) int {
	n64 := endOffset - sfr.pos
	if n64 > int64(len(b)) {
		n64 = int64(len(b))
	}
	n := int(n64)
	for i := 0; i < n; i++ {
		b[i] = 0
	}
	sfr.pos += n64
	return n
}

// Read reads the sparse file data in expanded form.
func (sfr *sparseFileReader) Read(b []byte) (n int, err error) {
	// Skip past all empty fragments.
	for len(sfr.sp) > 0 && sfr.sp[0].numBytes == 0 {
		sfr.sp = sfr.sp[1:]
	}

	// If there are not more fragments, then it is possible that there
	// is one last sparse hole.
	if len(sfr.sp) == 0 {
		if sfr.pos < sfr.tot {
			return sfr.readHole(b, sfr.tot), nil
		}
		return 0, io.EOF
	}

	// In front of a data fragment, so read a hole.
	if sfr.pos < sfr.sp[0].offset {
		return sfr.readHole(b, sfr.sp[0].offset), nil
	}

	// In a data fragment, so read from it.
	// This math is overflow free since we verify that offset and numBytes can
	// be safely added when creating the sparseFileReader.
	endPos := sfr.sp[0].offset + sfr.sp[0].numBytes // End offset of fragment
	bytesLeft := endPos - sfr.pos                   // Bytes left in fragment
	if int64(len(b)) > bytesLeft {
		b = b[:bytesLeft]
	}

	n, err = sfr.rfr.Read(b)
	sfr.pos += int64(n)
	if err == io.EOF {
		if sfr.pos < endPos {
			err = io.ErrUnexpectedEOF // There was supposed to be more data
		} else if sfr.pos < sfr.tot {
			err = nil // There is still an implicit sparse hole at the end
		}
	}

	if sfr.pos == endPos {
		sfr.sp = sfr.sp[1:] // We're done with this fragment, so pop it
	}
	return n, err
}

// numBytes returns the number of bytes left to read in the sparse file's
// sparse-encoded data in the tar archive.
func (sfr *sparseFileReader) numBytes() int64 {
	return sfr.rfr.numBytes()
}
