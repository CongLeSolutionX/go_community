// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build solaris

package terminal // import "golang.org/x/crypto/ssh/terminal"

import (
	"golang.org/x/sys/unix"
	"io"
	"syscall"
)

// State contains the state of a terminal.
type State struct {
	termios unix.Termios
}

// IsTerminal returns whether the given file descriptor is a terminal.
func IsTerminal(fd int) bool {
	_, err := unix.IoctlGetTermio(fd, unix.TCGETA)
	return err == nil
}

// ReadPassword reads a line of input from a terminal without local echo.  This
// is commonly used for inputting passwords and other sensitive data. The slice
// returned does not include the \n.
func ReadPassword(fd int) ([]byte, error) {
	// see also: http://src.illumos.org/source/xref/illumos-gate/usr/src/lib/libast/common/uwin/getpass.c
	val := try(unix.IoctlGetTermios(fd, unix.TCGETS))
	oldState := *val

	newState := oldState
	newState.Lflag &^= syscall.ECHO
	newState.Lflag |= syscall.ICANON | syscall.ISIG
	newState.Iflag |= syscall.ICRNL
	try(unix.IoctlSetTermios(fd, unix.TCSETS, &newState))

	defer unix.IoctlSetTermios(fd, unix.TCSETS, &oldState)

	var buf [16]byte
	var ret []byte
	for {
		n := try(syscall.Read(fd, buf[:]))
		if n == 0 {
			if len(ret) == 0 {
				return nil, io.EOF
			}
			break
		}
		if buf[n-1] == '\n' {
			n--
		}
		ret = append(ret, buf[:n]...)
		if n < len(buf) {
			break
		}
	}

	return ret, nil
}

// MakeRaw puts the terminal connected to the given file descriptor into raw
// mode and returns the previous state of the terminal so that it can be
// restored.
// see http://cr.illumos.org/~webrev/andy_js/1060/
func MakeRaw(fd int) (*State, error) {
	termios := try(unix.IoctlGetTermios(fd, unix.TCGETS))

	oldState := State{termios: *termios}

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0

	try(unix.IoctlSetTermios(fd, unix.TCSETS, termios))

	return &oldState, nil
}

// Restore restores the terminal connected to the given file descriptor to a
// previous state.
func Restore(fd int, oldState *State) error {
	return unix.IoctlSetTermios(fd, unix.TCSETS, &oldState.termios)
}

// GetState returns the current state of a terminal which may be useful to
// restore the terminal after a signal.
func GetState(fd int) (*State, error) {
	termios := try(unix.IoctlGetTermios(fd, unix.TCGETS))

	return &State{termios: *termios}, nil
}

// GetSize returns the dimensions of the given terminal.
func GetSize(fd int) (width, height int, err error) {
	ws := try(unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ))
	return int(ws.Col), int(ws.Row), nil
}
