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

//go:build go1.13
// +build go1.13

package jsutil

import (
	"fmt"
	"syscall/js"
)

func Uint8ArrayToSlice(value js.Value) []byte {
	s := make([]byte, value.Get("byteLength").Int())
	js.CopyBytesToGo(s, value)
	return s
}

func ArrayBufferToSlice(value js.Value) []byte {
	return Uint8ArrayToSlice(js.Global().Get("Uint8Array").New(value))
}

var temporaryBuffer = js.Global().Get("ArrayBuffer").New(16)

func getTemporaryUint8Array(size int) js.Value {
	if l := temporaryBuffer.Get("byteLength").Int(); l < size {
		for l < size {
			l *= 2
		}
		temporaryBuffer = js.Global().Get("ArrayBuffer").New(l)
	}
	return js.Global().Get("Uint8Array").New(temporaryBuffer, 0, size)
}

func SliceToTypedArray(s interface{}) js.Value {
	switch s := s.(type) {
	case []int8:
		a := getTemporaryUint8Array(len(s))
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Int8Array").New(buf, a.Get("byteOffset"), a.Get("byteLength"))
	case []int16:
		a := getTemporaryUint8Array(len(s) * 2)
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Int16Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/2)
	case []int32:
		a := getTemporaryUint8Array(len(s) * 4)
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Int32Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4)
	case []uint8:
		a := getTemporaryUint8Array(len(s))
		js.CopyBytesToJS(a, s)
		return a
	case []uint16:
		a := getTemporaryUint8Array(len(s) * 2)
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Uint16Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/2)
	case []uint32:
		a := getTemporaryUint8Array(len(s) * 4)
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Uint32Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4)
	case []float32:
		a := getTemporaryUint8Array(len(s) * 4)
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Float32Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/4)
	case []float64:
		a := getTemporaryUint8Array(len(s) * 8)
		js.CopyBytesToJS(a, sliceToByteSlice(s))
		buf := a.Get("buffer")
		return js.Global().Get("Float64Array").New(buf, a.Get("byteOffset"), a.Get("byteLength").Int()/8)
	default:
		panic(fmt.Sprintf("jsutil: unexpected value at SliceToTypedArray: %T", s))
	}
}
