// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tls

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// This is an implementation of a shim for the BoringSSL test suite, also known
// as BoGo. The shim is a binary invoked by the BoGo test runner. Use bogo.bash
// to run BoGo against this. A test binary for this package (as compiled by "go
// test -c") will behave as a BoGo shim if invoked with the -bogo flag, or if
// renamed to "crypto-tls-bogo-shim". Using a test binary allows us to access
// private testing hooks, and provides the standard coverage and profiling
// flags. No build tags because things hidden by build tags break.

func bogoMode() bool {
	flag.Usage = func() {
		// TODO(filippo): print better usage hints for the shim mode.
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		if path := os.Getenv("SHIM_UNIMPLEMENTED_FLAGS_LOG"); path != "" {
			f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				f.WriteString(strings.Join(os.Args[1:], " "))
				f.WriteString("\n")
				f.Close()
			}
		}

		os.Exit(89) // Signal unimplemented features to the BoGo runner.
	}

	var bogoFlag = flag.Bool("bogo", false, "run in BoGo shim mode")
	flag.Parse()
	return *bogoFlag || filepath.Base(os.Args[0]) == "crypto-tls-bogo-shim"
}

// TODO(filippo): use a separate FlagSet.
var (
	handshakerSupportedFlag = flag.Bool("is-handshaker-supported", false, "just say No [BoGo shim]")
	_                       = flag.String("handshaker-path", "", "ignored by this shim [BoGo shim]")
	_                       = flag.Bool("async", false, "ignored by this shim [BoGo shim]")
	portFlag                = flag.Int("port", 0, "runner port on localhost [BoGo shim]")
	serverFlag              = flag.Bool("server", false, "act as a TLS server [BoGo shim]")
	certFileFlag            = flag.String("cert-file", "", "path of the server certificate [BoGo shim]")
	keyFileFlag             = flag.String("key-file", "", "path of the server certificate key [BoGo shim]")
	resumeCountFlag         = flag.Int("resume-count", 0, "number of resumption connections to make [BoGo shim]")
)

// bogoShimMain is invoked by TestMain before anything else.
func bogoShimMain() {
	if !bogoMode() {
		return
	}

	if *handshakerSupportedFlag {
		fmt.Printf("No\n")
		os.Exit(0)
	}

	if *portFlag == 0 {
		log.Fatalf("BoGo shim mode requires -port.\n")
	}

	config := &Config{
		InsecureSkipVerify: true,
		ClientSessionCache: NewLRUClientSessionCache(32),
	}
	if *certFileFlag != "" {
		cert, err := LoadX509KeyPair(*certFileFlag, *keyFileFlag)
		if err != nil {
			log.Fatalf("Error loading server certificate: %v\n", err)
		}
		config.Certificates = []Certificate{cert}
	}

	for i := 0; i <= *resumeCountFlag; i++ {
		if i > 0 && !*serverFlag {
			if len(config.ClientSessionCache.(*lruSessionCache).m) == 0 {
				log.Fatalf("No sessions cached on resumption.")
			}
		}

		conn, err := net.Dial("tcp", net.JoinHostPort("localhost", strconv.Itoa(*portFlag)))
		if err != nil {
			log.Fatalf("Error establishing connection: %v\n", err)
		}

		var tlsConn *Conn
		if *serverFlag {
			tlsConn = Server(conn, config)
		} else {
			tlsConn = Client(conn, config)
		}

		if err := tlsConn.Handshake(); err != nil {
			log.Fatalf("Error running handshake: %v\n", err)
		}

		buf := make([]byte, 512)
		for {
			n, err := tlsConn.Read(buf)
			if err != nil && err != io.EOF {
				log.Fatalf("Error reading: %v", err)
			}

			for i := range buf[:n] {
				buf[i] ^= 0xff
			}

			if _, err := tlsConn.Write(buf[:n]); err != nil {
				log.Fatalf("Error writing: %v", err)
			}

			if err == io.EOF {
				break
			}
		}
	}

	os.Exit(0)
}
