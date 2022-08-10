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

// HashMap implements a read optimized hash map.
type HashMap[Key comparable, Value any] struct {
	hasher     func(Key) uintptr
	store      atomic.Pointer[store[Key, Value]] // pointer to a map instance that gets replaced if the map resizes
	linkedList atomic.Pointer[List[Key, Value]]  // key sorted linked list of elements
	// resizing marks a resizing operation in progress.
	// this is using uintptr instead of atomic.Bool to avoid using 32 bit int on 64 bit systems
	resizing atomic.Uintptr
}

// KeyValue represents a key/value that is returned by the iterator.
type KeyValue[Key comparable, Value any] struct {
	Key   Key
	Value Value
}

// New returns a new HashMap instance.
func New[Key comparable, Value any]() *HashMap[Key, Value] {
	return NewSized[Key, Value](DefaultSize)
}

// NewSized returns a new HashMap instance with a specific initialization size.
func NewSized[Key comparable, Value any](size uintptr) *HashMap[Key, Value] {
	m := &HashMap[Key, Value]{}
	m.allocate(size)
	m.setDefaultHasher()
	return m
}

// SetHasher sets a custom hasher.
func (m *HashMap[Key, Value]) SetHasher(hasher func(Key) uintptr) {
	m.hasher = hasher
}

// Len returns the number of elements within the map.
func (m *HashMap[Key, Value]) Len() int {
	list := m.linkedList.Load()
	return list.Len()
}

// Get retrieves an element from the map under given hash key.
func (m *HashMap[Key, Value]) Get(key Key) (Value, bool) {
	hash := m.hasher(key)
	store := m.store.Load()
	element := store.item(hash)

	// inline HashMap.searchItem()
	for element != nil {
		if element.keyHash == hash && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > hash {
			return *new(Value), false
		}

		element = element.Next()
	}
	return *new(Value), false
}

// GetOrInsert returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The returned bool is true if the value was loaded, false if stored.
func (m *HashMap[Key, Value]) GetOrInsert(key Key, value Value) (Value, bool) {
	hash := m.hasher(key)
	var newElement *ListElement[Key, Value]

	for {
		store := m.store.Load()
		element := store.item(hash)

		for element != nil {
			if element.keyHash == hash && element.key == key {
				actual := element.Value()
				return actual, true
			}

			if element.keyHash > hash {
				break
			}

			element = element.Next()
		}

		if newElement == nil { // allocate only once
			newElement = &ListElement[Key, Value]{
				key:     key,
				keyHash: hash,
			}
			newElement.value.Store(&value)
		}

		if m.insertElement(newElement, false) {
			return value, false
		}
	}
}

// FillRate returns the fill rate of the map as a percentage integer.
func (m *HashMap[Key, Value]) FillRate() int {
	store := m.store.Load()
	count := int(store.count.Load())
	l := len(store.index)
	return (count * 100) / l
}

// Del deletes the key from the map and returns whether the key was deleted.
func (m *HashMap[Key, Value]) Del(key Key) bool {
	hash := m.hasher(key)
	store := m.store.Load()
	element := store.item(hash)
	list := m.linkedList.Load()

	for ; element != nil; element = element.Next() {
		if element.keyHash == hash && element.key == key {
			m.deleteElement(element)
			list.Delete(element)
			return true
		}

		if element.keyHash > hash {
			return false
		}
	}
	return false
}

// Insert sets the value under the specified key to the map if it does not exist yet.
// If a resizing operation is happening concurrently while calling Insert, the item might show up in the map
// after the resize operation is finished.
// Returns true if the item was inserted or false if it existed.
func (m *HashMap[Key, Value]) Insert(key Key, value Value) bool {
	hash := m.hasher(key)
	element := &ListElement[Key, Value]{
		key:     key,
		keyHash: hash,
	}
	element.value.Store(&value)
	return m.insertElement(element, false)
}

// Set sets the value under the specified key to the map. An existing item for this key will be overwritten.
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map
// after the resize operation is finished.
func (m *HashMap[Key, Value]) Set(key Key, value Value) {
	hash := m.hasher(key)
	element := &ListElement[Key, Value]{
		key:     key,
		keyHash: hash,
	}
	element.value.Store(&value)
	m.insertElement(element, true)
}

// Grow resizes the hashmap to a new size, the size gets rounded up to next power of 2.
// To double the size of the hashmap use newSize 0.
// This function returns immediately, the resize operation is done in a goroutine.
// No resizing is done in case of another resize operation already being in progress.
func (m *HashMap[Key, Value]) Grow(newSize uintptr) {
	if m.resizing.CompareAndSwap(0, 1) {
		go m.grow(newSize, true)
	}
}

// String returns the map as a string, only hashed keys are printed.
func (m *HashMap[Key, Value]) String() string {
	list := m.linkedList.Load()

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

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *HashMap[Key, Value]) Range(f func(Key, Value) bool) {
	list := m.linkedList.Load()
	item := list.First()

	for item != nil {
		value := item.Value()
		if !f(item.key, value) {
			return
		}
		item = item.Next()
	}
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

// setDefaultHasher sets the default hasher depending on the key type.
func (m *HashMap[Key, Value]) setDefaultHasher() {
	var key Key
	switch any(key).(type) {
	case string:
		m.hasher = m.stringHasher
	case int, uint, uintptr:
		m.hasher = m.uintptrHasher
	case int8, uint8:
		m.hasher = m.byteHasher
	case int16, uint16:
		m.hasher = m.wordHasher
	case int32, uint32, float32:
		m.hasher = m.dwordHasher
	case int64, uint64, float64, complex64:
		m.hasher = m.qwordHasher
	case complex128:
		m.hasher = m.owordHasher
	default:
		panic(fmt.Errorf("unsupported key type %T", key))
	}
}

func (m *HashMap[Key, Value]) isResizeNeeded(store *store[Key, Value], count uintptr) bool {
	l := uintptr(len(store.index))
	if l == 0 {
		return false
	}
	fillRate := (count * 100) / l
	return fillRate > MaxFillRate
}

/* The Golang 1.10.1 compiler does not inline this function well
func (m *HashMap[Key, Value]) searchItem(item *ListElement[Key, Value], key Key, keyHash uintptr) (Value, bool) {
	for item != nil {
		if item.keyHash == keyHash && item.key == key {
			return item.Value(), true
		}

		if item.keyHash > keyHash {
			return *new(Value), false
		}

		item = item.Next()
	}
	return *new(Value), false
}
*/

func (m *HashMap[Key, Value]) insertElement(element *ListElement[Key, Value], update bool) bool {
	var existed, inserted bool

	for {
		store := m.store.Load()
		existing := store.item(element.keyHash)
		list := m.linkedList.Load()

		if update {
			inserted = list.AddOrUpdate(element, existing)
		} else if !inserted { // if retrying after insert during grow, do not add to list again
			existed, inserted = list.Add(element, existing)
			if existed {
				return false
			}
		}
		if !inserted {
			continue // a concurrent add did interfere, try again
		}

		count := store.addItem(element)
		currentStore := m.store.Load()
		if store != currentStore { // retry insert in case of insert during grow
			continue
		}

		if m.isResizeNeeded(store, count) && m.resizing.CompareAndSwap(0, 1) {
			go m.grow(0, true)
		}
		return true
	}
}

// deleteElement deletes an element from index.
func (m *HashMap[Key, Value]) deleteElement(element *ListElement[Key, Value]) {
	for {
		store := m.store.Load()
		index := element.keyHash >> store.keyShifts
		ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(store.array) + index*intSizeBytes))

		next := element.Next()
		if next != nil && element.keyHash>>store.keyShifts != index {
			next = nil // do not set index to next item if it's not the same slice index
		}
		atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(element), unsafe.Pointer(next))

		currentStore := m.store.Load()
		if store == currentStore { // check that no resize happened
			break
		}
	}
}

func (m *HashMap[Key, Value]) grow(newSize uintptr, loop bool) {
	defer m.resizing.CompareAndSwap(1, 0)

	for {
		currentStore := m.store.Load()
		if newSize == 0 {
			newSize = uintptr(len(currentStore.index)) << 1
		} else {
			newSize = roundUpPower2(newSize)
		}

		index := make([]*ListElement[Key, Value], newSize)
		header := (*reflect.SliceHeader)(unsafe.Pointer(&index))

		newStore := &store[Key, Value]{
			keyShifts: strconv.IntSize - log2(newSize),
			array:     unsafe.Pointer(header.Data), // use address of slice data storage
			index:     index,
		}

		m.fillIndexItems(newStore) // initialize new index slice with longer keys

		m.store.Store(newStore)

		m.fillIndexItems(newStore) // make sure that the new index is up-to-date with the current state of the linked list

		if !loop {
			return
		}

		// check if a new resize needs to be done already
		count := uintptr(m.Len())
		if !m.isResizeNeeded(newStore, count) {
			return
		}
		newSize = 0 // 0 means double the current size
	}
}

func (m *HashMap[Key, Value]) fillIndexItems(store *store[Key, Value]) {
	list := m.linkedList.Load()
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
