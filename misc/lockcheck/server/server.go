// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package server implements a runtime lock log server that collects
// the combined lock graph of all Go processes in a tree of processes
// and constructs a lockgraph.Graph.
package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"lockcheck/lockgraph"
	"lockcheck/locklog"
)

// Run starts a lock log server, runs the subcommand given by cmd, and
// collects the lock graph reported to the lock server by the
// subcommand. It returns true if it was able to collect any lock
// graph, even if there was an error (such as the subcommand failing).
func Run(cmd *exec.Cmd, builder *GraphBuilder) (bool, error) {
	// Start the lock log server.
	server, err := NewServer(builder)
	if err != nil {
		return false, err
	}
	defer server.Shutdown()

	// Start subcommand.
	if err := startCommand(cmd, server); err != nil {
		return false, err
	}
	defer func() {
		if cmd != nil {
			cmd.Process.Kill()
		}
	}()

	// Run server and monitor subcommand.
	stop := make(chan error, 2)
	go func() {
		err := server.Run()
		stop <- err
		cmd.Process.Kill()
	}()
	go func() {
		err := cmd.Wait()
		if err != nil {
			err = fmt.Errorf("subcommand exited with: %w", err)
		}
		stop <- err
		server.Shutdown()
	}()
	err = <-stop
	<-stop

	return true, err
}

// startCommand starts the subcommand given by cmd and tells it to
// connect to server.
func startCommand(cmd *exec.Cmd, server *Server) error {
	// Set up environment to tell the command where the log is.
	golocklog := fmt.Sprintf("GOLOCKLOG=%s", server.SockPath)
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, golocklog)

	// Start command.
	return cmd.Start()
}

type Server struct {
	SockPath  string
	sockDir   string
	l         net.Listener
	lockGraph *GraphBuilder
	gLock     sync.Mutex
	wg        sync.WaitGroup
}

func NewServer(graph *GraphBuilder) (*Server, error) {
	tmpDir, err := ioutil.TempDir("", "lockcheck-")
	if err != nil {
		return nil, fmt.Errorf("creating temporary directory for log socket: %w", err)
	}
	sockPath := filepath.Join(tmpDir, "s")
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		os.Remove(tmpDir)
		return nil, fmt.Errorf("creating log socket: %w", err)
	}
	return &Server{SockPath: sockPath, sockDir: tmpDir, l: l, lockGraph: graph}, nil
}

func (s *Server) Shutdown() {
	s.l.Close()
	os.RemoveAll(s.sockDir)
}

func (s *Server) Run() error {
	for {
		conn, err := s.l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				err = nil
			}
			return err
		}

		s.wg.Add(1)
		go s.run1(conn)
	}
}

func (s *Server) run1(conn net.Conn) {
	defer s.wg.Done()

	// Process the log.
	var proc *Proc
	load := func(exePath string, runtimeMain uint64) error {
		// Load the process.
		var err error
		proc, err = NewProc(string(exePath), "runtime.main", runtimeMain)
		return err
	}
	logReader, err := locklog.NewLogReader(conn, load)
	if err != nil {
		conn.Close()
		log.Print(err)
	}

	// Process the lock log.
	var rec locklog.Record
	g := s.lockGraph
	for {
		err := logReader.Next(&rec)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Print(err)
			break
		}

		thr := proc.Thread(rec.M)
		var stack lockgraph.Stack
		if rec.Op == locklog.LogOpAcquire || rec.Op == locklog.LogOpMayAcquire {
			// Resolve the stack outside of the lock.
			stack = proc.Stack(g.stacks, rec.Stack)
		}
		func() {
			s.gLock.Lock()
			defer s.gLock.Unlock()

			switch rec.Op {
			default:
				log.Fatal("unhandled log op", rec.Op)

			case locklog.LogOpAcquire:
				if rec.LockClass == "" {
					g.Acquire(thr, rec.LockAddr, stack)
				} else {
					g.AcquireLabeled(thr, rec.LockAddr, stack, rec.LockClass, rec.LockRank)
				}

			case locklog.LogOpRelease:
				g.Release(thr, rec.LockAddr)

			case locklog.LogOpMayAcquire:
				g.MayAcquire(thr, rec.LockAddr, stack)
			}
		}()
	}
}
