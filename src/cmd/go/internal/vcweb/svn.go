// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcweb

import (
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

type svnHandler struct {
	svnRoot string
	logger  *log.Logger

	pathOnce     sync.Once
	svnservePath string
	svnserveErr  error

	listenOnce sync.Once
	s          chan *svnState
}

type svnState struct {
	listener  net.Listener
	listenErr error
	conns     map[net.Conn]struct{}
	closing   bool
	done      chan struct{}
}

func (h *svnHandler) Available() bool {
	h.pathOnce.Do(func() {
		h.svnservePath, h.svnserveErr = exec.LookPath("svnserve")
	})
	return h.svnserveErr == nil
}

func (h *svnHandler) serve(c net.Conn) {
	defer func() {
		c.Close()

		s := <-h.s
		delete(s.conns, c)
		if len(s.conns) == 0 && s.listenErr != nil {
			close(s.done)
		}
		h.s <- s
	}()

	cmd := exec.Command(h.svnservePath, "--read-only", "--root="+h.svnRoot, "--inetd")
	cmd.Stdin = c
	cmd.Stdout = c
	stderr := new(strings.Builder)
	cmd.Stderr = stderr
	err := cmd.Run()

	var errFrag any = "ok"
	if err != nil {
		errFrag = err
	}
	stderrFrag := ""
	if stderr.Len() > 0 {
		stderrFrag = "\n" + stderr.String()
	}
	h.logger.Printf("%v: %s%s", cmd, errFrag, stderrFrag)
}

func (h *svnHandler) Handler(dir string, env []string, logger *log.Logger) (http.Handler, error) {
	if !h.Available() {
		return nil, ServerNotInstalledError{name: "hg"}
	}

	h.listenOnce.Do(func() {
		h.s = make(chan *svnState, 1)
		l, err := net.Listen("tcp", "localhost:0")
		done := make(chan struct{})

		h.s <- &svnState{
			listener:  l,
			listenErr: err,
			conns:     map[net.Conn]struct{}{},
			done:      done,
		}
		if err != nil {
			close(done)
			return
		}

		h.logger.Printf("serving svn on svn://%v", l.Addr())

		go func() {
			for {
				c, err := l.Accept()

				s := <-h.s
				if err != nil {
					s.listenErr = err
					if len(s.conns) == 0 {
						close(s.done)
					}
					h.s <- s
					return
				}
				if s.closing {
					c.Close()
				} else {
					s.conns[c] = struct{}{}
					go h.serve(c)
				}
				h.s <- s
			}
		}()
	})

	s := <-h.s
	addr := ""
	if s.listener != nil {
		addr = s.listener.Addr().String()
	}
	err := s.listenErr
	h.s <- s
	if err != nil {
		return nil, err
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.FormValue("vcwebsvn") != "" {
			w.Header().Add("Content-Type", "text/plain; charset=UTF-8")
			io.WriteString(w, "svn://"+addr+"\n")
			return
		}
		http.NotFound(w, req)
	})

	return handler, nil
}

func (h *svnHandler) Close() error {
	h.listenOnce.Do(func() {})
	if h.s == nil {
		return nil
	}

	var err error
	s := <-h.s
	s.closing = true
	if s.listener == nil {
		err = s.listenErr
	} else {
		err = s.listener.Close()
	}
	for c, _ := range s.conns {
		c.Close()
	}
	done := s.done
	h.s <- s

	<-done
	return err
}
