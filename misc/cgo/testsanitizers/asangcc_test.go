package sanitizers_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"fmt"
)

const testSource = `
#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <sanitizer/asan_interface.h>
#include <assert.h>

typedef unsigned long uptr;

struct __asan_global_source_location {
  const char *filename;
  int line_no;
  int column_no;
};

struct __asan_global {
  uptr beg;     
  uptr size;    
  uptr size_with_redzone;  
  const char *name;        
  const char *module_name; 
  uptr has_dynamic_init;   
  struct __asan_global_source_location *location;   // Source location of a global, or NULL if it is unknown.
  uptr odr_indicator;                        // The address of the ODR indicator symbol.
};

// actural size is ff[10], allocates 56 bytes as redzone.
// ff[10];
int ff[24] = {-1};

uint64_t getRedzoneSizeForGlobal(uint64_t size) {
  uint64_t maxRZ = 1<<18;
  uint64_t minRZ = 32;
  // Calculate RZ, where MinRZ <= RZ <= MaxRZ, and RZ ~ 1/4 * SizeInBytes.
  uint64_t RZ = (size / minRZ / 4)*minRZ;
  if (RZ > maxRZ) { RZ = maxRZ; }
  if (RZ < minRZ) { RZ = minRZ; }

  // Round up to multiple of minRZ.
  if (size % minRZ) { RZ += minRZ - (size % minRZ); }
  assert((RZ + size) % minRZ == 0);
  return RZ;
}


void __attribute__((constructor)) instrumentGlobals();

void instrumentGlobals() {
  printf("instrument...\n");
  struct __asan_global gs[1];
  // initialize a global as a instrumented global.
  uint64_t ffSize = 10*sizeof(int);
  gs[0].beg = &ff;
  gs[0].size = ffSize;
  gs[0].size_with_redzone = ffSize + getRedzoneSizeForGlobal(ffSize);
  gs[0].name = "ff";
  gs[0].module_name = "global.c";
  gs[0].odr_indicator = 0;
  gs[0].location = (struct __asan_global_source_location*)malloc(sizeof(struct __asan_global_source_location));
  gs[0].location->filename = "global.c";
  gs[0].location->line_no = 29;
  gs[0].location->column_no = 10;
  gs[0].has_dynamic_init = 0;

  // call __asan_register_global to instrument globals.
  __asan_register_globals(gs, 1);

}

int main() {
  printf("main\n");
   if (__asan_region_is_poisoned(&ff[11], sizeof(int))) {
    __asan_report_store4(&ff[11]);
  }
  ff[11] = 10;

  return 0;
}

`

func TestASANGcc(t *testing.T) {
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

	var buf bytes.Buffer
	buf.WriteString(testSource)
	src := buf.Bytes()
	dir := t.TempDir()
	fmt.Println("dir", dir)
	sourcefile := filepath.Join(dir, "global.c")
	objectfile := filepath.Join(dir, "global.o")
	execfile := filepath.Join(dir, "a.out")

	err = os.WriteFile(sourcefile, src, 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	_, err = exec.Command("gcc", "-c", sourcefile, "-o", objectfile).CombinedOutput()
	if err != nil {
		t.Errorf("gcc -c failed: %v", err)
	}

	_, err = exec.Command("gcc", "-fsanitize=address", objectfile, "-o", execfile).CombinedOutput()
	if err != nil {
		t.Errorf("gcc -o failed: %v", err)
	}

	out, err := exec.Command(execfile).CombinedOutput()

	if err != nil {
		t.Errorf("the output: %s\nerr: %v\n", out, err)
	}

}
