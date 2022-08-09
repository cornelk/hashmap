package hashmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestInsertCollision(t *testing.T) {
	m := New[string, int]()

	staticHasher := func(key string) uintptr {
		return 4 // chosen by fair dice roll. guaranteed to be random.
	}

	m.SetHasher(staticHasher)

	inserted := m.Insert("1", 1)
	require.True(t, inserted)
	inserted = m.Insert("2", 2)
	require.True(t, inserted)

	value, ok := m.Get("1")
	require.True(t, ok)
	assert.Equal(t, 1, value)

	value, ok = m.Get("2")
	require.True(t, ok)
	assert.Equal(t, 2, value)
}
