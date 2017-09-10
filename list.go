package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// List is a sorted doubly linked list.
type List struct {
	count uintptr
	root  *ListElement
}

// NewList returns an initialized list.
func NewList() *List {
	root := &ListElement{}
	return &List{root: root}
}

// Len returns the number of elements within the list.
func (l *List) Len() int {
	return int(atomic.LoadUintptr(&l.count))
}

// First returns the first item of the list.
func (l *List) First() *ListElement {
	return l.root.Next()
}

// Add adds an item to the list and returns false if an item for the hash existed.
func (l *List) Add(newElement *ListElement, searchStart *ListElement) (existed bool, inserted bool) {
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
func (l *List) AddOrUpdate(newElement *ListElement, searchStart *ListElement) bool {
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left? {
		searchStart = l.root
	}

	left, found, right := l.search(searchStart, newElement)
	if found != nil { // existing item found
		found.SetValue(newElement.value) // update the value
		return true
	}

	return l.insertAt(newElement, left, right)
}

// Cas compares and swaps the values and add an item to the list.
func (l *List) Cas(newElement *ListElement, oldValue unsafe.Pointer, searchStart *ListElement) bool {
	if searchStart == nil || newElement.keyHash < searchStart.keyHash { // key needs to be inserted on the left? {
		searchStart = l.root
	}

	_, found, _ := l.search(searchStart, newElement)
	if found == nil { // no existing item found
		return false
	}

	if found.CasValue(oldValue, newElement.value) {
		atomic.AddUintptr(&l.count, 1)
		return true
	}
	return false
}

func (l *List) search(searchStart *ListElement, item *ListElement) (left *ListElement, found *ListElement, right *ListElement) {
	if searchStart == l.root {
		found = searchStart.Next()
		if found == nil { // no items beside root?
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
		if !atomic.CompareAndSwapPointer(&l.root.nextElement, unsafe.Pointer(nil), unsafe.Pointer(newElement)) {
			return false // item was modified concurrently
		}
	} else {
		newElement.previousElement = unsafe.Pointer(left)
		newElement.nextElement = unsafe.Pointer(right)
		if !atomic.CompareAndSwapPointer(&left.nextElement, unsafe.Pointer(right), unsafe.Pointer(newElement)) {
			return false // item was modified concurrently
		}
	}

	atomic.AddUintptr(&l.count, 1)
	return true
}

// Delete marks the list element as deleted.
func (l *List) Delete(element *ListElement) {
	for {
		left := element.Previous()
		right := element.Next()
		if left != nil {
			if !atomic.CompareAndSwapPointer(&left.nextElement, unsafe.Pointer(element), unsafe.Pointer(right)) {
				continue // item was modified concurrently
			}
		}
		break
	}

	atomic.AddUintptr(&l.count, ^uintptr(0)) // decrease counter
}
