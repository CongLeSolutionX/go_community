package main_test

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"

	"cmd/go/internal/help"
)

func TestDocsUpToDate(t *testing.T) {
	buf := new(bytes.Buffer)
	// Match the command in mkalldocs.sh that generates alldocs.go.
	help.Help(buf, []string{"documentation"})
	data, err := ioutil.ReadFile("alldocs.go")
	if err != nil {
		t.Fatalf("error reading alldocs.go: %v", err)
	}
	if !reflect.DeepEqual(data, buf.Bytes()) {
		t.Errorf("alldocs.go is not up to date; run mkalldocs.sh to regenerate it")
	}
}
