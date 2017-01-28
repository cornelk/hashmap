package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// ListElement is an element of the list.
type ListElement struct {
	nextElement unsafe.Pointer
	keyHash     uint64
	value       unsafe.Pointer
	deleted     uint64 // the root and all deleted items are set to 1
}

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
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left?{
		searchStart = l.root
	}

	left, found, right := searchStart.search(newElement)
	if found != nil { // existing item found
		atomic.StorePointer(&found.value, unsafe.Pointer(newElement.value)) // update the value

		if atomic.CompareAndSwapUint64(&found.deleted, 1, 0) { // try to mark from deleted to not deleted
			atomic.AddUint64(&l.count, 1)
		}
		return true
	}

	return l.insertAt(newElement, left, right)
}

func (l *List) insertAt(newElement *ListElement, left *ListElement, right *ListElement) bool {
	if left == nil { // insert at root
		if !atomic.CompareAndSwapPointer(&l.root.nextElement, unsafe.Pointer(l.root), unsafe.Pointer(newElement)) {
			return false // item was already modified
		}
	} else {
		newElement.nextElement = unsafe.Pointer(right)
		if !atomic.CompareAndSwapPointer(&left.nextElement, unsafe.Pointer(right), unsafe.Pointer(newElement)) {
			return false // item was already modified
		}
	}

	atomic.AddUint64(&l.count, 1)
	return true
}

// Delete marks the list element as deleted.
func (l *List) Delete(element *ListElement) {
	if !atomic.CompareAndSwapUint64(&element.deleted, 0, 1) {
		return // element was already deleted
	}

	atomic.StorePointer(&element.value, nil) // clear the value for the GC
	atomic.AddUint64(&l.count, ^uint64(1-1))
}

// Next returns the item on the right.
func (e *ListElement) Next() *ListElement {
	return (*ListElement)(atomic.LoadPointer(&e.nextElement))
}

// search for entry starting from e (this)
func (e *ListElement) search(entry *ListElement) (left *ListElement, found *ListElement, right *ListElement) {
	if e == nil || unsafe.Pointer(e) == e.nextElement { // no items beside root
		return nil, nil, nil
	}

	found = e
	for {
		if atomic.LoadUint64(&e.deleted) == 0 { // skip reading values of root or deleted items
			right = found.Next()
			if entry.keyHash == found.keyHash { // key already exists
				return nil, found, nil
			}

			if entry.keyHash < found.keyHash { // new item needs to be inserted before the found value
				return left, nil, found
			}
		}

		// go to next entry in sorted linked list
		left = found
		found = left.Next()
		if found == nil { // no more items on the right
			return left, nil, nil
		}
	}
}
