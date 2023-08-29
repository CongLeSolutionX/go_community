// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"context"
	"internal/syscall/windows"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

// cgoAvailable set to true to indicate that the cgo resolver
// is available on Windows. Note that on Windows the cgo resolver
// does not actually use cgo.
const cgoAvailable = true

const (
	_WSAHOST_NOT_FOUND = syscall.Errno(11001)
	_WSATRY_AGAIN      = syscall.Errno(11002)
	_WSANO_DATA        = syscall.Errno(11004)
)

func winError(call string, err error) error {
	switch err {
	case _WSAHOST_NOT_FOUND, _WSANO_DATA:
		return errNoSuchHost
	}
	return os.NewSyscallError(call, err)
}

func getprotobyname(name string) (proto int, err error) {
	p, err := syscall.GetProtoByName(name)
	if err != nil {
		return 0, winError("getprotobyname", err)
	}
	return int(p.Proto), nil
}

// lookupProtocol looks up IP protocol name and returns correspondent protocol number.
func lookupProtocol(ctx context.Context, name string) (int, error) {
	// GetProtoByName return value is stored in thread local storage.
	// Start new os thread before the call to prevent races.
	type result struct {
		proto int
		err   error
	}
	ch := make(chan result) // unbuffered
	go func() {
		acquireThread()
		defer releaseThread()
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		proto, err := getprotobyname(name)
		select {
		case ch <- result{proto: proto, err: err}:
		case <-ctx.Done():
		}
	}()
	select {
	case r := <-ch:
		if r.err != nil {
			if proto, err := lookupProtocolMap(name); err == nil {
				return proto, nil
			}

			dnsError := &DNSError{Err: r.err.Error(), Name: name}
			if r.err == errNoSuchHost {
				dnsError.IsNotFound = true
			}
			r.err = dnsError
		}
		return r.proto, r.err
	case <-ctx.Done():
		return 0, mapErr(ctx.Err())
	}
}

func doWithRetryDNS[T any](ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	dnsConf := getSystemDNSConfig()
	ctx, cancel := context.WithTimeout(ctx, dnsConf.timeout)
	defer cancel()
	var err error
	var ret T
	for i := 0; i < dnsConf.attempts; i++ {
		if ret, err = fn(ctx); err == nil {
			return ret, nil
		}
		if err, ok := err.(*DNSError); !ok || !err.IsTemporary {
			break
		}
	}
	return ret, err
}

type getAddrInfoExParams struct {
	o syscall.Overlapped // must be the first field to work with GetAddrInfoEx

	fn func(code uint32)
}

// getAddrInfoExCallback implements the callback for GetAddrInfoEx as defined in
// https://learn.microsoft.com/en-us/windows/win32/api/winsock2/nc-winsock2-lpwsaoverlapped_completion_routine.
var getAddrInfoExCallback = syscall.NewCallback(func(code uint32, _ uint32, o uintptr, _ uint32) uintptr {
	params := (*getAddrInfoExParams)(unsafe.Pointer(o))
	params.fn(code)
	return 0
})

// getAddrInfoEx calls GetAddrInfoExW asynchronously passing the result to fn on completion.
// The result will be freed after calling fn, so it must not be used after that.
// If ctx is canceled before GetAddrInfoExW completes, the call will be canceled.
func getAddrInfoEx[T any](ctx context.Context, node, service string, hints *windows.AddrInfoExW, fn func(result *windows.AddrInfoExW) (T, error)) (T, error) {
	var err error
	var node16p, service16p *uint16
	if node != "" {
		node16p, err = syscall.UTF16PtrFromString(node)
		if err != nil {
			var zero T
			return zero, &DNSError{Err: err.Error()}
		}
	}
	if service != "" {
		service16p, err = syscall.UTF16PtrFromString(service)
		if err != nil {
			var zero T
			return zero, &DNSError{Err: err.Error()}
		}
	}
	var (
		result       *windows.AddrInfoExW
		cancelHandle syscall.Handle
		ch           chan error
	)
	if ctx.Err() == nil {
		ch = make(chan error, 1)
		o := getAddrInfoExParams{
			fn: func(code uint32) {
				var err error
				if code != 0 {
					err = syscall.Errno(code)
				}
				ch <- err
			},
		}
		err = windows.GetAddrInfoEx(node16p, service16p, windows.NS_DS, nil, hints, &result, nil, &o.o, getAddrInfoExCallback, &cancelHandle)
		if err != syscall.ERROR_IO_PENDING {
			// If the call to GetAddrInfoEx returns an error other than ERROR_IO_PENDING,
			// the callback will not be called, so we need to fulfill ch here.
			ch <- err
		}
	}
	select {
	case err := <-ch:
		if result != nil {
			defer windows.FreeAddrInfoExW(result)
		}
		if err != nil {
			isTemporary := err == _WSATRY_AGAIN
			err = winError("getaddrinfoexw", err)
			err = &DNSError{
				Err:         err.Error(),
				IsNotFound:  err == errNoSuchHost,
				IsTemporary: isTemporary,
			}
			var zero T
			return zero, err
		}
		return fn(result)
	case <-ctx.Done():
		if cancelHandle != syscall.InvalidHandle {
			// This is just an optimization, so ignore the error.
			windows.GetAddrInfoExCancel(&cancelHandle)
		}
		var zero T
		return zero, &DNSError{
			Err:       ctx.Err().Error(),
			IsTimeout: ctx.Err() == context.DeadlineExceeded,
		}
	}
}

func (r *Resolver) lookupHost(ctx context.Context, name string) ([]string, error) {
	ips, err := r.lookupIP(ctx, "ip", name)
	if err != nil {
		return nil, err
	}
	addrs := make([]string, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, ip.String())
	}
	return addrs, nil
}

func (r *Resolver) lookupIP(ctx context.Context, network, name string) ([]IPAddr, error) {
	if order, conf := systemConf().hostLookupOrder(r, name); order != hostLookupCgo {
		return r.goLookupIP(ctx, network, name, order, conf)
	}

	hints := windows.AddrInfoExW{
		Family:   syscall.AF_UNSPEC,
		Socktype: syscall.SOCK_STREAM,
		Protocol: syscall.IPPROTO_IP,
	}
	switch ipVersion(network) {
	case '4':
		hints.Family = syscall.AF_INET
	case '6':
		hints.Family = syscall.AF_INET6
	}
	addrs, err := doWithRetryDNS(ctx, func(ctx context.Context) ([]IPAddr, error) {
		return getAddrInfoEx(ctx, name, "", &hints, func(result *windows.AddrInfoExW) ([]IPAddr, error) {
			addrs := make([]IPAddr, 0, 5)
			for ; result != nil; result = result.Next {
				addr := unsafe.Pointer(result.Addr)
				switch result.Family {
				case syscall.AF_INET:
					a := (*syscall.RawSockaddrInet4)(addr).Addr
					addrs = append(addrs, IPAddr{IP: copyIP(a[:])})
				case syscall.AF_INET6:
					a := (*syscall.RawSockaddrInet6)(addr).Addr
					zone := zoneCache.name(int((*syscall.RawSockaddrInet6)(addr).Scope_id))
					addrs = append(addrs, IPAddr{IP: copyIP(a[:]), Zone: zone})
				default:
					return nil, &DNSError{Err: syscall.EWINDOWS.Error()}
				}
			}
			return addrs, nil
		})
	})
	if err != nil {
		if err, ok := err.(*DNSError); ok {
			err.Name = name
		}
		return nil, err
	}
	return addrs, nil
}

func (r *Resolver) lookupPort(ctx context.Context, network, service string) (int, error) {
	if systemConf().mustUseGoResolver(r) {
		return lookupPortMap(network, service)
	}

	hints := windows.AddrInfoExW{
		Family:   syscall.AF_UNSPEC,
		Protocol: syscall.IPPROTO_IP,
	}
	switch network {
	case "tcp4", "tcp6":
		hints.Socktype = syscall.SOCK_STREAM
	case "udp4", "udp6":
		hints.Socktype = syscall.SOCK_DGRAM
	}

	port, err := doWithRetryDNS(ctx, func(ctx context.Context) (int, error) {
		return getAddrInfoEx(ctx, "", service, &hints, func(result *windows.AddrInfoExW) (int, error) {
			addr := unsafe.Pointer(result.Addr)
			switch result.Family {
			case syscall.AF_INET:
				a := (*syscall.RawSockaddrInet4)(addr)
				return int(syscall.Ntohs(a.Port)), nil
			case syscall.AF_INET6:
				a := (*syscall.RawSockaddrInet6)(addr)
				return int(syscall.Ntohs(a.Port)), nil
			}
			return 0, &DNSError{Err: syscall.EINVAL.Error()}
		})
	})
	if err != nil {
		if port, err := lookupPortMap(network, service); err == nil {
			return port, nil
		}
		if err, ok := err.(*DNSError); ok {
			err.Name = network + "/" + service
		}
		return 0, err
	}
	return port, nil
}

func (r *Resolver) lookupCNAME(ctx context.Context, name string) (string, error) {
	if order, conf := systemConf().hostLookupOrder(r, name); order != hostLookupCgo {
		return r.goLookupCNAME(ctx, name, order, conf)
	}

	// TODO(bradfitz): finish ctx plumbing. Nothing currently depends on this.
	acquireThread()
	defer releaseThread()
	var rec *syscall.DNSRecord
	e := syscall.DnsQuery(name, syscall.DNS_TYPE_CNAME, 0, nil, &rec, nil)
	// windows returns DNS_INFO_NO_RECORDS if there are no CNAME-s
	if errno, ok := e.(syscall.Errno); ok && errno == syscall.DNS_INFO_NO_RECORDS {
		// if there are no aliases, the canonical name is the input name
		return absDomainName(name), nil
	}
	if e != nil {
		return "", &DNSError{Err: winError("dnsquery", e).Error(), Name: name}
	}
	defer syscall.DnsRecordListFree(rec, 1)

	resolved := resolveCNAME(syscall.StringToUTF16Ptr(name), rec)
	cname := windows.UTF16PtrToString(resolved)
	return absDomainName(cname), nil
}

func (r *Resolver) lookupSRV(ctx context.Context, service, proto, name string) (string, []*SRV, error) {
	if systemConf().mustUseGoResolver(r) {
		return r.goLookupSRV(ctx, service, proto, name)
	}
	// TODO(bradfitz): finish ctx plumbing. Nothing currently depends on this.
	acquireThread()
	defer releaseThread()
	var target string
	if service == "" && proto == "" {
		target = name
	} else {
		target = "_" + service + "._" + proto + "." + name
	}
	var rec *syscall.DNSRecord
	e := syscall.DnsQuery(target, syscall.DNS_TYPE_SRV, 0, nil, &rec, nil)
	if e != nil {
		return "", nil, &DNSError{Err: winError("dnsquery", e).Error(), Name: target}
	}
	defer syscall.DnsRecordListFree(rec, 1)

	srvs := make([]*SRV, 0, 10)
	for _, p := range validRecs(rec, syscall.DNS_TYPE_SRV, target) {
		v := (*syscall.DNSSRVData)(unsafe.Pointer(&p.Data[0]))
		srvs = append(srvs, &SRV{absDomainName(syscall.UTF16ToString((*[256]uint16)(unsafe.Pointer(v.Target))[:])), v.Port, v.Priority, v.Weight})
	}
	byPriorityWeight(srvs).sort()
	return absDomainName(target), srvs, nil
}

func (r *Resolver) lookupMX(ctx context.Context, name string) ([]*MX, error) {
	if systemConf().mustUseGoResolver(r) {
		return r.goLookupMX(ctx, name)
	}
	// TODO(bradfitz): finish ctx plumbing. Nothing currently depends on this.
	acquireThread()
	defer releaseThread()
	var rec *syscall.DNSRecord
	e := syscall.DnsQuery(name, syscall.DNS_TYPE_MX, 0, nil, &rec, nil)
	if e != nil {
		return nil, &DNSError{Err: winError("dnsquery", e).Error(), Name: name}
	}
	defer syscall.DnsRecordListFree(rec, 1)

	mxs := make([]*MX, 0, 10)
	for _, p := range validRecs(rec, syscall.DNS_TYPE_MX, name) {
		v := (*syscall.DNSMXData)(unsafe.Pointer(&p.Data[0]))
		mxs = append(mxs, &MX{absDomainName(windows.UTF16PtrToString(v.NameExchange)), v.Preference})
	}
	byPref(mxs).sort()
	return mxs, nil
}

func (r *Resolver) lookupNS(ctx context.Context, name string) ([]*NS, error) {
	if systemConf().mustUseGoResolver(r) {
		return r.goLookupNS(ctx, name)
	}
	// TODO(bradfitz): finish ctx plumbing. Nothing currently depends on this.
	acquireThread()
	defer releaseThread()
	var rec *syscall.DNSRecord
	e := syscall.DnsQuery(name, syscall.DNS_TYPE_NS, 0, nil, &rec, nil)
	if e != nil {
		return nil, &DNSError{Err: winError("dnsquery", e).Error(), Name: name}
	}
	defer syscall.DnsRecordListFree(rec, 1)

	nss := make([]*NS, 0, 10)
	for _, p := range validRecs(rec, syscall.DNS_TYPE_NS, name) {
		v := (*syscall.DNSPTRData)(unsafe.Pointer(&p.Data[0]))
		nss = append(nss, &NS{absDomainName(syscall.UTF16ToString((*[256]uint16)(unsafe.Pointer(v.Host))[:]))})
	}
	return nss, nil
}

func (r *Resolver) lookupTXT(ctx context.Context, name string) ([]string, error) {
	if systemConf().mustUseGoResolver(r) {
		return r.goLookupTXT(ctx, name)
	}
	// TODO(bradfitz): finish ctx plumbing. Nothing currently depends on this.
	acquireThread()
	defer releaseThread()
	var rec *syscall.DNSRecord
	e := syscall.DnsQuery(name, syscall.DNS_TYPE_TEXT, 0, nil, &rec, nil)
	if e != nil {
		return nil, &DNSError{Err: winError("dnsquery", e).Error(), Name: name}
	}
	defer syscall.DnsRecordListFree(rec, 1)

	txts := make([]string, 0, 10)
	for _, p := range validRecs(rec, syscall.DNS_TYPE_TEXT, name) {
		d := (*syscall.DNSTXTData)(unsafe.Pointer(&p.Data[0]))
		s := ""
		for _, v := range (*[1 << 10]*uint16)(unsafe.Pointer(&(d.StringArray[0])))[:d.StringCount:d.StringCount] {
			s += windows.UTF16PtrToString(v)
		}
		txts = append(txts, s)
	}
	return txts, nil
}

func (r *Resolver) lookupAddr(ctx context.Context, addr string) ([]string, error) {
	if order, conf := systemConf().addrLookupOrder(r, addr); order != hostLookupCgo {
		return r.goLookupPTR(ctx, addr, order, conf)
	}

	// TODO(bradfitz): finish ctx plumbing. Nothing currently depends on this.
	acquireThread()
	defer releaseThread()
	arpa, err := reverseaddr(addr)
	if err != nil {
		return nil, err
	}
	var rec *syscall.DNSRecord
	e := syscall.DnsQuery(arpa, syscall.DNS_TYPE_PTR, 0, nil, &rec, nil)
	if e != nil {
		return nil, &DNSError{Err: winError("dnsquery", e).Error(), Name: addr}
	}
	defer syscall.DnsRecordListFree(rec, 1)

	ptrs := make([]string, 0, 10)
	for _, p := range validRecs(rec, syscall.DNS_TYPE_PTR, arpa) {
		v := (*syscall.DNSPTRData)(unsafe.Pointer(&p.Data[0]))
		ptrs = append(ptrs, absDomainName(windows.UTF16PtrToString(v.Host)))
	}
	return ptrs, nil
}

const dnsSectionMask = 0x0003

// returns only results applicable to name and resolves CNAME entries.
func validRecs(r *syscall.DNSRecord, dnstype uint16, name string) []*syscall.DNSRecord {
	cname := syscall.StringToUTF16Ptr(name)
	if dnstype != syscall.DNS_TYPE_CNAME {
		cname = resolveCNAME(cname, r)
	}
	rec := make([]*syscall.DNSRecord, 0, 10)
	for p := r; p != nil; p = p.Next {
		// in case of a local machine, DNS records are returned with DNSREC_QUESTION flag instead of DNS_ANSWER
		if p.Dw&dnsSectionMask != syscall.DnsSectionAnswer && p.Dw&dnsSectionMask != syscall.DnsSectionQuestion {
			continue
		}
		if p.Type != dnstype {
			continue
		}
		if !syscall.DnsNameCompare(cname, p.Name) {
			continue
		}
		rec = append(rec, p)
	}
	return rec
}

// returns the last CNAME in chain.
func resolveCNAME(name *uint16, r *syscall.DNSRecord) *uint16 {
	// limit cname resolving to 10 in case of an infinite CNAME loop
Cname:
	for cnameloop := 0; cnameloop < 10; cnameloop++ {
		for p := r; p != nil; p = p.Next {
			if p.Dw&dnsSectionMask != syscall.DnsSectionAnswer {
				continue
			}
			if p.Type != syscall.DNS_TYPE_CNAME {
				continue
			}
			if !syscall.DnsNameCompare(name, p.Name) {
				continue
			}
			name = (*syscall.DNSPTRData)(unsafe.Pointer(&r.Data[0])).Host
			continue Cname
		}
		break
	}
	return name
}

// concurrentThreadsLimit returns the number of threads we permit to
// run concurrently doing DNS lookups.
func concurrentThreadsLimit() int {
	return 500
}
