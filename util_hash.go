package hashmap

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

import (
	"fmt"
	"math/bits"
	"reflect"
	"unsafe"

	"github.com/cespare/xxhash"
)

const (
	prime1 uint64 = 11400714785074694791
	prime2 uint64 = 14029467366897019727
	prime3 uint64 = 1609587929392839161
	prime4 uint64 = 9650029242287828579
	prime5 uint64 = 2870177450012600261
)

// Specialized xxhash hash functions, optimized for the bit size of the key where available,
// for all supported types beside string.

// setDefaultHasher sets the default hasher depending on the key type.
// Inlines hashing as anonymous functions for performance improvements, other options like
// returning an anonymous functions from another function turned out to not be as performant.
//
//nolint:funlen, maintidx
func (m *HashMap[Key, Value]) setDefaultHasher() {
	var key Key
	switch any(key).(type) {
	case string:
		m.hasher = func(key Key) uintptr {
			sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
			bh := reflect.SliceHeader{
				Data: sh.Data,
				Len:  sh.Len,
				Cap:  sh.Len, // cap needs to be set, otherwise xxhash fails on ARM Macs
			}
			buf := *(*[]byte)(unsafe.Pointer(&bh))
			return uintptr(xxhash.Sum64(buf))
		}

	case int, uint, uintptr:
		switch intSizeBytes {
		case 2:
			m.hasher = func(key Key) uintptr {
				bh := reflect.SliceHeader{
					Data: uintptr(unsafe.Pointer(&key)),
					Len:  2,
				}
				b := *(*[]byte)(unsafe.Pointer(&bh))

				var h = prime5 + 2

				h ^= uint64(b[0]) * prime5
				h = bits.RotateLeft64(h, 11) * prime1
				h ^= uint64(b[1]) * prime5
				h = bits.RotateLeft64(h, 11) * prime1

				h ^= h >> 33
				h *= prime2
				h ^= h >> 29
				h *= prime3
				h ^= h >> 32

				return uintptr(h)
			}

		case 4:
			m.hasher = func(key Key) uintptr {
				bh := reflect.SliceHeader{
					Data: uintptr(unsafe.Pointer(&key)),
					Len:  4,
				}
				b := *(*[]byte)(unsafe.Pointer(&bh))

				var h = prime5 + 4
				h ^= (uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24) * prime1
				h = bits.RotateLeft64(h, 23)*prime2 + prime3

				h ^= h >> 33
				h *= prime2
				h ^= h >> 29
				h *= prime3
				h ^= h >> 32

				return uintptr(h)
			}

		case 8:
			m.hasher = func(key Key) uintptr {
				bh := reflect.SliceHeader{
					Data: uintptr(unsafe.Pointer(&key)),
					Len:  8,
				}
				b := *(*[]byte)(unsafe.Pointer(&bh))

				var h = prime5 + 8

				val := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
					uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56

				// inline round()
				k1 := val * prime2
				k1 = bits.RotateLeft64(k1, 31)
				k1 *= prime1

				h ^= k1
				h = bits.RotateLeft64(h, 27)*prime1 + prime4

				h ^= h >> 33
				h *= prime2
				h ^= h >> 29
				h *= prime3
				h ^= h >> 32

				return uintptr(h)
			}

		default:
			panic(fmt.Errorf("unsupported integer byte size %d", intSizeBytes))
		}

	case int8, uint8:
		m.hasher = func(key Key) uintptr {
			bh := reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&key)),
				Len:  1,
			}
			b := *(*[]byte)(unsafe.Pointer(&bh))

			var h = prime5 + 1
			h ^= uint64(b[0]) * prime5
			h = bits.RotateLeft64(h, 11) * prime1

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case int16, uint16:
		m.hasher = func(key Key) uintptr {
			bh := reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&key)),
				Len:  2,
			}
			b := *(*[]byte)(unsafe.Pointer(&bh))

			var h = prime5 + 2

			h ^= uint64(b[0]) * prime5
			h = bits.RotateLeft64(h, 11) * prime1
			h ^= uint64(b[1]) * prime5
			h = bits.RotateLeft64(h, 11) * prime1

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case int32, uint32, float32:
		m.hasher = func(key Key) uintptr {
			bh := reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&key)),
				Len:  4,
			}
			b := *(*[]byte)(unsafe.Pointer(&bh))

			var h = prime5 + 4
			h ^= (uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24) * prime1
			h = bits.RotateLeft64(h, 23)*prime2 + prime3

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case int64, uint64, float64, complex64:
		m.hasher = func(key Key) uintptr {
			bh := reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&key)),
				Len:  8,
			}
			b := *(*[]byte)(unsafe.Pointer(&bh))

			var h = prime5 + 8

			val := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
				uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56

			// inline round()
			k1 := val * prime2
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= prime1

			h ^= k1
			h = bits.RotateLeft64(h, 27)*prime1 + prime4

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case complex128:
		m.hasher = func(key Key) uintptr {
			bh := reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&key)),
				Len:  16,
			}
			b := *(*[]byte)(unsafe.Pointer(&bh))

			var h = prime5 + 16

			val := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
				uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56

			// inline round()
			k1 := val * prime2
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= prime1

			h ^= k1
			h = bits.RotateLeft64(h, 27)*prime1 + prime4

			val = uint64(b[8]) | uint64(b[9])<<8 | uint64(b[10])<<16 | uint64(b[11])<<24 |
				uint64(b[12])<<32 | uint64(b[13])<<40 | uint64(b[14])<<48 | uint64(b[15])<<56

			// inline round()
			k1 = val * prime2
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= prime1

			h ^= k1
			h = bits.RotateLeft64(h, 27)*prime1 + prime4

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
