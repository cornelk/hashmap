package hashmap

import (
	"strconv"
	"testing"
)

type Animal struct {
	name string
}

func TestMapCreation(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("map is null.")
	}

	if m.Count() != 0 {
		t.Error("new map should be empty.")
	}
}

func TestInsert(t *testing.T) {
	m := New()
	elephant := Animal{"elephant"}
	monkey := Animal{"monkey"}

	m.Set(1, 1, elephant)
	m.Set(2, 2, monkey)

	if m.Count() != 2 {
		t.Error("map should contain exactly two elements.")
	}
}

func TestGet(t *testing.T) {
	m := New()

	// Get a missing element.
	val, ok := m.Get(1, 1)

	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}

	if val != nil {
		t.Error("Missing values should return as null.")
	}

	elephant := Animal{"elephant"}
	m.Set(2, 2, elephant)

	// Retrieve inserted element.

	tmp, ok := m.Get(2, 2)
	elephant = tmp.(Animal) // Type assertion.

	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	if &elephant == nil {
		t.Error("expecting an element, not null.")
	}

	if elephant.name != "elephant" {
		t.Error("item was modified.")
	}
}

func TestRemove(t *testing.T) {
	m := New()

	monkey := Animal{"monkey"}
	m.Set(1, 1, monkey)

	m.Remove(1, 1)

	if m.Count() != 0 {
		t.Error("Expecting count to be zero once item was removed.")
	}

	temp, ok := m.Get(1, 1)

	if ok != false {
		t.Error("Expecting ok to be false for missing items.")
	}

	if temp != nil {
		t.Error("Expecting item to be nil after its removal.")
	}

	// Remove a none existing element.
	m.Remove(2, 2)
}

func TestCount(t *testing.T) {
	m := New()
	for i := 0; i < 1024; i++ {
		m.Set(uint64(i), uint64(i), Animal{strconv.Itoa(i)})
	}

	if m.Count() != 1024 {
		t.Error("Expecting 100 element within map.")
	}
}
