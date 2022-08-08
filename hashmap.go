// Package hashmap provides a lock-free and thread-safe HashMap.
package hashmap

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"unsafe"
)

// DefaultSize is the default size for a zero allocated map.
const DefaultSize = 8

// MaxFillRate is the maximum fill rate for the slice before a resize  will happen.
const MaxFillRate = 50

type keyConstraint interface {
	string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr
}

// HashMap implements a read optimized hash map.
type HashMap[Key keyConstraint, Value any] struct {
	store      atomic.Pointer[store[Key, Value]] // pointer to a map instance that gets replaced if the map resizes
	linkedList atomic.Pointer[List[Key, Value]]  // key sorted linked list of elements
	// resizing marks a resizing operation in progress.
	// this is using uintptr instead of atomic.Bool to avoid using 32 bit int on 64 bit systems
	resizing atomic.Uintptr
}

// KeyValue represents a key/value that is returned by the iterator.
type KeyValue[Key keyConstraint, Value any] struct {
	Key   Key
	Value Value
}

// New returns a new HashMap instance.
func New[Key keyConstraint, Value any]() *HashMap[Key, Value] {
	m := &HashMap[Key, Value]{}
	return m
}

// NewSized returns a new HashMap instance with a specific initialization size.
func NewSized[Key keyConstraint, Value any](size uintptr) *HashMap[Key, Value] {
	m := &HashMap[Key, Value]{}
	m.allocate(size)
	return m
}

// Len returns the number of elements within the map.
func (m *HashMap[Key, Value]) Len() int {
	list := m.linkedList.Load()
	return list.Len()
}

func (m *HashMap[Key, Value]) allocate(newSize uintptr) {
	list := NewList[Key, Value]()
	// atomic swap in case of another allocation happening concurrently
	if m.linkedList.CompareAndSwap(nil, list) {
		if m.resizing.CompareAndSwap(0, 1) {
			m.grow(newSize, false)
		}
	}
}

// FillRate returns the fill rate of the map as an percentage integer.
func (m *HashMap[Key, Value]) FillRate() int {
	data := m.store.Load()
	count := int(data.count.Load())
	l := len(data.index)
	return (count * 100) / l
}

func (m *HashMap[Key, Value]) isResizeNeeded(store *store[Key, Value], count uintptr) bool {
	l := uintptr(len(store.index))
	if l == 0 {
		return false
	}
	fillRate := (count * 100) / l
	return fillRate > MaxFillRate
}

func (m *HashMap[Key, Value]) indexElement(hashedKey uintptr) (*store[Key, Value], *ListElement[Key, Value]) {
	store := m.store.Load()
	if store == nil {
		return nil, nil
	}
	index := hashedKey >> store.keyShifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(store.array) + index*intSizeBytes))
	item := (*ListElement[Key, Value])(atomic.LoadPointer(ptr))
	return store, item
}

/* The Golang 1.10.1 compiler does not inline this function well
func (m *HashMap[Key, Value]) searchItem(item *ListElement, key interface{}, keyHash uintptr) (value interface{}, ok bool) {
	for item != nil {
		if item.keyHash == keyHash && item.key == key {
			return item.Value(), true
		}

		if item.keyHash > keyHash {
			return nil, false
		}

		item = item.Next()
	}
	return nil, false
}
*/

// Del deletes the key from the map and returns whether the key was deleted.
func (m *HashMap[Key, Value]) Del(key interface{}) bool {
	list := m.linkedList.Load()
	if list == nil {
		return false
	}

	hash := getKeyHash(key)

	var element *ListElement[Key, Value]
ElementLoop:
	for _, element = m.indexElement(hash); element != nil; element = element.Next() {
		if element.keyHash == hash && element.key == key {
			break ElementLoop
		}

		if element.keyHash > hash {
			return false
		}
	}
	if element == nil {
		return false
	}

	m.deleteElement(element)
	list.Delete(element)
	return true
}

// deleteElement deletes an element from index.
func (m *HashMap[Key, Value]) deleteElement(element *ListElement[Key, Value]) {
	for {
		data := m.store.Load()
		index := element.keyHash >> data.keyShifts
		ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(data.array) + index*intSizeBytes))

		next := element.Next()
		if next != nil && element.keyHash>>data.keyShifts != index {
			next = nil // do not set index to next item if it's not the same slice index
		}
		atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(element), unsafe.Pointer(next))

		currentData := m.store.Load()
		if data == currentData { // check that no resize happened
			break
		}
	}
}

// Insert sets the value under the specified key to the map if it does not exist yet.
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map
// after the resize operation is finished.
// Returns true if the item was inserted or false if it existed.
func (m *HashMap[Key, Value]) Insert(key Key, value Value) bool {
	hash := getKeyHash(key)
	element := &ListElement[Key, Value]{
		key:     key,
		keyHash: hash,
	}
	element.value.Store(&value)
	return m.insertListElement(element, false)
}

// Set sets the value under the specified key to the map. An existing item for this key will be overwritten.
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map only after the resize operation is finished.
func (m *HashMap[Key, Value]) Set(key Key, value Value) {
	hash := getKeyHash(key)
	element := &ListElement[Key, Value]{
		key:     key,
		keyHash: hash,
	}
	element.value.Store(&value)
	m.insertListElement(element, true)
}

func (m *HashMap[Key, Value]) insertListElement(element *ListElement[Key, Value], update bool) bool {
	for {
		data, existing := m.indexElement(element.keyHash)
		if data == nil {
			m.allocate(DefaultSize)
			continue // read map data and slice item again
		}
		list := m.linkedList.Load()

		if update {
			if !list.AddOrUpdate(element, existing) {
				continue // a concurrent add did interfere, try again
			}
		} else {
			existed, inserted := list.Add(element, existing)
			if existed {
				return false
			}
			if !inserted {
				continue
			}
		}

		count := data.addItem(element)
		if m.isResizeNeeded(data, count) {
			if m.resizing.CompareAndSwap(0, 1) {
				go m.grow(0, true)
			}
		}
		return true
	}
}

// Cas performs a compare and swap operation sets the value under the specified hash key to the map. An existing item for this key will be overwritten.
func (m *HashMap[Key, Value]) Cas(key Key, from, to Value) bool {
	hash := getKeyHash(key)
	data, existing := m.indexElement(hash)
	if data == nil {
		return false
	}
	list := m.linkedList.Load()
	if list == nil {
		return false
	}

	element := &ListElement[Key, Value]{
		key:     key,
		keyHash: hash,
	}
	element.value.Store(&to)
	return list.Cas(element, &from, existing)
}

// Grow resizes the hashmap to a new size, gets rounded up to next power of 2.
// To double the size of the hashmap use newSize 0.
// This function returns immediately, the resize operation is done in a goroutine.
// No resizing is done in case of another resize operation already being in progress.
func (m *HashMap[Key, Value]) Grow(newSize uintptr) {
	if m.resizing.CompareAndSwap(0, 1) {
		go m.grow(newSize, true)
	}
}

func (m *HashMap[Key, Value]) grow(newSize uintptr, loop bool) {
	defer m.resizing.CompareAndSwap(1, 0)

	for {
		currentData := m.store.Load()
		if newSize == 0 {
			newSize = uintptr(len(currentData.index)) << 1
		} else {
			newSize = roundUpPower2(newSize)
		}

		index := make([]*ListElement[Key, Value], newSize)
		header := (*reflect.SliceHeader)(unsafe.Pointer(&index))

		newData := &store[Key, Value]{
			keyShifts: strconv.IntSize - log2(newSize),
			array:     unsafe.Pointer(header.Data), // use address of slice data storage
			index:     index,
		}

		m.fillIndexItems(newData) // initialize new index slice with longer keys

		m.store.Store(newData)

		m.fillIndexItems(newData) // make sure that the new index is up to date with the current state of the linked list

		if !loop {
			break
		}

		// check if a new resize needs to be done already
		count := uintptr(m.Len())
		if !m.isResizeNeeded(newData, count) {
			break
		}
		newSize = 0 // 0 means double the current size
	}
}

func (m *HashMap[Key, Value]) fillIndexItems(store *store[Key, Value]) {
	list := m.linkedList.Load()
	if list == nil {
		return
	}
	first := list.First()
	item := first
	lastIndex := uintptr(0)

	for item != nil {
		index := item.keyHash >> store.keyShifts
		if item == first || index != lastIndex { // store item with smallest hash key for every index
			store.addItem(item)
			lastIndex = index
		}
		item = item.Next()
	}
}

// String returns the map as a string, only hashed keys are printed.
func (m *HashMap[Key, Value]) String() string {
	list := m.linkedList.Load()
	if list == nil {
		return "[]"
	}

	buffer := bytes.NewBufferString("")
	buffer.WriteRune('[')

	first := list.First()
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

// Iter returns an iterator which could be used in a for range loop.
// The order of the items is sorted by hash keys.
func (m *HashMap[Key, Value]) Iter() <-chan KeyValue[Key, Value] {
	ch := make(chan KeyValue[Key, Value]) // do not use a size here since items can get added during iteration

	go func() {
		list := m.linkedList.Load()
		if list == nil {
			close(ch)
			return
		}

		item := list.First()
		for item != nil {
			value := item.Value()
			ch <- KeyValue[Key, Value]{
				Key:   item.key,
				Value: value,
			}
			item = item.Next()
		}
		close(ch)
	}()

	return ch
}
