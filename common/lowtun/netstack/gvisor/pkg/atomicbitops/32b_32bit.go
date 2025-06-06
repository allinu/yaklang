// Copyright 2022 The gVisor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build arm || mips || mipsle || 386
// +build arm mips mipsle 386

package atomicbitops

import (
	"sync/atomic"

	"github.com/yaklang/yaklang/common/lowtun/netstack/gvisor/pkg/sync"
)

// Note that this file is *identical* to 32b_64bit.go, as go_stateify gets
// confused about build tags if these are not separated.

// LINT.IfChange

// Int32 is an atomic int32.
//
// The default value is zero.
//
// Don't add fields to this struct. It is important that it remain the same
// size as its builtin analogue.
//
// +stateify savable
type Int32 struct {
	_     sync.NoCopy
	value int32
}

// FromInt32 returns an Int32 initialized to value v.
//
//go:nosplit
func FromInt32(v int32) Int32 {
	return Int32{value: v}
}

// Load is analogous to atomic.LoadInt32.
//
//go:nosplit
func (i *Int32) Load() int32 {
	return atomic.LoadInt32(&i.value)
}

// RacyLoad is analogous to reading an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (i *Int32) RacyLoad() int32 {
	return i.value
}

// Store is analogous to atomic.StoreInt32.
//
//go:nosplit
func (i *Int32) Store(v int32) {
	atomic.StoreInt32(&i.value, v)
}

// RacyStore is analogous to setting an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (i *Int32) RacyStore(v int32) {
	i.value = v
}

// Add is analogous to atomic.AddInt32.
//
//go:nosplit
func (i *Int32) Add(v int32) int32 {
	return atomic.AddInt32(&i.value, v)
}

// RacyAdd is analogous to adding to an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (i *Int32) RacyAdd(v int32) int32 {
	i.value += v
	return i.value
}

// Swap is analogous to atomic.SwapInt32.
//
//go:nosplit
func (i *Int32) Swap(v int32) int32 {
	return atomic.SwapInt32(&i.value, v)
}

// CompareAndSwap is analogous to atomic.CompareAndSwapInt32.
//
//go:nosplit
func (i *Int32) CompareAndSwap(oldVal, newVal int32) bool {
	return atomic.CompareAndSwapInt32(&i.value, oldVal, newVal)
}

//go:nosplit
func (i *Int32) ptr() *int32 {
	return &i.value
}

// Uint32 is an atomic uint32.
//
// Don't add fields to this struct. It is important that it remain the same
// size as its builtin analogue.
//
// See aligned_unsafe.go in this directory for justification.
//
// +stateify savable
type Uint32 struct {
	_     sync.NoCopy
	value uint32
}

// FromUint32 returns an Uint32 initialized to value v.
//
//go:nosplit
func FromUint32(v uint32) Uint32 {
	return Uint32{value: v}
}

// Load is analogous to atomic.LoadUint32.
//
//go:nosplit
func (u *Uint32) Load() uint32 {
	return atomic.LoadUint32(&u.value)
}

// RacyLoad is analogous to reading an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (u *Uint32) RacyLoad() uint32 {
	return u.value
}

// Store is analogous to atomic.StoreUint32.
//
//go:nosplit
func (u *Uint32) Store(v uint32) {
	atomic.StoreUint32(&u.value, v)
}

// RacyStore is analogous to setting an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (u *Uint32) RacyStore(v uint32) {
	u.value = v
}

// Add is analogous to atomic.AddUint32.
//
//go:nosplit
func (u *Uint32) Add(v uint32) uint32 {
	return atomic.AddUint32(&u.value, v)
}

// RacyAdd is analogous to adding to an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (u *Uint32) RacyAdd(v uint32) uint32 {
	u.value += v
	return u.value
}

// Swap is analogous to atomic.SwapUint32.
//
//go:nosplit
func (u *Uint32) Swap(v uint32) uint32 {
	return atomic.SwapUint32(&u.value, v)
}

// CompareAndSwap is analogous to atomic.CompareAndSwapUint32.
//
//go:nosplit
func (u *Uint32) CompareAndSwap(oldVal, newVal uint32) bool {
	return atomic.CompareAndSwapUint32(&u.value, oldVal, newVal)
}

//go:nosplit
func (u *Uint32) ptr() *uint32 {
	return &u.value
}

// Bool is an atomic Boolean.
//
// It is implemented by a Uint32, with value 0 indicating false, and 1
// indicating true.
//
// +stateify savable
type Bool struct {
	Uint32
}

// b32 returns a uint32 0 or 1 representing b.
func b32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// FromBool returns a Bool initialized to value val.
//
//go:nosplit
func FromBool(val bool) Bool {
	return Bool{
		Uint32: FromUint32(b32(val)),
	}
}

// Load is analogous to atomic.LoadBool, if such a thing existed.
//
//go:nosplit
func (b *Bool) Load() bool {
	return b.Uint32.Load() != 0
}

// RacyLoad is analogous to reading an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (b *Bool) RacyLoad() bool {
	return b.Uint32.RacyLoad() != 0
}

// Store is analogous to atomic.StoreBool, if such a thing existed.
//
//go:nosplit
func (b *Bool) Store(val bool) {
	b.Uint32.Store(b32(val))
}

// RacyStore is analogous to setting an atomic value without using
// synchronization.
//
// It may be helpful to document why a racy operation is permitted.
//
//go:nosplit
func (b *Bool) RacyStore(val bool) {
	b.Uint32.RacyStore(b32(val))
}

// Swap is analogous to atomic.SwapBool, if such a thing existed.
//
//go:nosplit
func (b *Bool) Swap(val bool) bool {
	return b.Uint32.Swap(b32(val)) != 0
}

// CompareAndSwap is analogous to atomic.CompareAndSwapBool, if such a thing
// existed.
//
//go:nosplit
func (b *Bool) CompareAndSwap(oldVal, newVal bool) bool {
	return b.Uint32.CompareAndSwap(b32(oldVal), b32(newVal))
}

// LINT.ThenChange(32b_64bit.go)
