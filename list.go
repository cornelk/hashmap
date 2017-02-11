package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// List is a sorted list.
type List struct {
	root  *ListElement
	count uint64
}

// NewList returns an initialized list.
func NewList() *List {
	e := &ListElement{deleted: 1}
	e.nextElement = unsafe.Pointer(e) // mark as root by pointing to itself

	return &List{root: e}
}

// Len returns the number of elements within the list.
func (l *List) Len() uint64 {
	return atomic.LoadUint64(&l.count)
}

// First returns the first item of the list.
func (l *List) First() *ListElement {
	item := l.root.Next()
	if item != l.root {
		return item
	}
	return nil
}

// Add adds or updates an item to the list.
func (l *List) Add(newElement *ListElement, searchStart *ListElement) bool {
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left? {
		searchStart = l.root
	}

	left, found, right := l.search(searchStart, newElement)
	if found != nil { // existing item found
		found.SetValue(newElement.value) // update the value
		if found.SetDeleted(false) {     // try to mark from deleted to not deleted
			atomic.AddUint64(&l.count, 1)
		}
		return true
	}

	return l.insertAt(newElement, left, right)
}

func (l *List) search(searchStart *ListElement, item *ListElement) (left *ListElement, found *ListElement, right *ListElement) {
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

func (l *List) insertAt(newElement *ListElement, left *ListElement, right *ListElement) bool {
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

	atomic.AddUint64(&l.count, 1)
	return true
}

// Delete marks the list element as deleted.
func (l *List) Delete(element *ListElement) {
	if !element.SetDeleted(true) {
		return // element was already deleted
	}

	element.SetValue(nil) // clear the value for the GC
	atomic.AddUint64(&l.count, ^uint64(0))
}
