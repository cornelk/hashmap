package hashmap

import (
	"reflect"
	"sync/atomic"
	"unsafe"

	"github.com/dchest/siphash"
)

// Get retrieves an element from the map under given hash key.
// Using interface{} adds a performance penalty.
// Please consider using GetUintKey or GetStringKey instead.
func (m *HashMap) Get(key interface{}) (value unsafe.Pointer, ok bool) {
	hashedKey := getKeyHash(key)

	// inline HashMap.getSliceItemForKey()
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey && entry.key == key {
			// inline ListElement.Value()
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetUintKey retrieves an element from the map under given integer key.
func (m *HashMap) GetUintKey(key uint64) (value unsafe.Pointer, ok bool) {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}

	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  8,
		Cap:  8,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	hashedKey := siphash.Hash(sipHashKey1, sipHashKey2, buf)

	// inline HashMap.getSliceItemForKey()
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey && entry.key == key {
			// inline ListElement.Value()
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetStringKey retrieves an element from the map under given string key.
func (m *HashMap) GetStringKey(key string) (value unsafe.Pointer, ok bool) {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}

	sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	hashedKey := siphash.Hash(sipHashKey1, sipHashKey2, buf)

	// inline HashMap.getSliceItemForKey()
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey && entry.key == key {
			// inline ListElement.Value()
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetHashedKey retrieves an element from the map under given hashed key.
func (m *HashMap) GetHashedKey(hashedKey uint64) (value unsafe.Pointer, ok bool) {
	// inline HashMap.getSliceItemForKey()
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey {
			// inline ListElement.Value()
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint64(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}
