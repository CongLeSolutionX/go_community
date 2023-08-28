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

func (r *Resolver) getAddrInfoEx(ctx context.Context, node, service string, hints *windows.AddrInfoExW, fn func(result *windows.AddrInfoExW) error) *DNSError {
	var cancelHandle syscall.Handle
	getaddr := func() *DNSError {
		acquireThread()
		defer releaseThread()

		ev, err := windows.CreateEvent(nil, 1, 0, nil)
		if err != nil {
			return &DNSError{Err: err.Error()}
		}
		defer syscall.CloseHandle(ev)
		var node16p, service16p *uint16
		if node != "" {
			node16p, err = syscall.UTF16PtrFromString(node)
			if err != nil {
				return &DNSError{Err: err.Error()}
			}
		}
		if service != "" {
			service16p, err = syscall.UTF16PtrFromString(service)
			if err != nil {
				return &DNSError{Err: err.Error()}
			}
		}

		dnsConf := getSystemDNSConfig()
		var result *windows.AddrInfoExW
		var o syscall.Overlapped
		o.HEvent = ev
		for i := 0; i < dnsConf.attempts; i++ {
			// There is no need to pass a timeout to GetAddrInfoExW nor to WaitForSingleObject,
			// it will be handled by the context.
			err = windows.GetAddrInfoEx(node16p, service16p, windows.NS_ALL, nil, hints, &result, nil, &o, 0, &cancelHandle)
			if err == syscall.ERROR_IO_PENDING {
				if s, err := syscall.WaitForSingleObject(ev, syscall.INFINITE); s != syscall.WAIT_OBJECT_0 {
					if err != nil {
						return &DNSError{Err: err.Error()}
					}
					return &DNSError{Err: "unexpected result from WaitForSingleObject"}
				}
				err = nil
				if code := windows.GetAddrInfoExOverlappedResult(&o); code != 0 {
					err = syscall.Errno(code)
				}
			}
			if err == nil || err != _WSATRY_AGAIN {
				break
			}
		}
		if err != nil {
			err := winError("getaddrinfoexw", err)
			dnsError := &DNSError{Err: err.Error()}
			if err == errNoSuchHost {
				dnsError.IsNotFound = true
			}
			return dnsError
		}
		defer windows.FreeAddrInfoExW(result)
		if err := fn(result); err != nil {
			if err, ok := err.(*DNSError); ok {
				return err
			}
			return &DNSError{Err: err.Error()}
		}
		return nil
	}

	var ch chan *DNSError
	if ctx.Err() == nil {
		ch = make(chan *DNSError, 1)
		go func() {
			err := getaddr()
			ch <- err
		}()
	}

	ctx, cancel := context.WithTimeout(ctx, getSystemDNSConfig().timeout)
	defer cancel()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		if cancelHandle != syscall.InvalidHandle {
			// This is just an optimization, so ignore the error.
			windows.GetAddrInfoExCancel(&cancelHandle)
		}
		return &DNSError{
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
	addrs := make([]IPAddr, 0, 5)
	err := r.getAddrInfoEx(ctx, name, "", &hints, func(result *windows.AddrInfoExW) error {
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
				return syscall.EWINDOWS
			}
		}
		return nil
	})
	if err != nil {
		err.Name = name
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

	var port int
	err := r.getAddrInfoEx(ctx, "", service, &hints, func(result *windows.AddrInfoExW) error {
		addr := unsafe.Pointer(result.Addr)
		switch result.Family {
		case syscall.AF_INET:
			a := (*syscall.RawSockaddrInet4)(addr)
			port = int(syscall.Ntohs(a.Port))
		case syscall.AF_INET6:
			a := (*syscall.RawSockaddrInet6)(addr)
			port = int(syscall.Ntohs(a.Port))
		}
		return syscall.EINVAL
	})
	if err != nil {
		if port, err := lookupPortMap(network, service); err == nil {
			return port, nil
		}
		err.Name = network + "/" + service
	}
	return port, err
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
