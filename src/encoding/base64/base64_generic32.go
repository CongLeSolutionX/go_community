// +build 386 arm armbe mips mipsle ppc s390 sparc

package base64

func (enc *Encoding) decode(dst, src []byte) (n int, end bool, err error) {
	var ninc, si, ns int

	if len(src) == 0 {
		return 0, true, nil
	}

	ilen := len(src)
	olen := len(dst)
	for ilen-si >= 4 && olen-n >= 4 {
		if ok := enc.decode32(dst[n:], src[si:]); ok {
			n += 3
			si += 4
			continue
		}
		si, ninc, end, err = enc.decode_quantum(dst[n:], src, si)
		if err != nil {
			return n + ninc, end, err
		}
		n += ninc
	}
	for si < len(src) {
		si, ns, end, err = enc.decode_quantum(dst[n:], src, si)
		n += ns
		if end || err != nil {
			return n, end, err
		}
	}
	return n, end, err
}
