package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// List32 is a sorted list.
type List32 struct {
	count uint32
	root  *ListElement32
}

// NewList32 returns an initialized list.
func NewList32() *List32 {
	e := &ListElement32{deleted: 1}
	e.nextElement = unsafe.Pointer(e) // mark as root by pointing to itself

	return &List32{root: e}
}

// Len returns the number of elements within the list.
func (l *List32) Len() uint32 {
	return atomic.LoadUint32(&l.count)
}

// First returns the first item of the list.
func (l *List32) First() *ListElement32 {
	item := l.root.Next()
	if item != l.root {
		return item
	}
	return nil
}

// Add adds an item to the list and returns false if an item for the hash existed.
func (l *List32) Add(newElement *ListElement32, searchStart *ListElement32) (existed bool, inserted bool) {
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left? {
		searchStart = l.root
	}

	left, found, right := l.search(searchStart, newElement)
	if found != nil { // existing item found
		return true, false
	}

	return false, l.insertAt(newElement, left, right)
}

// AddOrUpdate adds or updates an item to the list.
func (l *List32) AddOrUpdate(newElement *ListElement32, searchStart *ListElement32) bool {
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left? {
		searchStart = l.root
	}

	left, found, right := l.search(searchStart, newElement)
	if found != nil { // existing item found
		found.SetValue(newElement.value) // update the value
		if found.SetDeleted(false) {     // try to mark from deleted to not deleted
			atomic.AddUint32(&l.count, 1)
		}
		return true
	}

	return l.insertAt(newElement, left, right)
}

// Cas compares and swaps the values and add an item to the list.
func (l *List32) Cas(newElement *ListElement32, oldValue unsafe.Pointer, searchStart *ListElement32) bool {
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left? {
		searchStart = l.root
	}

	_, found, _ := l.search(searchStart, newElement)
	if found == nil { // no existing item found
		return false
	}

	if found.CasValue(oldValue, newElement.value) {
		if found.SetDeleted(false) { // try to mark from deleted to not deleted
			atomic.AddUint32(&l.count, 1)
		}
		return true
	}
	return false
}

func (l *List32) search(searchStart *ListElement32, item *ListElement32) (left *ListElement32, found *ListElement32, right *ListElement32) {
	if searchStart == l.root {
		found = searchStart.Next()
		if found == l.root { // no items beside root?
			return nil, nil, nil
		}
		left = searchStart
	} else {
		found = searchStart
	}

	for {
		if item.keyHash == found.keyHash { // key already exists
			return nil, found, nil
		}

		if item.keyHash < found.keyHash { // new item needs to be inserted before the found value
			return left, nil, found
		}

		// go to next entry in sorted linked list
		left = found
		found = left.Next()
		if found == nil { // no more items on the right
			return left, nil, nil
		}
	}
}

func (l *List32) insertAt(newElement *ListElement32, left *ListElement32, right *ListElement32) bool {
	if left == nil { // insert at root
		if !atomic.CompareAndSwapPointer(&l.root.nextElement, unsafe.Pointer(l.root), unsafe.Pointer(newElement)) {
			return false // item was modified concurrently
		}
	} else {
		newElement.nextElement = unsafe.Pointer(right)
		if !atomic.CompareAndSwapPointer(&left.nextElement, unsafe.Pointer(right), unsafe.Pointer(newElement)) {
			return false // item was modified concurrently
		}
	}

	atomic.AddUint32(&l.count, 1)
	return true
}

// Delete marks the list element as deleted.
func (l *List32) Delete(element *ListElement32) {
	if !element.SetDeleted(true) {
		return // element was already deleted
	}

	atomic.AddUint32(&l.count, ^uint32(0)) // decrease counter
}
