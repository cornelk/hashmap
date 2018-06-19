package hashmap

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
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

	m.Set(1, elephant)
	m.Set(1, monkey)

	if m.Len() != 1 {
		t.Errorf("map should contain exactly one element but has %v items.", m.Len())
	}

	item, ok := m.Get(1) // Retrieve inserted element.
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
	if item != monkey {
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
	ok = m.Insert(uintptr(0), c)
	if !ok {
		t.Error("insert did not succeed.")
	}
	ok = m.Insert(uintptr(128), c)
	if !ok {
		t.Error("insert did not succeed.")
	}
	ok = m.Insert(uintptr(128), c)
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

	m.Set(4, elephant)
	m.Set(3, elephant)
	m.Set(2, monkey)
	m.Set(1, monkey)

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

	m.Set("animal", elephant)

	_, ok = m.Get("human") // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}

	value, ok := m.Get("animal") // Retrieve inserted element.
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}

	if value != elephant {
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
		m.Set(uintptr(i), &Animal{strconv.Itoa(i)})
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

	m.Set(0, elephant)
	s = m.String()
	hashedKey0 := getKeyHash(0)
	if s != fmt.Sprintf("[%v]", hashedKey0) {
		t.Error("1 item map as string does not match:", s)
	}

	m.Set(1, monkey)
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
	m.Set(1, elephant)
	m.Set(2, monkey)
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

	for item := range m.Iter() {
		t.Errorf("map should be empty but got %v in the iterator.", item)
	}

	val, ok := m.Get(1) // Get a missing element.
	if ok {
		t.Error("ok should be false when item is missing from map.")
	}
	if val != nil {
		t.Error("Missing values should return as nil.")
	}

	m.Set(1, elephant)
}

func TestIterator(t *testing.T) {
	m := &HashMap{}

	for item := range m.Iter() {
		t.Errorf("Expected no object but got %v.", item)
	}

	itemCount := 16
	for i := itemCount; i > 0; i-- {
		m.Set(uintptr(i), &Animal{strconv.Itoa(i)})
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
		m.SetHashedKey(uintptr(i)<<(strconv.IntSize-log), &Animal{strconv.Itoa(i)})
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

	m.SetHashedKey(1<<(strconv.IntSize-2), elephant)
	if m.Len() != 1 {
		t.Error("map should contain exactly one element.")
	}
	if !m.CasHashedKey(1<<(strconv.IntSize-2), elephant, monkey) {
		t.Error("Cas should success if expectation met")
	}
	if m.CasHashedKey(1<<(strconv.IntSize-2), elephant, monkey) {
		t.Error("Cas should fail if expectation didn't meet")
	}
	item, ok := m.GetHashedKey(1 << (strconv.IntSize - 2))
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
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

	m.Set("animal", elephant)
	if m.Len() != 1 {
		t.Error("map should contain exactly one element.")
	}
	if !m.Cas("animal", elephant, monkey) {
		t.Error("Cas should success if expectation met")
	}
	if m.Cas("animal", elephant, monkey) {
		t.Error("Cas should fail if expectation didn't meet")
	}
	item, ok := m.Get("animal")
	if !ok {
		t.Error("ok should be true for item stored within the map.")
	}
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
			counter := int64(0)
			actual, _ := m.GetOrInsert(s, &counter)
			c := actual.(*int64)
			atomic.AddInt64(c, 1)
			break
		}
	}

	s := fmt.Sprintf("/api%d/", 0)
	val, _ := m.GetStringKey(s)
	c := val.(*int64)
	if *c != 25 {
		t.Error("wrong API call count.")
	}
}

func TestExample(t *testing.T) {
	m := &HashMap{}
	i := 123
	m.Set("amount", i)

	j, ok := m.Get("amount")
	if !ok {
		t.Fail()
	}

	if i != j {
		t.Fail()
	}
}

func TestHashMap_parallel(t *testing.T) {
	max := 10
	dur := 2 * time.Second
	m := &HashMap{}

	do := func(t *testing.T, max int, d time.Duration, fn func(*testing.T, int)) <-chan error {
		t.Helper()

		done := make(chan error)

		var times int64

		// This goroutines will terminate test in case if closure hangs.
		go func() {
			for {
				select {
				case <-time.After(d + 500*time.Millisecond):
					if atomic.LoadInt64(&times) == 0 {
						done <- fmt.Errorf("closure was not executed even once, something blocks it")
					}
					close(done)
				case <-done:
				}
			}
		}()

		go func() {
			timer := time.NewTimer(d)
			defer timer.Stop()

		InfLoop:
			for {
				for i := 0; i < max; i++ {
					select {
					case <-timer.C:
						break InfLoop
					default:
					}

					fn(t, i)
					atomic.AddInt64(&times, 1)
				}
			}

			close(done)
		}()

		return done
	}
	wait := func(t *testing.T, done <-chan error) {
		t.Helper()

		if err := <-done; err != nil {
			t.Error(err)
		}
	}

	// Initial fill.
	for i := 0; i < max; i++ {
		m.Set(i, i)
		m.Set(fmt.Sprintf("%d", i), i)
		m.SetHashedKey(uintptr(i), i)
	}

	t.Run("set_get", func(t *testing.T) {
		doneSet := do(t, max, dur, func(t *testing.T, i int) {
			m.Set(i, i)
		})
		doneGet := do(t, max, dur, func(t *testing.T, i int) {
			if _, ok := m.Get(i); !ok {
				t.Errorf("missing value for key: %d", i)
			}
		})
		doneGetStringKey := do(t, max, dur, func(t *testing.T, i int) {
			if _, ok := m.GetStringKey(fmt.Sprintf("%d", i)); !ok {
				t.Errorf("missing value for key: %d", i)
			}
		})
		doneGetHashedKey := do(t, max, dur, func(t *testing.T, i int) {
			if _, ok := m.GetHashedKey(uintptr(i)); !ok {
				t.Errorf("missing value for key: %d", i)
			}
		})

		wait(t, doneSet)
		wait(t, doneGet)
		wait(t, doneGetStringKey)
		wait(t, doneGetHashedKey)
	})

	t.Run("get-or-insert-and-delete", func(t *testing.T) {
		doneGetOrInsert := do(t, max, dur, func(t *testing.T, i int) {
			m.GetOrInsert(i, i)
		})
		doneDel := do(t, max, dur, func(t *testing.T, i int) {
			m.Del(i)
		})

		wait(t, doneGetOrInsert)
		wait(t, doneDel)
	})
}
