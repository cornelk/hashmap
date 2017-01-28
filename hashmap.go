package hashmap

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// MaxFillRate is the maximum fill rate for the slice before a resize  will happen
const MaxFillRate = 50

type (
	hashMapData struct {
		keyRightShifts uint64         // 64 - log2 of array size, to be used as index in the data array
		data           unsafe.Pointer // pointer to slice data array
		slice          []*ListElement // storage for the slice for the garbage collector to not clean it up
		count          uint64         // count of filled elements in the slice
	}

	// HashMap implements a read optimized hash map
	HashMap struct {
		mapData    unsafe.Pointer // pointer to a map instance that gets replaced if the map resizes
		linkedList *List          // key sorted linked list of elements
		sync.Mutex                // mutex that is only used for resize operations
	}
)

// New returns a new HashMap.
func New() *HashMap {
	return NewSize(8)
}

// NewSize returns a new HashMap instance with a specific initialization size.
func NewSize(size uint64) *HashMap {
	hashmap := &HashMap{
		linkedList: NewList(),
	}
	hashmap.Grow(size)
	return hashmap
}

// Len returns the number of elements within the map.
func (m *HashMap) Len() uint64 {
	return m.linkedList.Len()
}

// Fillrate returns the fill rate of the map as an percentage integer.
func (m *HashMap) Fillrate() uint64 {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	count := atomic.LoadUint64(&mapData.count)
	sliceLen := uint64(len(mapData.slice))
	return (count * 100) / sliceLen
}

func (m *HashMap) getSliceItemForKey(key uint64) *ListElement {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*ListElement)(atomic.LoadPointer(sliceDataIndexPointer))
	return entry
}

// Get retrieves an element from map under given key.
func (m *HashMap) Get(key uint64) (unsafe.Pointer, bool) {
	entry := m.getSliceItemForKey(key)

	for {
		if entry == nil {
			return nil, false
		}

		if entry.keyHash == key {
			if atomic.LoadUint64(&entry.deleted) == 0 {
				return entry.value, true
			}
			return nil, false
		}

		if entry.keyHash > key {
			return nil, false
		}

		entry = entry.Next()
	}
}

// Del deletes the key from the map.
func (m *HashMap) Del(key uint64) {
	entry := m.getSliceItemForKey(key)

	for {
		if entry == nil {
			return
		}

		if entry.keyHash == key {
			m.linkedList.Delete(entry)
			return
		}

		if entry.keyHash > key {
			return
		}

		entry = entry.Next()
	}
}

// Add adds the value under the specified key to the map. An existing item for this key will be overwritten.
func (m *HashMap) Add(key uint64, value unsafe.Pointer) {
	newEntry := &ListElement{
		keyHash: key,
		value:   value,
	}

	for {
		mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
		index := key >> mapData.keyRightShifts
		sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))

		sliceItem := (*ListElement)(atomic.LoadPointer(sliceDataIndexPointer))
		if !m.linkedList.Add(newEntry, sliceItem) {
			continue // a concurrent add did interfere, try again
		}

		if sliceItem == nil {
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(sliceDataIndexPointer), nil, unsafe.Pointer(newEntry)) {
				newSliceCount := atomic.AddUint64(&mapData.count, 1)
				sliceLen := uint64(len(mapData.slice))
				fillRate := (newSliceCount * 100) / sliceLen

				if fillRate > MaxFillRate { // check if the slice needs to be resized
					m.Lock()
					currentMapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
					if mapData != currentMapData { // double check, a concurrent resize happened
						m.Unlock()
						continue
					}

					m.grow(0)
					m.Unlock()
					return
				}
			}
		} else {
			if newEntry.keyHash < sliceItem.keyHash { // the new item is the smallest for this index
				atomic.CompareAndSwapPointer((*unsafe.Pointer)(sliceDataIndexPointer), unsafe.Pointer(sliceItem), unsafe.Pointer(newEntry))
			}
		}

		currentMapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
		if mapData != currentMapData { // a resize operation happened while we were inserting?
			continue // retry
		}
		return
	}
}

// Grow resizes the hashmap to a new size, gets rounded up to next power of 2.
// To double the size of the hashmap use newSize 0.
func (m *HashMap) Grow(newSize uint64) {
	m.Lock()
	m.grow(newSize)
	m.Unlock()
}

// Grow resizes the hashmap to a new size, gets rounded up to next power of 2.
// To double the size of the hashmap use newSize 0.
func (m *HashMap) grow(newSize uint64) {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	if newSize == 0 {
		newSize = uint64(len(mapData.slice)) << 1
	} else {
		newSize = roundUpPower2(newSize)
	}

	newSlice := make([]*ListElement, newSize)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&newSlice))

	newMapData := &hashMapData{
		keyRightShifts: 64 - log2(newSize),
		data:           unsafe.Pointer(header.Data), // use address of slice data storage
		slice:          newSlice,
	}

	if mapData != nil { // copy hashmap contents to new slice with longer key
		first := m.linkedList.First()
		item := first
		lastIndex := uint64(0)

		for item != nil {
			index := item.keyHash >> newMapData.keyRightShifts

			if item == first || index != lastIndex { // store item with smallest hash key for every index
				newSlice[index] = item
				newMapData.count++
				lastIndex = index
			}
			item = item.Next()
		}
	}

	atomic.StorePointer(&m.mapData, unsafe.Pointer(newMapData))
}

// String returns the map as a string, only keys are printed
func (m *HashMap) String() string {
	buffer := bytes.NewBufferString("")
	buffer.WriteRune('[')

	first := m.linkedList.First()
	item := first

	for item != nil {
		if item != first {
			buffer.WriteRune(',')
		}
		fmt.Fprint(buffer, item.keyHash)

		item = item.Next()
	}

	buffer.WriteRune(']')
	return buffer.String()
}
