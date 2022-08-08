//go:build customhash

package hashmap

import (
	"fmt"
	"reflect"
	"unsafe"
)

type customHashFunc func(b []byte) uint64

var customHash customHashFunc

// getKeyHash returns a hash for the key. Only string and number types are supported.
func getKeyHash(key interface{}) uintptr {
	switch x := key.(type) {
	case string:
		return getStringHash(x)
	case int:
		return getUintptrHash(uintptr(x))
	case int8:
		return getUintptrHash(uintptr(x))
	case int16:
		return getUintptrHash(uintptr(x))
	case int32:
		return getUintptrHash(uintptr(x))
	case int64:
		return getUintptrHash(uintptr(x))
	case uint:
		return getUintptrHash(uintptr(x))
	case uint8:
		return getUintptrHash(uintptr(x))
	case uint16:
		return getUintptrHash(uintptr(x))
	case uint32:
		return getUintptrHash(uintptr(x))
	case uint64:
		return getUintptrHash(uintptr(x))
	case uintptr:
		return getUintptrHash(x)
	}
	panic(fmt.Errorf("unsupported key type %T", key))
}

func getStringHash(s string) uintptr {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(customHash(buf))
}

func getUintptrHash(num uintptr) uintptr {
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&num)),
		Len:  intSizeBytes,
		Cap:  intSizeBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	return uintptr(customHash(buf))
}
