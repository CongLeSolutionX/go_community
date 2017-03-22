// +build amd64 amd64p32 arm64 arm64be ppc64 ppc64le mips64 mips64le mips64p32 mips64p32le s390x sparc64

package base64

func (enc *Encoding) decode(dst, src []byte) (n int, end bool, err error) {
	var ninc, si int

	if len(src) == 0 {
		return 0, true, nil
	}

	ilen := len(src)
	olen := len(dst)
	for ilen-si >= 8 && olen-n >= 8 {
		if ok := enc.decode64(dst[n:], src[si:]); ok {
			n += 6
			si += 8
			continue
		}
		si, ninc, end, err = enc.decode_quantum(dst[n:], src, si)
		if err != nil {
			return n + ninc, end, err
		}
		n += ninc
	}

	if ilen-si >= 4 && olen-n >= 4 {
		if ok := enc.decode32(dst[n:], src[si:]); ok {
			n += 3
			si += 4
		}
	}

	for si < len(src) {
		si, ninc, end, err = enc.decode_quantum(dst[n:], src, si)
		n += ninc
		if end || err != nil {
			return n, end, err
		}
	}
	return n, end, err
}
