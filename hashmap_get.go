package hashmap

import (
	"bytes"
	"reflect"
	"unsafe"

	"github.com/dchest/siphash"
)

// Get retrieves an element from the map under given hash key.
// Using interface{} adds a performance penalty.
// Please consider using GetUintKey or GetStringKey instead.
func (m *HashMap) Get(key interface{}) (value interface{}, ok bool) {
	h := getKeyHash(key)
	data, element := m.indexElement(h)
	if data == nil {
		return nil, false
	}

	// inline HashMap.searchItem()
	for element != nil {
		if element.keyHash == h {
			switch key.(type) {
			case []byte:
				if bytes.Compare(element.key.([]byte), key.([]byte)) == 0 {
					return element.Value(), true
				}
			default:
				if element.key == key {
					return element.Value(), true
				}
			}
		}

		if element.keyHash > h {
			return nil, false
		}

		element = element.Next()
	}
	return nil, false
}

// GetUintKey retrieves an element from the map under given integer key.
func (m *HashMap) GetUintKey(key uintptr) (value interface{}, ok bool) {
	// inline getUintptrHash()
	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  intSizeBytes,
		Cap:  intSizeBytes,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	h := uintptr(siphash.Hash(sipHashKey1, sipHashKey2, buf))

	data, element := m.indexElement(h)
	if data == nil {
		return nil, false
	}

	// inline HashMap.searchItem()
	for element != nil {
		if element.keyHash == h && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > h {
			return nil, false
		}

		element = element.Next()
	}
	return nil, false
}

// GetStringKey retrieves an element from the map under given string key.
func (m *HashMap) GetStringKey(key string) (value interface{}, ok bool) {
	// inline getStringHash()
	sh := (*reflect.StringHeader)(unsafe.Pointer(&key))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
	h := uintptr(siphash.Hash(sipHashKey1, sipHashKey2, buf))

	data, element := m.indexElement(h)
	if data == nil {
		return nil, false
	}

	// inline HashMap.searchItem()
	for element != nil {
		if element.keyHash == h && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > h {
			return nil, false
		}

		element = element.Next()
	}
	return nil, false
}

// GetHashedKey retrieves an element from the map under given hashed key.
func (m *HashMap) GetHashedKey(hashedKey uintptr) (value interface{}, ok bool) {
	data, element := m.indexElement(hashedKey)
	if data == nil {
		return nil, false
	}

	// inline HashMap.searchItem()
	for element != nil {
		if element.keyHash == hashedKey {
			return element.Value(), true
		}

		if element.keyHash > hashedKey {
			return nil, false
		}

		element = element.Next()
	}
	return nil, false
}

// GetOrInsert returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *HashMap) GetOrInsert(key interface{}, value interface{}) (actual interface{}, loaded bool) {
	h := getKeyHash(key)
	var newelement *ListElement

	for {
		data, element := m.indexElement(h)
		if data == nil {
			m.allocate(DefaultSize)
			continue
		}

		for element != nil {
			if element.keyHash == h {
				switch key.(type) {
				case []byte:
					if bytes.Compare(element.key.([]byte), key.([]byte)) == 0 {
						return element.Value(), true
					}
				default:
					if element.key == key {
						actual = element.Value()
						return actual, true
					}
				}
			}

			if element.keyHash > h {
				break
			}

			element = element.Next()
		}

		if newelement == nil { // allocate only once
			newelement = &ListElement{
				key:     key,
				keyHash: h,
				value:   unsafe.Pointer(&value),
			}
		}

		if m.insertListElement(newelement, false) {
			return value, false
		}
	}
}
