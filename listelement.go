package hashmap

import (
	"sync/atomic"
	"unsafe"
)

// ListElement is an element of the list.
type ListElement struct {
	keyHash     uintptr
	deleted     uintptr // for the root and all deleted items this is set to 1
	nextElement unsafe.Pointer
	key         interface{}
	value       unsafe.Pointer
}

// Value returns the value of the list item.
func (e *ListElement) Value() (value unsafe.Pointer, ok bool) {
	if atomic.LoadUintptr(&e.deleted) == 1 {
		return nil, false
	}
	value = atomic.LoadPointer(&e.value)
	// read again to make sure that the item has not been deleted between the
	// deleted check and reading of the value
	if atomic.LoadUintptr(&e.deleted) == 1 {
		return nil, false
	}
	return value, true
}

// Deleted returns whether the item was deleted.
func (e *ListElement) Deleted() bool {
	return atomic.LoadUintptr(&e.deleted) == 1
}

// Next returns the item on the right.
func (e *ListElement) Next() *ListElement {
	return (*ListElement)(atomic.LoadPointer(&e.nextElement))
}

// SetDeleted sets the deleted flag of the item.
func (e *ListElement) SetDeleted(deleted bool) bool {
	if !deleted {
		return atomic.CompareAndSwapUintptr(&e.deleted, 1, 0)
	}

	if !atomic.CompareAndSwapUintptr(&e.deleted, 0, 1) {
		return false
	}

	e.SetValue(nil) // clear the value for the GC
	return true
}

// SetValue sets the value of the item.
func (e *ListElement) SetValue(value unsafe.Pointer) {
	atomic.StorePointer(&e.value, value)
}

// CasValue compares and swaps the values of the item.
func (e *ListElement) CasValue(from, to unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&e.value, from, to)
}

// GetOrSetValue returns the value of the item.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (e *ListElement) GetOrSetValue(value unsafe.Pointer) (actual unsafe.Pointer, loaded bool) {
	for {
		if atomic.LoadUintptr(&e.deleted) == 0 { // inline ListElement.Deleted()
			actual = atomic.LoadPointer(&e.value)
			// read again to make sure that the item has not been deleted between the
			// deleted check and reading of the value
			if atomic.LoadUintptr(&e.deleted) == 0 {
				return actual, true
			}
		}

		if e.CasValue(nil, value) {
			if e.SetDeleted(false) {
				return value, false
			}
		}
	}
}
