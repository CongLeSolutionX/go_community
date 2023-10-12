// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package work

import (
	"fmt"
	"os"
	"sync"
)

// A Shell runs shell commands and performs shell-like file system operations.
//
// Shell tracks context related to running commands, and form a tree much like
// context.Context.
type Shell struct {
	workDir string // $WORK
	out     *shellOut
	action  *Action // nil for the root shell
}

// shellOut is a shell output printing stream. It may be shared between several
// Shell instances.
type shellOut struct {
	lock      sync.Mutex
	printFunc func(args ...any) (int, error)
	scriptDir string // current directory in printed script
}

// NewShell returns a new Shell that defaults to printing to stderr.
func NewShell(workDir string) *Shell {
	sh := Shell{
		workDir: workDir,
	}
	return sh.WithPrint(func(a ...any) (int, error) {
		return fmt.Fprint(os.Stderr, a...)
	})
}

// Print emits a to this Shell's output stream, formatting it like fmt.Print.
// It is safe to call concurrently.
func (sh *Shell) Print(a ...any) {
	sh.out.lock.Lock()
	defer sh.out.lock.Unlock()
	sh.out.printFunc(a...)
}

func (sh *Shell) printLocked(a ...any) {
	sh.out.printFunc(a...)
}

// WithPrint returns a Shell identical to sh, but with the given Print function.
//
// This function also introduces a new shell printing stream. Shell ensures that
// calls to print are serialized, but multiple calls to WithPrint with the same
// print function will result in Shells that are separately serialized.
func (sh *Shell) WithPrint(print func(a ...any) (int, error)) *Shell {
	sh2 := *sh
	sh2.out = &shellOut{printFunc: print}
	return &sh2
}

// WithAction returns a Shell identical to sh, but bound to Action a.
func (sh *Shell) WithAction(a *Action) *Shell {
	sh2 := *sh
	sh2.action = a
	return &sh2
}

// Sh returns a shell for running commands on behalf of Action a.
func (b *Builder) Sh(a *Action) *Shell {
	if a.sh == nil {
		a.sh = b.rootSh.WithAction(a)
	}
	return a.sh
}
