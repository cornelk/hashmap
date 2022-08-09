package hashmap

import (
	"reflect"
	"unsafe"

	"github.com/cespare/xxhash"
)

func (m *HashMap[Key, Value]) stringHasher(key Key) uintptr {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(xxhash.Sum64(buf))
}

func (m *HashMap[Key, Value]) uintptrHasher(key Key) uintptr {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  intSizeBytes,
		Cap:  intSizeBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(xxhash.Sum64(buf))
}

func (m *HashMap[Key, Value]) byteHasher(key Key) uintptr {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  1,
		Cap:  1,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(xxhash.Sum64(buf))
}

func (m *HashMap[Key, Value]) wordHasher(key Key) uintptr {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  2,
		Cap:  2,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(xxhash.Sum64(buf))
}

func (m *HashMap[Key, Value]) dwordHasher(key Key) uintptr {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  4,
		Cap:  4,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(xxhash.Sum64(buf))
}

func (m *HashMap[Key, Value]) qwordHasher(key Key) uintptr {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  8,
		Cap:  8,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(xxhash.Sum64(buf))
}

// used in unit test to test collision support.
func (m *HashMap[Key, Value]) staticHasher(key Key) uintptr {
	return 4 // chosen by fair dice roll. guaranteed to be random.
}
