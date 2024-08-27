package os_test

import (
	. "os"
	"strings"
	"syscall"
	"testing"
)

func benchmarkGetwd(b *testing.B) {
	b.ResetTimer()
	var wd string
	for range b.N {
		var err error
		wd, err = Getwd()
		if err != nil {
			b.Fatal(err)
		}
	}
	b.Logf("benchmarkGetwd: %q", wd)
}

func BenchmarkGetwd(b *testing.B) {
	benchmarkGetwd(b)
}

func BenchmarkGetwdNoPWD(b *testing.B) {
	// os.Getwd checks if PWD value is absolute, so setting it to "" has
	// the same effect as unsetting (but is easier to do from here).
	b.Setenv("PWD", "")
	benchmarkGetwd(b)
}

func BenchmarkGetwdBadPWD(b *testing.B) {
	b.Setenv("PWD", "/absolute/but/wrong/path")
	benchmarkGetwd(b)
}

func TestGetwdDeep(t *testing.T) {
	testGetwdDeep(t, false)
}

func TestGetwdDeepWithPWDSet(t *testing.T) {
	testGetwdDeep(t, true)
}

// testGetwdDeep checks that os.Getwd is able to return paths
// longer than syscall.PathMax (with or without PWD set).
func testGetwdDeep(t *testing.T, setPWD bool) {
	dir := t.TempDir()
	t.Chdir(dir)

	if setPWD {
		t.Setenv("PWD", dir)
	} else {
		// When testing os.Getwd, setting PWD to empty string
		// is the same as unsetting it, but the latter would
		// be more complicated since we don't have t.Unsetenv.
		t.Setenv("PWD", "")
	}

	name := strings.Repeat("a", 200)
	for {
		if err := Mkdir(name, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := Chdir(name); err != nil {
			t.Fatal(err)
		}
		if setPWD {
			dir += "/" + name
			if err := Setenv("PWD", dir); err != nil {
				t.Fatal(err)
			}
			t.Logf(" $PWD len: %d", len(dir))
		}

		wd, err := Getwd()
		t.Logf("Getwd len: %d", len(wd))
		if err != nil {
			t.Fatal(err)
		}
		if len(wd) > syscall.PathMax*2 { // Success.
			break
		}
	}
}
