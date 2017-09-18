package hashmap

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"
)

type Animal struct {
	name string
}

func TestMapCreation(t *testing.T) {
	m := &HashMap{}
	if m.Len() != 0 {
		t.Errorf("new map should be empty but has %d items.", m.Len())
	}
}

func TestOverwrite(t *testing.T) {
	m := &HashMap{}

	elephant := "elephant"
	monkey := "monkey"

	m.Set(1, unsafe.Pointer(&elephant))
	m.Set(1, unsafe.Pointer(&monkey))

	if m.Len() != 1 {
		t.Errorf("map should contain exactly one element but has %v items.", m.Len())
	}

	tmp, ok := m.Get(1) // Retrieve inserted element.
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}

	item := (*string)(tmp) // Type assertion.
	if *item != "monkey" {
		t.Error("wrong item returned.")
	}
}

func TestInsert(t *testing.T) {
	m := &HashMap{}
	_, ok := m.GetUintKey(0)
	if ok {
		t.Error("empty map should not return an item.")
	}
	c := uintptr(16)
	ok = m.Insert(uintptr(0), unsafe.Pointer(&c))
	if !ok {
		t.Error("insert did not succeed.")
	}
	ok = m.Insert(uintptr(128), unsafe.Pointer(&c))
	if !ok {
		t.Error("insert did not succeed.")
	}
	ok = m.Insert(uintptr(128), unsafe.Pointer(&c))
	if ok {
		t.Error("insert on existing item did succeed.")
	}
	_, ok = m.GetUintKey(128)
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
	_, ok = m.GetUintKey(127)
	if ok {
		t.Error("item for key should not exist.")
	}
	if m.Len() != 2 {
		t.Errorf("map should contain exactly 2 elements but has %v items.", m.Len())
	}
}

func TestSet(t *testing.T) {
	m := New(4)
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
	m := &HashMap{}
	elephant := "elephant"

	val, ok := m.Get("animal") // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}
	if val != nil {
		t.Error("Missing values should return as nil.")
	}

	m.Set("animal", unsafe.Pointer(&elephant))

	_, ok = m.Get("human") // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}

	val, ok = m.Get("animal") // Retrieve inserted element.
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}

	animal := (*string)(val)
	if *animal != elephant {
		t.Error("item was modified.")
	}
}

func TestGrow(t *testing.T) {
	m := &HashMap{}
	m.Grow(uintptr(63))

	for { // make sure to wait for resize operation to finish
		if atomic.LoadUintptr(&m.resizing) == 0 {
			break
		}
		time.Sleep(time.Microsecond * 50)
	}

	d := m.mapData()
	if d.keyshifts != 58 {
		t.Error("Grow operation did not result in correct internal map data structure.")
	}
}

func TestResize(t *testing.T) {
	m := New(2)
	itemCount := 50

	for i := 0; i < itemCount; i++ {
		m.Set(uintptr(i), unsafe.Pointer(&Animal{strconv.Itoa(i)}))
	}

	if m.Len() != itemCount {
		t.Error("Expected element count did not match.")
	}

	for { // make sure to wait for resize operation to finish
		if atomic.LoadUintptr(&m.resizing) == 0 {
			break
		}
		time.Sleep(time.Microsecond * 50)
	}

	if m.Fillrate() != 34 {
		t.Error("Expecting 34 percent fillrate.")
	}

	for i := 0; i < itemCount; i++ {
		_, ok := m.Get(uintptr(i))
		if !ok {
			t.Error("Getting inserted item failed.")
		}
	}
}

func TestStringer(t *testing.T) {
	m := &HashMap{}
	elephant := &Animal{"elephant"}
	monkey := &Animal{"monkey"}

	s := m.String()
	if s != "[]" {
		t.Error("empty map as string does not match.")
	}

	m.Set(0, unsafe.Pointer(elephant))
	s = m.String()
	hashedKey0 := getKeyHash(0)
	if s != fmt.Sprintf("[%v]", hashedKey0) {
		t.Error("1 item map as string does not match:", s)
	}

	m.Set(1, unsafe.Pointer(monkey))
	s = m.String()
	hashedKey1 := getKeyHash(1)
	if s != fmt.Sprintf("[%v,%v]", hashedKey1, hashedKey0) {
		t.Error("2 item map as string does not match:", s)
	}
}

func TestDelete(t *testing.T) {
	m := &HashMap{}
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
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}
	if val != nil {
		t.Error("Missing values should return as nil.")
	}

	m.Set(1, unsafe.Pointer(elephant))
}

func TestIterator(t *testing.T) {
	m := &HashMap{}

	for item := range m.Iter() {
		t.Errorf("Expected no object but got %v.", item)
	}

	itemCount := 16
	for i := itemCount; i > 0; i-- {
		m.Set(uintptr(i), unsafe.Pointer(&Animal{strconv.Itoa(i)}))
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
	m := &HashMap{}
	_, ok := m.GetHashedKey(uintptr(0))
	if ok {
		t.Error("empty map should not return an item.")
	}
	m.DelHashedKey(uintptr(0))
	m.allocate(uintptr(64))
	m.DelHashedKey(uintptr(0))

	itemCount := 16
	log := log2(uintptr(itemCount))

	for i := 0; i < itemCount; i++ {
		m.SetHashedKey(uintptr(i)<<(strconv.IntSize-log), unsafe.Pointer(&Animal{strconv.Itoa(i)}))
	}

	if m.Len() != itemCount {
		t.Error("Expected element count did not match.")
	}

	for i := 0; i < itemCount; i++ {
		_, ok = m.GetHashedKey(uintptr(i) << (strconv.IntSize - log))
		if !ok {
			t.Error("Getting inserted item failed.")
		}
	}

	for i := 0; i < itemCount; i++ {
		m.DelHashedKey(uintptr(i) << (strconv.IntSize - log))
	}
	_, ok = m.GetHashedKey(uintptr(0))
	if ok {
		t.Error("item for key should not exist.")
	}
	if m.Len() != 0 {
		t.Error("Map is not empty.")
	}
}

func TestCompareAndSwapHashedKey(t *testing.T) {
	m := &HashMap{}
	elephant := &Animal{"elephant"}
	monkey := &Animal{"monkey"}

	m.SetHashedKey(1<<(strconv.IntSize-2), unsafe.Pointer(elephant))
	if m.Len() != 1 {
		t.Error("map should contain exactly one element.")
	}
	if !m.CasHashedKey(1<<(strconv.IntSize-2), unsafe.Pointer(elephant), unsafe.Pointer(monkey)) {
		t.Error("Cas should success if expectation met")
	}
	if m.CasHashedKey(1<<(strconv.IntSize-2), unsafe.Pointer(elephant), unsafe.Pointer(monkey)) {
		t.Error("Cas should fail if expectation didn't meet")
	}
	tmp, ok := m.GetHashedKey(1 << (strconv.IntSize - 2))
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
	item := (*Animal)(tmp)
	if item != monkey {
		t.Error("wrong item returned.")
	}
}

func TestCompareAndSwap(t *testing.T) {
	m := &HashMap{}
	ok := m.CasHashedKey(uintptr(0), nil, nil)
	if ok {
		t.Error("empty map should not return an item.")
	}

	elephant := &Animal{"elephant"}
	monkey := &Animal{"monkey"}

	m.Set("animal", unsafe.Pointer(elephant))
	if m.Len() != 1 {
		t.Error("map should contain exactly one element.")
	}
	if !m.Cas("animal", unsafe.Pointer(elephant), unsafe.Pointer(monkey)) {
		t.Error("Cas should success if expectation met")
	}
	if m.Cas("animal", unsafe.Pointer(elephant), unsafe.Pointer(monkey)) {
		t.Error("Cas should fail if expectation didn't meet")
	}
	tmp, ok := m.Get("animal")
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
	item := (*Animal)(tmp)
	if item != monkey {
		t.Error("wrong item returned.")
	}
}

// TestAPICounter shows how to use the hashmap to count REST server API calls
func TestAPICounter(t *testing.T) {
	m := &HashMap{}

	for i := 0; i < 100; i++ {
		s := fmt.Sprintf("/api%d/", i%4)

		for {
			val, ok := m.GetStringKey(s)
			if !ok { // item does not exist yet
				c := uintptr(1)
				if !m.Insert(s, unsafe.Pointer(&c)) {
					continue // item was inserted concurrently, try to read it again
				}
				break
			}

			c := (*uintptr)(val)
			atomic.AddUintptr(c, 1)
			break
		}
	}

	s := fmt.Sprintf("/api%d/", 0)
	val, _ := m.GetStringKey(s)
	c := (*uintptr)(val)
	if *c != 25 {
		t.Error("wrong API call count.")
	}
}

func TestExample(t *testing.T) {
	m := &HashMap{}
	i := 123
	m.Set("amount", unsafe.Pointer(&i))

	val, ok := m.Get("amount")
	if !ok {
		t.Fail()
	}

	j := *(*int)(val)
	if i != j {
		t.Fail()
	}
}

func TestGetOrInsert(t *testing.T) {
	m := &HashMap{}

	var i, j uintptr
	actual, loaded := m.GetOrInsert("api1", unsafe.Pointer(&i))
	if loaded {
		t.Error("item should have been inserted.")
	}

	counter := (*uintptr)(actual)
	if *counter != 0 {
		t.Error("item should be 0.")
	}

	atomic.AddUintptr(counter, 1) // increase counter

	actual, loaded = m.GetOrInsert("api1", unsafe.Pointer(&j))
	if !loaded {
		t.Error("item should have been loaded.")
	}

	counter = (*uintptr)(actual)
	if *counter != 1 {
		t.Error("item should be 1.")
	}
}
