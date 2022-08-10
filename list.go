package hashmap

import (
	"sync/atomic"
)

// List is a sorted linked list.
type List[Key comparable, Value any] struct {
	count atomic.Uintptr
	head  *ListElement[Key, Value]
}

// NewList returns an initialized list.
func NewList[Key comparable, Value any]() *List[Key, Value] {
	return &List[Key, Value]{
		head: &ListElement[Key, Value]{},
	}
}

// Len returns the number of elements within the list.
func (l *List[Key, Value]) Len() int {
	return int(l.count.Load())
}

// First returns the first item of the list.
func (l *List[Key, Value]) First() *ListElement[Key, Value] {
	return l.head.Next()
}

// Add adds an item to the list and returns false if an item for the hash existed.
// searchStart = nil will start to search at the head item.
func (l *List[Key, Value]) Add(element, searchStart *ListElement[Key, Value]) (existed bool, inserted bool) {
	left, found, right := l.search(searchStart, element)
	if found != nil { // existing item found
		return true, false
	}

	return false, l.insertAt(element, left, right)
}

// AddOrUpdate adds or updates an item to the list.
func (l *List[Key, Value]) AddOrUpdate(element, searchStart *ListElement[Key, Value]) bool {
	left, found, right := l.search(searchStart, element)
	if found != nil { // existing item found
		found.value.Store(element.value.Load()) // update the value
		return true
	}

	return l.insertAt(element, left, right)
}

// Delete deletes an element from the list.
func (l *List[Key, Value]) Delete(element *ListElement[Key, Value]) {
	if !element.deleted.CompareAndSwap(0, 1) {
		return // concurrent delete of the item is in progress
	}

	right := element.Next()
	// point head to next element if element to delete was head
	l.head.next.CompareAndSwap(element, right)

	// element left from the deleted element will replace its next
	// pointer to the next valid element on call of Next().

	l.count.Add(^uintptr(0)) // decrease counter
}

func (l *List[Key, Value]) search(searchStart, item *ListElement[Key, Value]) (left, found, right *ListElement[Key, Value]) {
	if searchStart != nil && item.keyHash < searchStart.keyHash { // key would remain left from item? {
		searchStart = nil // start search at head
	}

	if searchStart == nil { // start search at head?
		left = l.head
		found = left.Next()
		if found == nil { // no items beside head?
			return nil, nil, nil
		}
	} else {
		found = searchStart
	}

	for {
		if item.keyHash == found.keyHash && item.key == found.key { // key hash already exists, compare keys
			return nil, found, nil
		}

		if item.keyHash < found.keyHash { // new item needs to be inserted before the found value
			if l.head == left {
				return nil, nil, found
			}
			return left, nil, found
		}

		// go to next element in sorted linked list
		left = found
		found = left.Next()
		if found == nil { // no more items on the right
			return left, nil, nil
		}
	}
}

func (l *List[Key, Value]) insertAt(element, left, right *ListElement[Key, Value]) bool {
	if left == nil {
		left = l.head
	}

	element.next.Store(right)

	if !left.next.CompareAndSwap(right, element) {
		return false // item was modified concurrently
	}

	l.count.Add(1)
	return true
}
