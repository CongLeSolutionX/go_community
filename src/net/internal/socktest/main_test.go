// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris windows

package socktest_test

import (
	"fmt"
	"net/internal/socktest"
	"os"
	"sync"
	"syscall"
	"testing"
)

var sw socktest.Switch

func TestMain(m *testing.M) {
	installTestHooks()
	sw.Register("TestSwitch", func(s uintptr, cookie socktest.Cookie) {
		sw.AddFilter(s, socktest.FilterClose, func(st *socktest.State) (socktest.AfterFilter, error) {
			sw.Cookie(s)
			if testing.Verbose() {
				fmt.Println(st)
			}
			return nil, nil
		})
	})

	st := m.Run()

	if n := len(sw.Sockets()); n != 0 {
		panic(fmt.Sprintf("got %d; want 0", n))
	}
	sw.Disable()
	sw.Deregister("TestSwitch")
	os.Exit(st)
}

func TestSwitch(t *testing.T) {
	done := make(chan struct{})
	var wg1, wg2 sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg1.Add(2)
		go func() {
			defer wg1.Done()
			socketFunc(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
		}()
		go func() {
			defer wg1.Done()
			socketFunc(syscall.AF_INET6, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
		}()
	}
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for {
			select {
			case <-done:
				for _, st := range sw.Sockets() {
					closeFunc(st.Sysfd)
				}
				return
			default:
				for _, st := range sw.Sockets() {
					closeFunc(st.Sysfd)
				}
			}
		}
	}()
	wg1.Wait()
	close(done)
	wg2.Wait()
}
