package net

import (
	"runtime"
	"syscall"
)

type Poller struct {
	fd *netFD
}

type PollerType int

const (
	PollReader PollerType = iota
	PollWriter
)

// Call calls fn for non-IO operation.
//
// The user-defined function fn takes a socket descriptor and does
// non-IO operation.
// In fn, calling any method or function provided by the net package
// is prohibited.
func (p *Poller) Call(fn func(s uintptr) (err error)) error {
	if err := p.fd.pfd.Incref(); err != nil {
		return err
	}
	defer p.fd.pfd.Decref()
	err := fn(uintptr(p.fd.pfd.Sysfd))
	runtime.KeepAlive(p.fd)
	return err
}

// Run starts IO operation associated with fn.
//
// The user-defined function fn takes a socket descriptor, does read
// or write operation, and returns true when accomplished.
// In fn, calling any method or function provided by the net package
// is prohibited.
func (p *Poller) Run(typ PollerType, fn func(s uintptr) (done bool, err error)) error {
	switch typ {
	case PollReader:
		if err := p.fd.pfd.ReadLock(); err != nil {
			return err
		}
		defer p.fd.pfd.ReadUnlock()
		if err := p.fd.pfd.PrepareRead(); err != nil {
			return err
		}
	case PollWriter:
		if err := p.fd.pfd.WriteLock(); err != nil {
			return err
		}
		defer p.fd.pfd.WriteUnlock()
		if err := p.fd.pfd.PrepareWrite(); err != nil {
			return err
		}
	default:
		return syscall.EINVAL
	}
	for {
		iocomp, err := fn(uintptr(p.fd.pfd.Sysfd))
		runtime.KeepAlive(p.fd)
		if iocomp {
			return err
		}
		if typ == PollReader {
			if err := p.fd.pfd.WaitRead(); err != nil {
				return err
			}
		} else {
			if err := p.fd.pfd.WaitWrite(); err != nil {
				return err
			}
		}
	}
}

func (c *UDPConn) Poller() *Poller {
	return &Poller{fd: c.fd}
}
