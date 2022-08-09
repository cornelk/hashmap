package hashmap

import (
	"sync/atomic"
)

// ListElement is an element of a list.
type ListElement[Key keyConstraint, Value any] struct {
	keyHash uintptr
	// deleted marks the item as deleting or deleted
	// this is using uintptr instead of atomic.Bool to avoid using 32 bit int on 64 bit systems
	deleted atomic.Uintptr

	previousElement atomic.Pointer[ListElement[Key, Value]] // is nil for the first item in list
	nextElement     atomic.Pointer[ListElement[Key, Value]] // is nil for the last item in list

	value atomic.Pointer[Value]
	key   Key
}

// Value returns the value of the list item.
func (e *ListElement[Key, Value]) Value() *Value {
	return e.value.Load()
}

// Next returns the item on the right.
func (e *ListElement[Key, Value]) Next() *ListElement[Key, Value] {
	return e.nextElement.Load()
}

// Previous returns the item on the left.
func (e *ListElement[Key, Value]) Previous() *ListElement[Key, Value] {
	return e.previousElement.Load()
}

// setValue sets the value of the item.
func (e *ListElement[Key, Value]) setValue(value *Value) {
	e.value.Store(value)
}

// casValue compares and swaps the values of the item.
func (e *ListElement[Key, Value]) casValue(from, to *Value) bool {
	return e.value.CompareAndSwap(from, to)
}
