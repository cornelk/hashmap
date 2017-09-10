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
	h := getKeyHash(key)

	// inline HashMap.indexElement()
	data := (*hashMapData)(atomic.LoadPointer(&m.datamap))
	if data == nil {
		return nil, false
	}
	index := h >> data.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + uintptr(index*intSizeBytes)))
	element := (*ListElement)(atomic.LoadPointer(ptr))

	for element != nil {
		if element.keyHash == h && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > h {
			return nil, false
		}

		element = (*ListElement)(atomic.LoadPointer(&element.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetUintKey retrieves an element from the map under given integer key.
func (m *HashMap) GetUintKey(key uintptr) (value unsafe.Pointer, ok bool) {
	data := (*hashMapData)(atomic.LoadPointer(&m.datamap))
	if data == nil {
		return nil, false
	}

	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  intSizeBytes,
		Cap:  intSizeBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	h := uintptr(siphash.Hash(sipHashKey1, sipHashKey2, buf))

	// inline HashMap.indexElement()
	index := h >> data.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + uintptr(index*intSizeBytes)))
	element := (*ListElement)(atomic.LoadPointer(ptr))

	for element != nil {
		if element.keyHash == h && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > h {
			return nil, false
		}

		element = (*ListElement)(atomic.LoadPointer(&element.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetStringKey retrieves an element from the map under given string key.
func (m *HashMap) GetStringKey(key string) (value unsafe.Pointer, ok bool) {
	data := (*hashMapData)(atomic.LoadPointer(&m.datamap))
	if data == nil {
		return nil, false
	}

	sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	h := uintptr(siphash.Hash(sipHashKey1, sipHashKey2, buf))

	// inline HashMap.indexElement()
	index := h >> data.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + uintptr(index*intSizeBytes)))
	element := (*ListElement)(atomic.LoadPointer(ptr))

	for element != nil {
		if element.keyHash == h && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > h {
			return nil, false
		}

		element = (*ListElement)(atomic.LoadPointer(&element.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetHashedKey retrieves an element from the map under given hashed key.
func (m *HashMap) GetHashedKey(hashedKey uintptr) (value unsafe.Pointer, ok bool) {
	// inline HashMap.indexElement()
	data := (*hashMapData)(atomic.LoadPointer(&m.datamap))
	if data == nil {
		return nil, false
	}
	index := hashedKey >> data.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + uintptr(index*intSizeBytes)))
	element := (*ListElement)(atomic.LoadPointer(ptr))

	for element != nil {
		if element.keyHash == hashedKey {
			return element.Value(), true
		}

		if element.keyHash > hashedKey {
			return nil, false
		}

		element = (*ListElement)(atomic.LoadPointer(&element.nextElement)) // inline ListElement.Next()
	}
	return nil, false
}

// GetOrInsert returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *HashMap) GetOrInsert(key interface{}, value unsafe.Pointer) (actual unsafe.Pointer, loaded bool) {
	h := getKeyHash(key)
	var newelement *ListElement

	for {
		// inline HashMap.indexElement()
		data := (*hashMapData)(atomic.LoadPointer(&m.datamap))
		if data == nil {
			m.allocate(DefaultSize)
			data = (*hashMapData)(atomic.LoadPointer(&m.datamap))
		}
		index := h >> data.keyshifts
		ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.data) + uintptr(index*intSizeBytes)))
		element := (*ListElement)(atomic.LoadPointer(ptr))

		for element != nil {
			if element.keyHash == h && element.key == key {
				actual = element.Value()
				return actual, true
			}

			if element.keyHash > h {
				break
			}

			element = (*ListElement)(atomic.LoadPointer(&element.nextElement)) // inline ListElement.Next()
		}

		if newelement == nil { // allocate only once
			newelement = &ListElement{
				key:     key,
				keyHash: h,
				value:   value,
			}
		}

		if m.insertListElement(newelement, false) {
			return value, false
		}
	}
}
