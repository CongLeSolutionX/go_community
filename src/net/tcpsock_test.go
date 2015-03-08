// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"internal/testenv"
	"io"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func BenchmarkTCP4OneShot(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, false, false, ln, &d)
}

func BenchmarkTCP4OneShotTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, false, true, ln, &d)
}

func BenchmarkTCP4Persistent(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, true, false, ln, &d)
}

func BenchmarkTCP4PersistentTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, true, true, ln, &d)
}

func BenchmarkTCP6OneShot(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, false, false, ln, &d)
}

func BenchmarkTCP6OneShotTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, false, true, ln, &d)
}

func BenchmarkTCP6Persistent(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, true, false, ln, &d)
}

func BenchmarkTCP6PersistentTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, true, true, ln, &d)
}

func benchmarkTCP(b *testing.B, persistent, timeout bool, ln Listener, d *Dialer) {
	defer ln.Close()

	const msgLen = 512
	conns := b.N
	numConcurrent := runtime.GOMAXPROCS(-1) * 2
	msgs := 1
	if persistent {
		conns = numConcurrent
		msgs = b.N / conns
		if msgs == 0 {
			msgs = 1
		}
		if conns > b.N {
			conns = b.N
		}
	}
	sendMsg := func(c Conn, buf []byte) bool {
		n, err := c.Write(buf)
		if n != len(buf) || err != nil {
			b.Log(err)
			return false
		}
		return true
	}
	recvMsg := func(c Conn, buf []byte) bool {
		for read := 0; read != len(buf); {
			n, err := c.Read(buf)
			read += n
			if err != nil {
				b.Log(err)
				return false
			}
		}
		return true
	}
	serverSem := make(chan bool, numConcurrent)
	// Acceptor.
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				break
			}
			serverSem <- true
			// Server connection.
			go func(c Conn) {
				defer func() {
					c.Close()
					<-serverSem
				}()
				if timeout {
					c.SetDeadline(time.Now().Add(time.Hour)) // Not intended to fire.
				}
				var buf [msgLen]byte
				for m := 0; m < msgs; m++ {
					if !recvMsg(c, buf[:]) || !sendMsg(c, buf[:]) {
						break
					}
				}
			}(c)
		}
	}()
	clientSem := make(chan bool, numConcurrent)
	for i := 0; i < conns; i++ {
		clientSem <- true
		// Client connection.
		go func() {
			defer func() {
				<-clientSem
			}()
			c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
			if err != nil {
				b.Log(err)
				return
			}
			defer c.Close()
			if timeout {
				c.SetDeadline(time.Now().Add(time.Hour)) // Not intended to fire.
			}
			var buf [msgLen]byte
			for m := 0; m < msgs; m++ {
				if !sendMsg(c, buf[:]) || !recvMsg(c, buf[:]) {
					break
				}
			}
		}()
	}
	for i := 0; i < numConcurrent; i++ {
		clientSem <- true
		serverSem <- true
	}
}

func BenchmarkTCP4ConcurrentReadWrite(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCPConcurrentReadWrite(b, ln, &d)
}

func BenchmarkTCP6ConcurrentReadWrite(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCPConcurrentReadWrite(b, ln, &d)
}

func benchmarkTCPConcurrentReadWrite(b *testing.B, ln Listener, d *Dialer) {
	defer ln.Close()

	// The benchmark creates GOMAXPROCS client/server pairs.
	// Each pair creates 4 goroutines: client reader/writer and server reader/writer.
	// The benchmark stresses concurrent reading and writing to the same connection.
	// Such pattern is used in net/http and net/rpc.

	P := runtime.GOMAXPROCS(0)
	N := b.N / P
	W := 1000

	// Setup P client/server connections.
	clients := make([]Conn, P)
	servers := make([]Conn, P)

	done := make(chan bool)
	go func() {
		for p := 0; p < P; p++ {
			s, err := ln.Accept()
			if err != nil {
				b.Error(err)
				return
			}
			var buf [1]byte
			if _, err := s.Read(buf[:]); err != nil {
				s.Close()
				b.Error(err)
				return
			}
			servers[p] = s
		}
		done <- true
	}()
	for p := 0; p < P; p++ {
		c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			b.Fatal(err)
		}
		var buf [1]byte
		if _, err := c.Write(buf[:]); err != nil {
			b.Fatal(err)
		}
		clients[p] = c
	}
	<-done

	b.ResetTimer()

	var wg sync.WaitGroup
	wg.Add(4 * P)
	for p := 0; p < P; p++ {
		// Client writer.
		go func(c Conn) {
			defer wg.Done()
			var buf [1]byte
			for i := 0; i < N; i++ {
				v := byte(i)
				for w := 0; w < W; w++ {
					v *= v
				}
				buf[0] = v
				if _, err := c.Write(buf[:]); err != nil {
					b.Error(err)
					return
				}
			}
		}(clients[p])

		// Pipe between server reader and server writer.
		pipe := make(chan byte, 128)

		// Server reader.
		go func(s Conn) {
			defer wg.Done()
			var buf [1]byte
			for i := 0; i < N; i++ {
				if _, err := s.Read(buf[:]); err != nil {
					b.Error(err)
					return
				}
				pipe <- buf[0]
			}
		}(servers[p])

		// Server writer.
		go func(s Conn) {
			defer wg.Done()
			var buf [1]byte
			for i := 0; i < N; i++ {
				v := <-pipe
				for w := 0; w < W; w++ {
					v *= v
				}
				buf[0] = v
				if _, err := s.Write(buf[:]); err != nil {
					b.Error(err)
					return
				}
			}
			s.Close()
		}(servers[p])

		// Client reader.
		go func(c Conn) {
			defer c.Close()
			defer wg.Done()
			var buf [1]byte
			for i := 0; i < N; i++ {
				if _, err := c.Read(buf[:]); err != nil {
					b.Error(err)
					return
				}
			}
		}(clients[p])
	}
	wg.Wait()
}

type resolveTCPAddrTest struct {
	network       string
	litAddrOrName string
	addr          *TCPAddr
	err           error
}

var resolveTCPAddrTests = []resolveTCPAddrTest{
	{"tcp", "127.0.0.1:0", &TCPAddr{IP: IPv4(127, 0, 0, 1), Port: 0}, nil},
	{"tcp4", "127.0.0.1:65535", &TCPAddr{IP: IPv4(127, 0, 0, 1), Port: 65535}, nil},

	{"tcp", "[::1]:0", &TCPAddr{IP: ParseIP("::1"), Port: 0}, nil},
	{"tcp6", "[::1]:65535", &TCPAddr{IP: ParseIP("::1"), Port: 65535}, nil},

	{"tcp", "[::1%en0]:1", &TCPAddr{IP: ParseIP("::1"), Port: 1, Zone: "en0"}, nil},
	{"tcp6", "[::1%911]:2", &TCPAddr{IP: ParseIP("::1"), Port: 2, Zone: "911"}, nil},

	{"", "127.0.0.1:0", &TCPAddr{IP: IPv4(127, 0, 0, 1), Port: 0}, nil}, // Go 1.0 behavior
	{"", "[::1]:0", &TCPAddr{IP: ParseIP("::1"), Port: 0}, nil},         // Go 1.0 behavior

	{"tcp", ":12345", &TCPAddr{Port: 12345}, nil},

	{"http", "127.0.0.1:0", nil, UnknownNetworkError("http")},
}

func TestResolveTCPAddr(t *testing.T) {
	origTestHookLookupIP := testHookLookupIP
	defer func() { testHookLookupIP = origTestHookLookupIP }()
	testHookLookupIP = lookupLocalhost

	for i, tt := range resolveTCPAddrTests {
		addr, err := ResolveTCPAddr(tt.network, tt.litAddrOrName)
		if err != tt.err {
			t.Errorf("#%d: %v", i, err)
		} else if !reflect.DeepEqual(addr, tt.addr) {
			t.Errorf("#%d: got %#v; want %#v", i, addr, tt.addr)
		}
		if err != nil {
			continue
		}
		rtaddr, err := ResolveTCPAddr(addr.Network(), addr.String())
		if err != nil {
			t.Errorf("#%d: %v", i, err)
		} else if !reflect.DeepEqual(rtaddr, addr) {
			t.Errorf("#%d: got %#v; want %#v", i, rtaddr, addr)
		}
	}
}

var tcpListenerNameTests = []struct {
	net   string
	laddr *TCPAddr
}{
	{"tcp4", &TCPAddr{IP: IPv4(127, 0, 0, 1)}},
	{"tcp4", &TCPAddr{}},
	{"tcp4", nil},
}

func TestTCPListenerName(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)

	for _, tt := range tcpListenerNameTests {
		ln, err := ListenTCP(tt.net, tt.laddr)
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		la := ln.Addr()
		if a, ok := la.(*TCPAddr); !ok || a.Port == 0 {
			t.Fatalf("got %v; expected a proper address with non-zero port number", la)
		}
	}
}

func TestIPv6LinkLocalUnicastTCP(t *testing.T) {
	testenv.MustHaveExternalNetwork(t)

	if !supportsIPv6 {
		t.Skip("IPv6 is not supported")
	}

	for i, tt := range ipv6LinkLocalUnicastTCPTests {
		ln, err := Listen(tt.network, tt.address)
		if err != nil {
			// It might return "LookupHost returned no
			// suitable address" error on some platforms.
			t.Log(err)
			continue
		}
		ls, err := (&streamListener{Listener: ln}).newLocalServer()
		if err != nil {
			t.Fatal(err)
		}
		defer ls.teardown()
		ch := make(chan error, 1)
		handler := func(ls *localServer, ln Listener) { transponder(ln, ch) }
		if err := ls.buildup(handler); err != nil {
			t.Fatal(err)
		}
		if la, ok := ln.Addr().(*TCPAddr); !ok || !tt.nameLookup && la.Zone == "" {
			t.Fatalf("got %v; expected a proper address with zone identifier", la)
		}

		c, err := Dial(tt.network, ls.Listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer c.Close()
		if la, ok := c.LocalAddr().(*TCPAddr); !ok || !tt.nameLookup && la.Zone == "" {
			t.Fatalf("got %v; expected a proper address with zone identifier", la)
		}
		if ra, ok := c.RemoteAddr().(*TCPAddr); !ok || !tt.nameLookup && ra.Zone == "" {
			t.Fatalf("got %v; expected a proper address with zone identifier", ra)
		}

		if _, err := c.Write([]byte("TCP OVER IPV6 LINKLOCAL TEST")); err != nil {
			t.Fatal(err)
		}
		b := make([]byte, 32)
		if _, err := c.Read(b); err != nil {
			t.Fatal(err)
		}

		for err := range ch {
			t.Errorf("#%d: %v", i, err)
		}
	}
}

func TestTCPConcurrentAccept(t *testing.T) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))
	ln, err := Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	const N = 10
	var wg sync.WaitGroup
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			for {
				c, err := ln.Accept()
				if err != nil {
					break
				}
				c.Close()
			}
			wg.Done()
		}()
	}
	attempts := 10 * N
	fails := 0
	d := &Dialer{Timeout: 200 * time.Millisecond}
	for i := 0; i < attempts; i++ {
		c, err := d.Dial("tcp", ln.Addr().String())
		if err != nil {
			fails++
		} else {
			c.Close()
		}
	}
	ln.Close()
	wg.Wait()
	if fails > attempts/9 { // see issues 7400 and 7541
		t.Fatalf("too many Dial failed: %v", fails)
	}
	if fails > 0 {
		t.Logf("# of failed Dials: %v", fails)
	}
}

func TestTCPReadWriteAllocs(t *testing.T) {
	switch runtime.GOOS {
	case "nacl", "windows":
		// NaCl needs to allocate pseudo file descriptor
		// stuff. See syscall/fd_nacl.go.
		// Windows uses closures and channels for IO
		// completion port-based netpoll. See fd_windows.go.
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	ln, err := Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	var server Conn
	errc := make(chan error)
	go func() {
		var err error
		server, err = ln.Accept()
		errc <- err
	}()
	client, err := Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	var buf [128]byte
	allocs := testing.AllocsPerRun(1000, func() {
		_, err := server.Write(buf[:])
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.ReadFull(client, buf[:])
		if err != nil {
			t.Fatal(err)
		}
	})
	if allocs > 0 {
		t.Fatalf("got %v; want 0", allocs)
	}
}

func TestTCPStress(t *testing.T) {
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	nmsgs := int(1e4)
	if testing.Short() {
		nmsgs = 1e2
	}
	testTCPStress(t, ln, 2, nmsgs, 512)
}

func testTCPStress(t *testing.T, ln Listener, nclients, nmsgs, msgLen int) {
	defer ln.Close()
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		t.Fatal(err)
	}
	defer ls.teardown()
	ch := make(chan error, nclients)
	handler := func(ls *localServer, ln Listener) {
		persistentTransponder(ln, nmsgs, msgLen, nil, ch)
	}
	if err := ls.buildup(handler); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(nclients)
	for i := 0; i < nclients; i++ {
		go func() {
			defer wg.Done()
			c, err := Dial(ln.Addr().Network(), ln.Addr().String())
			if err != nil {
				if perr := parseDialError(err); perr != nil {
					t.Error(perr)
				}
				t.Error(err)
				return
			}
			defer c.Close()
			b := make([]byte, msgLen)
			for m := 0; m < nmsgs; m++ {
				nw, err := c.Write(b)
				if err != nil {
					if perr := parseWriteError(err); perr != nil {
						t.Error(perr)
					}
					t.Error(err)
					return
				}
				nr, err := io.ReadFull(c, b)
				if err != nil {
					if perr := parseReadError(err); perr != nil {
						t.Error(perr)
					}
					t.Error(err)
					return
				}
				if nr != nw {
					t.Errorf("got %d bytes written; want %d", nr, nw)
				}
			}
		}()
	}
	wg.Wait()

	ls.teardown()
	for err := range ch {
		t.Error(err)
	}
}

func TestTCPSelfConnect(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO(brainman): do not know why it hangs.
		t.Skip("known-broken test on windows")
	}

	// Try to connect to that address repeatedly.
	nattempts := 10000
	if testing.Short() {
		nattempts = 1000
	}
	if runtime.GOOS != "linux" {
		// Non-Linux systems take a long time to figure out
		// that there is nothing listening on localhost.
		nattempts = 100
	}
	var d Dialer
	testTCPSelfConnect(t, &d, nattempts)
}

func testTCPSelfConnect(t *testing.T, d *Dialer, nattempts int) {
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	type twsock struct { // time-wait state socket
		TCPAddr
		error
	}
	ch := make(chan twsock)
	go func() {
		var tws twsock
		defer func() { ch <- tws }()
		c, err := ln.Accept()
		if err != nil {
			tws.error = err
			return
		}
		addr := c.LocalAddr().(*TCPAddr)
		tws.TCPAddr.IP = make(IP, len(addr.IP))
		copy(tws.TCPAddr.IP, addr.IP)
		tws.TCPAddr.Port = addr.Port
		tws.TCPAddr.Zone = addr.Zone
		c.Close()
	}()

	c, err := Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		ln.Close()
		t.Fatal(err)
	}
	tws := <-ch
	if tws.error != nil {
		t.Fatal(tws.error)
	}
	c.Close()
	network := ln.Addr().Network()
	ln.Close()

	d.Timeout = time.Millisecond
	if tws.TCPAddr.IP.To4() != nil {
		d.LocalAddr = &TCPAddr{IP: IPv4(127, 0, 0, 1)}
	}
	if tws.TCPAddr.IP.To16() != nil && tws.TCPAddr.IP.To4() == nil {
		d.LocalAddr = &TCPAddr{IP: IPv6loopback}
	}
	for i := 0; i < nattempts; i++ {
		c, err := d.Dial(network, tws.TCPAddr.String())
		if d.FastOpen {
			_, err = c.Write([]byte("TCP FAST OPEN"))
		}
		if err != nil {
			continue
		}
		laddr := c.LocalAddr().(*TCPAddr)
		if laddr.Port == tws.TCPAddr.Port && laddr.IP.Equal(tws.TCPAddr.IP) {
			t.Errorf("Dial %v->%v should fail", laddr, &tws.TCPAddr)
		}
		if !d.FastOpen {
			t.Logf("Dial %v->%v succeeded - possibly racing with other listener", laddr, &tws.TCPAddr)
		}
		c.Close()
	}
}
