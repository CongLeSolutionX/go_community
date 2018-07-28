package modfetch

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"cmd/go/internal/modfetch/codehost"
)

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	dir, err := ioutil.TempDir("", "gitrepo-test-")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	codehost.WorkRoot = dir
	return m.Run()
}
