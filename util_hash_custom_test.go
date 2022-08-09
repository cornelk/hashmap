//go:build customhash

package hashmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertCollision(t *testing.T) {
	m := New[string, int]()

	customHash = func(b []byte) uint64 {
		return 4 // chosen by fair dice roll. guaranteed to be random.
	}

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
