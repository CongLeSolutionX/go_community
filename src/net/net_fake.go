// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Fake networking for js/wasm and wasip1/wasm.
// It is intended to allow tests of other package to pass.

//go:build js || wasip1

package net

import (
	"context"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	sockets       sync.Map // fakeSockAddr → *netFD
	fakeSocketIDs sync.Map // fakeNetFD.id → *netFD
	fakePorts     sync.Map // port # → *netFD
	nextPort      atomic.Int32
)

const defaultBuffer = 65535

type fakeSockAddr struct {
	family  int
	address string
}

func fakeAddr(sa sockaddr) fakeSockAddr {
	return fakeSockAddr{
		family:  sa.family(),
		address: sa.String(),
	}
}

type fakeNetFD struct {
	fd           *netFD
	reservedPort int // serves as both reserved port and fd

	queue         *packetQueue // incoming packets
	peer          *netFD       // nil for listeners and PacketConns
	readDeadline  atomic.Pointer[deadlineTimer]
	writeDeadline atomic.Pointer[deadlineTimer]

	fakeAddr      fakeSockAddr
	incomingEmpty chan bool
	incoming      chan []*netFD // closed when the FD's Listener is closed
}

// socket returns a network file descriptor that is ready for
// I/O using the fake network.
func socket(ctx context.Context, net string, family, sotype, proto int, ipv6only bool, laddr, raddr sockaddr, ctrlCtxFn func(context.Context, string, string, syscall.RawConn) error) (*netFD, error) {
	if raddr != nil && ctrlCtxFn != nil {
		return nil, &AddrError{Err: "socket: Dialer.Control not supported on " + runtime.GOOS, Addr: raddr.String()}
	}
	switch sotype {
	case syscall.SOCK_STREAM, syscall.SOCK_SEQPACKET, syscall.SOCK_DGRAM:
	default:
		return nil, os.NewSyscallError("socket", syscall.ENOTSUP)
	}

	fd := &netFD{
		family: family,
		sotype: sotype,
		net:    net,
	}
	fd.fakeNetFD = newFakeNetFD(fd)

	if raddr == nil {
		if err := fakeListen(fd, laddr); err != nil {
			fd.Close()
			return nil, err
		}
		return fd, nil
	}

	if err := fakeConnect(fd, laddr, raddr); err != nil {
		fd.Close()
		return nil, err
	}
	return fd, nil
}

func resolvedAddrIsValid(sa sockaddr) bool {
	switch sa := sa.(type) {
	case *TCPAddr:
		return (len(sa.IP) == 4 || len(sa.IP) == 16) && sa.Port > 0 && sa.Port < 1<<16

	case *UDPAddr:
		return (len(sa.IP) == 4 || len(sa.IP) == 16) && sa.Port > 0 && sa.Port < 1<<16

	case *UnixAddr:
		if sa.Name != "" {
			sepi := len(sa.Name) - 1
			for sepi >= 0 && !os.IsPathSeparator(sa.Name[sepi]) {
				sepi--
			}
			if sepi <= 0 {
				return false
			}
			if _, err := os.Stat(sa.Name[:sepi]); err != nil {
				return false
			}
		}
		return true

	default:
		return true
	}
}

func assignFakeAddr(net string, family int, addr sockaddr, defaultPort int) (sockaddr, error) {
	validate := func(sa sockaddr) (sockaddr, error) {
		if !resolvedAddrIsValid(sa) {
			return nil, syscall.EINVAL
		}
		return sa, nil
	}

	assignIP := func() (ip IP, port int, zone string, err error) {
		switch addr := addr.(type) {
		case *TCPAddr:
			if len(net) < 3 || net[:3] != "tcp" {
				return nil, 0, "", syscall.EINVAL
			}
			if addr != nil {
				ip = addr.IP
				port = addr.Port
				zone = addr.Zone
			}
		case *UDPAddr:
			if len(net) < 3 || net[:3] != "udp" {
				return nil, 0, "", syscall.EINVAL
			}
			if addr != nil {
				ip = addr.IP
				port = addr.Port
				zone = addr.Zone
			}
		case nil:
		}

		if ip == nil {
			ip = IPv4(127, 0, 0, 1)
		}
		switch family {
		case syscall.AF_INET:
			ip = ip.To4()
		case syscall.AF_INET6:
			ip = ip.To16()
		default:
			return nil, 0, "", syscall.EINVAL
		}
		if ip == nil {
			return nil, 0, "", syscall.EINVAL
		}

		if port == 0 {
			port = defaultPort
		}
		return ip, port, zone, nil
	}

	switch net {
	case "tcp", "tcp4", "tcp6":
		ip, port, zone, err := assignIP()
		if err != nil {
			return nil, err
		}
		return validate(&TCPAddr{IP: ip, Port: port, Zone: zone})

	case "udp", "udp4", "udp6":
		ip, port, zone, err := assignIP()
		if err != nil {
			return nil, err
		}
		return validate(&UDPAddr{IP: ip, Port: port, Zone: zone})

	case "unix", "unixgram", "unixpacket":
		name := ""
		if addr != nil {
			uaddr, ok := addr.(*UnixAddr)
			if !ok {
				return nil, syscall.EINVAL
			}
			if uaddr != nil {
				name = uaddr.Name
			}
		}
		return validate(&UnixAddr{Net: net, Name: name})

	default:
		return nil, syscall.EAFNOSUPPORT
	}
}

func newFakeNetFD(fd *netFD) *fakeNetFD {
	ffd := &fakeNetFD{fd: fd}
	ffd.readDeadline.Store(newDeadlineTimer(noDeadline))
	ffd.writeDeadline.Store(newDeadlineTimer(noDeadline))
	for {
		ffd.reservedPort = int(nextPort.Add(1))
		if ffd.reservedPort == 0 {
			continue
		}
		if _, dup := fakePorts.LoadOrStore(ffd.reservedPort, ffd.fd); !dup {
			break
		}
	}
	return ffd
}

func (ffd *fakeNetFD) Read(p []byte) (n int, err error) {
	n, _, err = ffd.queue.recvfrom(ffd.readDeadline.Load(), p, false, nil)
	return n, err
}

func (ffd *fakeNetFD) Write(p []byte) (nn int, err error) {
	peer := ffd.peer
	if peer == nil {
		if ffd.fd.raddr == nil {
			return 0, os.NewSyscallError("write", syscall.ENOTCONN)
		}
		peeri, _ := sockets.Load(fakeAddr(ffd.fd.raddr.(sockaddr)))
		if peeri == nil {
			return 0, os.NewSyscallError("write", syscall.ECONNRESET)
		}
		peer = peeri.(*netFD)
		if peer.queue == nil {
			return 0, os.NewSyscallError("write", syscall.ECONNRESET)
		}
	}

	if peer.fakeNetFD == nil {
		return 0, os.NewSyscallError("write", syscall.EINVAL)
	}
	return peer.queue.write(ffd.writeDeadline.Load(), p, ffd.fd.laddr.(sockaddr))
}

func (ffd *fakeNetFD) Close() (err error) {
	if ffd.fakeAddr != (fakeSockAddr{}) {
		if !sockets.CompareAndDelete(ffd.fakeAddr, ffd.fd) && err == nil {
			err = ErrClosed
		}
	}

	if ffd.queue != nil {
		if closeErr := ffd.queue.closeRead(); err == nil {
			err = closeErr
		}
	}
	if ffd.peer != nil {
		if closeErr := ffd.peer.queue.closeWrite(); err == nil {
			err = closeErr
		}
	}
	ffd.readDeadline.Load().Reset(noDeadline)
	ffd.writeDeadline.Load().Reset(noDeadline)

	if ffd.incoming != nil {
		select {
		case incoming, ok := <-ffd.incoming:
			if !ok {
				return ErrClosed
			}
			for _, c := range incoming {
				c.Close()
			}
		case <-ffd.incomingEmpty:
		}
		close(ffd.incoming)
	}

	fakePorts.CompareAndDelete(ffd.reservedPort, ffd.fd)

	return err
}

func (ffd *fakeNetFD) closeRead() error {
	return ffd.queue.closeRead()
}

func (ffd *fakeNetFD) closeWrite() error {
	if ffd.peer == nil {
		return os.NewSyscallError("closeWrite", syscall.ENOTCONN)
	}
	return ffd.peer.queue.closeWrite()
}

func (ffd *fakeNetFD) accept(laddr Addr) (*netFD, error) {
	if ffd.incoming == nil {
		return nil, os.NewSyscallError("accept", syscall.EINVAL)
	}

	var (
		incoming []*netFD
		ok       bool
	)
	select {
	case <-ffd.readDeadline.Load().expired:
		return nil, os.ErrDeadlineExceeded
	case incoming, ok = <-ffd.incoming:
		if !ok {
			return nil, syscall.EINVAL
		}
	}

	peer := incoming[0]
	incoming = incoming[1:]
	if len(incoming) > 0 {
		ffd.incoming <- incoming
	} else {
		ffd.incomingEmpty <- true
	}
	return peer, nil
}

func (ffd *fakeNetFD) SetDeadline(t time.Time) error {
	err1 := ffd.SetReadDeadline(t)
	err2 := ffd.SetWriteDeadline(t)
	if err1 != nil {
		return err1
	}
	return err2
}

func (ffd *fakeNetFD) SetReadDeadline(t time.Time) error {
	dt := ffd.readDeadline.Load()
	if !dt.Reset(t) {
		ffd.readDeadline.Store(newDeadlineTimer(t))
	}
	return nil
}

func (ffd *fakeNetFD) SetWriteDeadline(t time.Time) error {
	dt := ffd.writeDeadline.Load()
	if !dt.Reset(t) {
		ffd.writeDeadline.Store(newDeadlineTimer(t))
	}
	return nil
}

const maxPacketSize = 65535

type packet struct {
	buf       []byte
	bufOffset int
	next      *packet
	from      sockaddr
}

func (p *packet) clear() {
	p.buf = p.buf[:0]
	p.bufOffset = 0
	p.next = nil
	p.from = nil
}

var packetPool = sync.Pool{
	New: func() any { return new(packet) },
}

type packetQueueState struct {
	head, tail      *packet
	nBytes          int
	readBufferBytes int
	readClosed      bool
	writeClosed     bool
	noLinger        bool
}

type packetQueue struct {
	empty chan packetQueueState // contains configuration parameters when the queue is empty and not closed
	ready chan packetQueueState // contains the packets when non-empty or closed
	full  chan packetQueueState // contains the packets when buffer is full and not closed
}

func newPacketQueue(readBufferBytes int) *packetQueue {
	pq := &packetQueue{
		empty: make(chan packetQueueState, 1),
		ready: make(chan packetQueueState, 1),
		full:  make(chan packetQueueState, 1),
	}
	pq.put(packetQueueState{
		readBufferBytes: readBufferBytes,
	})
	return pq
}

func (pq *packetQueue) get() packetQueueState {
	var q packetQueueState
	select {
	case q = <-pq.empty:
	case q = <-pq.ready:
	case q = <-pq.full:
	}
	return q
}

func (pq *packetQueue) put(q packetQueueState) {
	switch {
	case q.readClosed || q.writeClosed:
		pq.ready <- q
	case q.nBytes >= q.readBufferBytes:
		pq.full <- q
	case q.head == nil:
		if q.nBytes > 0 {
			defer panic("net: put with nil packet list and nonzero nBytes")
		}
		pq.empty <- q
	default:
		pq.ready <- q
	}
}

func (pq *packetQueue) closeRead() error {
	q := pq.get()

	// Discard any unread packets.
	for q.head != nil {
		p := q.head
		q.head = p.next
		p.clear()
		packetPool.Put(p)
	}
	q.nBytes = 0

	q.readClosed = true
	pq.put(q)
	return nil
}

func (pq *packetQueue) closeWrite() error {
	q := pq.get()
	q.writeClosed = true
	pq.put(q)
	return nil
}

func (pq *packetQueue) setLinger(linger bool) error {
	q := pq.get()
	defer func() { pq.put(q) }()

	if q.writeClosed {
		return ErrClosed
	}
	q.noLinger = !linger
	return nil
}

func (pq *packetQueue) write(dt *deadlineTimer, b []byte, from sockaddr) (n int, err error) {
	for {
		dn := len(b)
		if dn > maxPacketSize {
			dn = maxPacketSize
		}

		dn, err = pq.send(dt, b[:dn], from, true)
		n += dn
		if err != nil {
			return n, err
		}

		b = b[dn:]
		if len(b) == 0 {
			return n, nil
		}
	}
}

func (pq *packetQueue) send(dt *deadlineTimer, b []byte, from sockaddr, block bool) (n int, err error) {
	if from == nil {
		return 0, os.NewSyscallError("send", syscall.EINVAL)
	}
	if len(b) > maxPacketSize {
		return 0, os.NewSyscallError("send", syscall.EMSGSIZE)
	}

	var q packetQueueState
	var full chan packetQueueState
	if !block {
		full = pq.full
	}
	select {
	case <-dt.expired:
		return 0, os.ErrDeadlineExceeded

	case q = <-full:
		pq.put(q)
		return 0, os.NewSyscallError("send", syscall.ENOBUFS)

	case q = <-pq.empty:
	case q = <-pq.ready:
	}
	defer func() { pq.put(q) }()

	// Don't allow a packet to be sent if the deadline has expired,
	// even if the select above chose a different branch.
	select {
	case <-dt.expired:
		return 0, os.ErrDeadlineExceeded
	default:
	}
	if q.writeClosed {
		return 0, ErrClosed
	} else if q.readClosed {
		return 0, os.NewSyscallError("send", syscall.ECONNRESET)
	}

	p := packetPool.Get().(*packet)
	p.buf = append(p.buf[:0], b...)
	p.from = from

	if q.head == nil {
		q.head = p
	} else {
		q.tail.next = p
	}
	q.tail = p
	q.nBytes += len(p.buf)

	return len(b), nil
}

func (pq *packetQueue) recvfrom(dt *deadlineTimer, b []byte, wholePacket bool, checkFrom func(sockaddr) error) (n int, from sockaddr, err error) {
	var q packetQueueState
	var empty chan packetQueueState
	if len(b) == 0 {
		// For consistency with the implementation on Unix platforms,
		// allow a zero-length Read to proceed if the queue is empty.
		// (Without this, TestZeroByteRead deadlocks.)
		empty = pq.empty
	}
	select {
	case <-dt.expired:
		return 0, nil, os.ErrDeadlineExceeded
	case q = <-empty:
	case q = <-pq.ready:
	case q = <-pq.full:
	}
	defer func() { pq.put(q) }()

	p := q.head
	if p == nil {
		switch {
		case q.readClosed:
			return 0, nil, ErrClosed
		case q.writeClosed:
			if q.noLinger {
				return 0, nil, os.NewSyscallError("recvfrom", syscall.ECONNRESET)
			}
			return 0, nil, io.EOF
		case len(b) == 0:
			return 0, nil, nil
		default:
			// This should be impossible: pq.full should only contain a non-empty list,
			// pq.ready should either contain a non-empty list or indicate that the
			// connection is closed, and we should only receive from pq.empty if
			// len(b) == 0.
			panic("net: nil packet list from non-closed packetQueue")
		}
	}

	select {
	case <-dt.expired:
		return 0, nil, os.ErrDeadlineExceeded
	default:
	}

	if checkFrom != nil {
		if err := checkFrom(p.from); err != nil {
			return 0, nil, err
		}
	}

	n = copy(b, p.buf[p.bufOffset:])
	from = p.from
	if wholePacket || p.bufOffset+n == len(p.buf) {
		q.head = p.next
		q.nBytes -= len(p.buf)
		p.clear()
		packetPool.Put(p)
	} else {
		p.bufOffset += n
	}

	return n, from, nil
}

// setReadBuffer sets a soft limit on the number of bytes available to read
// from the pipe.
func (pq *packetQueue) setReadBuffer(bytes int) error {
	if bytes <= 0 {
		return os.NewSyscallError("setReadBuffer", syscall.EINVAL)
	}
	q := pq.get() // Use the queue as a lock.
	q.readBufferBytes = bytes
	pq.put(q)
	return nil
}

type deadlineTimer struct {
	timer   chan *time.Timer
	expired chan struct{}
}

func newDeadlineTimer(deadline time.Time) *deadlineTimer {
	dt := &deadlineTimer{
		timer:   make(chan *time.Timer, 1),
		expired: make(chan struct{}),
	}
	dt.timer <- nil
	dt.Reset(deadline)
	return dt
}

// Reset attempts to reset the timer.
// If the timer has already expired, Reset returns false.
func (dt *deadlineTimer) Reset(deadline time.Time) bool {
	timer := <-dt.timer
	defer func() { dt.timer <- timer }()

	if deadline.Equal(noDeadline) {
		if timer != nil && timer.Stop() {
			timer = nil
		}
		return timer == nil
	}

	d := time.Until(deadline)
	if d < 0 {
		// Ensure that a deadline in the past takes effect immediately.
		defer func() { <-dt.expired }()
	}

	if timer == nil {
		timer = time.AfterFunc(d, func() { close(dt.expired) })
		return true
	}
	if !timer.Stop() {
		return false
	}
	timer.Reset(d)
	return true
}

func sysSocket(family, sotype, proto int) (int, error) {
	return 0, os.NewSyscallError("sysSocket", syscall.ENOSYS)
}

func fakeListen(fd *netFD, laddr sockaddr) (err error) {
	ffd := newFakeNetFD(fd)
	defer func() {
		if fd.fakeNetFD != ffd {
			// Failed to register listener; clean up.
			ffd.Close()
		}
	}()

	fd.laddr, err = assignFakeAddr(fd.net, fd.family, laddr, ffd.reservedPort)
	if err != nil {
		return os.NewSyscallError("listen", err)
	}

	ffd.fakeAddr = fakeAddr(fd.laddr.(sockaddr))
	switch fd.sotype {
	case syscall.SOCK_STREAM, syscall.SOCK_SEQPACKET:
		ffd.incoming = make(chan []*netFD, 1)
		ffd.incomingEmpty = make(chan bool, 1)
		ffd.incomingEmpty <- true
	case syscall.SOCK_DGRAM:
		ffd.queue = newPacketQueue(defaultBuffer)
	default:
		return os.NewSyscallError("listen", syscall.EINVAL)
	}

	fd.fakeNetFD = ffd
	if _, dup := sockets.LoadOrStore(ffd.fakeAddr, fd); dup {
		fd.fakeNetFD = nil
		return os.NewSyscallError("listen", syscall.EADDRINUSE)
	}

	return nil
}

func fakeConnect(fd *netFD, laddr, raddr sockaddr) error {
	if fd.isConnected {
		return os.NewSyscallError("connect", syscall.EISCONN)
	}

	var err error
	fd.laddr, err = assignFakeAddr(fd.net, fd.family, laddr, fd.fakeNetFD.reservedPort)
	if err != nil {
		return os.NewSyscallError("connect", err)
	}

	if !resolvedAddrIsValid(raddr) {
		return os.NewSyscallError("connect", syscall.EINVAL)
	}
	fd.raddr = raddr

	fd.fakeNetFD.queue = newPacketQueue(defaultBuffer)

	switch fd.sotype {
	case syscall.SOCK_DGRAM:
		if ua, ok := fd.laddr.(*UnixAddr); !ok || ua.Name != "" {
			fd.fakeNetFD.fakeAddr = fakeAddr(fd.laddr.(sockaddr))
			if _, dup := sockets.LoadOrStore(fd.fakeNetFD.fakeAddr, fd); dup {
				return os.NewSyscallError("connect", syscall.EADDRINUSE)
			}
		}
		fd.isConnected = true
		return nil

	case syscall.SOCK_STREAM, syscall.SOCK_SEQPACKET:
	default:
		return os.NewSyscallError("connect", syscall.EINVAL)
	}

	fa := fakeAddr(raddr)
	lni, ok := sockets.Load(fa)
	if !ok {
		return os.NewSyscallError("connect", syscall.ECONNREFUSED)
	}
	ln := lni.(*netFD)
	if ln.sotype != fd.sotype {
		return os.NewSyscallError("connect", syscall.EPROTOTYPE)
	}
	if ln.incoming == nil {
		return os.NewSyscallError("connect", syscall.ECONNREFUSED)
	}

	peer := &netFD{
		family:      ln.family,
		sotype:      ln.sotype,
		net:         ln.net,
		laddr:       ln.laddr,
		raddr:       fd.laddr,
		isConnected: true,
	}
	peer.fakeNetFD = newFakeNetFD(fd)
	peer.fakeNetFD.queue = newPacketQueue(defaultBuffer)
	defer func() {
		if fd.peer != peer {
			// Failed to connect; clean up.
			peer.Close()
		}
	}()

	var incoming []*netFD
	select {
	case ok = <-ln.incomingEmpty:
	case incoming, ok = <-ln.incoming:
	}
	if !ok {
		return os.NewSyscallError("connect", syscall.ECONNREFUSED)
	}
	defer func() {
		ln.incoming <- append(incoming, peer)
	}()

	fd.isConnected = true
	fd.peer = peer
	peer.peer = fd
	return nil
}

func (ffd *fakeNetFD) readFrom(p []byte) (n int, sa syscall.Sockaddr, err error) {
	if ffd.queue == nil {
		return 0, nil, os.NewSyscallError("readFrom", syscall.EINVAL)
	}

	n, from, err := ffd.queue.recvfrom(ffd.readDeadline.Load(), p, true, nil)

	if from != nil {
		// Convert the net.sockaddr to a syscall.Sockaddr type.
		var saErr error
		sa, saErr = from.sockaddr(ffd.fd.family)
		if err == nil {
			err = saErr
		}
	}

	return n, sa, err
}

func (ffd *fakeNetFD) readFromInet4(p []byte, sa *syscall.SockaddrInet4) (n int, err error) {
	n, _, err = ffd.queue.recvfrom(ffd.readDeadline.Load(), p, true, func(from sockaddr) error {
		fromSA, err := from.sockaddr(syscall.AF_INET)
		if err != nil {
			return err
		}
		if fromSA == nil {
			return os.NewSyscallError("readFromInet4", syscall.EINVAL)
		}
		*sa = *(fromSA.(*syscall.SockaddrInet4))
		return nil
	})
	return n, err
}

func (ffd *fakeNetFD) readFromInet6(p []byte, sa *syscall.SockaddrInet6) (n int, err error) {
	n, _, err = ffd.queue.recvfrom(ffd.readDeadline.Load(), p, true, func(from sockaddr) error {
		fromSA, err := from.sockaddr(syscall.AF_INET6)
		if err != nil {
			return err
		}
		if fromSA == nil {
			return os.NewSyscallError("readFromInet6", syscall.EINVAL)
		}
		*sa = *(fromSA.(*syscall.SockaddrInet6))
		return nil
	})
	return n, err
}

func (ffd *fakeNetFD) readMsg(p []byte, oob []byte, flags int) (n, oobn, retflags int, sa syscall.Sockaddr, err error) {
	if flags != 0 {
		return 0, 0, 0, nil, os.NewSyscallError("readMsg", syscall.ENOTSUP)
	}
	n, sa, err = ffd.readFrom(p)
	return n, 0, 0, sa, err
}

func (ffd *fakeNetFD) readMsgInet4(p []byte, oob []byte, flags int, sa *syscall.SockaddrInet4) (n, oobn, retflags int, err error) {
	if flags != 0 {
		return 0, 0, 0, os.NewSyscallError("readMsgInet4", syscall.ENOTSUP)
	}
	n, err = ffd.readFromInet4(p, sa)
	return n, 0, 0, err
}

func (ffd *fakeNetFD) readMsgInet6(p []byte, oob []byte, flags int, sa *syscall.SockaddrInet6) (n, oobn, retflags int, err error) {
	if flags != 0 {
		return 0, 0, 0, os.NewSyscallError("readMsgInet6", syscall.ENOTSUP)
	}
	n, err = ffd.readFromInet6(p, sa)
	return n, 0, 0, err
}

func (ffd *fakeNetFD) writeMsg(p []byte, oob []byte, sa syscall.Sockaddr) (n int, oobn int, err error) {
	if len(oob) > 0 {
		return 0, 0, os.NewSyscallError("writeMsg", syscall.ENOTSUP)
	}
	n, err = ffd.writeTo(p, sa)
	return n, 0, err
}

func (ffd *fakeNetFD) writeMsgInet4(p []byte, oob []byte, sa *syscall.SockaddrInet4) (n int, oobn int, err error) {
	return ffd.writeMsg(p, oob, sa)
}

func (ffd *fakeNetFD) writeMsgInet6(p []byte, oob []byte, sa *syscall.SockaddrInet6) (n int, oobn int, err error) {
	return ffd.writeMsg(p, oob, sa)
}

func (ffd *fakeNetFD) writeTo(p []byte, sa syscall.Sockaddr) (n int, err error) {
	raddr := ffd.fd.raddr
	if sa != nil {
		if ffd.fd.isConnected {
			return 0, os.NewSyscallError("writeTo", syscall.EISCONN)
		}
		raddr = ffd.fd.addrFunc()(sa)
	}
	if raddr == nil {
		return 0, os.NewSyscallError("writeTo", syscall.EINVAL)
	}

	peeri, _ := sockets.Load(fakeAddr(raddr.(sockaddr)))
	if peeri == nil {
		if len(ffd.fd.net) >= 3 && ffd.fd.net[:3] == "udp" {
			return len(p), nil
		}
		return 0, os.NewSyscallError("writeTo", syscall.ECONNRESET)
	}
	peer := peeri.(*netFD)
	if peer.queue == nil {
		if len(ffd.fd.net) >= 3 && ffd.fd.net[:3] == "udp" {
			return len(p), nil
		}
		return 0, os.NewSyscallError("writeTo", syscall.ECONNRESET)
	}

	block := true
	if len(ffd.fd.net) >= 3 && ffd.fd.net[:3] == "udp" {
		block = false
	}
	return peer.queue.send(ffd.writeDeadline.Load(), p, ffd.fd.laddr.(sockaddr), block)
}

func (ffd *fakeNetFD) writeToInet4(p []byte, sa *syscall.SockaddrInet4) (n int, err error) {
	return ffd.writeTo(p, sa)
}

func (ffd *fakeNetFD) writeToInet6(p []byte, sa *syscall.SockaddrInet6) (n int, err error) {
	return ffd.writeTo(p, sa)
}

func (ffd *fakeNetFD) dup() (f *os.File, err error) {
	return nil, os.NewSyscallError("dup", syscall.ENOSYS)
}

func (ffd *fakeNetFD) setReadBuffer(bytes int) error {
	if ffd.queue == nil {
		return os.NewSyscallError("setReadBuffer", syscall.EINVAL)
	}
	ffd.queue.setReadBuffer(bytes)
	return nil
}

func (ffd *fakeNetFD) setWriteBuffer(bytes int) error {
	return os.NewSyscallError("setWriteBuffer", syscall.ENOTSUP)
}

func (ffd *fakeNetFD) setLinger(sec int) error {
	if sec < 0 || ffd.peer == nil {
		return os.NewSyscallError("setLinger", syscall.EINVAL)
	}
	ffd.peer.queue.setLinger(sec > 0)
	return nil
}
