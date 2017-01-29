package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// ListElement is an element of the list.
type ListElement struct {
	nextElement unsafe.Pointer
	key         interface{}
	keyHash     uint64
	value       unsafe.Pointer
	deleted     uint64 // for the root and all deleted items this is set to 1
}

// Value returns the value of the list item.
func (e *ListElement) Value() unsafe.Pointer {
	return atomic.LoadPointer(&e.value)
}

// Deleted returns whether the item was deleted.
func (e *ListElement) Deleted() bool {
	return atomic.LoadUint64(&e.deleted) == 1
}

// Next returns the item on the right.
func (e *ListElement) Next() *ListElement {
	return (*ListElement)(atomic.LoadPointer(&e.nextElement))
}

// search for entry starting from e (this)
func (e *ListElement) search(entry *ListElement) (left *ListElement, found *ListElement, right *ListElement) {
	eNext := (atomic.LoadPointer(&e.nextElement))
	if eNext == unsafe.Pointer(e) { // no items beside root
		return nil, nil, nil
	}

	found = e
	for {
		right = found.Next()
		if entry.keyHash == found.keyHash { // key already exists
			return nil, found, nil
		}

		if entry.keyHash < found.keyHash { // new item needs to be inserted before the found value
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
