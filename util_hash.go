package hashmap

import (
	"fmt"
	"math/bits"
	"reflect"
)

const (
	prime1 uint64 = 11400714785074694791
	prime2 uint64 = 14029467366897019727
	prime3 uint64 = 1609587929392839161
	prime4 uint64 = 9650029242287828579
	prime5 uint64 = 2870177450012600261
)

var prime1v = prime1

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

// Specialized xxhash hash functions, optimized for the bit size of the key where available,
// for all supported types beside string.

// setDefaultHasher sets the default hasher depending on the key type.
// Inlines hashing as anonymous functions for performance improvements, other options like
// returning an anonymous functions from another function turned out to not be as performant.
//
//nolint:funlen, maintidx
func (m *Map[Key, Value]) setDefaultHasher() {
	var key Key
	kind := reflect.ValueOf(&key).Elem().Type().Kind()

	switch kind {
	case reflect.Int, reflect.Uint, reflect.Uintptr:
		switch intSizeBytes {
		case 2:
			m.hasher = func(key Key) uintptr {
				h := prime5 + 2
				h ^= (uint64(key) & 0xff) * prime5
				h = bits.RotateLeft64(h, 11) * prime1
				h ^= ((uint64(key) >> 8) & 0xff) * prime5
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
				h := prime5 + 4
				h ^= uint64(key) * prime1
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
				k1 := uint64(key) * prime2
				k1 = bits.RotateLeft64(k1, 31)
				k1 *= prime1
				h := (prime5 + 8) ^ k1
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

	case reflect.Int8, reflect.Uint8:
		m.hasher = func(key Key) uintptr {
			h := prime5 + 1
			h ^= uint64(key) * prime5
			h = bits.RotateLeft64(h, 11) * prime1

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case reflect.Int16, reflect.Uint16:
		m.hasher = func(key Key) uintptr {
			h := prime5 + 2
			h ^= (uint64(key) & 0xff) * prime5
			h = bits.RotateLeft64(h, 11) * prime1
			h ^= ((uint64(key) >> 8) & 0xff) * prime5
			h = bits.RotateLeft64(h, 11) * prime1

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case reflect.Int32, reflect.Uint32, reflect.Float32:
		m.hasher = func(key Key) uintptr {
			h := prime5 + 4
			h ^= uint64(key) * prime1
			h = bits.RotateLeft64(h, 23)*prime2 + prime3

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	case reflect.Int64, reflect.Uint64, reflect.Float64:
		m.hasher = func(key Key) uintptr {
			k1 := uint64(key) * prime2
			k1 = bits.RotateLeft64(k1, 31)
			k1 *= prime1
			h := (prime5 + 8) ^ k1
			h = bits.RotateLeft64(h, 27)*prime1 + prime4

			h ^= h >> 33
			h *= prime2
			h ^= h >> 29
			h *= prime3
			h ^= h >> 32

			return uintptr(h)
		}

	default:
		panic(fmt.Errorf("unsupported key type %T of kind %v", key, kind))
	}
}
