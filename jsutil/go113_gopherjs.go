// Copyright 2019 The Ebiten Authors
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

// +build go1.13
// +build js
// +build !wasm

package jsutil

import (
	"syscall/js"
	"unsafe"

	gojs "github.com/gopherjs/gopherjs/js"
)

func Uint8ArrayToSlice(value js.Value) []byte {
	// Note that TypedArrayOf cannot work correcly on Wasm.
	// See https://github.com/golang/go/issues/31980

	s := make([]byte, value.Get("byteLength").Int())
	a := TypedArrayOf(s)
	a.Call("set", value)
	a.Release()
	return s
}

func ArrayBufferToSlice(value js.Value) []byte {
	return Uint8ArrayToSlice(js.Global().Get("Uint8Array").New(value))
}

func SliceToTypedArray(s interface{}) (js.Value, func()) {
	// Note that TypedArrayOf cannot work correcly on Wasm.
	// See https://github.com/golang/go/issues/31980

	a := TypedArrayOf(s)
	return a.Value, func() { a.Release() }
}

type Value struct {
	v *gojs.Object

	// inited represents whether Value is non-zero value. true represents the value is not 'undefined'.
	inited bool
}

func objectToValue(obj *gojs.Object) js.Value {
	if obj == gojs.Undefined {
		return *(*js.Value)(unsafe.Pointer(&Value{}))
	}
	return *(*js.Value)(unsafe.Pointer(&Value{obj, true}))
}

type TypedArray struct {
	js.Value
}

func (t *TypedArray) Release() {
	t.Value = js.Value{}
}

func TypedArrayOf(slice interface{}) TypedArray {
	switch slice := slice.(type) {
	case []int8, []int16, []int32, []uint8, []uint16, []uint32, []float32, []float64:
		return TypedArray{objectToValue(id.Invoke(slice))}
	default:
		panic("TypedArrayOf: not a supported slice")
	}
}

var (
	id *gojs.Object
)

func init() {
	id = gojs.Global.Call("eval", "(function(x) { return x; })")
}
