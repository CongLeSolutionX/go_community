package js

import (
	"unsafe"
)

var int8Array = Global().Get("Int8Array")
var int16Array = Global().Get("Int16Array")
var int32Array = Global().Get("Int32Array")
var uint8Array = Global().Get("Uint8Array")
var uint16Array = Global().Get("Uint16Array")
var uint32Array = Global().Get("Uint32Array")
var float32Array = Global().Get("Float32Array")
var float64Array = Global().Get("Float64Array")

type TypedArray struct {
	Value
}

// Close frees up resources allocated for the typed array.
// The typed array and its buffer must not be accessed after calling Close.
func (a TypedArray) Close() {
	delete(openTypedArrays, a)
}

var openTypedArrays = make(map[TypedArray]struct{})

// TypedArrayOf returns a JavaScript typed array backed by the slice's underlying array.
// It can be passed to functions of this package that accept interface{}, for example Value.Set and Value.Call.
//
// Supported are []int8, []int16, []int32, []uint8, []uint16, []uint32, []float32 and []float64.
// Passing an unsupported value causes a panic.
//
// TypedArray.Close must be called to free up resources when the typed array will not be used any more.
func TypedArrayOf(slice interface{}) TypedArray {
	a := TypedArray{typedArrayOf(slice)}
	openTypedArrays[a] = struct{}{}
	return a
}

func typedArrayOf(slice interface{}) Value {
	switch slice := slice.(type) {
	case []int8:
		if len(slice) == 0 {
			return int8Array.New(memory.Get("buffer"), 0, 0)
		}
		return int8Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []int16:
		if len(slice) == 0 {
			return int16Array.New(memory.Get("buffer"), 0, 0)
		}
		return int16Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []int32:
		if len(slice) == 0 {
			return int32Array.New(memory.Get("buffer"), 0, 0)
		}
		return int32Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []uint8:
		if len(slice) == 0 {
			return uint8Array.New(memory.Get("buffer"), 0, 0)
		}
		return uint8Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []uint16:
		if len(slice) == 0 {
			return uint16Array.New(memory.Get("buffer"), 0, 0)
		}
		return uint16Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []uint32:
		if len(slice) == 0 {
			return uint32Array.New(memory.Get("buffer"), 0, 0)
		}
		return uint32Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []float32:
		if len(slice) == 0 {
			return float32Array.New(memory.Get("buffer"), 0, 0)
		}
		return float32Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	case []float64:
		if len(slice) == 0 {
			return float64Array.New(memory.Get("buffer"), 0, 0)
		}
		return float64Array.New(memory.Get("buffer"), unsafe.Pointer(&slice[0]), len(slice))
	default:
		panic("TypedArrayOf: not a supported slice")
	}
}
