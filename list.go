package hashmap

import (
	"sync/atomic"
)

// List is a sorted doubly linked list.
type List[Key keyConstraint, Value any] struct {
	count atomic.Uintptr
	head  *ListElement[Key, Value]
}

// NewList returns an initialized list.
func NewList[Key keyConstraint, Value any]() *List[Key, Value] {
	return &List[Key, Value]{head: &ListElement[Key, Value]{}}
}

// Len returns the number of elements within the list.
func (l *List[Key, Value]) Len() int {
	return int(l.count.Load())
}

// Head returns the head item of the list.
func (l *List[Key, Value]) Head() *ListElement[Key, Value] {
	return l.head
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
		found.setValue(element.value.Load()) // update the value
		return true
	}

	return l.insertAt(element, left, right)
}

// Delete deletes an element from the list.
func (l *List[Key, Value]) Delete(element *ListElement[Key, Value]) {
	if !element.deleted.CompareAndSwap(0, 1) {
		return // concurrent delete of the item is in progress
	}

	for {
		left := element.Previous()
		right := element.Next()

		if left == nil { // element is first item in list?
			if !l.head.nextElement.CompareAndSwap(element, right) {
				continue // now head item was inserted concurrently
			}
		} else {
			if !left.nextElement.CompareAndSwap(element, right) {
				continue // item was modified concurrently
			}
		}
		if right != nil {
			right.previousElement.CompareAndSwap(element, left)
		}
		break
	}

	l.count.Add(^uintptr(0)) // decrease counter
}

// Cas compares and swaps the value of an item in the list.
func (l *List[Key, Value]) Cas(element *ListElement[Key, Value], oldValue *Value, searchStart *ListElement[Key, Value]) bool {
	_, found, _ := l.search(searchStart, element)
	if found == nil { // no existing item found
		return false
	}

	if found.casValue(oldValue, element.value.Load()) {
		return true
	}
	return false
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

	element.previousElement.Store(left)
	element.nextElement.Store(right)

	if !left.nextElement.CompareAndSwap(right, element) {
		return false // item was modified concurrently
	}

	if right != nil && !right.previousElement.CompareAndSwap(left, element) {
		return false // item was modified concurrently
	}

	l.count.Add(1)
	return true
}
