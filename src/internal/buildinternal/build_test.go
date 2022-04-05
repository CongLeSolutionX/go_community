package buildinternal

import (
	"path/filepath"
	"testing"
)

func TestShellSafety(t *testing.T) {
	tests := []struct {
		input, srcdir, expected string
		result                  bool
	}{
		{"-I${SRCDIR}/../include", "/projects/src/issue 11868", "-I/projects/src/issue 11868/../include", true},
		{"-I${SRCDIR}", "~wtf$@%^", "-I~wtf$@%^", true},
		{"-X${SRCDIR}/1,${SRCDIR}/2", "/projects/src/issue 11868", "-X/projects/src/issue 11868/1,/projects/src/issue 11868/2", true},
		{"-I/tmp -I/tmp", "/tmp2", "-I/tmp -I/tmp", true},
		{"-I/tmp", "/tmp/[0]", "-I/tmp", true},
		{"-I${SRCDIR}/dir", "/tmp/[0]", "-I/tmp/[0]/dir", false},
		{"-I${SRCDIR}/dir", "/tmp/go go", "-I/tmp/go go/dir", true},
		{"-I${SRCDIR}/dir dir", "/tmp/go", "-I/tmp/go/dir dir", true},
	}
	for _, test := range tests {
		output, ok := expandSrcDir(test.input, test.srcdir)
		if ok != test.result {
			t.Errorf("Expected %t while %q expands to %q with SRCDIR=%q; got %t", test.result, test.input, output, test.srcdir, ok)
		}
		if output != test.expected {
			t.Errorf("Expected %q while %q expands with SRCDIR=%q; got %q", test.expected, test.input, test.srcdir, output)
		}
	}
}

var (
	expandSrcDirPath = filepath.Join(string(filepath.Separator)+"projects", "src", "add")
)

var expandSrcDirTests = []struct {
	input, expected string
}{
	{"-L ${SRCDIR}/libs -ladd", "-L /projects/src/add/libs -ladd"},
	{"${SRCDIR}/add_linux_386.a -pthread -lstdc++", "/projects/src/add/add_linux_386.a -pthread -lstdc++"},
	{"Nothing to expand here!", "Nothing to expand here!"},
	{"$", "$"},
	{"$$", "$$"},
	{"${", "${"},
	{"$}", "$}"},
	{"$FOO ${BAR}", "$FOO ${BAR}"},
	{"Find me the $SRCDIRECTORY.", "Find me the $SRCDIRECTORY."},
	{"$SRCDIR is missing braces", "$SRCDIR is missing braces"},
}

func TestExpandSrcDir(t *testing.T) {
	for _, test := range expandSrcDirTests {
		output, _ := expandSrcDir(test.input, expandSrcDirPath)
		if output != test.expected {
			t.Errorf("%q expands to %q with SRCDIR=%q when %q is expected", test.input, output, expandSrcDirPath, test.expected)
		} else {
			t.Logf("%q expands to %q with SRCDIR=%q", test.input, output, expandSrcDirPath)
		}
	}
}
