// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program runs go vet on the standard library and commands.
// It compares the output against a set of whitelists maintained in this directory.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"internal/testenv"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type whitelist map[string]int

var (
	bytesGOOS     = []byte("GOOS")
	bytesGOARCH   = []byte("GOARCH")
	bytesARCHSUFF = []byte("ARCHSUFF")
	bytesComment  = []byte("//")
)

// load adds entries from the whitelist file, if present, for os/arch to w.
func (w whitelist) load(goos string, goarch string) {
	// Look up whether goarch is a 32 bit or 64 bit architecture.
	archbits, ok := nbits[goarch]
	if !ok {
		log.Fatal("unknown bitwidth for arch %q", goarch)
	}

	// Look up whether goarch has a shared arch suffix,
	// such as mips64x for mips64 and mips64le.
	archsuff := goarch
	if x, ok := archAsmX[goarch]; ok {
		archsuff = x
	}

	// Load whitelists.
	filenames := []string{
		"all.txt",
		goos + ".txt",
		goarch + ".txt",
		goos + "_" + goarch + ".txt",
		fmt.Sprintf("%dbit.txt", archbits),
	}
	if goarch != archsuff {
		filenames = append(filenames,
			archsuff+".txt",
			goos+"_"+archsuff+".txt",
		)
	}

	// We allow error message templates using GOOS and GOARCH.
	goosb := []byte(goos)
	if goos == "android" {
		goosb = []byte("linux") // so many special cases :(
	}
	goarchb := []byte(goarch)
	archsuffb := []byte(archsuff)

	// Read whitelists and do template substitution.
	for _, filename := range filenames {
		buf, err := ioutil.ReadFile(filepath.Join("whitelist", filename))
		if err != nil {
			// Allow not-exist errors; not all combinations have whitelists.
			if os.IsNotExist(err) {
				continue
			}
			log.Fatal(err)
		}

		lines := bytes.Split(buf, []byte{'\n'})
		for _, line := range lines {
			if len(line) == 0 || bytes.HasPrefix(line, bytesComment) {
				continue
			}
			// Where art thou, bytes.Replacer?
			line = bytes.Replace(line, bytesGOOS, goosb, -1)
			line = bytes.Replace(line, bytesGOARCH, goarchb, -1)
			line = bytes.Replace(line, bytesARCHSUFF, archsuffb, -1)
			w[string(line)]++
		}
	}
}

type platform struct {
	os   string
	arch string
}

func (p platform) String() string {
	return p.os + "/" + p.arch
}

type platformsFlag []platform

func (p *platformsFlag) String() string {
	var buf bytes.Buffer
	for i, s := range *p {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s.String())
	}
	return buf.String()
}

func (p *platformsFlag) Set(value string) error {
	vv := strings.Split(value, "/")
	if len(vv) != 2 {
		return fmt.Errorf("could not parse platform %s, must be of form goos/goarch", value)
	}
	*p = append(*p, platform{os: vv[0], arch: vv[1]})
	return nil
}

var (
	flagPlatforms platformsFlag
	flagAll       = flag.Bool("all", false, "run all platforms")
	flagNoLines   = flag.Bool("n", false, "don't print line numbers")
)

func init() {
	flag.Var(&flagPlatforms, "p", "platform to use, e.g. linux/amd64; can be repeated")
}

// ignorePathPrefixes are file path prefixes that should be ignored wholesale.
var ignorePathPrefixes = [...]string{
	// These testdata dirs have lots of intentionally broken/bad code for tests.
	"cmd/go/testdata/",
	"cmd/vet/testdata/",
	"go/printer/testdata/",
	// cmd/compile/internal/big is a vendored copy of math/big.
	// Ignore it so that we only have to deal with math/big issues once.
	"cmd/compile/internal/big/",
}

func (p platform) vet() {
	if p.arch == "s390x" {
		// TODO: reinstate when s390x gets vet support (issue 15454)
		return
	}
	fmt.Printf("go run main.go -p %s\n", p)

	// Load whitelist(s).
	w := make(whitelist)
	w.load(p.os, p.arch)

	env := append(os.Environ(), "GOOS="+p.os, "GOARCH="+p.arch)

	// Do 'go install std' before running vet.
	// It is cheap when already installed.
	// Not installing leads to non-obvious failures due to inability to typecheck.
	// TODO: If go/loader ever makes it to the standard library, have vet use it,
	// at which point vet can work off source rather than compiled packages.
	cmd := exec.Command("go", "install", "std")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("failed to run GOOS=%s GOARCH=%s 'go install std': %v\n%s", p.os, p.arch, err, out)
	}

	// 'go tool vet .' is considerably faster than 'go vet ./...'
	// The unsafeptr checks are disabled for now,
	// because there are so many false positives,
	// and no clear way to improve vet to eliminate large chunks of them.
	// And having them in the whitelists will just cause annoyance
	// and churn when working on the runtime.
	cmd = exec.Command("go", "tool", "vet", "-unsafeptr=false", ".")
	cmd.Dir = filepath.Join(runtime.GOROOT(), "src")
	cmd.Env = env
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Process vet output.
	scan := bufio.NewScanner(stderr)
NextLine:
	for scan.Scan() {
		line := scan.Text()
		if strings.HasPrefix(line, "vet: ") {
			// Typecheck failure: Malformed syntax or multiple packages or the like.
			// This will yield nicer error messages elsewhere, so ignore them here.
			continue
		}

		fields := strings.SplitN(line, ":", 3)
		var file, lineno, msg string
		switch len(fields) {
		case 2:
			// vet message with no line number
			file, msg = fields[0], fields[1]
		case 3:
			file, lineno, msg = fields[0], fields[1], fields[2]
		default:
			log.Fatalf("could not parse vet output line:\n%s", line)
		}
		msg = strings.TrimSpace(msg)

		for _, ignore := range ignorePathPrefixes {
			if strings.HasPrefix(file, filepath.FromSlash(ignore)) {
				continue NextLine
			}
		}

		key := file + ": " + msg
		if w[key] == 0 {
			// Vet error with no match in the whitelist. Print it.
			if *flagNoLines {
				fmt.Printf("%s: %s\n", file, msg)
			} else {
				fmt.Printf("%s:%s: %s\n", file, lineno, msg)
			}
			continue
		}
		w[key]--
	}
	if scan.Err() != nil {
		log.Fatalf("failed to scan vet output: %v", scan.Err())
	}
	err = cmd.Wait()
	// We expect vet to fail.
	// Make sure it has failed appropriately, though (for example, not a PathError).
	if _, ok := err.(*exec.ExitError); !ok {
		log.Fatalf("unexpected go vet execution failure: %v", err)
	}
	printedHeader := false
	if len(w) > 0 {
		for k, v := range w {
			if v != 0 {
				if !printedHeader {
					fmt.Println("unmatched whitelist entries")
					printedHeader = true
				}
				for i := 0; i < v; i++ {
					fmt.Println(k)
				}
			}
		}
	}
}

func main() {
	if !testenv.HasGoBuild() {
		log.Print("no cmd/go; skipping")
		return
	}

	flag.Parse()
	if *flagAll && len(flagPlatforms) != 0 {
		log.Fatalf("-all and -p flags are incompatible\n")
	}
	if len(flagPlatforms) == 0 {
		flagPlatforms = append(flagPlatforms, platform{os: build.Default.GOOS, arch: build.Default.GOARCH})
	}
	if *flagAll {
		cmd := exec.Command("go", "tool", "dist", "list")
		out, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		lines := bytes.Split(out, []byte{'\n'})
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			flag.Set("p", string(line))
		}
	}

	for _, p := range flagPlatforms {
		p.vet()
	}
}

// nbits maps from architecture names to the number of bits in a pointer.
// TODO: figure out a clean way to avoid get this info rather than listing it here yet again.
var nbits = map[string]int{
	"386":      32,
	"amd64":    64,
	"amd64p32": 32,
	"arm":      32,
	"arm64":    64,
	"mips64":   64,
	"mips64le": 64,
	"ppc64":    64,
	"ppc64le":  64,
}

// archAsmX maps architectures to the suffix usually used for their assembly files,
// if different than the arch name itself.
var archAsmX = map[string]string{
	"android":  "linux",
	"mips64":   "mips64x",
	"mips64le": "mips64x",
	"ppc64":    "ppc64x",
	"ppc64le":  "ppc64x",
}
