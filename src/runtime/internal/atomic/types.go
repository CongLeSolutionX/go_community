// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package atomic

import "unsafe"

// Int32 is an atomically accessed int32 value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Int32 struct {
	noCopy noCopy
	value  int32
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into i.
func (i *Int32) Load() int32 {
	return Loadint32(&i.value)
}

// Store updates the value atomically.
func (i *Int32) Store(value int32) {
	Storeint32(&i.value, value)
}

// Cas atomically (with respect to other methods)
// compares i's value with old, and if they're equal,
// swaps i's value with new.
//
// Returns true if the operation succeeded.
func (i *Int32) Cas(old, new int32) bool {
	return Casint32(&i.value, old, new)
}

// Swap replaces i's value with new, returning
// i's value before the replacement.
func (i *Int32) Swap(new int32) int32 {
	return Xchgint32(&i.value, new)
}

// Add adds delta to i atomically, returning
// the new updated value.
//
// This operation wraps around in the usual
// two's-complement way.
func (i *Int32) Add(delta int32) int32 {
	return Xaddint32(&i.value, delta)
}

// Int64 is an atomically accessed int64 value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Int64 struct {
	noCopy noCopy
	value  int64
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into i.
func (i *Int64) Load() int64 {
	return Loadint64(&i.value)
}

// Store updates the value atomically.
func (i *Int64) Store(value int64) {
	Storeint64(&i.value, value)
}

// Cas atomically (with respect to other methods)
// compares i's value with old, and if they're equal,
// swaps i's value with new.
//
// Returns true if the operation succeeded.
func (i *Int64) Cas(old, new int64) bool {
	return Casint64(&i.value, old, new)
}

// Swap replaces i's value with new, returning
// i's value before the replacement.
func (i *Int64) Swap(new int64) int64 {
	return Xchgint64(&i.value, new)
}

// Add adds delta to i atomically, returning
// the new updated value.
//
// This operation wraps around in the usual
// two's-complement way.
func (i *Int64) Add(delta int64) int64 {
	return Xaddint64(&i.value, delta)
}

// Uint8 is an atomically accessed uint8 value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Uint8 struct {
	noCopy noCopy
	value  uint8
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into u.
func (u *Uint8) Load() uint8 {
	return Load8(&u.value)
}

// Store updates the value atomically.
func (u *Uint8) Store(value uint8) {
	Store8(&u.value, value)
}

// And takes value and performs a bit-wise
// "and" operation with the value of u, storing
// the result into u.
//
// The full process is performed atomically.
func (u *Uint8) And(value uint8) {
	And8(&u.value, value)
}

// Or takes value and performs a bit-wise
// "or" operation with the value of u, storing
// the result into u.
//
// The full process is performed atomically.
func (u *Uint8) Or(value uint8) {
	Or8(&u.value, value)
}

// Uint32 is an atomically accessed uint32 value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Uint32 struct {
	noCopy noCopy
	value  uint32
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into u.
func (u *Uint32) Load() uint32 {
	return Load(&u.value)
}

// LoadAcquire is a partially unsynchronized version
// of Load that relaxes ordering constraints. Other threads
// may observe method calls that precede this operation to
// occur after it, but no method call that occurs after it
// on this thread can be observed to occur before it.
//
// WARNING: Use sparingly and with great care.
func (u *Uint32) LoadAcquire() uint32 {
	return LoadAcq(&u.value)
}

// Store updates the value atomically.
func (u *Uint32) Store(value uint32) {
	Store(&u.value, value)
}

// StoreRelease is a partially unsynchronized version
// of Store that relaxes ordering constraints. Other threads
// may observe method calls that occur after this operation to
// precede it, but no method call that precedes it
// on this thread can be observed to occur after it.
//
// WARNING: Use sparingly and with great care.
func (u *Uint32) StoreRelease(value uint32) {
	StoreRel(&u.value, value)
}

// Cas atomically (with respect to other methods)
// compares u's value with old, and if they're equal,
// swaps u's value with new.
//
// Returns true if the operation succeeded.
func (u *Uint32) Cas(old, new uint32) bool {
	return Cas(&u.value, old, new)
}

// CasRelease is a partially unsynchronized version
// of Cas that relaxes ordering constraints. Other threads
// may observe method calls that occur after this operation to
// precede it, but no method call that precedes it
// on this thread can be observed to occur after it.
//
// Returns true if the operation succeeded.
//
// WARNING: Use sparingly and with great care.
func (u *Uint32) CasRelease(old, new uint32) bool {
	return CasRel(&u.value, old, new)
}

// Swap replaces u's value with new, returning
// u's value before the replacement.
func (u *Uint32) Swap(value uint32) uint32 {
	return Xchg(&u.value, value)
}

// And takes value and performs a bit-wise
// "and" operation with the value of u, storing
// the result into u.
//
// The full process is performed atomically.
func (u *Uint32) And(value uint32) {
	And(&u.value, value)
}

// Or takes value and performs a bit-wise
// "or" operation with the value of u, storing
// the result into u.
//
// The full process is performed atomically.
func (u *Uint32) Or(value uint32) {
	Or(&u.value, value)
}

// Add adds delta to u atomically, returning
// the new updated value.
//
// This operation wraps around in the usual
// two's-complement way.
func (u *Uint32) Add(delta int32) uint32 {
	return Xadd(&u.value, delta)
}

// Uint64 is an atomically accessed uint64 value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Uint64 struct {
	noCopy noCopy
	value  uint64
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into u.
func (u *Uint64) Load() uint64 {
	return Load64(&u.value)
}

// Store updates the value atomically.
func (u *Uint64) Store(value uint64) {
	Store64(&u.value, value)
}

// Cas atomically (with respect to other methods)
// compares u's value with old, and if they're equal,
// swaps u's value with new.
//
// Returns true if the operation succeeded.
func (u *Uint64) Cas(old, new uint64) bool {
	return Cas64(&u.value, old, new)
}

// Swap replaces u's value with new, returning
// u's value before the replacement.
func (u *Uint64) Swap(value uint64) uint64 {
	return Xchg64(&u.value, value)
}

// Add adds delta to u atomically, returning
// the new updated value.
//
// This operation wraps around in the usual
// two's-complement way.
func (u *Uint64) Add(delta int64) uint64 {
	return Xadd64(&u.value, delta)
}

// Uintptr is an atomically accessed uintptr value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Uintptr struct {
	noCopy noCopy
	value  uintptr
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into u.
func (u *Uintptr) Load() uintptr {
	return Loaduintptr(&u.value)
}

// LoadAcquire is a partially unsynchronized version
// of Load that relaxes ordering constraints. Other threads
// may observe method calls that precede this operation to
// occur after it, but no method call that occurs after it
// on this thread can be observed to occur before it.
//
// WARNING: Use sparingly and with great care.
func (u *Uintptr) LoadAcquire() uintptr {
	return LoadAcquintptr(&u.value)
}

// Store updates the value atomically.
func (u *Uintptr) Store(value uintptr) {
	Storeuintptr(&u.value, value)
}

// StoreRelease is a partially unsynchronized version
// of Store that relaxes ordering constraints. Other threads
// may observe method calls that occur after this operation to
// precede it, but no method call that precedes it
// on this thread can be observed to occur after it.
//
// WARNING: Use sparingly and with great care.
func (u *Uintptr) StoreRelease(value uintptr) {
	StoreReluintptr(&u.value, value)
}

// Cas atomically (with respect to other methods)
// compares u's value with old, and if they're equal,
// swaps u's value with new.
//
// Returns true if the operation succeeded.
func (u *Uintptr) Cas(old, new uintptr) bool {
	return Casuintptr(&u.value, old, new)
}

// Swap replaces u's value with new, returning
// u's value before the replacement.
func (u *Uintptr) Swap(value uintptr) uintptr {
	return Xchguintptr(&u.value, value)
}

// Add adds delta to u atomically, returning
// the new updated value.
//
// This operation wraps around in the usual
// two's-complement way.
func (u *Uintptr) Add(delta uintptr) uintptr {
	return Xadduintptr(&u.value, delta)
}

// Float64 is an atomically accessed float64 value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type Float64 struct {
	u Uint64
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into f.
func (f *Float64) Load() float64 {
	r := f.u.Load()
	return *(*float64)(unsafe.Pointer(&r))
}

// Store updates the value atomically.
func (f *Float64) Store(value float64) {
	f.u.Store(*(*uint64)(unsafe.Pointer(&value)))
}

// UnsafePointer is an atomically accessed unsafe.Pointer value.
//
// Operations on this type, unless otherwise noted, are
// sequentially consistent across threads. More specifically,
// operations that happen in a specific order on one thread,
// will always be observed to happen in exactly that order
// by another thread.
//
// Note that because of the atomicity guarantees, stores to values
// of this type never trigger a write barrier, and the relevant
// methods are suffixed with "NoWB" to indicate that explicitly.
// As a result, this type should be used carefully, and sparingly,
// mostly with values that do not live in the Go heap anyway.
//
// A value of this type should never be copied and should only be
// accessed via its methods in order to retain its guarantees.
type UnsafePointer struct {
	noCopy noCopy
	value  unsafe.Pointer
}

// Load accesses and returns the value atomically;
// the resulting value will always be a valid past
// value stored into u.
func (u *UnsafePointer) Load() unsafe.Pointer {
	return Loadp(unsafe.Pointer(&u.value))
}

// StoreNoWB updates the value atomically.
//
// WARNING: As the name implies this operation does *not*
// perform a write barrier on value, and so this operation may
// hide pointers from the GC. Use with care and sparingly.
// It is safe to use with values not found in the Go heap.
func (u *UnsafePointer) StoreNoWB(value unsafe.Pointer) {
	StorepNoWB(unsafe.Pointer(&u.value), value)
}

// CasNoWB atomically (with respect to other methods)
// compares u's value with old, and if they're equal,
// swaps u's value with new.
//
// Returns true if the operation succeeded.
//
// WARNING: As the name implies this operation does *not*
// perform a write barrier on value, and so this operation may
// hide pointers from the GC. Use with care and sparingly.
// It is safe to use with values not found in the Go heap.
func (u *UnsafePointer) CasNoWB(old, new unsafe.Pointer) bool {
	return Casp1(&u.value, old, new)
}

// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}
