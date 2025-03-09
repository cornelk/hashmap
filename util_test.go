package hashmap

import (
	"testing"

	"github.com/cornelk/hashmap/assert"
)

func TestLog2(t *testing.T) {
	var fixtures = map[uintptr]uintptr{
		0: 0,
		1: 0,
		2: 1,
		3: 2,
		4: 2,
		5: 3,
	}

	for input, result := range fixtures {
		output := log2(input)
		assert.Equal(t, output, result)
	}
}

func TestHashCollision(t *testing.T) {
	m := New[string, int]()

	staticHasher := func(_ string) uintptr {
		return 4 // chosen by fair dice roll. guaranteed to be random.
	}

	m.SetHasher(staticHasher)

	inserted := m.Insert("1", 1)
	assert.True(t, inserted)
	inserted = m.Insert("2", 2)
	assert.True(t, inserted)

	value, ok := m.Get("1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)

	value, ok = m.Get("2")
	assert.True(t, ok)
	assert.Equal(t, 2, value)
}

func TestAliasTypeSupport(t *testing.T) {
	type alias uintptr

	m := New[alias, alias]()

	inserted := m.Insert(1, 1)
	assert.True(t, inserted)

	value, ok := m.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 1, value)
}
