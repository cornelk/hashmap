package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// ListElement is an element of a list.
type ListElement struct {
	keyHash         uintptr
	previousElement unsafe.Pointer
	nextElement     unsafe.Pointer
	key             interface{}
	value           unsafe.Pointer
}

// Value returns the value of the list item.
func (e *ListElement) Value() (value unsafe.Pointer) {
	return atomic.LoadPointer(&e.value)
}

// Next returns the item on the right.
func (e *ListElement) Next() *ListElement {
	return (*ListElement)(atomic.LoadPointer(&e.nextElement))
}

// Previous returns the item on the left.
func (e *ListElement) Previous() *ListElement {
	return (*ListElement)(atomic.LoadPointer(&e.previousElement))
}

// SetValue sets the value of the item.
func (e *ListElement) SetValue(value unsafe.Pointer) {
	atomic.StorePointer(&e.value, value)
}

// CasValue compares and swaps the values of the item.
func (e *ListElement) CasValue(from, to unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&e.value, from, to)
}
