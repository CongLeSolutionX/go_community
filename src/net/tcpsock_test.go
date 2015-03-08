// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"fmt"
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

	b.StopTimer()
	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, false, false)
}

func BenchmarkTCP4OneShotTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, false, true)
}

func BenchmarkTCP4Persistent(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, true, false)
}

func BenchmarkTCP4PersistentTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv4 {
		b.Skip("IPv4 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp4")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, true, true)
}

func BenchmarkTCP6OneShot(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, false, false)
}

func BenchmarkTCP6OneShotTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, false, true)
}

func BenchmarkTCP6Persistent(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, true, false)
}

func BenchmarkTCP6PersistentTimeout(b *testing.B) {
	testHookUninstaller.Do(uninstallTestHooks)

	if !supportsIPv6 {
		b.Skip("IPv6 is not supported")
	}

	b.StopTimer()
	ln, err := newLocalListener("tcp6")
	if err != nil {
		b.Fatal(err)
	}
	var d Dialer
	benchmarkTCP(b, ln, &d, true, true)
}

func benchmarkTCP(b *testing.B, ln Listener, d *Dialer, persistent, timeout bool) {
	const msgLen = 512
	clients := b.N
	numConcurrent := runtime.GOMAXPROCS(-1) * 2
	msgs := 1
	if persistent {
		clients = numConcurrent
		msgs = b.N / clients
		if msgs == 0 {
			msgs = 1
		}
		if clients > b.N {
			clients = b.N
		}
	}

	serverThrottle := make(chan struct{}, numConcurrent)
	ch := make(chan error, 1)
	handler := func(ls *localServer, ln Listener) {
		tcpTransponder(ln, msgs, msgLen, serverThrottle, ch)
	}
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		ln.Close()
		b.Fatal(err)
	}
	if err := ls.buildup(handler); err != nil {
		ls.teardown()
		b.Fatal(err)
	}

	b.StartTimer()

	clientThrottle := make(chan struct{}, numConcurrent)
	var wg sync.WaitGroup
	wg.Add(clients)
	for i := 0; i < clients; i++ {
		clientThrottle <- struct{}{}
		go func(i int) {
			defer func() {
				<-clientThrottle
				wg.Done()
			}()
			c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
			if err != nil {
				b.Logf("TR#%d: %v", i, err)
				return
			}
			defer c.Close()
			if timeout {
				c.SetDeadline(time.Now().Add(time.Hour)) // Not intended to fire.
			}
			var buf [msgLen]byte
			tc := &testTCPConn{Conn: c, prefix: fmt.Sprintf("TR#%d", i), ch: ch}
			for m := 0; m < msgs; m++ {
				if !tc.write(buf[:]) || !tc.read(buf[:tc.nw]) {
					return
				}
				tc.reset()
			}
		}(i)
	}
	wg.Wait()

	ls.teardown()
	for err := range ch {
		b.Error(err)
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
	benchmarkTCPConcurrentReadWrite(b, ln)
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
	benchmarkTCPConcurrentReadWrite(b, ln)
}

func benchmarkTCPConcurrentReadWrite(b *testing.B, ln Listener) {
	// The benchmark creates GOMAXPROCS client/server pairs.
	// Each pair creates 4 goroutines: client reader/writer and server reader/writer.
	// The benchmark stresses concurrent reading and writing to the same connection.
	// Such pattern is used in net/http and net/rpc.

	b.StopTimer()

	P := runtime.GOMAXPROCS(0)
	N := b.N / P
	W := 1000

	// Setup P client/server connections.
	clients := make([]Conn, P)
	servers := make([]Conn, P)
	done := make(chan bool)
	defer ln.Close()
	go func() {
		for p := 0; p < P; p++ {
			s, err := ln.Accept()
			if err != nil {
				b.Error(err)
				return
			}
			servers[p] = s
		}
		done <- true
	}()
	for p := 0; p < P; p++ {
		c, err := Dial(ln.Addr().Network(), ln.Addr().String())
		if err != nil {
			b.Fatal(err)
		}
		clients[p] = c
	}
	<-done

	b.StartTimer()

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
				_, err := c.Write(buf[:])
				if err != nil {
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
				_, err := s.Read(buf[:])
				if err != nil {
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
				_, err := s.Write(buf[:])
				if err != nil {
					b.Error(err)
					return
				}
			}
			s.Close()
		}(servers[p])

		// Client reader.
		go func(c Conn) {
			defer wg.Done()
			var buf [1]byte
			for i := 0; i < N; i++ {
				_, err := c.Read(buf[:])
				if err != nil {
					b.Error(err)
					return
				}
			}
			c.Close()
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
	if testing.Short() || !*testExternal {
		t.Skip("avoid external network")
	}

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
	if testing.Short() || !*testExternal {
		t.Skip("avoid external network")
	}
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
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	var d Dialer
	testTCPConcurrentAccept(t, ln, &d)
}

func testTCPConcurrentAccept(t *testing.T, ln Listener, d *Dialer) {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(4))
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
	d.Timeout = 200 * time.Millisecond
	for i := 0; i < attempts; i++ {
		c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
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
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	var d Dialer
	testTCPReadWriteAllocs(t, ln, &d)
}

func testTCPReadWriteAllocs(t *testing.T, ln Listener, d *Dialer) {
	switch runtime.GOOS {
	case "nacl", "windows":
		// NaCl needs to allocate pseudo file descriptor
		// stuff. See syscall/fd_nacl.go.
		// Windows uses closures and channels for IO
		// completion port-based netpoll. See fd_windows.go.
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	defer ln.Close()
	var b [128]byte
	var server Conn
	errc := make(chan error)
	go func() {
		var err error
		server, err = ln.Accept()
		errc <- err
	}()
	client, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if d.FastOpen {
		if _, err := client.Write(b[:1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := <-errc; err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if d.FastOpen {
		if _, err := server.Read(b[:1]); err != nil {
			t.Fatal(err)
		}
	}
	allocs := testing.AllocsPerRun(1000, func() {
		_, err := server.Write(b[:])
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.ReadFull(client, b[:])
		if err != nil {
			t.Fatal(err)
		}
	})
	if allocs > 0 {
		t.Fatalf("got %v; want 0", allocs)
	}
}

func TestTCPSelfConnect(t *testing.T) {
	if runtime.GOOS == "windows" {
		// TODO(brainman): do not know why it hangs.
		t.Skip("known-broken test on windows")
	}

	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	var d Dialer
	testTCPSelfConnect(t, ln, &d)
}

func testTCPSelfConnect(t *testing.T, ln Listener, d *Dialer) {
	c, err := d.Dial(ln.Addr().Network(), ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	network := c.LocalAddr().Network()
	laddr := *c.LocalAddr().(*TCPAddr)
	c.Close()
	ln.Close()

	// Try to connect to that address repeatedly.
	n := 100000
	if testing.Short() {
		n = 1000
	}
	switch runtime.GOOS {
	case "darwin", "dragonfly", "freebsd", "netbsd", "openbsd", "plan9", "solaris", "windows":
		// Non-Linux systems take a long time to figure
		// out that there is nothing listening on localhost.
		n = 100
	}
	for i := 0; i < n; i++ {
		d.Timeout = time.Millisecond
		c, err := d.Dial(network, laddr.String())
		if err == nil {
			addr := c.LocalAddr().(*TCPAddr)
			if addr.Port == laddr.Port || !d.FastOpen && addr.IP.Equal(laddr.IP) {
				t.Errorf("#%d: Dial %q should fail", i, addr)
			} else if !d.FastOpen {
				t.Logf("#%d: Dial %q succeeded - possibly racing with other listener", i, addr)
			}
			c.Close()
		}
	}
}

func TestTCPStress(t *testing.T) {
	ln, err := newLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}
	msgs := int(1e4)
	if testing.Short() {
		msgs = 1e2
	}
	testTCPStress(t, ln, 2, msgs, 512)
}

func testTCPStress(t *testing.T, ln Listener, clients, msgs, msgLen int) {
	ch := make(chan error, 1)
	handler := func(ls *localServer, ln Listener) {
		tcpTransponder(ln, msgs, msgLen, nil, ch)
	}
	ls, err := (&streamListener{Listener: ln}).newLocalServer()
	if err != nil {
		ln.Close()
		t.Fatal(err)
	}
	if err := ls.buildup(handler); err != nil {
		ls.teardown()
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(clients)
	for i := 0; i < clients; i++ {
		go func(i int) {
			defer wg.Done()
			c, err := Dial(ln.Addr().Network(), ln.Addr().String())
			if err != nil {
				if perr := parseDialError(err); perr != nil {
					ch <- fmt.Errorf("TR#%d: %v", i, perr)
				}
				t.Logf("TR#%d: %v", i, err)
				return
			}
			defer c.Close()
			b := make([]byte, msgLen)
			tc := &testTCPConn{Conn: c, prefix: fmt.Sprintf("TR#%d", i), ch: ch}
			for m := 0; m < msgs; m++ {
				if !tc.write(b) || !tc.read(b[:tc.nw]) {
					return
				}
				if tc.nr != tc.nw {
					ch <- fmt.Errorf("TR#%d: got %d bytes read; want %d", i, tc.nr, tc.nw)
				}
				tc.reset()
			}
		}(i)
	}
	wg.Wait()

	ls.teardown()
	for err := range ch {
		t.Error(err)
	}
}
