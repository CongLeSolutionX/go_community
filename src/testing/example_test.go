package testing

import "os"

var f = os.Stdout

func TestCaptureStdout(t *T) {
	matchString := func(string, string) (bool, error) { return true, nil }
	examples := []InternalExample{{
		F: func() {
			f.WriteString("test output")
		},
		Output: "test output",
	}}

	stdout := *os.Stdout
	ok := RunExamples(matchString, examples)
	if !ok {
		t.Error("stdout was not captured")
	}
	if stdout != *os.Stdout {
		*os.Stdout = stdout
		t.Error("stdout was not restored ")
	}
}
