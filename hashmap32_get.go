package hashmap

import (
	"reflect"
	"sync/atomic"
	"unsafe"

//	"github.com/dchest/siphash"
)

// Get retrieves an element from the map under given hash key.
// Using interface{} adds a performance penalty.
// Please consider using GetUintKey or GetStringKey instead.
func (m *HashMap32) Get(key interface{}) (value unsafe.Pointer, ok bool) {
	hashedKey := getKeyHash32(key)

	// inline HashMap32.getSliceItemForKey()
	mapData := (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey && entry.key == key {
			// inline ListElement32.Value()
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement32)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement32.Next()
	}
	return nil, false
}

// GetUintKey retrieves an element from the map under given integer key.
func (m *HashMap32) GetUintKey(key uint32) (value unsafe.Pointer, ok bool) {
	mapData := (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}

	bh := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(&key)),
		Len:  4,
		Cap:  4,
	}
	buf := *(*[]byte)(unsafe.Pointer(&bh))
//	hashedKey := siphash.Hash(sipHashKey1, sipHashKey2, buf)
	hashedKey := XXHash_GoChecksum32(buf)

	// inline HashMap32.getSliceItemForKey()
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey && entry.key == key {
			// inline ListElement32.Value()
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement32)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement32.Next()
	}
	return nil, false
}

// GetStringKey retrieves an element from the map under given string key.
func (m *HashMap32) GetStringKey(key string) (value unsafe.Pointer, ok bool) {
	mapData := (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
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
	// hashedKey := siphash.Hash(sipHashKey1, sipHashKey2, buf)
	hashedKey := XXHash_GoChecksum32(buf)

	// inline HashMap32.getSliceItemForKey()
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey && entry.key == key {
			// inline ListElement32.Value()
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement32)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement32.Next()
	}
	return nil, false
}

// GetHashedKey retrieves an element from the map under given hashed key.
func (m *HashMap32) GetHashedKey(hashedKey uint32) (value unsafe.Pointer, ok bool) {
	// inline HashMap32.getSliceItemForKey()
	mapData := (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
	if mapData == nil {
		return nil, false
	}
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer))

	for entry != nil {
		if entry.keyHash == hashedKey {
			// inline ListElement32.Value()
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			value = atomic.LoadPointer(&entry.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUint32(&entry.deleted) == 1 {
				return nil, false
			}
			return value, true
		}

		if entry.keyHash > hashedKey {
			return nil, false
		}

		entry = (*ListElement32)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement32.Next()
	}
	return nil, false
}

// GetOrInsert returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *HashMap32) GetOrInsert(key interface{}, value unsafe.Pointer) (actual unsafe.Pointer, loaded bool) {
	hashedKey := getKeyHash32(key)
	var newEntry *ListElement32

	for {
		// inline HashMap32.getSliceItemForKey()
		mapData := (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
		if mapData == nil {
			m.allocate(DefaultSize)
			mapData = (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
		}
		index := hashedKey >> mapData.keyRightShifts
		sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
		sliceItem := (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer))

		entry := sliceItem
		for entry != nil {
			if entry.keyHash == hashedKey && entry.key == key {
				actual, loaded := entry.GetOrSetValue(value)
				if loaded {
					return actual, true
				}

				list := m.list()
				atomic.AddUint32(&list.count, 1)
				return value, false
			}

			if entry.keyHash > hashedKey {
				break
			}

			entry = (*ListElement32)(atomic.LoadPointer(&entry.nextElement)) // inline ListElement32.Next()
		}

		if newEntry == nil { // allocate only once
			newEntry = &ListElement32{
				key:     key,
				keyHash: hashedKey,
				value:   value,
			}
		}

		if m.insertListElement(newEntry, false) {
			return value, false
		}
	}
}
