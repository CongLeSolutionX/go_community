package cgotest

/*
#include <inttypes.h>

// packed struct to cause misaligned struct size
typedef struct {
	uint64_t u;
	uint8_t v;
} __attribute__((__packed__)) misaligned_size;

typedef struct {
	uint64_t a;
	misaligned_size m;
	uint32_t b;
	misaligned_size m_arr[8];
	uint32_t c;
} st_misaligned;

misaligned_size arr_misaligned_size[3];

uint64_t get_a(st_misaligned *sm)
{
	return sm->a;
}

uint32_t get_b(st_misaligned *sm)
{
	return sm->b;
}

uint32_t get_c(st_misaligned *sm)
{
	return sm->c;
}
*/
import "C"
import (
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func test46675(t *testing.T) {
	if C.sizeof_misaligned_size == unsafe.Sizeof(C.misaligned_size{}) {
		t.Skip("C compiler did not pack struct")
	}

	t.Run("type", func(t *testing.T) {
		misalignedSizeName := reflect.TypeOf(C.misaligned_size{}).Name()

		// Test type of st_misaligned
		typ := reflect.TypeOf(C.st_misaligned{})
		if typ.Kind() != reflect.Struct {
			t.Fatalf("st_misaligned is of kind %s, expected struct", typ.Kind())
		}

		if _, ok := typ.FieldByName("a"); !ok {
			t.Errorf("st_misaligned.a should be visible")
		}
		if _, ok := typ.FieldByName("b"); !ok {
			t.Errorf("st_misaligned.b should be visible")
		}
		if _, ok := typ.FieldByName("c"); !ok {
			t.Errorf("st_misaligned.c should be visible")
		}

		if f, ok := typ.FieldByName("m"); ok && f.Type.Kind() == reflect.Struct && f.Type.Name() == misalignedSizeName {
			t.Errorf("st_misaligned.m should not expose misaligned_size")
		}

		if f, ok := typ.FieldByName("m_arr"); ok && f.Type.Kind() == reflect.Array && f.Type.Elem().Kind() == reflect.Struct && f.Type.Elem().Name() == misalignedSizeName {
			t.Errorf("st_misaligned.m_arr should not expose an array of misaligned_size")
		}

		// Test type of arr_misaligned_size
		// arr_misaligned_size should be opaque or an array of opaque elements (i.e. elements should be byte arrays)
		// In particular, arr_misaligned_size should not be an array of misaligned_size.
		typ = reflect.TypeOf(C.arr_misaligned_size)
		if typ.Kind() == reflect.Array && typ.Elem().Kind() == reflect.Struct && typ.Elem().Name() == misalignedSizeName {
			t.Errorf("arr_misaligned_size should not expose array of misaligned_size")
		}
	})

	t.Run("value", func(t *testing.T) {
		// Test write and read of st_misaligned
		sm := C.st_misaligned{
			a: math.MaxUint64,
			b: math.MaxUint32,
			c: math.MaxUint32 - 1,
		}

		gotA := C.get_a(&sm)
		if gotA != sm.a {
			t.Errorf("st_misaligned.a: got %d, expected %d", gotA, sm.a)
		}

		gotB := C.get_b(&sm)
		if gotB != sm.b {
			t.Errorf("st_misaligned.b: got %d, expected %d", gotB, sm.b)
		}

		gotC := C.get_c(&sm)
		if gotC != sm.c {
			t.Errorf("st_misaligned.c: got %d, expected %d", gotC, sm.c)
		}
	})

	t.Run("size", func(t *testing.T) {
		testcases := []struct {
			name   string
			cSize  uintptr
			goSize uintptr
		}{
			{"st_misaligned", C.sizeof_st_misaligned, unsafe.Sizeof(C.st_misaligned{})},
			{"arr_misaligned_size", C.sizeof_arr_misaligned_size, unsafe.Sizeof(C.arr_misaligned_size)},
			{"element of arr_misaligned_size", C.sizeof_misaligned_size, unsafe.Sizeof(C.arr_misaligned_size[0])},
		}

		for _, tc := range testcases {
			if tc.cSize != tc.goSize {
				t.Errorf("%s: C size=%d, Go size=%d", tc.name, tc.cSize, tc.goSize)
			}
		}
	})

}
