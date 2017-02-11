package hashmap

import (
	"strconv"
	"testing"
	"unsafe"
)

type Animal struct {
	name string
}

func TestMapCreation(t *testing.T) {
	m := New()
	if m == nil {
		t.Error("map is null.")
	}

	if m.Len() != 0 {
		t.Error("new map should be empty.")
	}
}

func TestOverwrite(t *testing.T) {
	m := New()

	elephant := "elephant"
	monkey := "monkey"

	m.Set(1, unsafe.Pointer(&elephant))
	m.Set(1, unsafe.Pointer(&monkey))

	if m.Len() != 1 {
		t.Error("map should contain exactly one element.")
	}

	tmp, ok := m.Get(1) // Retrieve inserted element.
	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	item := (*string)(tmp) // Type assertion.
	if *item != "monkey" {
		t.Error("wrong item returned.")
	}
}

func TestInsert(t *testing.T) {
	m := NewSize(4)
	elephant := "elephant"
	monkey := "monkey"

	m.Set(4, unsafe.Pointer(&elephant))
	m.Set(3, unsafe.Pointer(&elephant))
	m.Set(2, unsafe.Pointer(&monkey))
	m.Set(1, unsafe.Pointer(&monkey))

	if m.Len() != 4 {
		t.Error("map should contain exactly 4 elements.")
	}
}

func TestGet(t *testing.T) {
	m := New()
	elephant := "elephant"

	val, ok := m.Get("animal") // Get a missing element.
	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}
	if val != nil {
		t.Error("Missing values should return as nil.")
	}

	m.Set("animal", unsafe.Pointer(&elephant))

	_, ok = m.Get("human") // Get a missing element.
	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}

	val, ok = m.Get("animal") // Retrieve inserted element.
	if ok == false {
		t.Error("ok should be true for item stored within the map.")
	}

	animal := (*string)(val)
	if *animal != elephant {
		t.Error("item was modified.")
	}
}

func TestResize(t *testing.T) {
	m := NewSize(2)
	itemCount := 50

	for i := 0; i < itemCount; i++ {
		m.Set(uint64(i), unsafe.Pointer(&Animal{strconv.Itoa(i)}))
	}

	if m.Len() != uint64(itemCount) {
		t.Error("Expected element count did not match.")
	}

	if m.Fillrate() < 30 {
		t.Error("Expecting >= 30 percent fillrate.")
	}

	for i := 0; i < itemCount; i++ {
		_, ok := m.Get(uint64(i))
		if !ok {
			t.Error("Getting inserted item failed.")
		}
	}
}

func TestStringer(t *testing.T) {
	m := New()
	elephant := &Animal{"elephant"}
	monkey := &Animal{"monkey"}

	s := m.String()
	if s != "[]" {
		t.Error("empty map as string does not match.")
	}

	m.Set(0, unsafe.Pointer(elephant))
	s = m.String()
	if s != "[9257401834698437112]" {
		t.Error("1 item map as string does not match:", s)
	}

	m.Set(1, unsafe.Pointer(monkey))
	s = m.String()
	if s != "[1754102016959854353,9257401834698437112]" {
		t.Error("2 item map as string does not match:", s)
	}
}

func TestDelete(t *testing.T) {
	m := New()
	m.Del(0)

	elephant := &Animal{"elephant"}
	monkey := &Animal{"monkey"}
	m.Set(1, unsafe.Pointer(elephant))
	m.Set(2, unsafe.Pointer(monkey))
	m.Del(0)
	m.Del(3)
	if m.Len() != 2 {
		t.Error("map should contain exactly two elements.")
	}

	m.Del(1)
	m.Del(1)
	m.Del(2)
	if m.Len() != 0 {
		t.Error("map should be empty.")
	}

	val, ok := m.Get(1) // Get a missing element.
	if ok == true {
		t.Error("ok should be false when item is missing from map.")
	}
	if val != nil {
		t.Error("Missing values should return as nil.")
	}

	m.Set(1, unsafe.Pointer(elephant))
}

func TestIterator(t *testing.T) {
	m := New()
	itemCount := 16

	for i := itemCount; i > 0; i-- {
		m.Set(uint64(i), unsafe.Pointer(&Animal{strconv.Itoa(i)}))
	}

	counter := 0
	for item := range m.Iter() {
		val := item.Value
		if val == nil {
			t.Error("Expecting an object.")
		}
		counter++
	}

	if counter != itemCount {
		t.Error("Returned item count did not match.")
	}
}

func TestHashedKey(t *testing.T) {
	m := New()
	itemCount := 16
	log := log2(uint64(itemCount))

	for i := 0; i < itemCount; i++ {
		m.SetHashedKey(uint64(i)<<(64-log), unsafe.Pointer(&Animal{strconv.Itoa(i)}))
	}

	if m.Len() != uint64(itemCount) {
		t.Error("Expected element count did not match.")
	}

	for i := 0; i < itemCount; i++ {
		_, ok := m.GetHashedKey(uint64(i) << (64 - log))
		if !ok {
			t.Error("Getting inserted item failed.")
		}
	}

	for i := 0; i < itemCount; i++ {
		m.DelHashedKey(uint64(i) << (64 - log))
	}

	if m.Len() != uint64(0) {
		t.Error("Map is not empty.")
	}
}
