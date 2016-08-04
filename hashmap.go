package hashmap

import (
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"unsafe"
)

const intSizeBytes = strconv.IntSize >> 3

type (
	hashMapEntry struct {
		key1  uint64
		key2  uint64
		value interface{}
	}

	hashMapData struct {
		andMask uint64
		data    unsafe.Pointer
		size    uint64
		count   uint64
		slice   []*hashMapEntry
	}

	// HashMap implements a read optimized hash map
	HashMap struct {
		mapData unsafe.Pointer
		sync.Mutex
	}
)

// New returns a new HashMap.
func New() *HashMap {
	return NewSize(8)
}

// NewSize returns a new HashMap instance with a specific initialization size.
func NewSize(size uint64) *HashMap {
	hashmap := &HashMap{}
	hashmap.Resize(size)
	return hashmap
}

// Count returns the number of elements within the map.
func (m *HashMap) Count() uint64 {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	return atomic.LoadUint64(&mapData.count)
}

// Get retrieves an element from map under given key.
func (m *HashMap) Get(key1 uint64, key2 uint64) (interface{}, bool) {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key1 & mapData.andMask
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))

	if entry == nil || key1 != entry.key1 || key2 != entry.key2 {
		return nil, false
	}

	return entry.value, true
}

// Set sets the given value under the specified key.
func (m *HashMap) Set(key1 uint64, key2 uint64, value interface{}) {
	m.Lock()
	defer m.Unlock()

	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key1 & mapData.andMask
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))

	if entry != nil { // space in slice is used?
		if key1 == entry.key1 && key2 == entry.key2 { // slice entry keys match what we are looking for?
			if value == entry.value { // trying to set the same key and value?
				return
			}
		} else {
			m.Resize(mapData.size + 1) // collision found with shortened key, resize
		}

		for {
			existingEntry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))
			if existingEntry == nil || existingEntry.key1 == key1 { // last resizing operation fixed the collision?
				break
			}
			m.Resize(mapData.size + 1)

			mapData = (*hashMapData)(atomic.LoadPointer(&m.mapData))                                                       // update pointer
			index = key1 & mapData.andMask                                                                                 // update index key
			sliceDataIndexPointer = (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes))) // update index pointer
		}
	}

	entry = &hashMapEntry{ // create a new instance in the update case as well, updating value would not be thread-safe
		key1:  key1,
		key2:  key2,
		value: value,
	}

	atomic.StorePointer((*unsafe.Pointer)(sliceDataIndexPointer), unsafe.Pointer(entry))
	atomic.AddUint64(&mapData.count, 1)
}

// Remove removes an element from the map.
func (m *HashMap) Remove(key1 uint64, key2 uint64) {
	m.Lock()
	defer m.Unlock()

	_, exists := m.Get(key1, key2)
	if !exists {
		return
	}

	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key1 & mapData.andMask

	sliceDataIndexPointer := unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes))
	atomic.StorePointer((*unsafe.Pointer)(sliceDataIndexPointer), nil)
	atomic.AddUint64(&mapData.count, ^uint64(0))
}

// Resize resizes the hashmap to a new size, gets rounded up to next power of 2
// Locking of the hashmap needs to be done outside of this function
func (m *HashMap) Resize(newSize uint64) {
	newSize = roundUpPower2(newSize)
	newSlice := make([]*hashMapEntry, newSize)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&newSlice))

	newMapData := &hashMapData{
		andMask: newSize - 1,
		data:    unsafe.Pointer(header.Data),
		size:    newSize,
		count:   0,
		slice:   newSlice,
	}

	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	if mapData != nil { // copy hashmap contents to new slice with longer key
		newMapData.count = mapData.count
		for _, entry := range mapData.slice {
			if entry == nil {
				continue
			}

			index := entry.key1 & mapData.andMask
			newSlice[index] = entry
		}
	}

	atomic.StorePointer(&m.mapData, unsafe.Pointer(newMapData))
}

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
