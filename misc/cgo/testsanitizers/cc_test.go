package sanitizers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
)

var overcommit struct {
	sync.Once
	value int
	err   error
}

// requireOvercommit skips t if the kernel does not allow overcommit.
func requireOvercommit(t *testing.T) {
	t.Helper()

	overcommit.Once.Do(func() {
		var out []byte
		out, overcommit.err = exec.Command("sysctl", "-n", "vm.overcommit_memory").Output()
		if overcommit.err != nil {
			return
		}
		overcommit.value, overcommit.err = strconv.Atoi(string(bytes.TrimSpace(out)))
	})

	if overcommit.err != nil {
		t.Logf("couldn't determine vm.overcommit_memory (%v); assuming no overcommit", overcommit.err)
	}
	if overcommit.value == 2 {
		t.Skip("vm.overcommit_memory=2")
	}
}

var env struct {
	sync.Once
	m   map[string]string
	err error
}

// goEnv returns the output of $(go env) as a map.
func goEnv(key string) (string, error) {
	env.Once.Do(func() {
		var out []byte
		out, env.err = exec.Command("go", "env", "-json").Output()
		if env.err != nil {
			return
		}

		env.m = make(map[string]string)
		env.err = json.Unmarshal(out, &env.m)
	})
	if env.err != nil {
		return "", env.err
	}

	v, ok := env.m[key]
	if !ok {
		return "", fmt.Errorf("`go env`: no entry for %v", key)
	}
	return v, nil
}

// replaceEnv sets the key environment variable to value in cmd.
func replaceEnv(cmd *exec.Cmd, key, value string) {
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	prefix := key + "="
	for i, entry := range cmd.Env {
		if strings.HasPrefix(entry, prefix) {
			cmd.Env = append(cmd.Env[:i], cmd.Env[i:]...)
			break
		}
	}
	cmd.Env = append(cmd.Env, prefix+value)
}

// mustRun executes t and fails cmd with a well-formatted message if it fails.
func mustRun(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%#q exited with %v\n%s", strings.Join(cmd.Args, " "), err, out)
	}
}

// cc returns a cmd that executes `$(go env CC) $(go env GOGCCFLAGS) $args`.
func cc(args ...string) (*exec.Cmd, error) {
	CC, err := goEnv("CC")
	if err != nil {
		return nil, err
	}

	GOGCCFLAGS, err := goEnv("GOGCCFLAGS")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(CC, strings.Split(GOGCCFLAGS, " ")...)
	cmd.Args = append(cmd.Args, args...)
	return cmd, nil
}

type version struct {
	name         string
	major, minor int
}

var compiler struct {
	sync.Once
	version
	err error
}

// compilerVersion detects the version of $(go env CC).
//
// It returns a non-nil error if the compiler matches a known version schema but
// the version could not be parsed, or if $(go env CC) could not be determined.
func compilerVersion() (version, error) {
	compiler.Once.Do(func() {
		compiler.err = func() error {
			compiler.name = "unknown"

			cmd, err := cc("--version")
			if err != nil {
				return err
			}
			out, err := cmd.Output()
			if err != nil {
				// Compiler does not support "--version" flag: not Clang or GCC.
				return nil
			}

			var match [][]byte
			if bytes.HasPrefix(out, []byte("gcc")) {
				compiler.name = "gcc"

				cmd, err := cc("-dumpversion")
				if err != nil {
					return err
				}
				out, err := cmd.Output()
				if err != nil {
					// gcc, but does not support gcc's "-dumpversion" flag?!
					return err
				}
				gccRE := regexp.MustCompile(`(\d+)\.(\d+)`)
				match = gccRE.FindSubmatch(out)
			} else {
				clangRE := regexp.MustCompile(`clang version (\d+)\.(\d+)`)
				if match = clangRE.FindSubmatch(out); len(match) > 0 {
					compiler.name = "clang"
				}
			}

			if len(match) < 3 {
				return nil // "unknown"
			}
			if compiler.major, err = strconv.Atoi(string(match[1])); err != nil {
				return err
			}
			if compiler.minor, err = strconv.Atoi(string(match[2])); err != nil {
				return err
			}
			return nil
		}()
	})
	return compiler.version, compiler.err
}

type config struct {
	sanitizer                string
	cFlags, ldFlags, goFlags []string

	checkCompilerOnce sync.Once
	err               error
	skip              bool // If true, skip with err instead of failing with it.
}

var configs sync.Map // map[string]

// configure returns the configuration for the given sanitizer.
func configure(sanitizer string) *config {
	if i, ok := configs.Load(sanitizer); ok {
		return i.(*config)
	}

	c := &config{
		sanitizer: sanitizer,
		cFlags:    []string{"-fsanitize=" + sanitizer},
		ldFlags:   []string{"-fsanitize=" + sanitizer},
	}

	switch sanitizer {
	case "memory":
		c.goFlags = append(c.goFlags, "-msan")

	case "thread":
		c.goFlags = append(c.goFlags, "--installsuffix=tsan")
		compiler, _ := compilerVersion()
		if compiler.name == "gcc" {
			c.cFlags = append(c.cFlags, "-fPIE")
			c.ldFlags = append(c.ldFlags, "-fPIE", "-pie")
		}

	default:
		panic(fmt.Sprintf("unrecognized sanitizer: %q", sanitizer))
	}

	i, _ := configs.LoadOrStore(sanitizer, c)
	c = i.(*config)
	return c
}

// goCmd returns a Cmd that executes "go $subcommand $args" with appropriate
// additional flags and environment.
func (c *config) goCmd(subcommand string, args ...string) *exec.Cmd {
	cmd := exec.Command("go", subcommand)
	cmd.Args = append(cmd.Args, c.goFlags...)
	cmd.Args = append(cmd.Args, args...)
	replaceEnv(cmd, "CGO_CFLAGS", strings.Join(c.cFlags, " "))
	replaceEnv(cmd, "CGO_LDFLAGS", strings.Join(c.ldFlags, " "))
	return cmd
}

// skipIfCCompilerBroken skips t if the C compiler does not produce working
// binaries as configured.
func (c *config) skipIfCCompilerBroken(t *testing.T) {
	c.checkCompilerOnce.Do(func() {
		c.skip, c.err = c.checkCCompiler()
	})
	if c.err != nil {
		t.Helper()
		if c.skip {
			t.Skip(c.err)
		}
		t.Fatal(c.err)
	}
}

// skipIfRuntimeIncompatible skips t if the Go runtime is suspected not to work
// with cgo as configured.
func (c *config) skipIfRuntimeIncompatible(t *testing.T) {
	if c.sanitizer != "thread" {
		return
	}

	t.Helper()

	compiler, err := compilerVersion()
	if err != nil {
		t.Logf("failed to detect compiler version: %v", err)
	}
	if compiler.name == "gcc" && compiler.major < 7 {
		// GCC before version 7 does not set __SANITIZE_THREAD__ when TSAN is in
		// use, which prevents runtime/cgo/libcgo.h from installing runtime hooks.
		// See https://golang.org/issue/15983.
		t.Skipf("gcc %d.%d (older than 7.0) does not set __SANITIZE_THREAD__", compiler.major, compiler.minor)
	}
}

var cMain = []byte(`
int main() {
	return 0;
}
`)

func (c *config) checkCCompiler() (skip bool, err error) {
	dir, err := ioutil.TempDir("", c.sanitizer)
	if err != nil {
		return false, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	src := filepath.Join(dir, "return0.c")
	if err := ioutil.WriteFile(src, cMain, 0600); err != nil {
		return false, fmt.Errorf("failed to write C source file: %v", err)
	}

	dst := filepath.Join(dir, "return0")
	cmd, err := cc(c.cFlags...)
	if err != nil {
		return false, err
	}
	cmd.Args = append(cmd.Args, c.ldFlags...)
	cmd.Args = append(cmd.Args, "-o", dst, src)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if bytes.Contains(out, []byte("unrecognized")) &&
			bytes.Contains(out, []byte("-fsanitize")) {
			return true, errors.New(string(out))
		}
		return true, fmt.Errorf("%#q failed: %v\n%s", strings.Join(cmd.Args, " "), err, out)
	}

	if _, err := exec.Command(dst).CombinedOutput(); err != nil {
		if os.IsNotExist(err) {
			return true, fmt.Errorf("%#q failed to produce executable: %v", strings.Join(cmd.Args, " "), err)
		}
		return true, fmt.Errorf("%#q generated broken executable: %v", strings.Join(cmd.Args, " "), err)
	}

	return false, nil
}

// srcPath returns the path to the given file relative to this test's source tree.
func srcPath(path string) string {
	return filepath.Join("src", path)
}

// A tempDir manages a temporary directory within a test.
type tempDir struct {
	base string
}

func (d *tempDir) RemoveAll(t *testing.T) {
	t.Helper()
	if d.base == "" {
		return
	}
	if err := os.RemoveAll(d.base); err != nil {
		t.Fatal("Failed to remove temp dir: %v", err)
	}
}

func (d *tempDir) Join(name string) string {
	return filepath.Join(d.base, name)
}

func newTempDir(t *testing.T) *tempDir {
	t.Helper()
	dir, err := ioutil.TempDir("", filepath.Dir(t.Name()))
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return &tempDir{base: dir}
}
