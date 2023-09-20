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
//
// TODO: Add a RemoveAll method. "rm -rf" is pretty common.
type Shell struct {
	workDir string // $WORK
	out     *shellOut
	action  *Action      // nil for the root shell
	caches  *shellCaches // per-Builder caches
}

// shellOut is a shell output printing stream. It may be shared between several
// Shell instances.
type shellOut struct {
	lock      sync.Mutex
	printFunc func(args ...any) (int, error)
	scriptDir string // current directory in printed script
}

// shellCaches is the caches shared by all Shells derived by a Builder.
type shellCaches struct {
	mkdirLock  sync.Mutex
	mkdirCache map[string]bool // a cache of created directories
}

// NewShell returns a new Shell that defaults to printing to stderr.
func NewShell(workDir string) *Shell {
	sh := Shell{
		workDir: workDir,
	}
	sh.caches = &shellCaches{
		mkdirCache: make(map[string]bool),
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

// Shell returns a shell for running commands on behalf of Action a.
func (b *Builder) Shell(a *Action) *Shell {
	if a == nil {
		// The root shell has a nil Action. The point of this method is to
		// create a Shell bound to an Action, so disallow nil Actions here.
		panic("nil Action")
	}
	if a.sh == nil {
		a.sh = b.backgroundSh.WithAction(a)
	}
	return a.sh
}

// BackgroundShell returns a Builder-wide Shell that's not bound to any Action.
// Try not to use this unless there's really no sensible Action available.
func (b *Builder) BackgroundShell() *Shell {
	return b.backgroundSh
}
