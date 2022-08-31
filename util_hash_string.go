package hashmap

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"reflect"
	"unsafe"
)

/*
Copyright (c) 2016 Caleb Spare

MIT License

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

// setDefaultHasher sets the default hasher for the string map.
//
//nolint:funlen
func (m *MapString[Key, Value]) setDefaultHasher() {
	var key Key
	kind := reflect.ValueOf(&key).Elem().Type().Kind()

	switch kind {
	case reflect.String:
		m.hasher = func(key Key) uintptr {
			sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
			bh := reflect.SliceHeader{
				Data: sh.Data,
				Len:  sh.Len,
				Cap:  sh.Len, // cap needs to be set, otherwise xxhash fails on ARM Macs
			}

			b := *(*[]byte)(unsafe.Pointer(&bh))
			var h uint64

			if sh.Len >= 32 {
				v1 := prime1v + prime2
				v2 := prime2
				v3 := uint64(0)
				v4 := -prime1v
				for len(b) >= 32 {
					v1 = round(v1, binary.LittleEndian.Uint64(b[0:8:len(b)]))
					v2 = round(v2, binary.LittleEndian.Uint64(b[8:16:len(b)]))
					v3 = round(v3, binary.LittleEndian.Uint64(b[16:24:len(b)]))
					v4 = round(v4, binary.LittleEndian.Uint64(b[24:32:len(b)]))
					b = b[32:len(b):len(b)]
				}
				h = rol1(v1) + rol7(v2) + rol12(v3) + rol18(v4)
				h = mergeRound(h, v1)
				h = mergeRound(h, v2)
				h = mergeRound(h, v3)
				h = mergeRound(h, v4)
			} else {
				h = prime5
			}

			h += uint64(sh.Len)

			i, end := 0, len(b)
			for ; i+8 <= end; i += 8 {
				k1 := round(0, binary.LittleEndian.Uint64(b[i:i+8:len(b)]))
				h ^= k1
				h = rol27(h)*prime1 + prime4
			}
			if i+4 <= end {
				h ^= uint64(binary.LittleEndian.Uint32(b[i:i+4:len(b)])) * prime1
				h = rol23(h)*prime2 + prime3
				i += 4
			}
			for ; i < end; i++ {
				h ^= uint64(b[i]) * prime5
				h = rol11(h) * prime1
			}

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	default:
		panic(fmt.Errorf("unsupported key type %T", key))
	}
}

func round(acc, input uint64) uint64 {
	acc += input * prime2
	acc = rol31(acc)
	acc *= prime1
	return acc
}

func mergeRound(acc, val uint64) uint64 {
	val = round(0, val)
	acc ^= val
	acc = acc*prime1 + prime4
	return acc
}

func rol1(x uint64) uint64  { return bits.RotateLeft64(x, 1) }
func rol7(x uint64) uint64  { return bits.RotateLeft64(x, 7) }
func rol11(x uint64) uint64 { return bits.RotateLeft64(x, 11) }
func rol12(x uint64) uint64 { return bits.RotateLeft64(x, 12) }
func rol18(x uint64) uint64 { return bits.RotateLeft64(x, 18) }
func rol23(x uint64) uint64 { return bits.RotateLeft64(x, 23) }
func rol27(x uint64) uint64 { return bits.RotateLeft64(x, 27) }
func rol31(x uint64) uint64 { return bits.RotateLeft64(x, 31) }
