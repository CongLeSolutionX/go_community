package sanitizers_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const testGoSource = `
package main

/*
#cgo CFLAGS: -fsanitize=address
#cgo LDFLAGS: -fsanitize=address

#include <stdio.h>
#include <stdint.h>
#include <sanitizer/asan_interface.h>

typedef unsigned long uptr;

// This structure is used to describe the source location of
// a place where global was defined.
struct asan_global_source_location {
  const char *filename;
  int line_no;
  int column_no;
};

// This structure describes an instrumented global variable.
struct asan_global {
  uptr beg;
  uptr size;
  uptr size_with_redzone;
  const char *name;
  const char *module_name;
  uptr has_dynamic_init;
  struct asan_global_source_location *location;
  uptr odr_indicator;
};

// Register an array of globals.
void __asan_register_globals_go(void *g, uintptr_t n) {
        struct asan_global *globals = (struct asan_global *)(g);
        for (int i = 0; i < n; i ++ ) {
                printf("lile = %d\n", globals[i].location->line_no);
                printf("column = %d\n", globals[i].location->column_no);
                printf("filename = %s\n", globals[i].location->filename);
                printf("module_name = %s\n", globals[i].module_name);
                printf("name = %s\n", globals[i].name);
        }
        __asan_register_globals(globals, n);
}

void __asan_read_go(void *addr, uintptr_t sz) {
        if (__asan_region_is_poisoned(addr, sz)) {
                switch (sz) {
                case 1: __asan_report_load1(addr); break;
                case 2: __asan_report_load2(addr); break;
                case 4: __asan_report_load4(addr); break;
                case 8: __asan_report_load8(addr); break;
                default: __asan_report_load_n(addr, sz); break;
                }
        }
}
*/
import "C"
import "fmt"
import "unsafe"

// actural size is ff[10], allocates 56 bytes as redzone.
var ff [24]int32

type asanG struct {
  beg        uintptr
  size       uintptr
  sizeWithR  uintptr
  name       uintptr
  mName      uintptr
  hasDymInit uintptr
  location   uintptr
  odrInd     uintptr
}

type asanL struct {
  filename uintptr
  line     int32
  column   int32
}

type goStringDef struct {
  Data uintptr
  Len  int
}

var gs [2]asanG
var ll asanL
var sn string
var mn string

func init() {
  fmt.Println("intrument init ....")
  gs[0].beg = uintptr(unsafe.Pointer(&ff))
  gs[0].size = 40
  gs[0].sizeWithR = 40 + uintptr(getRedzoneSizeForGlobal(40))
  sn = "ff\000"
  gs[0].name = (*goStringDef)(unsafe.Pointer(&sn)).Data
  mn = "main\000"
  gs[0].mName = (*goStringDef)(unsafe.Pointer(&mn)).Data
  ll.filename = (*goStringDef)(unsafe.Pointer(&mn)).Data
  ll.line = int32(81)
  ll.column = int32(200)
  gs[0].location = uintptr(unsafe.Pointer(&ll))

  C.__asan_register_globals_go(unsafe.Pointer(&gs[0]), 1)
}

func getRedzoneSizeForGlobal(size int) int {
  maxRZ := 1 << 18
  minRZ := 32
  redZone := (size / minRZ / 4) * minRZ
  switch {
  case redZone > maxRZ:
          redZone = maxRZ
  case redZone < minRZ:
          redZone = minRZ
  }
  // Round up to multiple of minRZ.
  if size%minRZ != 0 {
          redZone += minRZ - (size % minRZ)
  }
  return redZone
}

func main() {
  fmt.Println("main...")
  C.__asan_read_go(unsafe.Pointer(&ff[11]), 4)
  ff[11] = 10 // BOOM
  fmt.Println("finish ...")
}
  
`

func TestASANGccGo(t *testing.T) {
	goos, err := goEnv("GOOS")
	if err != nil {
		t.Fatal(err)
	}
	goarch, err := goEnv("GOARCH")
	if err != nil {
		t.Fatal(err)
	}

	if !aSanSupported(goos, goarch) {
		t.Skipf("skipping on %s/%s; -asan option is not supported.", goos, goarch)
	}

	goCC, err := goEnv("CC")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CC", goCC)

	var buf bytes.Buffer
	buf.WriteString(testGoSource)
	src := buf.Bytes()
	dir := t.TempDir()
	sourcefile := filepath.Join(dir, "global.go")

	err = os.WriteFile(sourcefile, src, 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	out, err := exec.Command("go", "run", sourcefile).CombinedOutput()
	fmt.Println("test...")
	if err != nil {
		t.Errorf("the output: %s\nerr: %v\n", out, err)
	} else {
		t.Errorf("error: %v\n", err)
	}

}
