package hashmap

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/cespare/xxhash"
)

const (
	// intSizeBytes is the size in byte of an int or uint value.
	intSizeBytes = strconv.IntSize >> 3
)

// roundUpPower2 rounds a number to the next power of 2.
func roundUpPower2(i uint64) uint64 {
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}

// log2 computes the binary logarithm of x, rounded up to the next integer.
func log2(i uint64) uint64 {
	var n, p uint64
	for p = 1; p < i; p += p {
		n++
	}
	return n
}

// getKeyHash returns a 64 bit hash for the key
func getKeyHash(key interface{}) uint64 {
	var num uint64
	hash := xxhash.New()
	switch x := key.(type) {
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		sh := (*reflect.StringHeader)(unsafe.Pointer(&x))
		bh := reflect.SliceHeader{
			Data: sh.Data,
			Len:  sh.Len,
			Cap:  sh.Len,
		}
		buf := *(*[]byte)(unsafe.Pointer(&bh))
		hash.Write(buf)
		return hash.Sum64()
	case int:
		num = uint64(x)
	case int8:
		num = uint64(x)
	case int16:
		num = uint64(x)
	case int32:
		num = uint64(x)
	case int64:
		num = uint64(x)
	case uint:
		num = uint64(x)
	case uint8:
		num = uint64(x)
	case uint16:
		num = uint64(x)
	case uint32:
		num = uint64(x)
	case uint64:
		num = uint64(x)
	case uintptr:
		num = uint64(x)
	default:
		panic(fmt.Errorf("unsupported key type %T", key))
	}

	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&num)),
		Len:  8,
		Cap:  8,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	hash.Write(buf)
	return hash.Sum64()
}
