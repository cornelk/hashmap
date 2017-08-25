package hashmap

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

// DefaultSize is the default size for a zero allocated map
const DefaultSize32 = 8

// MaxFillRate is the maximum fill rate for the slice before a resize  will happen.
const MaxFillRate32 = 50

type (
	hashMapData32 struct {
		keyRightShifts uint32         // 32 now --> 64 - log2 of array size, to be used as index in the data array
		count          uint32         // count of filled elements in the slice
		data           unsafe.Pointer // pointer to slice data array
		slice          []*ListElement32 // storage for the slice for the garbage collector to not clean it up
	}

	// HashMap32 implements a read optimized hash map.
	HashMap32 struct {
		mapDataPtr unsafe.Pointer // pointer to a map instance that gets replaced if the map resizes
		linkedList unsafe.Pointer // key sorted linked list of elements
		sync.Mutex                // mutex that is only used for resize operations
	}

	// KeyValue represents a key/value that is returned by the iterator.

	// KeyValue struct {
	// 	Key   interface{}
	// 	Value unsafe.Pointer
	// }
)

// New returns a new HashMap32 instance with a specific initialization size.
func New32(size uint32) *HashMap32 {
	m := &HashMap32{}
	m.allocate(size)
	return m
}

// Len returns the number of elements within the map.
func (m *HashMap32) Len() uint32 {
	list := m.list()
	if list == nil {
		return 0
	}
	return list.Len()
}

func (m *HashMap32) mapData() *hashMapData32 {
	return (*hashMapData32)(atomic.LoadPointer(&m.mapDataPtr))
}

func (m *HashMap32) list() *List32 {
	return (*List32)(atomic.LoadPointer(&m.linkedList))
}

func (m *HashMap32) allocate(newSize uint32) {
	m.Lock()
	defer m.Unlock()

	mapData := m.mapData()
	if mapData != nil { // check that no other allocation happened
		return
	}

	list := NewList32()
	atomic.StorePointer(&m.linkedList, unsafe.Pointer(list))
	m.grow(newSize)
}

// Fillrate returns the fill rate of the map as an percentage integer.
func (m *HashMap32) Fillrate() uint32 {
	mapData := m.mapData()
	count := atomic.LoadUint32(&mapData.count)
	sliceLen := uint32(len(mapData.slice))
	return (count * 100) / sliceLen
}

func (m *HashMap32) getSliceItemForKey(hashedKey uint32) (mapData *hashMapData32, item *ListElement32) {
	mapData = m.mapData()
	if mapData == nil {
		return nil, nil
	}
	index := hashedKey >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))
	item = (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer))
	return mapData, item
}

// Del deletes the hashed key from the map.
func (m *HashMap32) Del(key interface{}) {
	list := m.list()
	if list == nil {
		return
	}

	hashedKey := getKeyHash32(key)
	for _, entry := m.getSliceItemForKey(hashedKey); entry != nil; entry = entry.Next() {
		if entry.keyHash == hashedKey && entry.key == key {
			list.Delete(entry)
			return
		}

		if entry.keyHash > hashedKey {
			return
		}
	}
}

// DelHashedKey deletes the hashed key from the map.
func (m *HashMap32) DelHashedKey(hashedKey uint32) {
	list := m.list()
	if list == nil {
		return
	}

	for _, entry := m.getSliceItemForKey(hashedKey); entry != nil; entry = entry.Next() {
		if entry.keyHash == hashedKey {
			list.Delete(entry)
			return
		}

		if entry.keyHash > hashedKey {
			return
		}
	}
}

// Insert sets the value under the specified key to the map if it does not exist yet.
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map only after the resize operation is finished.
// Returns true if the item was inserted or false if it existed.
func (m *HashMap32) Insert(key interface{}, value unsafe.Pointer) bool {
	hashedKey := getKeyHash32(key)

	newEntry := &ListElement32{
		key:     key,
		keyHash: hashedKey,
		value:   value,
	}

	return m.insertListElement(newEntry, false)
}

// Set sets the value under the specified key to the map. An existing item for this key will be overwritten.
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map only after the resize operation is finished.
func (m *HashMap32) Set(key interface{}, value unsafe.Pointer) {
	hashedKey := getKeyHash32(key)

	newEntry := &ListElement32{
		key:     key,
		keyHash: hashedKey,
		value:   value,
	}

	m.insertListElement(newEntry, true)
}

// SetHashedKey sets the value under the specified hash key to the map. An existing item for this key will be overwritten.
// You can use this function if your keys are already hashes and you want to avoid another hashing of the key.
// Do not use non hashes as keys for this function, the performance would decrease!
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map only after the resize operation is finished.
func (m *HashMap32) SetHashedKey(hashedKey uint32, value unsafe.Pointer) {
	newEntry := &ListElement32{
		key:     hashedKey,
		keyHash: hashedKey,
		value:   value,
	}

	m.insertListElement(newEntry, true)
}

func (m *HashMap32) insertListElement(newEntry *ListElement32, update bool) bool {
	for {
		mapData, sliceItem := m.getSliceItemForKey(newEntry.keyHash)
		if mapData == nil {
			m.allocate(DefaultSize32)
			continue // read mapdata and slice item again
		}
		list := m.list()

		if update {
			if !list.AddOrUpdate(newEntry, sliceItem) {
				continue // a concurrent add did interfere, try again
			}
		} else {
			existed, inserted := list.Add(newEntry, sliceItem)
			if existed {
				return false
			}
			if !inserted {
				continue
			}
		}

		newSliceCount := mapData.addItemToIndex(newEntry)
		if newSliceCount != 0 {
			sliceLen := uint32(len(mapData.slice))
			fillRate := (newSliceCount * 100) / sliceLen

			if fillRate > MaxFillRate32 { // check if the slice needs to be resized
				m.Lock()
				currentMapData := m.mapData()
				if mapData == currentMapData { // double check that no other resize happened
					m.grow(0)
				}
				m.Unlock()
			}
		}
		return true
	}
}

// CasHashedKey performs a compare and swap operation sets the value under the specified hash key to the map. An existing item for this key will be overwritten.
func (m *HashMap32) CasHashedKey(hashedKey uint32, from, to unsafe.Pointer) bool {
	newEntry := &ListElement32{
		key:     hashedKey,
		keyHash: hashedKey,
		value:   to,
	}

	for {
		mapData, sliceItem := m.getSliceItemForKey(hashedKey)
		if mapData == nil {
			return false
		}
		list := m.list()
		if list == nil {
			return false
		}
		if !list.Cas(newEntry, from, sliceItem) {
			return false
		}

		newSliceCount := mapData.addItemToIndex(newEntry)
		if newSliceCount != 0 {
			sliceLen := uint32(len(mapData.slice))
			fillRate := (newSliceCount * 100) / sliceLen

			if fillRate > MaxFillRate32 { // check if the slice needs to be resized
				m.Lock()
				currentMapData := m.mapData()
				if mapData == currentMapData { // double check that no other resize happened
					m.grow(0)
				}
				m.Unlock()
			}
		}
		return true
	}
}

// Cas performs a compare and swap operation sets the value under the specified hash key to the map. An existing item for this key will be overwritten.
func (m *HashMap32) Cas(key interface{}, from, to unsafe.Pointer) bool {
	hashedKey := getKeyHash32(key)
	return m.CasHashedKey(hashedKey, from, to)
}

// adds an item to the index if needed and returns the new item counter if it changed, otherwise 0
func (mapData *hashMapData32) addItemToIndex(item *ListElement32) uint32 {
	index := item.keyHash >> mapData.keyRightShifts
	sliceDataIndexPointer := (*unsafe.Pointer)(unsafe.Pointer(uintptr(mapData.data) + uintptr(index*intSizeBytes)))

	for { // loop until the smallest key hash is in the index
		sliceItem := (*ListElement32)(atomic.LoadPointer(sliceDataIndexPointer)) // get the current item in the index
		if sliceItem == nil {                                                  // no item yet at this index
			if atomic.CompareAndSwapPointer(sliceDataIndexPointer, nil, unsafe.Pointer(item)) {
				return atomic.AddUint32(&mapData.count, 1)
			}
			continue // a new item was inserted concurrently, retry
		}

		if item.keyHash < sliceItem.keyHash {
			// the new item is the smallest for this index?
			if !atomic.CompareAndSwapPointer(sliceDataIndexPointer, unsafe.Pointer(sliceItem), unsafe.Pointer(item)) {
				continue // a new item was inserted concurrently, retry
			}
		}
		return 0
	}
}

// Grow resizes the hashmap to a new size, gets rounded up to next power of 2.
// To double the size of the hashmap use newSize 0.
func (m *HashMap32) Grow(newSize uint32) {
	m.Lock()
	m.grow(newSize)
	m.Unlock()
}

func (m *HashMap32) grow(newSize uint32) {
	mapData := m.mapData()
	if newSize == 0 {
		newSize = uint32(len(mapData.slice)) << 1
	} else {
		newSize = roundUpPower2_32(newSize)
	}

	newSlice := make([]*ListElement32, newSize)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&newSlice))

	newMapData := &hashMapData32{
		keyRightShifts: 32 - log2_32(newSize),
		data:           unsafe.Pointer(header.Data), // use address of slice data storage
		slice:          newSlice,
	}

	m.fillIndexItems(newMapData) // initialize new index slice with longer keys

	atomic.StorePointer(&m.mapDataPtr, unsafe.Pointer(newMapData))

	m.fillIndexItems(newMapData) // make sure that the new index is up to date with the current state of the linked list
}

func (m *HashMap32) fillIndexItems(mapData *hashMapData32) {
	list := m.list()
	if list == nil {
		return
	}
	first := list.First()
	item := first
	lastIndex := uint32(0)

	for item != nil {
		index := item.keyHash >> mapData.keyRightShifts
		if item == first || index != lastIndex { // store item with smallest hash key for every index
			if !item.Deleted() {
				mapData.addItemToIndex(item)
				lastIndex = index
			}
		}
		item = item.Next()
	}
}

// String returns the map as a string, only hashed keys are printed.
func (m *HashMap32) String() string {
	list := m.list()
	if list == nil {
		return "[]"
	}

	buffer := bytes.NewBufferString("")
	buffer.WriteRune('[')

	first := list.First()
	item := first

	for item != nil {
		if !item.Deleted() {
			if item != first {
				buffer.WriteRune(',')
			}
			fmt.Fprint(buffer, item.keyHash)
		}

		item = item.Next()
	}
	buffer.WriteRune(']')
	return buffer.String()
}

// Iter returns an iterator which could be used in a for range loop.
// The order of the items is sorted by hash keys.
func (m *HashMap32) Iter() <-chan KeyValue {
	ch := make(chan KeyValue) // do not use a size here since items can get added during iteration

	go func() {
		list := m.list()
		if list == nil {
			close(ch)
			return
		}
		item := list.First()
		for item != nil {
			value, ok := item.Value()
			if ok {
				ch <- KeyValue{item.key, value}
			}
			item = item.Next()
		}
		close(ch)
	}()

	return ch
}
