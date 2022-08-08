//go:build customhash

package hashmap

import (
	"testing"
)

func TestInsertCollision(t *testing.T) {
	m := New[string, int]()

	customHash = func(b []byte) uint64 {
		return 4 // chosen by fair dice roll. guaranteed to be random.
	}

	ok := m.Insert("1", 1)
	if !ok {
		t.Error("first insert did not succeed.")
	}
	ok = m.Insert("2", 2)
	if !ok {
		t.Error("second insert did not succeed.")
	}

	value1, ok := m.Get("1")
	if !ok || value1.(int) != 1 {
		t.Error("correct first item is not found in map.")
	}

	value2, ok := m.Get("2")
	if !ok || value2.(int) != 2 {
		t.Error("correct second item is not found in map.")
	}
}
