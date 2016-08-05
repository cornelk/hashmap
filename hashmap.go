package hashmap

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

type (
	hashMapEntry struct {
		key   uint64
		value interface{}
	}

	hashMapData struct {
		andMask uint64
		data    unsafe.Pointer
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

// Len returns the number of elements within the map.
func (m *HashMap) Len() uint64 {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	return atomic.LoadUint64(&mapData.count)
}

// Get retrieves an element from map under given key.
func (m *HashMap) Get(key uint64) (interface{}, bool) {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key & mapData.andMask
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))

	if entry == nil || key != entry.key {
		return nil, false
	}

	return entry.value, true
}

// Add adds the value under the specified key to the map. An existing item for this key will be overwritten.
func (m *HashMap) Add(key uint64, value interface{}) {
	m.Lock()
	defer m.Unlock()

	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key & mapData.andMask
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))

	if entry != nil { // space in slice is used?
		if key == entry.key { // slice entry keys match what we are looking for?
			if value == entry.value { // trying to set the same key and value?
				return
			}
		} else {
			m.Resize(0) // collision found with shortened key, resize
		}

		for {
			existingEntry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))
			if existingEntry == nil || existingEntry.key == key { // last resizing operation fixed the collision?
				break
			}
			m.Resize(0)

			mapData = (*hashMapData)(atomic.LoadPointer(&m.mapData))                                                       // update pointer
			index = key & mapData.andMask                                                                                  // update index key
			sliceDataIndexPointer = (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes))) // update index pointer
		}
	}

	entry = &hashMapEntry{ // create a new instance in the update case as well, updating value would not be thread-safe
		key:   key,
		value: value,
	}

	atomic.StorePointer((*unsafe.Pointer)(sliceDataIndexPointer), unsafe.Pointer(entry))
	atomic.AddUint64(&mapData.count, 1)
}

// Del removes an element from the map.
func (m *HashMap) Del(key uint64) {
	m.Lock() // lock before getting the value to make sure to work on the latest data slice
	defer m.Unlock()

	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	index := key & mapData.andMask
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	entry := (*hashMapEntry)(atomic.LoadPointer(sliceDataIndexPointer))
	if entry == nil || key != entry.key {
		return
	}

	atomic.StorePointer((*unsafe.Pointer)(sliceDataIndexPointer), nil) // clear item in slice
	atomic.AddUint64(&mapData.count, ^uint64(0))                       // decrement item counter
}

// Resize resizes the hashmap to a new size, gets rounded up to next power of 2.
// To double the size of the hashmap use newSize 0.
// Locking of the hashmap needs to be done outside of this function.
func (m *HashMap) Resize(newSize uint64) {
	mapData := (*hashMapData)(atomic.LoadPointer(&m.mapData))
	if newSize == 0 {
		newSize = uint64(len(mapData.slice)) << 1
	} else {
		newSize = roundUpPower2(newSize)
	}
	newSlice := make([]*hashMapEntry, newSize)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&newSlice))

	newMapData := &hashMapData{
		andMask: newSize - 1,
		data:    unsafe.Pointer(header.Data), // use address of slice storage
		count:   0,
		slice:   newSlice,
	}

	if mapData != nil && mapData.count > 0 { // copy hashmap contents to new slice with longer key
		newMapData.count = mapData.count
		for _, entry := range mapData.slice {
			if entry == nil {
				continue
			}

			index := entry.key & mapData.andMask
			newSlice[index] = entry
		}
	}

	atomic.StorePointer(&m.mapData, unsafe.Pointer(newMapData))
}
